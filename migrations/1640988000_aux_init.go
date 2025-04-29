package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

func init() {
	core.SystemMigrations.Add(&core.Migration{
		Up: func(txApp core.App) error {
			_, execErr := txApp.AuxDB().NewQuery(`
				CREATE TABLE IF NOT EXISTS {{_logs}} (
					[[id]]      TEXT PRIMARY KEY DEFAULT length(substr(md5(random()::text), 1, 15)) NOT NULL,
					[[level]]   INT DEFAULT 0 NOT NULL,
					[[message]] TEXT DEFAULT '' NOT NULL,
					[[data]]    JSONB DEFAULT '{}' NOT NULL,
					[[created]] TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
				);

				CREATE INDEX IF NOT EXISTS idx_logs_level ON {{_logs}} ([[level]]);
				CREATE INDEX IF NOT EXISTS idx_logs_message ON {{_logs}} ([[message]]);
				CREATE INDEX IF NOT EXISTS idx_logs_created ON {{_logs}} ([[created]]);

				CREATE COLLATION IF NOT EXISTS nocase (provider = icu, locale = 'und-u-ks-level1', deterministic = false);

				create or replace function public.pb_json_each(input_data anyelement)
				    returns TABLE(value text)
				    immutable
				    language plpgsql
				as
				$$
				DECLARE
				    json_data jsonb;
				BEGIN
				    IF input_data IS NULL THEN
				        RETURN;
				    END IF;

				    IF pg_typeof(input_data) IN ('json'::regtype, 'jsonb'::regtype) THEN
				        json_data := input_data::jsonb;
				    ELSE
				        BEGIN
				            json_data := input_data::text::jsonb;
				        EXCEPTION WHEN others THEN
				            RETURN QUERY SELECT input_data::text;
				            RETURN;
				        END;
				    END IF;

				    CASE jsonb_typeof(json_data)
				        WHEN 'array' THEN
				            RETURN QUERY
				            SELECT elem
				            FROM jsonb_array_elements_text(json_data) AS t(elem);

				        WHEN 'object' THEN
				            RETURN QUERY
				            SELECT val
				            FROM jsonb_each_text(json_data) AS t(key, val);

				        ELSE
				            RETURN QUERY
				            SELECT trim(both '"' from json_data::text);
				    END CASE;
				END;
				$$;

				comment on function public.pb_json_each(anyelement) is 'Extracts values from JSON data and returns them as rows. Handles JSON arrays, objects, primitive values, and non-JSON inputs. Returns a table with a single text column "value".';

				create or replace function public.pb_json_array_length(input_data anyelement) returns integer
				    immutable
				    language plpgsql
				as
				$$
				DECLARE
				    input_jsonb jsonb;
				BEGIN
				    -- NULL or empty input → 0
				    IF input_data IS NULL THEN
				        RETURN 0;
				    END IF;

				    -- Fast-path for jsonb and json types
				    IF pg_typeof(input_data) = 'jsonb'::regtype THEN
				        IF jsonb_typeof(input_data::jsonb) = 'array' THEN
				            RETURN jsonb_array_length(input_data::jsonb);
				        ELSE
				            RETURN 0;
				        END IF;

				    ELSIF pg_typeof(input_data) = 'json'::regtype THEN
				        IF json_typeof(input_data::json) = 'array' THEN
				            RETURN json_array_length(input_data::json);
				        ELSE
				            RETURN 0;
				        END IF;
				    END IF;

				    -- Other types → try to parse as JSON
				    BEGIN
				        -- Avoid empty string parsing
				        IF input_data::text IS NULL OR input_data::text = '' THEN
				            RETURN 0;
				        END IF;

				        input_jsonb := input_data::text::jsonb;

				        IF jsonb_typeof(input_jsonb) = 'array' THEN
				            RETURN jsonb_array_length(input_jsonb);
				        ELSE
				            RETURN 0;
				        END IF;

				    EXCEPTION WHEN others THEN
				        -- If not valid JSON → 0
				        RETURN 0;
				    END;
				END;
				$$;

				comment on function public.pb_json_array_length(anyelement) is 'Returns the length of a JSON array. Returns 0 for non-arrays, NULL values, empty strings, or invalid JSON.
								Handles various input formats including native JSON/JSONB types and JSON string literals.';

				create or replace function public.pb_json_extract(data anyelement, path text) returns text
				    immutable
				    language plpgsql
				as
				$$
				DECLARE
				    input_jsonb jsonb;
				    json_path jsonpath;
				    result text;
				BEGIN
				    IF data IS NULL OR path IS NULL THEN
				        RETURN NULL;
				    END IF;

				    -- Try pre-cast JSON path
				    BEGIN
				        json_path := path::jsonpath;
				    EXCEPTION
				        WHEN others THEN
				            RETURN NULL; -- Invalid path
				    END;

				    -- Detect type only once
				    IF pg_typeof(data) IN ('jsonb'::regtype, 'json'::regtype) THEN
				        input_jsonb := data::jsonb;
				    ELSE
				        BEGIN
				            input_jsonb := data::text::jsonb;
				        EXCEPTION
				            WHEN others THEN
				                RETURN data::text; -- Not JSON, just return as text
				        END;
				    END IF;

				    -- Safe JSON query
				    BEGIN
				        result := jsonb_path_query_first(input_jsonb, json_path) #>> '{}';
				        RETURN result;
				    EXCEPTION
				        WHEN others THEN
				            RETURN NULL; -- In case JSON path query fails
				    END;
				END;
				$$;

				comment on function public.pb_json_extract(anyelement, text) is 'Extracts values from JSON data as text using a path in the format $.x.y (similar to SQLite''s JSON_EXTRACT).
								Returns NULL for NULL inputs or when the path doesn''t exist.
								For non-JSON inputs, attempts to convert to JSON or returns the input as text.';
			`).Execute()

			if execErr != nil {
				return fmt.Errorf("_mfas error: %w", execErr)
			}

			return nil
		},
		Down: func(txApp core.App) error {
			_, err := txApp.AuxDB().DropTable("_logs").Execute()
			return err
		},
		ReapplyCondition: func(txApp core.App, runner *core.MigrationsRunner, fileName string) (bool, error) {
			// reapply only if the _logs table doesn't exist
			exists := txApp.AuxHasTable("_logs")
			return !exists, nil
		},
	})
}
