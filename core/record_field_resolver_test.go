package core_test

import (
	"encoding/json"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/thewandererbg/pgbase/core"
	"github.com/thewandererbg/pgbase/tests"
	"github.com/thewandererbg/pgbase/tools/list"
	"github.com/thewandererbg/pgbase/tools/search"
	"github.com/thewandererbg/pgbase/tools/types"
)

func TestRecordFieldResolverAllowedFields(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	collection, err := app.FindCollectionByNameOrId("demo1")
	if err != nil {
		t.Fatal(err)
	}

	r := core.NewRecordFieldResolver(app, collection, nil, false)

	fields := r.AllowedFields()
	if len(fields) != 8 {
		t.Fatalf("Expected %d original allowed fields, got %d", 8, len(fields))
	}

	// change the allowed fields
	newFields := []string{"a", "b", "c"}
	expected := slices.Clone(newFields)
	r.SetAllowedFields(newFields)

	// change the new fields to ensure that the slice was cloned
	newFields[2] = "d"

	fields = r.AllowedFields()
	if len(fields) != len(expected) {
		t.Fatalf("Expected %d changed allowed fields, got %d", len(expected), len(fields))
	}

	for i, v := range expected {
		if fields[i] != v {
			t.Errorf("[%d] Expected field %q", i, v)
		}
	}
}

func TestRecordFieldResolverAllowHiddenFields(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	collection, err := app.FindCollectionByNameOrId("demo1")
	if err != nil {
		t.Fatal(err)
	}

	r := core.NewRecordFieldResolver(app, collection, nil, false)

	allowHiddenFields := r.AllowHiddenFields()
	if allowHiddenFields {
		t.Fatalf("Expected original allowHiddenFields %v, got %v", allowHiddenFields, !allowHiddenFields)
	}

	// change the flag
	expected := !allowHiddenFields
	r.SetAllowHiddenFields(expected)

	allowHiddenFields = r.AllowHiddenFields()
	if allowHiddenFields != expected {
		t.Fatalf("Expected changed allowHiddenFields %v, got %v", expected, allowHiddenFields)
	}
}

