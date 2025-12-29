package core_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/thewandererbg/pgbase/core"
	"github.com/thewandererbg/pgbase/tests"
	"github.com/thewandererbg/pgbase/tools/security"
)

func TestPubSubStartNoDatabaseURL(t *testing.T) {
	t.Parallel()
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("NewTestApp failed: %v", err)
	}
	defer app.Cleanup()

	_ = os.Unsetenv("PB_DATA_URI")

	err = app.StartPubSub()
	if err == nil {
		t.Fatal("Expected error when PB_DATA_URI is not set")
	}

	expectedErr := "PB_DATA_URI environment variable is required when pub/sub is enabled"
	if err.Error() != expectedErr {
		t.Fatalf("Expected error %q, got %q", expectedErr, err.Error())
	}

	if state := app.PubSubState(); state != nil {
		t.Fatal("Expected no pubsub state when PB_DATA_URI is not set")
	}
}

func TestPubSubStartInvalidURL(t *testing.T) {
	t.Parallel()
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	_ = os.Setenv("PB_DATA_URI", "postgresql://invalid:password@nonexistent:5432/db")
	defer os.Unsetenv("PB_DATA_URI")

	// With reconnect loop, StartPubSub no longer fails synchronously.
	if err := app.StartPubSub(); err != nil {
		t.Fatalf("Expected StartPubSub to succeed (async reconnect), got: %v", err)
	}

	state := app.PubSubState()
	if state == nil {
		t.Fatal("Expected pubsub state to be created")
	}

	// Give it a bit of time to attempt connect and fail.
	time.Sleep(150 * time.Millisecond)

	// Connection should still be nil (or at least not guaranteed to be non-nil).
	if state.Conn != nil {
		t.Fatal("Expected no connection to be established with invalid DB URL")
	}

	// Stop should clean up without hanging.
	app.StopPubSub()

	if state := app.PubSubState(); state != nil {
		t.Fatal("Expected pubsub state to be removed after stop")
	}
}

func TestPubSubStartAndStop(t *testing.T) {
	t.Parallel()
	app, cleanup, _ := setupPubSubApp(t)
	defer cleanup()

	state := app.PubSubState()
	if state == nil {
		t.Fatal("Expected pubsub state to be created")
	}

	if state.CancelFunc == nil {
		t.Fatal("Expected cancel function to be set")
	}

	// Stop should always succeed (even if connection never came up).
	app.StopPubSub()

	if state := app.PubSubState(); state != nil {
		t.Fatal("Expected pubsub state to be removed after stop")
	}

	select {
	case <-state.Done:
	case <-time.After(1 * time.Second):
		t.Fatal("Expected done channel to be closed within 1 seconds")
	}
}

func TestPubSubStopWhenNotStarted(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	app.StopPubSub()

	if app.PubSubState() != nil {
		t.Fatal("Expected no pubsub state")
	}
}

func TestPubSubReconnect(t *testing.T) {
	t.Parallel()
	app, cleanup, dbURL := setupPubSubApp(t)
	defer cleanup()

	state := app.PubSubState()
	if state == nil || state.Conn == nil {
		t.Fatal("Expected pubsub connection to be established")
	}

	// Create separate connection to kill the listener
	killConn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("Failed to create kill connection: %v", err)
	}
	defer killConn.Close(context.Background())

	// Find the listener PID by application_name
	var pid int
	err = killConn.QueryRow(context.Background(),
		"SELECT pid FROM pg_stat_activity WHERE application_name = 'pgbase' LIMIT 1",
	).Scan(&pid)
	if err != nil {
		t.Fatalf("Failed to find listener pid: %v", err)
	}

	// Kill the listener connection
	_, err = killConn.Exec(context.Background(), "SELECT pg_terminate_backend($1)", pid)
	if err != nil {
		t.Fatalf("Failed to terminate backend: %v", err)
	}

	// Wait for reconnection (5s interval + buffer)
	time.Sleep(6 * time.Second)

	// Verify reconnected
	state = app.PubSubState()
	if state == nil || state.Conn == nil {
		t.Fatal("Expected pubsub to reconnect")
	}

	// Verify new connection exists in pg_stat_activity
	var newPid int
	err = killConn.QueryRow(context.Background(),
		"SELECT pid FROM pg_stat_activity WHERE application_name = 'pgbase' LIMIT 1",
	).Scan(&newPid)
	if err != nil {
		t.Fatalf("Reconnected connection not found: %v", err)
	}

	if newPid == pid {
		t.Fatal("Expected new PID after reconnection")
	}
}

