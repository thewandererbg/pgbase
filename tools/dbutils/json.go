package dbutils

import (
	"fmt"
	"strings"
)

// JSONEach returns JSON_EACH SQLite string expression with
// some normalizations for non-json columns.
func JSONEach(column string) string {
	return fmt.Sprintf(
		`pb_json_each([[%s]])`,
		column)
}

// JSONArrayLength returns JSON_ARRAY_LENGTH SQLite string expression
// with some normalizations for non-json columns.
//
// It works with both json and non-json column values.
//
// Returns 0 for empty string or NULL column values.
func JSONArrayLength(column string) string {
	return fmt.Sprintf(
		`pb_json_array_length([[%s]])`,
		column,
	)
}

// JSONExtract returns a JSON_EXTRACT SQLite string expression with
// some normalizations for non-json columns.
func JSONExtract(column string, path string) string {
	// prefix the path with dot if it is not starting with array notation
	if path != "" && !strings.HasPrefix(path, "[") {
		path = "." + path
	}

	return fmt.Sprintf(
		`pb_json_extract([[%s]], '$%s')`,
		column,
		path,
	)
}
