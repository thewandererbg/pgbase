package dbutils_test

import (
	"testing"

	"github.com/thewandererbg/pgbase/tools/dbutils"
)

func TestJSONEach(t *testing.T) {
	result := dbutils.JSONEach("a.b")

	expected := "pb_json_each([[a.b]])"

	if result != expected {
		t.Fatalf("Expected\n%v\ngot\n%v", expected, result)
	}
}

func TestJSONArrayLength(t *testing.T) {
	result := dbutils.JSONArrayLength("a.b")

	expected := "pb_json_array_length([[a.b]])"

	if result != expected {
		t.Fatalf("Expected\n%v\ngot\n%v", expected, result)
	}
}

func TestJSONExtract(t *testing.T) {
	scenarios := []struct {
		name     string
		column   string
		path     string
		expected string
	}{
		{
			"empty path",
			"a.b",
			"",
			"pb_json_extract([[a.b]], '$')",
		},
		{
			"starting with array index",
			"a.b",
			"[1].a[2]",
			"pb_json_extract([[a.b]], '$[1].a[2]')",
		},
		{
			"starting with key",
			"a.b",
			"a.b[2].c",
			"pb_json_extract([[a.b]], '$.a.b[2].c')",
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			result := dbutils.JSONExtract(s.column, s.path)

			if result != s.expected {
				t.Fatalf("Expected\n%v\ngot\n%v", s.expected, result)
			}
		})
	}
}