func TestPubSubCollectionsSync(t *testing.T) {
	t.Parallel()
	app1, app2, cleanup, _ := setupPubSubApps(t)
	defer cleanup()

	// Create a new collection in app1
	collection := core.NewBaseCollection("test_sync_" + security.RandomString(5))
	collection.Fields.Add(&core.TextField{Name: "title"})

	if err := app1.Save(collection); err != nil {
		t.Fatalf("Failed to create collection in app1: %v", err)
	}

	// Wait for notification propagation
	time.Sleep(150 * time.Millisecond)

	// Verify app2 has the collection in cache
	cachedCollection, _ := app2.FindCollectionByNameOrId(collection.Name)
	if cachedCollection == nil {
		t.Fatalf("Expected app2 to have collection %s in cache", collection.Name)
	}

	// Verify title field exists and matches
	field1 := collection.Fields.GetByName("title")
	if field1 == nil {
		t.Fatal("Expected title field in app1")
	}

	// Update collection in app1 - add description field
	collection.Fields.Add(&core.TextField{Name: "description"})

	if err := app1.Save(collection); err != nil {
		t.Fatalf("Failed to update collection in app1: %v", err)
	}

	// Wait for notification propagation
	time.Sleep(150 * time.Millisecond)

	// Verify app2 has the updated collection
	updatedCollection, _ := app2.FindCollectionByNameOrId(collection.Name)
	if updatedCollection == nil {
		t.Fatal("Expected app2 to have updated collection in cache")
	}

	// Verify description field exists in app2
	descField := updatedCollection.Fields.GetByName("description")
	if descField == nil {
		t.Fatal("Expected description field in app2 cache")
	}

	// Delete the collection in app1
	if err := app1.Delete(collection); err != nil {
		t.Fatalf("Failed to delete collection in app1: %v", err)
	}

	// Wait for notification propagation
	time.Sleep(150 * time.Millisecond)

	// Verify app2 no longer has the collection in cache
	deletedCollection, _ := app2.FindCollectionByNameOrId(collection.Name)
	if deletedCollection != nil {
		t.Fatalf("Expected app2 to not have collection %s in cache after deletion", collection.Name)
	}
}

func TestPubSubSettingsSync(t *testing.T) {
	t.Parallel()
	app1, app2, cleanup, _ := setupPubSubApps(t)
	defer cleanup()

	// Update settings in app1
	testAppName := "test_app_" + security.RandomString(5)
	app1.Settings().Meta.AppName = testAppName
	app1.Settings().Meta.AppURL = "https://test.example.com"

	if err := app1.Save(app1.Settings()); err != nil {
		t.Fatalf("Failed to save settings in app1: %v", err)
	}

	// Wait for notification propagation
	time.Sleep(150 * time.Millisecond)

	// Verify app2 settings are updated
	if app2.Settings().Meta.AppName != app1.Settings().Meta.AppName {
		t.Fatalf("Expected app2 AppName %s, got %s",
			app1.Settings().Meta.AppName,
			app2.Settings().Meta.AppName)
	}

	if app2.Settings().Meta.AppURL != app1.Settings().Meta.AppURL {
		t.Fatalf("Expected app2 AppURL %s, got %s",
			app1.Settings().Meta.AppURL,
			app2.Settings().Meta.AppURL)
	}
}

func setupPubSubApps(t *testing.T) (*tests.TestApp, *tests.TestApp, func(), string) {
	id := core.GenerateDefaultRandomId()[0:5]
	dataDB := fmt.Sprintf("pbdb_%s", id)
	dbURL := fmt.Sprintf("postgres://postgres:postgrespassword@localhost:5432/%s?sslmode=disable&default_query_exec_mode=simple_protocol", dataDB)

	app1, _ := tests.NewTestAppWithOptions(
		tests.TestAppConfig{
			MultiInstanceEnabled: true,
			PubSubDataURI:        dbURL,
		},
		"/tmp/"+dataDB,
	)

	if err := app1.StartPubSub(); err != nil {
		app1.Cleanup()
		t.Fatalf("Failed to start pubsub for app1: %v", err)
	}

	app2, _ := tests.NewTestAppWithOptions(
		tests.TestAppConfig{
			MultiInstanceEnabled: true,
			PubSubDataURI:        dbURL,
		},
		"/tmp/"+dataDB,
	)

	if err := app2.StartPubSub(); err != nil {
		app1.StopPubSub()
		app1.Cleanup()
		app2.Cleanup()
		t.Fatalf("Failed to start pubsub for app2: %v", err)
	}

	cleanup := func() {
		app1.StopPubSub()
		app2.StopPubSub()
		app1.Cleanup()
		app2.Cleanup()
	}

	// Wait for both to connect
	time.Sleep(100 * time.Millisecond)

	return app1, app2, cleanup, dbURL
}

func setupPubSubApp(t *testing.T) (*tests.TestApp, func(), string) {
	id := core.GenerateDefaultRandomId()[0:5]
	dataDB := "pbdb_" + id
	dbURL := fmt.Sprintf("postgres://postgres:postgrespassword@localhost:5432/%s?sslmode=disable&default_query_exec_mode=simple_protocol", dataDB)

	app, _ := tests.NewTestAppWithOptions(
		tests.TestAppConfig{
			MultiInstanceEnabled: true,
			PubSubDataURI:        dbURL,
		},
		"/tmp/"+dataDB,
	)

	if err := app.StartPubSub(); err != nil {
		app.Cleanup()
		t.Fatalf("Failed to start pubsub: %v", err)
	}

	cleanup := func() {
		app.StopPubSub()
		app.Cleanup()
	}

	// Wait for initial connection
	time.Sleep(100 * time.Millisecond)

	return app, cleanup, dbURL
}
