package core

import (
	"database/sql"
	"fmt"

	"github.com/pocketbase/dbx"
)

// TableColumns returns all column names of a single table by its name.
func (app *BaseApp) TableColumns(tableName string) ([]string, error) {
	columns := []string{}

	err := app.DB().NewQuery("SELECT column_name FROM information_schema.columns WHERE table_schema = 'public' AND table_name = {:tableName}").
		Bind(dbx.Params{"tableName": tableName}).
		Column(&columns)

	return columns, err
}

type TableInfoRow struct {
	// the `db:"pk"` tag has special semantic so we cannot rename
	// the original field without specifying a custom mapper
	PK int

	Index        int            `db:"cid"`
	Name         string         `db:"name"`
	Type         string         `db:"type"`
	NotNull      bool           `db:"notnull"`
	DefaultValue sql.NullString `db:"dflt_value"`
}

// TableInfo returns the column information for the specified table.
func (app *BaseApp) TableInfo(tableName string) ([]*TableInfoRow, error) {
	info := []*TableInfoRow{}

	query := `
        SELECT
            a.attnum as cid,
            a.attname as name,
            format_type(a.atttypid, a.atttypmod) as type,
            a.attnotnull as notnull,
            pg_get_expr(d.adbin, d.adrelid) as dflt_value,
            CASE WHEN pk.contype = 'p' THEN 1 ELSE 0 END as pk
        FROM pg_class c
        JOIN pg_attribute a ON a.attrelid = c.oid
        LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
        LEFT JOIN (
            SELECT conrelid, contype, unnest(conkey) as conkey
            FROM pg_constraint
            WHERE contype = 'p'
        ) pk ON pk.conrelid = a.attrelid AND pk.conkey = a.attnum
        WHERE c.relname = {:tableName}
        AND (c.relkind = 'r' OR c.relkind = 'v')  -- 'r' for tables, 'v' for views
        AND a.attnum > 0
        AND NOT a.attisdropped
        ORDER BY a.attnum
    `

	err := app.DB().NewQuery(query).
		Bind(dbx.Params{"tableName": tableName}).
		All(&info)
	if err != nil {
		return nil, err
	}

	if len(info) == 0 {
		return nil, fmt.Errorf("empty table info probably due to invalid or missing table %s", tableName)
	}

	return info, nil
}

// TableIndexes returns a name grouped map with all non empty index of the specified table.
//
// Note: This method doesn't return an error on nonexisting table.
func (app *BaseApp) TableIndexes(tableName string) (map[string]string, error) {
	indexes := []struct {
		Name       string
		Definition string
	}{}

	err := app.DB().Select("indexname as name", "indexdef as definition").
		From("pg_indexes").
		AndWhere(dbx.HashExp{"tablename": tableName}).
		AndWhere(dbx.NewExp("schemaname = current_schema")).
		AndWhere(dbx.NewExp("indexname NOT LIKE '%_pkey'")).
		All(&indexes)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(indexes))
	for _, idx := range indexes {
		result[idx.Name] = idx.Definition
	}

	return result, nil
}

// DeleteTable drops the specified table.
//
// This method is a no-op if a table with the provided name doesn't exist.
//
// NB! Be aware that this method is vulnerable to SQL injection and the
// "tableName" argument must come only from trusted input!
func (app *BaseApp) DeleteTable(tableName string) error {
	_, err := app.DB().NewQuery(fmt.Sprintf(
		"DROP TABLE IF EXISTS {{%s}} CASCADE",
		tableName,
	)).Execute()

	return err
}

// HasTable checks if a table (or view) with the provided name exists (case insensitive).
// in the current app.DB() instance.
func (app *BaseApp) HasTable(tableName string) bool {
	return app.hasTable(app.DB(), tableName)
}

// AuxHasTable checks if a table (or view) with the provided name exists (case insensitive)
// in the current app.AuxDB() instance.
func (app *BaseApp) AuxHasTable(tableName string) bool {
	return app.hasTable(app.AuxDB(), tableName)
}

func (app *BaseApp) hasTable(db dbx.Builder, tableName string) bool {
	var exists int

	err := db.Select("(1)").
		From("information_schema.tables").
		AndWhere(dbx.HashExp{"table_schema": "public"}).
		AndWhere(dbx.NewExp("LOWER(table_name)=LOWER({:tableName})", dbx.Params{"tableName": tableName})).
		Limit(1).
		Row(&exists)

	return err == nil && exists > 0
}

// Vacuum executes VACUUM on the current app.DB() instance
// in order to reclaim unused data db disk space.
func (app *BaseApp) Vacuum() error {
	return app.vacuum(app.DB())
}

// AuxVacuum executes VACUUM on the current app.AuxDB() instance
// in order to reclaim unused auxiliary db disk space.
func (app *BaseApp) AuxVacuum() error {
	return app.vacuum(app.AuxDB())
}

func (app *BaseApp) vacuum(db dbx.Builder) error {
	_, err := db.NewQuery("VACUUM").Execute()

	return err
}