func TestRecordFieldResolverUpdateQuery(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	authRecord, err := app.FindRecordById("users", "4q1xlclmfloku33")
	if err != nil {
		t.Fatal(err)
	}

	requestInfo := &core.RequestInfo{
		Context: "ctx",
		Headers: map[string]string{
			"a": "123",
			"b": "456",
		},
		Query: map[string]string{
			"a": "", // to ensure that :isset returns true because the key exists
			"b": "123",
		},
		Body: map[string]any{
			"a":                  nil, // to ensure that :isset returns true because the key exists
			"b":                  123,
			"number":             10,
			"select_many":        []string{"optionA", "optionC"},
			"rel_one":            "test",
			"rel_many":           []string{"test1", "test2"},
			"file_one":           "test",
			"file_many":          []string{"test1", "test2", "test3"},
			"self_rel_one":       "test",
			"self_rel_many":      []string{"test1"},
			"rel_many_cascade":   []string{"test1", "test2"},
			"rel_one_cascade":    "test1",
			"rel_one_no_cascade": "test1",
		},
		Auth: authRecord,
	}

	scenarios := []struct {
		name               string
		collectionIdOrName string
		rule               string
		allowHiddenFields  bool
		expectQuery        string
	}{
		{
			"non relation field (with all default operators)",
			"demo4",
			"title = true || title != 'test' || title ~ 'test1' || title !~ '%test2' || title > 1 || title >= 2 || title < 3 || title <= 4",
			false,
			`SELECT "demo4".* FROM "demo4" WHERE ([[demo4.title]] = true OR [[demo4.title]] IS DISTINCT FROM {:TEST} OR [[demo4.title]]::text LIKE {:TEST} ESCAPE '\' OR [[demo4.title]]::text NOT LIKE {:TEST} ESCAPE '\' OR [[demo4.title]] > {:TEST} OR [[demo4.title]] >= {:TEST} OR [[demo4.title]] < {:TEST} OR [[demo4.title]] <= {:TEST})`,
		},
		{
			"non relation field (with all opt/any operators)",
			"demo4",
			"title ?= true || title ?!= 'test' || title ?~ 'test1' || title ?!~ '%test2' || title ?> 1 || title ?>= 2 || title ?< 3 || title ?<= 4",
			false,
			`SELECT "demo4".* FROM "demo4" WHERE ([[demo4.title]] = true OR [[demo4.title]] IS DISTINCT FROM {:TEST} OR [[demo4.title]]::text LIKE {:TEST} ESCAPE '\' OR [[demo4.title]]::text NOT LIKE {:TEST} ESCAPE '\' OR [[demo4.title]] > {:TEST} OR [[demo4.title]] >= {:TEST} OR [[demo4.title]] < {:TEST} OR [[demo4.title]] <= {:TEST})`,
		},
		{
			"single direct rel",
			"demo4",
			"self_rel_one > true",
			false,
			`SELECT "demo4".* FROM "demo4" WHERE [[demo4.self_rel_one]] > true`,
		},
		{
			"single direct rel (with id)",
			"demo4",
			"self_rel_one.id > true",
			false,
			`SELECT "demo4".* FROM "demo4" WHERE [[demo4.self_rel_one]] > true`,
		},
		{
			"single direct rel (with non-id field)",
			"demo4",
			"self_rel_one.created > true",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN "demo4" "demo4_self_rel_one" ON [[demo4_self_rel_one.id]] = [[demo4.self_rel_one]] WHERE [[demo4_self_rel_one.created]] > true`,
		},
		{
			"multiple direct rel",
			"demo4",
			"self_rel_many ?> true",
			false,
			`SELECT "demo4".* FROM "demo4" WHERE [[demo4.self_rel_many]] > true`,
		},
		{
			"multiple direct rel (with id)",
			"demo4",
			"self_rel_many.id ?> true",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN pb_json_each([[demo4.self_rel_many]]) "demo4_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "demo4_self_rel_many" ON [[demo4_self_rel_many.id]] = [[demo4_self_rel_many_je.value]] WHERE [[demo4_self_rel_many.id]] > true`,
		},
		{
			"nested single rel (self rel)",
			"demo4",
			"self_rel_one.title > true",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN "demo4" "demo4_self_rel_one" ON [[demo4_self_rel_one.id]] = [[demo4.self_rel_one]] WHERE [[demo4_self_rel_one.title]] > true`,
		},
		{
			"nested single rel (other collection)",
			"demo4",
			"rel_one_cascade.title > true",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN "demo3" "demo4_rel_one_cascade" ON [[demo4_rel_one_cascade.id]] = [[demo4.rel_one_cascade]] WHERE [[demo4_rel_one_cascade.title]] > true`,
		},
		{
			"non-relation field + single rel",
			"demo4",
			"title > true || self_rel_one.title > true",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN "demo4" "demo4_self_rel_one" ON [[demo4_self_rel_one.id]] = [[demo4.self_rel_one]] WHERE ([[demo4.title]] > true OR [[demo4_self_rel_one.title]] > true)`,
		},
		{
			"nested incomplete relations (opt/any operator)",
			"demo4",
			"self_rel_many.self_rel_one ?> true",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN pb_json_each([[demo4.self_rel_many]]) "demo4_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "demo4_self_rel_many" ON [[demo4_self_rel_many.id]] = [[demo4_self_rel_many_je.value]] WHERE [[demo4_self_rel_many.self_rel_one]] > true`,
		},
		{
			"nested incomplete relations (multi-match operator)",
			"demo4",
			"self_rel_many.self_rel_one > true",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN pb_json_each([[demo4.self_rel_many]]) "demo4_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "demo4_self_rel_many" ON [[demo4_self_rel_many.id]] = [[demo4_self_rel_many_je.value]] WHERE ((([[demo4_self_rel_many.self_rel_one]] > true) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm_demo4_self_rel_many.self_rel_one]] as [[multiMatchValue]] FROM "demo4" "__mm_demo4" LEFT JOIN pb_json_each([[__mm_demo4.self_rel_many]]) "__mm_demo4_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "__mm_demo4_self_rel_many" ON [[__mm_demo4_self_rel_many.id]] = [[__mm_demo4_self_rel_many_je.value]] WHERE "__mm_demo4"."id" = "demo4"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] > true)))))`,
		},
		{
			"nested complete relations (opt/any operator)",
			"demo4",
			"self_rel_many.self_rel_one.title ?> true",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN pb_json_each([[demo4.self_rel_many]]) "demo4_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "demo4_self_rel_many" ON [[demo4_self_rel_many.id]] = [[demo4_self_rel_many_je.value]] LEFT JOIN "demo4" "demo4_self_rel_many_self_rel_one" ON [[demo4_self_rel_many_self_rel_one.id]] = [[demo4_self_rel_many.self_rel_one]] WHERE [[demo4_self_rel_many_self_rel_one.title]] > true`,
		},
		{
			"nested complete relations (multi-match operator)",
			"demo4",
			"self_rel_many.self_rel_one.title > true",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN pb_json_each([[demo4.self_rel_many]]) "demo4_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "demo4_self_rel_many" ON [[demo4_self_rel_many.id]] = [[demo4_self_rel_many_je.value]] LEFT JOIN "demo4" "demo4_self_rel_many_self_rel_one" ON [[demo4_self_rel_many_self_rel_one.id]] = [[demo4_self_rel_many.self_rel_one]] WHERE ((([[demo4_self_rel_many_self_rel_one.title]] > true) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm_demo4_self_rel_many_self_rel_one.title]] as [[multiMatchValue]] FROM "demo4" "__mm_demo4" LEFT JOIN pb_json_each([[__mm_demo4.self_rel_many]]) "__mm_demo4_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "__mm_demo4_self_rel_many" ON [[__mm_demo4_self_rel_many.id]] = [[__mm_demo4_self_rel_many_je.value]] LEFT JOIN "demo4" "__mm_demo4_self_rel_many_self_rel_one" ON [[__mm_demo4_self_rel_many_self_rel_one.id]] = [[__mm_demo4_self_rel_many.self_rel_one]] WHERE "__mm_demo4"."id" = "demo4"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] > true)))))`,
		},
		{
			"repeated nested relations (opt/any operator)",
			"demo4",
			"self_rel_many.self_rel_one.self_rel_many.self_rel_one.title ?> true",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN pb_json_each([[demo4.self_rel_many]]) "demo4_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "demo4_self_rel_many" ON [[demo4_self_rel_many.id]] = [[demo4_self_rel_many_je.value]] LEFT JOIN "demo4" "demo4_self_rel_many_self_rel_one" ON [[demo4_self_rel_many_self_rel_one.id]] = [[demo4_self_rel_many.self_rel_one]] LEFT JOIN pb_json_each([[demo4_self_rel_many_self_rel_one.self_rel_many]]) "demo4_self_rel_many_self_rel_one_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "demo4_self_rel_many_self_rel_one_self_rel_many" ON [[demo4_self_rel_many_self_rel_one_self_rel_many.id]] = [[demo4_self_rel_many_self_rel_one_self_rel_many_je.value]] LEFT JOIN "demo4" "demo4_self_rel_many_self_rel_one_self_rel_many_self_rel_one" ON [[demo4_self_rel_many_self_rel_one_self_rel_many_self_rel_one.id]] = [[demo4_self_rel_many_self_rel_one_self_rel_many.self_rel_one]] WHERE [[demo4_self_rel_many_self_rel_one_self_rel_many_self_rel_one.title]] > true`,
		},
		{
			"repeated nested relations (multi-match operator)",
			"demo4",
			"self_rel_many.self_rel_one.self_rel_many.self_rel_one.title > true",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN pb_json_each([[demo4.self_rel_many]]) "demo4_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "demo4_self_rel_many" ON [[demo4_self_rel_many.id]] = [[demo4_self_rel_many_je.value]] LEFT JOIN "demo4" "demo4_self_rel_many_self_rel_one" ON [[demo4_self_rel_many_self_rel_one.id]] = [[demo4_self_rel_many.self_rel_one]] LEFT JOIN pb_json_each([[demo4_self_rel_many_self_rel_one.self_rel_many]]) "demo4_self_rel_many_self_rel_one_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "demo4_self_rel_many_self_rel_one_self_rel_many" ON [[demo4_self_rel_many_self_rel_one_self_rel_many.id]] = [[demo4_self_rel_many_self_rel_one_self_rel_many_je.value]] LEFT JOIN "demo4" "demo4_self_rel_many_self_rel_one_self_rel_many_self_rel_one" ON [[demo4_self_rel_many_self_rel_one_self_rel_many_self_rel_one.id]] = [[demo4_self_rel_many_self_rel_one_self_rel_many.self_rel_one]] WHERE ((([[demo4_self_rel_many_self_rel_one_self_rel_many_self_rel_one.title]] > true) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm_demo4_self_rel_many_self_rel_one_self_rel_many_self_rel_one.title]] as [[multiMatchValue]] FROM "demo4" "__mm_demo4" LEFT JOIN pb_json_each([[__mm_demo4.self_rel_many]]) "__mm_demo4_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "__mm_demo4_self_rel_many" ON [[__mm_demo4_self_rel_many.id]] = [[__mm_demo4_self_rel_many_je.value]] LEFT JOIN "demo4" "__mm_demo4_self_rel_many_self_rel_one" ON [[__mm_demo4_self_rel_many_self_rel_one.id]] = [[__mm_demo4_self_rel_many.self_rel_one]] LEFT JOIN pb_json_each([[__mm_demo4_self_rel_many_self_rel_one.self_rel_many]]) "__mm_demo4_self_rel_many_self_rel_one_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "__mm_demo4_self_rel_many_self_rel_one_self_rel_many" ON [[__mm_demo4_self_rel_many_self_rel_one_self_rel_many.id]] = [[__mm_demo4_self_rel_many_self_rel_one_self_rel_many_je.value]] LEFT JOIN "demo4" "__mm_demo4_self_rel_many_self_rel_one_self_rel_many_self_rel_one" ON [[__mm_demo4_self_rel_many_self_rel_one_self_rel_many_self_rel_one.id]] = [[__mm_demo4_self_rel_many_self_rel_one_self_rel_many.self_rel_one]] WHERE "__mm_demo4"."id" = "demo4"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] > true)))))`,
		},
		{
			"multiple relations (opt/any operators)",
			"demo4",
			"self_rel_many.title ?= 'test' || self_rel_one.json_object.a ?> true",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN pb_json_each([[demo4.self_rel_many]]) "demo4_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "demo4_self_rel_many" ON [[demo4_self_rel_many.id]] = [[demo4_self_rel_many_je.value]] LEFT JOIN "demo4" "demo4_self_rel_one" ON [[demo4_self_rel_one.id]] = [[demo4.self_rel_one]] WHERE ([[demo4_self_rel_many.title]] = {:TEST} OR pb_json_extract([[demo4_self_rel_one.json_object]], '$.a') > true)`,
		},
		{
			"multiple relations (multi-match operators)",
			"demo4",
			"self_rel_many.title = 'test' || self_rel_one.json_object.a > true",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN pb_json_each([[demo4.self_rel_many]]) "demo4_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "demo4_self_rel_many" ON [[demo4_self_rel_many.id]] = [[demo4_self_rel_many_je.value]] LEFT JOIN "demo4" "demo4_self_rel_one" ON [[demo4_self_rel_one.id]] = [[demo4.self_rel_one]] WHERE ((([[demo4_self_rel_many.title]] = {:TEST}) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm_demo4_self_rel_many.title]] as [[multiMatchValue]] FROM "demo4" "__mm_demo4" LEFT JOIN pb_json_each([[__mm_demo4.self_rel_many]]) "__mm_demo4_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "__mm_demo4_self_rel_many" ON [[__mm_demo4_self_rel_many.id]] = [[__mm_demo4_self_rel_many_je.value]] WHERE "__mm_demo4"."id" = "demo4"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] = {:TEST})))) OR pb_json_extract([[demo4_self_rel_one.json_object]], '$.a') > true)`,
		},
		{
			"back relations via single relation field (without unique index)",
			"demo3",
			"demo4_via_rel_one_cascade.id = true",
			false,
			`SELECT DISTINCT "demo3".* FROM "demo3" LEFT JOIN "demo4" "demo3_demo4_via_rel_one_cascade" ON [[demo3.id]] IN (SELECT [[demo3_demo4_via_rel_one_cascade_je.value]] FROM pb_json_each([[demo3_demo4_via_rel_one_cascade.rel_one_cascade]]) {{demo3_demo4_via_rel_one_cascade_je}}) WHERE ((([[demo3_demo4_via_rel_one_cascade.id]] = true) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm_demo3_demo4_via_rel_one_cascade.id]] as [[multiMatchValue]] FROM "demo3" "__mm_demo3" LEFT JOIN "demo4" "__mm_demo3_demo4_via_rel_one_cascade" ON [[__mm_demo3.id]] IN (SELECT [[__mm_demo3_demo4_via_rel_one_cascade_je.value]] FROM pb_json_each([[__mm_demo3_demo4_via_rel_one_cascade.rel_one_cascade]]) {{__mm_demo3_demo4_via_rel_one_cascade_je}}) WHERE "__mm_demo3"."id" = "demo3"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] = true)))))`,
		},
		{
			"back relations via single relation field (with unique index)",
			"demo3",
			"demo4_via_rel_one_unique.id = true",
			false,
			`SELECT DISTINCT "demo3".* FROM "demo3" LEFT JOIN "demo4" "demo3_demo4_via_rel_one_unique" ON [[demo3_demo4_via_rel_one_unique.rel_one_unique]] = [[demo3.id]] WHERE [[demo3_demo4_via_rel_one_unique.id]] = true`,
		},
		{
			"back relations via multiple relation field (opt/any operators)",
			"demo3",
			"demo4_via_rel_many_cascade.id ?= true",
			false,
			`SELECT DISTINCT "demo3".* FROM "demo3" LEFT JOIN "demo4" "demo3_demo4_via_rel_many_cascade" ON [[demo3.id]] IN (SELECT [[demo3_demo4_via_rel_many_cascade_je.value]] FROM pb_json_each([[demo3_demo4_via_rel_many_cascade.rel_many_cascade]]) {{demo3_demo4_via_rel_many_cascade_je}}) WHERE [[demo3_demo4_via_rel_many_cascade.id]] = true`,
		},
		{
			"back relations via multiple relation field (multi-match operators)",
			"demo3",
			"demo4_via_rel_many_cascade.id = true",
			false,
			`SELECT DISTINCT "demo3".* FROM "demo3" LEFT JOIN "demo4" "demo3_demo4_via_rel_many_cascade" ON [[demo3.id]] IN (SELECT [[demo3_demo4_via_rel_many_cascade_je.value]] FROM pb_json_each([[demo3_demo4_via_rel_many_cascade.rel_many_cascade]]) {{demo3_demo4_via_rel_many_cascade_je}}) WHERE ((([[demo3_demo4_via_rel_many_cascade.id]] = true) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm_demo3_demo4_via_rel_many_cascade.id]] as [[multiMatchValue]] FROM "demo3" "__mm_demo3" LEFT JOIN "demo4" "__mm_demo3_demo4_via_rel_many_cascade" ON [[__mm_demo3.id]] IN (SELECT [[__mm_demo3_demo4_via_rel_many_cascade_je.value]] FROM pb_json_each([[__mm_demo3_demo4_via_rel_many_cascade.rel_many_cascade]]) {{__mm_demo3_demo4_via_rel_many_cascade_je}}) WHERE "__mm_demo3"."id" = "demo3"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] = true)))))`,
		},
		{
			"back relations via unique multiple relation field (should be the same as multi-match)",
			"demo3",
			"demo4_via_rel_many_unique.id = true",
			false,
			`SELECT DISTINCT "demo3".* FROM "demo3" LEFT JOIN "demo4" "demo3_demo4_via_rel_many_unique" ON [[demo3.id]] IN (SELECT [[demo3_demo4_via_rel_many_unique_je.value]] FROM pb_json_each([[demo3_demo4_via_rel_many_unique.rel_many_unique]]) {{demo3_demo4_via_rel_many_unique_je}}) WHERE ((([[demo3_demo4_via_rel_many_unique.id]] = true) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm_demo3_demo4_via_rel_many_unique.id]] as [[multiMatchValue]] FROM "demo3" "__mm_demo3" LEFT JOIN "demo4" "__mm_demo3_demo4_via_rel_many_unique" ON [[__mm_demo3.id]] IN (SELECT [[__mm_demo3_demo4_via_rel_many_unique_je.value]] FROM pb_json_each([[__mm_demo3_demo4_via_rel_many_unique.rel_many_unique]]) {{__mm_demo3_demo4_via_rel_many_unique_je}}) WHERE "__mm_demo3"."id" = "demo3"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] = true)))))`,
		},
		{
			"recursive back relations",
			"demo3",
			"demo4_via_rel_many_cascade.rel_one_cascade.demo4_via_rel_many_cascade.id ?= true",
			false,
			`SELECT DISTINCT "demo3".* FROM "demo3" LEFT JOIN "demo4" "demo3_demo4_via_rel_many_cascade" ON [[demo3.id]] IN (SELECT [[demo3_demo4_via_rel_many_cascade_je.value]] FROM pb_json_each([[demo3_demo4_via_rel_many_cascade.rel_many_cascade]]) {{demo3_demo4_via_rel_many_cascade_je}}) LEFT JOIN "demo3" "demo3_demo4_via_rel_many_cascade_rel_one_cascade" ON [[demo3_demo4_via_rel_many_cascade_rel_one_cascade.id]] = [[demo3_demo4_via_rel_many_cascade.rel_one_cascade]] LEFT JOIN "demo4" "demo3_demo4_via_rel_many_cascade_rel_one_cascade_demo4_via_rel_many_cascade" ON [[demo3_demo4_via_rel_many_cascade_rel_one_cascade.id]] IN (SELECT [[demo3_demo4_via_rel_many_cascade_rel_one_cascade_demo4_via_rel_many_cascade_je.value]] FROM pb_json_each([[demo3_demo4_via_rel_many_cascade_rel_one_cascade_demo4_via_rel_many_cascade.rel_many_cascade]]) {{demo3_demo4_via_rel_many_cascade_rel_one_cascade_demo4_via_rel_many_cascade_je}}) WHERE [[demo3_demo4_via_rel_many_cascade_rel_one_cascade_demo4_via_rel_many_cascade.id]] = true`,
		},
		{
			"@collection join (opt/any operators)",
			"demo4",
			"@collection.demo1.text ?> true || @collection.demo2.active ?> true || @collection.demo1:demo1_alias.file_one ?> true",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN "demo1" "__collection_demo1" ON TRUE LEFT JOIN "demo2" "__collection_demo2" ON TRUE LEFT JOIN "demo1" "__collection_alias_demo1_alias" ON TRUE WHERE ([[__collection_demo1.text]] > true OR [[__collection_demo2.active]] > true OR [[__collection_alias_demo1_alias.file_one]] > true)`,
		},
		{
			"@collection join (multi-match operators)",
			"demo4",
			"@collection.demo1.text > true || @collection.demo2.active > true || @collection.demo1.file_one > true",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN "demo1" "__collection_demo1" ON TRUE LEFT JOIN "demo2" "__collection_demo2" ON TRUE WHERE ((([[__collection_demo1.text]] > true) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm__collection_demo1.text]] as [[multiMatchValue]] FROM "demo4" "__mm_demo4" LEFT JOIN "demo1" "__mm__collection_demo1" WHERE "__mm_demo4"."id" = "demo4"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] > true)))) OR (([[__collection_demo2.active]] > true) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm__collection_demo2.active]] as [[multiMatchValue]] FROM "demo4" "__mm_demo4" LEFT JOIN "demo2" "__mm__collection_demo2" WHERE "__mm_demo4"."id" = "demo4"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] > true)))) OR (([[__collection_demo1.file_one]] > true) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm__collection_demo1.file_one]] as [[multiMatchValue]] FROM "demo4" "__mm_demo4" LEFT JOIN "demo1" "__mm__collection_demo1" WHERE "__mm_demo4"."id" = "demo4"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] > true)))))`,
		},
		{
			"@request.auth fields",
			"demo4",
			"@request.auth.id > true || @request.auth.username > true || @request.auth.rel.title > true || @request.body.demo < true || @request.auth.missingA.missingB > false",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN "users" "__auth_users" ON "__auth_users"."id"={:TEST} LEFT JOIN "demo2" "__auth_users_rel" ON [[__auth_users_rel.id]] = [[__auth_users.rel]] WHERE ({:TEST} > true OR [[__auth_users.username]] > true OR [[__auth_users_rel.title]] > true OR NULL < true OR NULL > false)`,
		},
		{
			"@request.* static fields",
			"demo4",
			"@request.context = true || @request.query.a = true || @request.query.b = true || @request.query.missing = true || @request.headers.a = true || @request.headers.missing = true",
			false,
			`SELECT "demo4".* FROM "demo4" WHERE ({:TEST} = true OR '' = true OR {:TEST} = true OR '' = true OR {:TEST} = true OR '' = true)`,
		},
		{
			"hidden field with system filters (multi-match and ignore emailVisibility)",
			"demo4",
			"@collection.users.email > true || @request.auth.email > true",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN "users" "__collection_users" ON TRUE WHERE ((([[__collection_users.email]] > true) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm__collection_users.email]] as [[multiMatchValue]] FROM "demo4" "__mm_demo4" LEFT JOIN "users" "__mm__collection_users" WHERE "__mm_demo4"."id" = "demo4"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] > true)))) OR {:TEST} > true)`,
		},
		{
			"hidden field (add emailVisibility)",
			"users",
			"id > true || email > true || email:lower > false",
			false,
			`SELECT "users".* FROM "users" WHERE ([[users.id]] > true OR (([[users.email]] > true) AND ([[users.emailVisibility]] = TRUE)) OR ((LOWER([[users.email]]) > false) AND ([[users.emailVisibility]] = TRUE)))`,
		},
		{
			"hidden field (force ignore emailVisibility)",
			"users",
			"email > true",
			true,
			`SELECT "users".* FROM "users" WHERE [[users.email]] > true`,
		},
		{
			"static @request fields with :lower modifier",
			"demo1",
			"@request.body.a:lower > true ||@request.body.b:lower > true ||@request.body.c:lower > true ||@request.query.a:lower > true ||@request.query.b:lower > true ||@request.query.c:lower > true ||@request.headers.a:lower > true ||@request.headers.c:lower > true",
			false,
			`SELECT "demo1".* FROM "demo1" WHERE (NULL > true OR LOWER({:TEST}) > true OR NULL > true OR LOWER({:TEST}) > true OR LOWER({:TEST}) > true OR NULL > true OR LOWER({:TEST}) > true OR NULL > true)`,
		},
		{
			"collection fields with :lower modifier",
			"demo1",
			"@request.body.rel_one:lower > true ||@request.body.rel_many:lower > true ||@request.body.rel_many.email:lower > true ||text:lower > true ||bool:lower > true ||url:lower > true ||select_one:lower > true ||select_many:lower > true ||file_one:lower > true ||file_many:lower > true ||number:lower > true ||email:lower > true ||datetime:lower > true ||json:lower > true ||rel_one:lower > true ||rel_many:lower > true ||rel_many.name:lower > true ||created:lower > true",
			false,
			`SELECT DISTINCT "demo1".* FROM "demo1" LEFT JOIN "users" "__data_users_rel_many" ON [[__data_users_rel_many.id]] IN ({:TEST}, {:TEST}) LEFT JOIN pb_json_each([[demo1.rel_many]]) "demo1_rel_many_je" ON TRUE LEFT JOIN "users" "demo1_rel_many" ON [[demo1_rel_many.id]] = [[demo1_rel_many_je.value]] WHERE (LOWER({:TEST}) > true OR LOWER({:TEST}) > true OR ((LOWER([[__data_users_rel_many.email]]) > true) AND (NOT EXISTS (SELECT 1 FROM (SELECT LOWER([[__data_mm_users_rel_many.email]]) as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN "users" "__data_mm_users_rel_many" ON [[__data_mm_users_rel_many.id]] IN ({:TEST}, {:TEST}) WHERE "__mm_demo1"."id" = "demo1"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] > true)))) OR LOWER([[demo1.text]]) > true OR LOWER([[demo1.bool]]) > true OR LOWER([[demo1.url]]) > true OR LOWER([[demo1.select_one]]) > true OR LOWER([[demo1.select_many]]) > true OR LOWER([[demo1.file_one]]) > true OR LOWER([[demo1.file_many]]) > true OR LOWER([[demo1.number]]) > true OR LOWER([[demo1.email]]) > true OR LOWER([[demo1.datetime]]) > true OR LOWER(pb_json_extract([[demo1.json]], '$')) > true OR LOWER([[demo1.rel_one]]) > true OR LOWER([[demo1.rel_many]]) > true OR ((LOWER([[demo1_rel_many.name]]) > true) AND (NOT EXISTS (SELECT 1 FROM (SELECT LOWER([[__mm_demo1_rel_many.name]]) as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN pb_json_each([[__mm_demo1.rel_many]]) "__mm_demo1_rel_many_je" ON TRUE LEFT JOIN "users" "__mm_demo1_rel_many" ON [[__mm_demo1_rel_many.id]] = [[__mm_demo1_rel_many_je.value]] WHERE "__mm_demo1"."id" = "demo1"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] > true)))) OR LOWER([[demo1.created]]) > true)`,
		},
		{
			"isset modifier",
			"demo1",
			"@request.body.a:isset > true ||@request.body.b:isset > true ||@request.body.c:isset > true ||@request.query.a:isset > true ||@request.query.b:isset > true ||@request.query.c:isset > true ||@request.headers.a:isset > true ||@request.headers.c:isset > true",
			false,
			`SELECT "demo1".* FROM "demo1" WHERE (TRUE > true OR TRUE > true OR FALSE > true OR TRUE > true OR TRUE > true OR FALSE > true OR TRUE > true OR FALSE > true)`,
		},
		{
			"@request.body.rel.* fields",
			"demo4",
			"@request.body.rel_one_cascade.title > true &&@request.body.rel_one_no_cascade.title < true &&@request.body.self_rel_many.title = true",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN "demo3" "__data_demo3_rel_one_cascade" ON [[__data_demo3_rel_one_cascade.id]]={:TEST} LEFT JOIN "demo3" "__data_demo3_rel_one_no_cascade" ON [[__data_demo3_rel_one_no_cascade.id]]={:TEST} LEFT JOIN "demo4" "__data_demo4_self_rel_many" ON [[__data_demo4_self_rel_many.id]]={:TEST} WHERE ([[__data_demo3_rel_one_cascade.title]] > true AND [[__data_demo3_rel_one_no_cascade.title]] < true AND (([[__data_demo4_self_rel_many.title]] = true) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__data_mm_demo4_self_rel_many.title]] as [[multiMatchValue]] FROM "demo4" "__mm_demo4" LEFT JOIN "demo4" "__data_mm_demo4_self_rel_many" ON [[__data_mm_demo4_self_rel_many.id]]={:TEST} WHERE "__mm_demo4"."id" = "demo4"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] = true)))))`,
		},
		{
			"@request.body.arrayble:each fields",
			"demo1",
			"@request.body.select_one:each > true &&@request.body.select_one:each ?< true &&@request.body.select_many:each > true &&@request.body.select_many:each ?< true &&@request.body.file_one:each > true &&@request.body.file_one:each ?< true &&@request.body.file_many:each > true &&@request.body.file_many:each ?< true &&@request.body.rel_one:each > true &&@request.body.rel_one:each ?< true &&@request.body.rel_many:each > true &&@request.body.rel_many:each ?< true",
			false,
			`SELECT DISTINCT "demo1".* FROM "demo1" LEFT JOIN json_each({:TEST}) "__dataEach_select_one_je" ON TRUE LEFT JOIN json_each({:TEST}) "__dataEach_select_many_je" ON TRUE LEFT JOIN json_each({:TEST}) "__dataEach_file_one_je" ON TRUE LEFT JOIN json_each({:TEST}) "__dataEach_file_many_je" ON TRUE LEFT JOIN json_each({:TEST}) "__dataEach_rel_one_je" ON TRUE LEFT JOIN json_each({:TEST}) "__dataEach_rel_many_je" ON TRUE WHERE ([[__dataEach_select_one_je.value]] > true AND [[__dataEach_select_one_je.value]] < true AND (([[__dataEach_select_many_je.value]] > true) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm__dataEach_select_many_je.value]] as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN json_each({:TEST}) "__mm__dataEach_select_many_je" WHERE "__mm_demo1"."id" = "demo1"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] > true)))) AND [[__dataEach_select_many_je.value]] < true AND [[__dataEach_file_one_je.value]] > true AND [[__dataEach_file_one_je.value]] < true AND (([[__dataEach_file_many_je.value]] > true) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm__dataEach_file_many_je.value]] as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN json_each({:TEST}) "__mm__dataEach_file_many_je" WHERE "__mm_demo1"."id" = "demo1"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] > true)))) AND [[__dataEach_file_many_je.value]] < true AND [[__dataEach_rel_one_je.value]] > true AND [[__dataEach_rel_one_je.value]] < true AND (([[__dataEach_rel_many_je.value]] > true) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm__dataEach_rel_many_je.value]] as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN json_each({:TEST}) "__mm__dataEach_rel_many_je" WHERE "__mm_demo1"."id" = "demo1"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] > true)))) AND [[__dataEach_rel_many_je.value]] < true)`,
		},
		{
			"regular arrayble:each fields",
			"demo1",
			"select_one:each > true &&select_one:each ?< true &&select_many:each > true &&select_many:each ?< true &&file_one:each > true &&file_one:each ?< true &&file_many:each > true &&file_many:each ?< true &&rel_one:each > true &&rel_one:each ?< true &&rel_many:each > true &&rel_many:each ?< true",
			false,
			`SELECT DISTINCT "demo1".* FROM "demo1" LEFT JOIN pb_json_each([[demo1.select_one]]) "demo1_select_one_je" ON TRUE LEFT JOIN pb_json_each([[demo1.select_many]]) "demo1_select_many_je" ON TRUE LEFT JOIN pb_json_each([[demo1.file_one]]) "demo1_file_one_je" ON TRUE LEFT JOIN pb_json_each([[demo1.file_many]]) "demo1_file_many_je" ON TRUE LEFT JOIN pb_json_each([[demo1.rel_one]]) "demo1_rel_one_je" ON TRUE LEFT JOIN pb_json_each([[demo1.rel_many]]) "demo1_rel_many_je" ON TRUE WHERE ([[demo1_select_one_je.value]] > true AND [[demo1_select_one_je.value]] < true AND (([[demo1_select_many_je.value]] > true) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm_demo1_select_many_je.value]] as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN pb_json_each([[__mm_demo1.select_many]]) "__mm_demo1_select_many_je" WHERE "__mm_demo1"."id" = "demo1"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] > true)))) AND [[demo1_select_many_je.value]] < true AND [[demo1_file_one_je.value]] > true AND [[demo1_file_one_je.value]] < true AND (([[demo1_file_many_je.value]] > true) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm_demo1_file_many_je.value]] as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN pb_json_each([[__mm_demo1.file_many]]) "__mm_demo1_file_many_je" WHERE "__mm_demo1"."id" = "demo1"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] > true)))) AND [[demo1_file_many_je.value]] < true AND [[demo1_rel_one_je.value]] > true AND [[demo1_rel_one_je.value]] < true AND (([[demo1_rel_many_je.value]] > true) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm_demo1_rel_many_je.value]] as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN pb_json_each([[__mm_demo1.rel_many]]) "__mm_demo1_rel_many_je" WHERE "__mm_demo1"."id" = "demo1"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] > true)))) AND [[demo1_rel_many_je.value]] < true)`,
		},
		{
			"arrayble:each vs arrayble:each",
			"demo1",
			"select_one:each != select_many:each &&select_many:each > select_one:each &&select_many:each ?< select_one:each &&select_many:each = @request.body.select_many:each",
			false,
			`SELECT DISTINCT "demo1".* FROM "demo1" LEFT JOIN pb_json_each([[demo1.select_one]]) "demo1_select_one_je" ON TRUE LEFT JOIN pb_json_each([[demo1.select_many]]) "demo1_select_many_je" ON TRUE LEFT JOIN json_each({:TEST}) "__dataEach_select_many_je" ON TRUE WHERE (((COALESCE([[demo1_select_one_je.value]], '') IS DISTINCT FROM COALESCE([[demo1_select_many_je.value]], '')) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm_demo1_select_many_je.value]] as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN pb_json_each([[__mm_demo1.select_many]]) "__mm_demo1_select_many_je" WHERE "__mm_demo1"."id" = "demo1"."id") {{__smTEST}} WHERE NOT (COALESCE([[demo1_select_one_je.value]], '') IS DISTINCT FROM COALESCE([[__smTEST.multiMatchValue]], ''))))) AND (([[demo1_select_many_je.value]] > [[demo1_select_one_je.value]]) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm_demo1_select_many_je.value]] as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN pb_json_each([[__mm_demo1.select_many]]) "__mm_demo1_select_many_je" WHERE "__mm_demo1"."id" = "demo1"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] > [[demo1_select_one_je.value]])))) AND [[demo1_select_many_je.value]] < [[demo1_select_one_je.value]] AND (([[demo1_select_many_je.value]] = [[__dataEach_select_many_je.value]]) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm_demo1_select_many_je.value]] as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN pb_json_each([[__mm_demo1.select_many]]) "__mm_demo1_select_many_je" WHERE "__mm_demo1"."id" = "demo1"."id") {{__mlTEST}} LEFT JOIN (SELECT [[__mm__dataEach_select_many_je.value]] as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN json_each({:TEST}) "__mm__dataEach_select_many_je" WHERE "__mm_demo1"."id" = "demo1"."id") {{__mrTEST}} WHERE NOT (COALESCE([[__mlTEST.multiMatchValue]], '') = COALESCE([[__mrTEST.multiMatchValue]], ''))))))`,
		},
		{
			"mixed multi-match vs multi-match",
			"demo1",
			"rel_many.rel.active != rel_many.name &&rel_many.rel.active ?= rel_many.name &&rel_many.rel.title ~ rel_one.email &&@collection.demo2.active = rel_many.rel.active &&@collection.demo2.active ?= rel_many.rel.active &&rel_many.email > @request.body.rel_many.email",
			false,
			`SELECT DISTINCT "demo1".* FROM "demo1" LEFT JOIN pb_json_each([[demo1.rel_many]]) "demo1_rel_many_je" ON TRUE LEFT JOIN "users" "demo1_rel_many" ON [[demo1_rel_many.id]] = [[demo1_rel_many_je.value]] LEFT JOIN "demo2" "demo1_rel_many_rel" ON [[demo1_rel_many_rel.id]] = [[demo1_rel_many.rel]] LEFT JOIN "demo1" "demo1_rel_one" ON [[demo1_rel_one.id]] = [[demo1.rel_one]] LEFT JOIN "demo2" "__collection_demo2" ON TRUE LEFT JOIN "users" "__data_users_rel_many" ON [[__data_users_rel_many.id]] IN ({:TEST}, {:TEST}) WHERE (((COALESCE([[demo1_rel_many_rel.active]], '') IS DISTINCT FROM COALESCE([[demo1_rel_many.name]], '')) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm_demo1_rel_many_rel.active]] as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN pb_json_each([[__mm_demo1.rel_many]]) "__mm_demo1_rel_many_je" ON TRUE LEFT JOIN "users" "__mm_demo1_rel_many" ON [[__mm_demo1_rel_many.id]] = [[__mm_demo1_rel_many_je.value]] LEFT JOIN "demo2" "__mm_demo1_rel_many_rel" ON [[__mm_demo1_rel_many_rel.id]] = [[__mm_demo1_rel_many.rel]] WHERE "__mm_demo1"."id" = "demo1"."id") {{__mlTEST}} LEFT JOIN (SELECT [[__mm_demo1_rel_many.name]] as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN pb_json_each([[__mm_demo1.rel_many]]) "__mm_demo1_rel_many_je" ON TRUE LEFT JOIN "users" "__mm_demo1_rel_many" ON [[__mm_demo1_rel_many.id]] = [[__mm_demo1_rel_many_je.value]] WHERE "__mm_demo1"."id" = "demo1"."id") {{__mrTEST}} WHERE NOT (COALESCE([[__mlTEST.multiMatchValue]], '') IS DISTINCT FROM COALESCE([[__mrTEST.multiMatchValue]], ''))))) AND COALESCE([[demo1_rel_many_rel.active]], '') = COALESCE([[demo1_rel_many.name]], '') AND (([[demo1_rel_many_rel.title]]::text LIKE ('%' || [[demo1_rel_one.email]] || '%') ESCAPE '\') AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm_demo1_rel_many_rel.title]] as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN pb_json_each([[__mm_demo1.rel_many]]) "__mm_demo1_rel_many_je" ON TRUE LEFT JOIN "users" "__mm_demo1_rel_many" ON [[__mm_demo1_rel_many.id]] = [[__mm_demo1_rel_many_je.value]] LEFT JOIN "demo2" "__mm_demo1_rel_many_rel" ON [[__mm_demo1_rel_many_rel.id]] = [[__mm_demo1_rel_many.rel]] WHERE "__mm_demo1"."id" = "demo1"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]]::text LIKE ('%' || [[demo1_rel_one.email]] || '%') ESCAPE '\')))) AND ((COALESCE([[__collection_demo2.active]], '') = COALESCE([[demo1_rel_many_rel.active]], '')) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm__collection_demo2.active]] as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN "demo2" "__mm__collection_demo2" WHERE "__mm_demo1"."id" = "demo1"."id") {{__mlTEST}} LEFT JOIN (SELECT [[__mm_demo1_rel_many_rel.active]] as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN pb_json_each([[__mm_demo1.rel_many]]) "__mm_demo1_rel_many_je" ON TRUE LEFT JOIN "users" "__mm_demo1_rel_many" ON [[__mm_demo1_rel_many.id]] = [[__mm_demo1_rel_many_je.value]] LEFT JOIN "demo2" "__mm_demo1_rel_many_rel" ON [[__mm_demo1_rel_many_rel.id]] = [[__mm_demo1_rel_many.rel]] WHERE "__mm_demo1"."id" = "demo1"."id") {{__mrTEST}} WHERE NOT (COALESCE([[__mlTEST.multiMatchValue]], '') = COALESCE([[__mrTEST.multiMatchValue]], ''))))) AND COALESCE([[__collection_demo2.active]], '') = COALESCE([[demo1_rel_many_rel.active]], '') AND (((([[demo1_rel_many.email]] > [[__data_users_rel_many.email]]) AND (NOT EXISTS (SELECT 1 FROM (SELECT [[__mm_demo1_rel_many.email]] as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN pb_json_each([[__mm_demo1.rel_many]]) "__mm_demo1_rel_many_je" ON TRUE LEFT JOIN "users" "__mm_demo1_rel_many" ON [[__mm_demo1_rel_many.id]] = [[__mm_demo1_rel_many_je.value]] WHERE "__mm_demo1"."id" = "demo1"."id") {{__mlTEST}} LEFT JOIN (SELECT [[__data_mm_users_rel_many.email]] as [[multiMatchValue]] FROM "demo1" "__mm_demo1" LEFT JOIN "users" "__data_mm_users_rel_many" ON [[__data_mm_users_rel_many.id]] IN ({:TEST}, {:TEST}) WHERE "__mm_demo1"."id" = "demo1"."id") {{__mrTEST}} WHERE NOT ([[__mlTEST.multiMatchValue]] > [[__mrTEST.multiMatchValue]]))))) AND ([[demo1_rel_many.emailVisibility]] = TRUE)))`,
		},
		{
			"@request.body.arrayable:length fields",
			"demo1",
			"@request.body.select_one:length > 1 &&@request.body.select_one:length ?> 2 &&@request.body.select_many:length < 3 &&@request.body.select_many:length ?> 4 &&@request.body.rel_one:length = 5 &&@request.body.rel_one:length ?= 6 &&@request.body.rel_many:length != 7 &&@request.body.rel_many:length ?!= 8 &&@request.body.file_one:length = 9 &&@request.body.file_one:length ?= 0 &&@request.body.file_many:length != 1 &&@request.body.file_many:length ?!= 2",
			false,
			`SELECT "demo1".* FROM "demo1" WHERE (0 > {:TEST} AND 0 > {:TEST} AND 2 < {:TEST} AND 2 > {:TEST} AND 1 = {:TEST} AND 1 = {:TEST} AND 2 IS DISTINCT FROM {:TEST} AND 2 IS DISTINCT FROM {:TEST} AND 1 = {:TEST} AND 1 = {:TEST} AND 3 IS DISTINCT FROM {:TEST} AND 3 IS DISTINCT FROM {:TEST})`,
		},
		{
			"regular arrayable:length fields",
			"demo4",
			"@request.body.self_rel_one.self_rel_many:length > 1 &&@request.body.self_rel_one.self_rel_many:length ?> 2 &&@request.body.rel_many_cascade.files:length ?< 3 &&@request.body.rel_many_cascade.files:length < 4 &&@request.body.rel_one_cascade.files:length < 4.1 &&self_rel_one.self_rel_many:length = 5 &&self_rel_one.self_rel_many:length ?= 6 &&self_rel_one.rel_many_cascade.files:length != 7 &&self_rel_one.rel_many_cascade.files:length ?!= 8",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN "demo4" "__data_demo4_self_rel_one" ON [[__data_demo4_self_rel_one.id]]={:TEST} LEFT JOIN "demo3" "__data_demo3_rel_many_cascade" ON [[__data_demo3_rel_many_cascade.id]] IN ({:TEST}, {:TEST}) LEFT JOIN "demo3" "__data_demo3_rel_one_cascade" ON [[__data_demo3_rel_one_cascade.id]]={:TEST} LEFT JOIN "demo4" "demo4_self_rel_one" ON [[demo4_self_rel_one.id]] = [[demo4.self_rel_one]] LEFT JOIN pb_json_each([[demo4_self_rel_one.rel_many_cascade]]) "demo4_self_rel_one_rel_many_cascade_je" ON TRUE LEFT JOIN "demo3" "demo4_self_rel_one_rel_many_cascade" ON [[demo4_self_rel_one_rel_many_cascade.id]] = [[demo4_self_rel_one_rel_many_cascade_je.value]] WHERE (pb_json_array_length([[__data_demo4_self_rel_one.self_rel_many]]) > {:TEST} AND pb_json_array_length([[__data_demo4_self_rel_one.self_rel_many]]) > {:TEST} AND pb_json_array_length([[__data_demo3_rel_many_cascade.files]]) < {:TEST} AND ((pb_json_array_length([[__data_demo3_rel_many_cascade.files]]) < {:TEST}) AND (NOT EXISTS (SELECT 1 FROM (SELECT pb_json_array_length([[__data_mm_demo3_rel_many_cascade.files]]) as [[multiMatchValue]] FROM "demo4" "__mm_demo4" LEFT JOIN "demo3" "__data_mm_demo3_rel_many_cascade" ON [[__data_mm_demo3_rel_many_cascade.id]] IN ({:TEST}, {:TEST}) WHERE "__mm_demo4"."id" = "demo4"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] < {:TEST})))) AND pb_json_array_length([[__data_demo3_rel_one_cascade.files]]) < {:TEST} AND pb_json_array_length([[demo4_self_rel_one.self_rel_many]]) = {:TEST} AND pb_json_array_length([[demo4_self_rel_one.self_rel_many]]) = {:TEST} AND ((pb_json_array_length([[demo4_self_rel_one_rel_many_cascade.files]]) IS DISTINCT FROM {:TEST}) AND (NOT EXISTS (SELECT 1 FROM (SELECT pb_json_array_length([[__mm_demo4_self_rel_one_rel_many_cascade.files]]) as [[multiMatchValue]] FROM "demo4" "__mm_demo4" LEFT JOIN "demo4" "__mm_demo4_self_rel_one" ON [[__mm_demo4_self_rel_one.id]] = [[__mm_demo4.self_rel_one]] LEFT JOIN pb_json_each([[__mm_demo4_self_rel_one.rel_many_cascade]]) "__mm_demo4_self_rel_one_rel_many_cascade_je" ON TRUE LEFT JOIN "demo3" "__mm_demo4_self_rel_one_rel_many_cascade" ON [[__mm_demo4_self_rel_one_rel_many_cascade.id]] = [[__mm_demo4_self_rel_one_rel_many_cascade_je.value]] WHERE "__mm_demo4"."id" = "demo4"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] IS DISTINCT FROM {:TEST})))) AND pb_json_array_length([[demo4_self_rel_one_rel_many_cascade.files]]) IS DISTINCT FROM {:TEST})`,
		},
		{
			"json_extract and json_array_length COALESCE equal normalizations",
			"demo4",
			"json_object.a.b = '' && self_rel_many:length != 2 && json_object.a.b > 3 && self_rel_many:length <= 4",
			false,
			`SELECT "demo4".* FROM "demo4" WHERE (pb_json_extract([[demo4.json_object]], '$.a.b') IS NOT DISTINCT FROM {:TEST} AND pb_json_array_length([[demo4.self_rel_many]]) IS DISTINCT FROM {:TEST} AND pb_json_extract([[demo4.json_object]], '$.a.b') > {:TEST} AND pb_json_array_length([[demo4.self_rel_many]]) <= {:TEST})`,
		},
		{
			"json field equal normalization checks",
			"demo4",
			"json_object = '' || json_object != '' || '' = json_object || '' != json_object ||json_object = null || json_object != null || null = json_object || null != json_object ||json_object = true || json_object != true || true = json_object || true != json_object ||json_object = json_object || json_object != json_object ||json_object = title || title != json_object ||self_rel_many.json_object = '' || null = self_rel_many.json_object ||self_rel_many.json_object = self_rel_many.json_object",
			false,
			`SELECT DISTINCT "demo4".* FROM "demo4" LEFT JOIN pb_json_each([[demo4.self_rel_many]]) "demo4_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "demo4_self_rel_many" ON [[demo4_self_rel_many.id]] = [[demo4_self_rel_many_je.value]] WHERE (pb_json_extract([[demo4.json_object]], '$') IS NOT DISTINCT FROM {:TEST} OR pb_json_extract([[demo4.json_object]], '$') IS DISTINCT FROM {:TEST} OR {:TEST} IS NOT DISTINCT FROM pb_json_extract([[demo4.json_object]], '$') OR {:TEST} IS DISTINCT FROM pb_json_extract([[demo4.json_object]], '$') OR pb_json_extract([[demo4.json_object]], '$') IS NOT DISTINCT FROM NULL OR pb_json_extract([[demo4.json_object]], '$') IS DISTINCT FROM NULL OR NULL IS NOT DISTINCT FROM pb_json_extract([[demo4.json_object]], '$') OR NULL IS DISTINCT FROM pb_json_extract([[demo4.json_object]], '$') OR pb_json_extract([[demo4.json_object]], '$') IS NOT DISTINCT FROM true OR pb_json_extract([[demo4.json_object]], '$') IS DISTINCT FROM true OR true IS NOT DISTINCT FROM pb_json_extract([[demo4.json_object]], '$') OR true IS DISTINCT FROM pb_json_extract([[demo4.json_object]], '$') OR pb_json_extract([[demo4.json_object]], '$') IS NOT DISTINCT FROM pb_json_extract([[demo4.json_object]], '$') OR pb_json_extract([[demo4.json_object]], '$') IS DISTINCT FROM pb_json_extract([[demo4.json_object]], '$') OR pb_json_extract([[demo4.json_object]], '$') IS NOT DISTINCT FROM [[demo4.title]] OR [[demo4.title]] IS DISTINCT FROM pb_json_extract([[demo4.json_object]], '$') OR ((pb_json_extract([[demo4_self_rel_many.json_object]], '$') IS NOT DISTINCT FROM {:TEST}) AND (NOT EXISTS (SELECT 1 FROM (SELECT pb_json_extract([[__mm_demo4_self_rel_many.json_object]], '$') as [[multiMatchValue]] FROM "demo4" "__mm_demo4" LEFT JOIN pb_json_each([[__mm_demo4.self_rel_many]]) "__mm_demo4_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "__mm_demo4_self_rel_many" ON [[__mm_demo4_self_rel_many.id]] = [[__mm_demo4_self_rel_many_je.value]] WHERE "__mm_demo4"."id" = "demo4"."id") {{__smTEST}} WHERE NOT ([[__smTEST.multiMatchValue]] IS NOT DISTINCT FROM {:TEST})))) OR ((NULL IS NOT DISTINCT FROM pb_json_extract([[demo4_self_rel_many.json_object]], '$')) AND (NOT EXISTS (SELECT 1 FROM (SELECT pb_json_extract([[__mm_demo4_self_rel_many.json_object]], '$') as [[multiMatchValue]] FROM "demo4" "__mm_demo4" LEFT JOIN pb_json_each([[__mm_demo4.self_rel_many]]) "__mm_demo4_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "__mm_demo4_self_rel_many" ON [[__mm_demo4_self_rel_many.id]] = [[__mm_demo4_self_rel_many_je.value]] WHERE "__mm_demo4"."id" = "demo4"."id") {{__smTEST}} WHERE NOT (NULL IS NOT DISTINCT FROM [[__smTEST.multiMatchValue]])))) OR ((pb_json_extract([[demo4_self_rel_many.json_object]], '$') IS NOT DISTINCT FROM pb_json_extract([[demo4_self_rel_many.json_object]], '$')) AND (NOT EXISTS (SELECT 1 FROM (SELECT pb_json_extract([[__mm_demo4_self_rel_many.json_object]], '$') as [[multiMatchValue]] FROM "demo4" "__mm_demo4" LEFT JOIN pb_json_each([[__mm_demo4.self_rel_many]]) "__mm_demo4_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "__mm_demo4_self_rel_many" ON [[__mm_demo4_self_rel_many.id]] = [[__mm_demo4_self_rel_many_je.value]] WHERE "__mm_demo4"."id" = "demo4"."id") {{__mlTEST}} LEFT JOIN (SELECT pb_json_extract([[__mm_demo4_self_rel_many.json_object]], '$') as [[multiMatchValue]] FROM "demo4" "__mm_demo4" LEFT JOIN pb_json_each([[__mm_demo4.self_rel_many]]) "__mm_demo4_self_rel_many_je" ON TRUE LEFT JOIN "demo4" "__mm_demo4_self_rel_many" ON [[__mm_demo4_self_rel_many.id]] = [[__mm_demo4_self_rel_many_je.value]] WHERE "__mm_demo4"."id" = "demo4"."id") {{__mrTEST}} WHERE NOT ([[__mlTEST.multiMatchValue]] IS NOT DISTINCT FROM [[__mrTEST.multiMatchValue]])))))`,
		},
		{
			"geoPoint props access",
			"demo1",
			"point = '' || point.lat > 1 || point.lon < 2 || point.something > 3",
			false,
			`SELECT "demo1".* FROM "demo1" WHERE (([[demo1.point]] = '' OR [[demo1.point]] IS NULL) OR pb_json_extract([[demo1.point]], '$.lat') > {:TEST} OR pb_json_extract([[demo1.point]], '$.lon') < {:TEST} OR pb_json_extract([[demo1.point]], '$.something') > {:TEST})`,
		},
	}

	// var output strings.Builder
	// output.WriteString("[\n")

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			collection, err := app.FindCollectionByNameOrId(s.collectionIdOrName)
			if err != nil {
				t.Fatalf("[%s] Failed to load collection %s: %v", s.name, s.collectionIdOrName, err)
			}

			query := app.RecordQuery(collection)

			r := core.NewRecordFieldResolver(app, collection, requestInfo, s.allowHiddenFields)

			expr, err := search.FilterData(s.rule).BuildExpr(r)
			if err != nil {
				t.Fatalf("[%s] BuildExpr failed with error %v", s.name, err)
			}

			if err := r.UpdateQuery(query); err != nil {
				t.Fatalf("[%s] UpdateQuery failed with error %v", s.name, err)
			}

			rawQuery := query.AndWhere(expr).Build().SQL()

			// rawQuery = regexp.MustCompile(`__ml\w+`).ReplaceAllString(rawQuery, "__mlTEST")
			// rawQuery = regexp.MustCompile(`__mr\w+`).ReplaceAllString(rawQuery, "__mrTEST")
			// rawQuery = regexp.MustCompile(`__sm\w+`).ReplaceAllString(rawQuery, "__smTEST")
			// rawQuery = regexp.MustCompile(`\{\:\w+\}`).ReplaceAllString(rawQuery, "{:TEST}")

			// output.WriteString("{\n")
			// output.WriteString("\"" + s.name + "\",\n")
			// output.WriteString("\"" + s.collectionIdOrName + "\",\n")
			// output.WriteString("\"" + s.rule + "\",\n")
			// output.WriteString(fmt.Sprintf("%t", s.allowHiddenFields) + ",\n")
			// output.WriteString("`" + rawQuery + "`,\n")
			// output.WriteString("},\n")

			// replace TEST placeholder with .+ regex pattern
			expectQuery := strings.ReplaceAll(
				"^"+regexp.QuoteMeta(s.expectQuery)+"$",
				"TEST",
				`\w+`,
			)

			if !list.ExistInSliceWithRegex(rawQuery, []string{expectQuery}) {
				t.Fatalf("[%s] Expected query\n %v \ngot:\n %v", s.name, expectQuery, rawQuery)
			}
		})
	}
	// output.WriteString("]\n")
	// // write output to file txt
	// os.WriteFile("./test.txt", []byte(output.String()), 0644)
}

func TestRecordFieldResolverResolveCollectionFields(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	collection, err := app.FindCollectionByNameOrId("demo4")
	if err != nil {
		t.Fatal(err)
	}

	authRecord, err := app.FindRecordById("users", "4q1xlclmfloku33")
	if err != nil {
		t.Fatal(err)
	}

	requestInfo := &core.RequestInfo{
		Auth: authRecord,
	}

	r := core.NewRecordFieldResolver(app, collection, requestInfo, true)

	scenarios := []struct {
		fieldName   string
		expectError bool
		expectName  string
	}{
		{"", true, ""},
		{" ", true, ""},
		{"unknown", true, ""},
		{"invalid format", true, ""},
		{"id", false, "[[demo4.id]]"},
		{"created", false, "[[demo4.created]]"},
		{"updated", false, "[[demo4.updated]]"},
		{"title", false, "[[demo4.title]]"},
		{"title.test", true, ""},
		{"self_rel_many", false, "[[demo4.self_rel_many]]"},
		{"self_rel_many.", true, ""},
		{"self_rel_many.unknown", true, ""},
		{"self_rel_many.title", false, "[[demo4_self_rel_many.title]]"},
		{"self_rel_many.self_rel_one.self_rel_many.title", false, "[[demo4_self_rel_many_self_rel_one_self_rel_many.title]]"},

		// max relations limit
		{"self_rel_many.self_rel_many.self_rel_many.self_rel_many.self_rel_many.self_rel_many.id", false, "[[demo4_self_rel_many_self_rel_many_self_rel_many_self_rel_many_self_rel_many_self_rel_many.id]]"},
		{"self_rel_many.self_rel_many.self_rel_many.self_rel_many.self_rel_many.self_rel_many.self_rel_many.id", true, ""},

		// back relations
		{"rel_one_cascade.demo4_via_title.id", true, ""},        // not a relation field
		{"rel_one_cascade.demo4_via_self_rel_one.id", true, ""}, // relation field but to a different collection
		{"rel_one_cascade.demo4_via_rel_one_cascade.id", false, "[[demo4_rel_one_cascade_demo4_via_rel_one_cascade.id]]"},
		{"rel_one_cascade.demo4_via_rel_one_cascade.rel_one_cascade.demo4_via_rel_one_cascade.id", false, "[[demo4_rel_one_cascade_demo4_via_rel_one_cascade_rel_one_cascade_demo4_via_rel_one_cascade.id]]"},

		// json_extract
		{"json_array.0", false, `pb_json_extract([[demo4.json_array]], '$[0]')`},
		{"json_object.a.b.c", false, `pb_json_extract([[demo4.json_object]], '$.a.b.c')`},

		// max relations limit shouldn't apply for json paths
		{"json_object.a.b.c.e.f.g.h.i.j.k.l.m.n.o.p", false, `pb_json_extract([[demo4.json_object]], '$.a.b.c.e.f.g.h.i.j.k.l.m.n.o.p')`},

		// @request.auth relation join
		{"@request.auth.rel", false, "[[__auth_users.rel]]"},
		{"@request.auth.rel.title", false, "[[__auth_users_rel.title]]"},
		{"@request.auth.demo1_via_rel_many.id", false, "[[__auth_users_demo1_via_rel_many.id]]"},
		{"@request.auth.rel.missing", false, "NULL"},
		{"@request.auth.missing_via_rel", false, "NULL"},
		{"@request.auth.demo1_via_file_one.id", false, "NULL"}, // not a relation field
		{"@request.auth.demo1_via_rel_one.id", false, "NULL"},  // relation field but to a different collection

		// @collection fieds
		{"@collect", true, ""},
		{"collection.demo4.title", true, ""},
		{"@collection", true, ""},
		{"@collection.unknown", true, ""},
		{"@collection.demo2", true, ""},
		{"@collection.demo2.", true, ""},
		{"@collection.demo2:someAlias", true, ""},
		{"@collection.demo2:someAlias.", true, ""},
		{"@collection.demo2.title", false, "[[__collection_demo2.title]]"},
		{"@collection.demo2:someAlias.title", false, "[[__collection_alias_someAlias.title]]"},
		{"@collection.demo4.id", false, "[[__collection_demo4.id]]"},
		{"@collection.demo4.created", false, "[[__collection_demo4.created]]"},
		{"@collection.demo4.updated", false, "[[__collection_demo4.updated]]"},
		{"@collection.demo4.self_rel_many.missing", true, ""},
		{"@collection.demo4.self_rel_many.self_rel_one.self_rel_many.self_rel_one.title", false, "[[__collection_demo4_self_rel_many_self_rel_one_self_rel_many_self_rel_one.title]]"},
	}

	for _, s := range scenarios {
		t.Run(s.fieldName, func(t *testing.T) {
			r, err := r.Resolve(s.fieldName)

			hasErr := err != nil
			if hasErr != s.expectError {
				t.Fatalf("Expected hasErr %v, got %v (%v)", s.expectError, hasErr, err)
			}

			if hasErr {
				return
			}

			if r.Identifier != s.expectName {
				t.Fatalf("Expected r.Identifier\n%q\ngot\n%q", s.expectName, r.Identifier)
			}

			// params should be empty for non @request fields
			if len(r.Params) != 0 {
				t.Fatalf("Expected 0 r.Params, got\n%v", r.Params)
			}
		})
	}
}

func TestRecordFieldResolverResolveStaticRequestInfoFields(t *testing.T) {
	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	collection, err := app.FindCollectionByNameOrId("demo1")
	if err != nil {
		t.Fatal(err)
	}

	authRecord, err := app.FindRecordById("users", "4q1xlclmfloku33")
	if err != nil {
		t.Fatal(err)
	}

	requestInfo := &core.RequestInfo{
		Context: "ctx",
		Method:  "get",
		Query: map[string]string{
			"a": "123",
		},
		Body: map[string]any{
			"number":          "10",
			"number_unknown":  "20",
			"raw_json_obj":    types.JSONRaw(`{"a":123}`),
			"raw_json_arr1":   types.JSONRaw(`[123, 456]`),
			"raw_json_arr2":   types.JSONRaw(`[{"a":123},{"b":456}]`),
			"raw_json_simple": types.JSONRaw(`123`),
			"b":               456,
			"c":               map[string]int{"sub": 1},
		},
		Headers: map[string]string{
			"d": "789",
		},
		Auth: authRecord,
	}

	r := core.NewRecordFieldResolver(app, collection, requestInfo, true)

	scenarios := []struct {
		fieldName        string
		expectError      bool
		expectParamValue string // encoded json
	}{
		{"@request", true, ""},
		{"@request.invalid format", true, ""},
		{"@request.invalid_format2!", true, ""},
		{"@request.missing", true, ""},
		{"@request.context", false, `"ctx"`},
		{"@request.method", false, `"get"`},
		{"@request.query", true, ``},
		{"@request.query.a", false, `"123"`},
		{"@request.query.a.missing", false, ``},
		{"@request.headers", true, ``},
		{"@request.headers.missing", false, ``},
		{"@request.headers.d", false, `"789"`},
		{"@request.headers.d.sub", false, ``},
		{"@request.body", true, ``},
		{"@request.body.b", false, `456`},
		{"@request.body.number", false, `10`},           // number field normalization
		{"@request.body.number_unknown", false, `"20"`}, // no numeric normalizations for unknown fields
		{"@request.body.b.missing", false, ``},
		{"@request.body.c", false, `"{\"sub\":1}"`},
		{"@request.auth", true, ""},
		{"@request.auth.id", false, `"4q1xlclmfloku33"`},
		{"@request.auth.collectionId", false, `"` + authRecord.Collection().Id + `"`},
		{"@request.auth.collectionName", false, `"` + authRecord.Collection().Name + `"`},
		{"@request.auth.verified", false, `false`},
		{"@request.auth.emailVisibility", false, `false`},
		{"@request.auth.email", false, `"test@example.com"`}, // should always be returned no matter of the emailVisibility state
		{"@request.auth.missing", false, `NULL`},
		{"@request.body.raw_json_simple", false, `"123"`},
		{"@request.body.raw_json_simple.a", false, `NULL`},
		{"@request.body.raw_json_obj.a", false, `123`},
		{"@request.body.raw_json_obj.b", false, `NULL`},
		{"@request.body.raw_json_arr1.1", false, `456`},
		{"@request.body.raw_json_arr1.3", false, `NULL`},
		{"@request.body.raw_json_arr2.0.a", false, `123`},
		{"@request.body.raw_json_arr2.0.b", false, `NULL`},
	}

	for _, s := range scenarios {
		t.Run(s.fieldName, func(t *testing.T) {
			r, err := r.Resolve(s.fieldName)

			hasErr := err != nil
			if hasErr != s.expectError {
				t.Fatalf("Expected hasErr %v, got %v (%v)", s.expectError, hasErr, err)
			}

			if hasErr {
				return
			}

			// missing key
			// ---
			if len(r.Params) == 0 {
				if r.Identifier != "NULL" {
					t.Fatalf("Expected 0 placeholder parameters for %v, got %v", r.Identifier, r.Params)
				}
				return
			}

			// existing key
			// ---
			if len(r.Params) != 1 {
				t.Fatalf("Expected 1 placeholder parameter for %v, got %v", r.Identifier, r.Params)
			}

			var paramName string
			var paramValue any
			for k, v := range r.Params {
				paramName = k
				paramValue = v
			}

			if r.Identifier != ("{:" + paramName + "}") {
				t.Fatalf("Expected parameter r.Identifier %q, got %q", paramName, r.Identifier)
			}

			encodedParamValue, _ := json.Marshal(paramValue)
			if string(encodedParamValue) != s.expectParamValue {
				t.Fatalf("Expected r.Params %#v for %s, got %#v", s.expectParamValue, r.Identifier, string(encodedParamValue))
			}
		})
	}

	// ensure that the original email visibility was restored
	if authRecord.EmailVisibility() {
		t.Fatal("Expected the original authRecord emailVisibility to remain unchanged")
	}
	if v, ok := authRecord.PublicExport()[core.FieldNameEmail]; ok {
		t.Fatalf("Expected the original authRecord email to not be exported, got %q", v)
	}
}
