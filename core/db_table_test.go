package core_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/thewandererbg/pgbase/core"
	"github.com/thewandererbg/pgbase/tests"
)

func TestHasTable(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	scenarios := []struct {
		tableName string
		expected  bool
	}{
		{"", false},
		{"test", false},
		{core.CollectionNameSuperusers, true},
		{"demo3", true},
		{"DEMO3", true}, // table names are case insensitives by default
		{"view1", true}, // view
	}

	for _, s := range scenarios {
		t.Run(s.tableName, func(t *testing.T) {
			result := app.HasTable(s.tableName)
			if result != s.expected {
				t.Fatalf("Expected %v, got %v", s.expected, result)
			}
		})
	}
}

func TestAuxHasTable(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	scenarios := []struct {
		tableName string
		expected  bool
	}{
		{"", false},
		{"test", false},
		{"_lOGS", true}, // table names are case insensitives by default
	}

	for _, s := range scenarios {
		t.Run(s.tableName, func(t *testing.T) {
			result := app.AuxHasTable(s.tableName)
			if result != s.expected {
				t.Fatalf("Expected %v, got %v", s.expected, result)
			}
		})
	}
}

func TestTableColumns(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	scenarios := []struct {
		tableName string
		expected  []string
	}{
		{"", nil},
		{"_params", []string{"id", "value", "created", "updated"}},
	}

	for i, s := range scenarios {
		t.Run(fmt.Sprintf("%d_%s", i, s.tableName), func(t *testing.T) {
			columns, _ := app.TableColumns(s.tableName)

			if len(columns) != len(s.expected) {
				t.Fatalf("Expected columns %v, got %v", s.expected, columns)
			}

			for _, c := range columns {
				if !slices.Contains(s.expected, c) {
					t.Errorf("Didn't expect column %s", c)
				}
			}
		})
	}
}

func TestTableInfo(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	scenarios := []struct {
		tableName string
		expected  string
	}{
		{"", "null"},
		{"missing", "null"},
		{
			"_params",
			`[{"PK":1,"Index":1,"Name":"id","Type":"character varying(15)","NotNull":true,"DefaultValue":{"String":"length(substr(md5((random())::text), 1, 15))","Valid":true}},{"PK":0,"Index":2,"Name":"value","Type":"text","NotNull":false,"DefaultValue":{"String":"","Valid":false}},{"PK":0,"Index":3,"Name":"created","Type":"timestamp with time zone","NotNull":true,"DefaultValue":{"String":"CURRENT_TIMESTAMP","Valid":true}},{"PK":0,"Index":4,"Name":"updated","Type":"timestamp with time zone","NotNull":true,"DefaultValue":{"String":"CURRENT_TIMESTAMP","Valid":true}}]`,
		},
	}

	for i, s := range scenarios {
		t.Run(fmt.Sprintf("%d_%s", i, s.tableName), func(t *testing.T) {
			rows, _ := app.TableInfo(s.tableName)

			raw, err := json.Marshal(rows)
			if err != nil {
				t.Fatal(err)
			}

			if str := string(raw); str != s.expected {
				t.Fatalf("Expected\n%s\ngot\n%s", s.expected, str)
			}
		})
	}
}

func TestTableIndexes(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	scenarios := []struct {
		tableName string
		expected  []string
	}{
		{"", nil},
		{"missing", nil},
		{
			core.CollectionNameSuperusers,
			[]string{"idx_email_pbc_3142635823", "idx_tokenKey_pbc_3142635823"},
		},
	}

	for i, s := range scenarios {
		t.Run(fmt.Sprintf("%d_%s", i, s.tableName), func(t *testing.T) {
			indexes, _ := app.TableIndexes(s.tableName)

			if len(indexes) != len(s.expected) {
				t.Fatalf("Expected %d indexes, got %d\n%v", len(s.expected), len(indexes), indexes)
			}

			for _, name := range s.expected {
				if v, ok := indexes[name]; !ok || v == "" {
					t.Fatalf("Expected non-empty index %q in \n%v", name, indexes)
				}
			}
		})
	}
}

func TestDeleteTable(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	scenarios := []struct {
		tableName   string
		expectError bool
	}{
		{"", true},
		{"test", false}, // missing tables are ignored
		{"_admins", false},
		{"demo3", false},
	}

	for i, s := range scenarios {
		t.Run(fmt.Sprintf("%d_%s", i, s.tableName), func(t *testing.T) {
			err := app.DeleteTable(s.tableName)

			hasErr := err != nil
			if hasErr != s.expectError {
				t.Fatalf("Expected hasErr %v, got %v", s.expectError, hasErr)
			}
		})
	}
}

func TestVacuum(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	calledQueries := []string{}
	app.DB().(*dbx.DB).QueryLogFunc = func(ctx context.Context, t time.Duration, sql string, rows *sql.Rows, err error) {
		calledQueries = append(calledQueries, sql)
	}
	app.DB().(*dbx.DB).ExecLogFunc = func(ctx context.Context, t time.Duration, sql string, result sql.Result, err error) {
		calledQueries = append(calledQueries, sql)
	}

	if err := app.Vacuum(); err != nil {
		t.Fatal(err)
	}

	if total := len(calledQueries); total != 1 {
		t.Fatalf("Expected 1 query, got %d", total)
	}

	if calledQueries[0] != "VACUUM" {
		t.Fatalf("Expected VACUUM query, got %s", calledQueries[0])
	}
}

func TestAuxVacuum(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	calledQueries := []string{}
	app.AuxDB().(*dbx.DB).QueryLogFunc = func(ctx context.Context, t time.Duration, sql string, rows *sql.Rows, err error) {
		calledQueries = append(calledQueries, sql)
	}
	app.AuxDB().(*dbx.DB).ExecLogFunc = func(ctx context.Context, t time.Duration, sql string, result sql.Result, err error) {
		calledQueries = append(calledQueries, sql)
	}

	if err := app.AuxVacuum(); err != nil {
		t.Fatal(err)
	}

	if total := len(calledQueries); total != 1 {
		t.Fatalf("Expected 1 query, got %d", total)
	}

	if calledQueries[0] != "VACUUM" {
		t.Fatalf("Expected VACUUM query, got %s", calledQueries[0])
	}
}
