package core

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pocketbase/dbx"
)

// -------------------------------------------------------------------
// PostgreSQL LISTEN/NOTIFY support
// -------------------------------------------------------------------

const (
	pubSubStoreKey          = "pbPubSub"
	pubSubReconnectInterval = 5 * time.Second
	chMetaCollections       = "pb:meta:collections"
	chMetaSettings          = "pb:meta:settings"
)

type PubSubState struct {
	mu         sync.Mutex
	Conn       *pgx.Conn
	CancelFunc context.CancelFunc
	Done       chan struct{}
}

func (app *BaseApp) startPubSub() error {
	app.Logger().Info("starting postgres listener")

	dbURI := app.config.PubSubDataURI
	if dbURI == "" {
		dbURI = os.Getenv("PB_DATA_URI")
	}
	if dbURI == "" {
		return fmt.Errorf("PB_DATA_URI environment variable is required when pub/sub is enabled")
	}

	ctx, cancel := context.WithCancel(context.Background())

	state := &PubSubState{
		CancelFunc: cancel,
		Done:       make(chan struct{}),
	}
	app.Store().Set(pubSubStoreKey, state)

	connectAndListen := func() (*pgx.Conn, error) {
		config, err := pgx.ParseConfig(dbURI)
		if err != nil {
			return nil, err
		}
		config.RuntimeParams["application_name"] = "pgbase"

		conn, err := pgx.ConnectConfig(ctx, config)
		if err != nil {
			return nil, err
		}

		for _, ch := range []string{chMetaCollections, chMetaSettings} {
			if _, err := conn.Exec(ctx, `LISTEN `+pgx.Identifier{ch}.Sanitize()); err != nil {
				_ = conn.Close(ctx)
				return nil, fmt.Errorf("LISTEN %s: %w", ch, err)
			}
		}

		return conn, nil
	}

	go func() {
		defer close(state.Done)

		for {
			if ctx.Err() != nil {
				return
			}

			conn, err := connectAndListen()
			if err != nil {
				app.Logger().Warn("pubsub connect/listen failed", "error", err)
			} else {
				app.Logger().Info("pubsub listener connected")
				state.mu.Lock()
				state.Conn = conn
				state.mu.Unlock()

				app.subscribeToEvents(ctx, conn) // returns on ctx cancel or conn error

				_ = conn.Close(ctx)

				state.mu.Lock()
				if state.Conn == conn {
					state.Conn = nil
				}
				state.mu.Unlock()
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(pubSubReconnectInterval):
			}
		}
	}()

	return nil
}

func (app *BaseApp) subscribeToEvents(ctx context.Context, conn *pgx.Conn) {
	for {
		n, err := conn.WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			app.Logger().Warn("pubsub listener connection lost, reconnecting", "error", err)
			return
		}

		switch n.Channel {
		case chMetaCollections:
			app.Logger().Info("reloading cached collections")
			app.ReloadCachedCollections()

		case chMetaSettings:
			app.Logger().Info("reloading settings")
			app.ReloadSettings()

		default:
			// ignore unknown channels
			continue
		}
	}
}

// stopPubSub closes the LISTEN connection and waits for the goroutine to exit.
func (app *BaseApp) stopPubSub() {
	state := app.PubSubState()
	if state == nil {
		return
	}

	if state.CancelFunc != nil {
		state.CancelFunc()
	}

	state.mu.Lock()
	conn := state.Conn
	state.Conn = nil
	state.mu.Unlock()

	if conn != nil {
		_ = conn.Close(context.Background())
	}

	<-state.Done
	app.Store().Remove(pubSubStoreKey)
}

// PublishCollectionsReload notifies all pods to reload cached collections.
func (app *BaseApp) PublishCollectionsReload() error {
	if app.PubSubState() == nil {
		return nil
	}

	_, err := app.DB().NewQuery("SELECT pg_notify({:channel}, '')").
		Bind(dbx.Params{
			"channel": chMetaCollections,
		}).
		Execute()

	return err
}

// PublishSettingsReload notifies all pods to reload application settings.
func (app *BaseApp) PublishSettingsReload() error {
	if app.PubSubState() == nil {
		return nil
	}

	_, err := app.DB().NewQuery("SELECT pg_notify({:channel}, '')").
		Bind(dbx.Params{
			"channel": chMetaSettings,
		}).
		Execute()

	return err
}

// EnsureCollectionsCacheFresh reloads collections locally when pubsub is disabled
// or not running, otherwise broadcasts a reload signal.
func (app *BaseApp) EnsureCollectionsCacheFresh() {
	if app.MultiInstanceEnabled() && app.PubSubState() != nil {
		if err := app.PublishCollectionsReload(); err != nil {
			app.Logger().Warn("Failed to publish collections reload", "error", err)
		}
		return
	}

	if err := app.ReloadCachedCollections(); err != nil {
		app.Logger().Warn("Failed to reload collections cache", "error", err)
	}
}

// EnsureSettingsCacheFresh reloads settings locally when pubsub is disabled
// or not running, otherwise broadcasts a reload signal.
func (app *BaseApp) EnsureSettingsCacheFresh() {
	if app.MultiInstanceEnabled() && app.PubSubState() != nil {
		if err := app.PublishSettingsReload(); err != nil {
			app.Logger().Warn("Failed to publish settings reload", "error", err)
		}
		return
	}

	if err := app.ReloadSettings(); err != nil {
		app.Logger().Warn("Failed to reload settings", "error", err)
	}
}

// PubSubState returns the current pubsub state (nil if not started).
func (app *BaseApp) PubSubState() *PubSubState {
	v, _ := app.Store().Get(pubSubStoreKey).(*PubSubState)
	return v
}
