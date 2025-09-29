package migrations

import (
	"fmt"

	"github.com/thewandererbg/pgbase/core"
)

func init() {
	core.SystemMigrations.Register(func(txApp core.App) error {
		if err := registerPostgresFunctions(txApp); err != nil {
			return fmt.Errorf("register postgres functions error: %w", err)
		}
		return nil
	}, func(txApp core.App) error {
		if err := unregisterPostgresFunctions(txApp); err != nil {
			return fmt.Errorf("unregister postgres functions error: %w", err)
		}
		return nil
	})
}

func registerPostgresFunctions(txApp core.App) error {
	_, execErr := txApp.DB().NewQuery(`
		CREATE COLLATION IF NOT EXISTS nocase (provider = icu, locale = 'und-u-ks-level1', deterministic = false);

		CREATE OR REPLACE FUNCTION public.pb_json_each(input_data jsonb)
		    RETURNS TABLE(value text)
		    IMMUTABLE
		    LANGUAGE plpgsql
		AS $$
		DECLARE
		    json_type text;
		BEGIN
		    IF input_data IS NULL THEN
		        RETURN;
		    END IF;

		    json_type := jsonb_typeof(input_data);

		    IF json_type = 'array' THEN
		        RETURN QUERY SELECT jsonb_array_elements_text(input_data);
		    ELSIF json_type = 'object' THEN
		        RETURN QUERY SELECT val FROM jsonb_each_text(input_data) AS t(key, val);
		    ELSE
		        RETURN QUERY SELECT trim(both '"' from input_data::text);
		    END IF;
		END;
		$$;

		CREATE OR REPLACE FUNCTION public.pb_json_array_length(input_data jsonb)
		    RETURNS integer
		    IMMUTABLE
		    LANGUAGE sql
		AS $$
		    SELECT CASE
		        WHEN input_data IS NULL THEN 0
		        WHEN jsonb_typeof(input_data) = 'array' THEN jsonb_array_length(input_data)
		        ELSE 0
		    END;
		$$;

		CREATE OR REPLACE FUNCTION public.pb_json_extract(data jsonb, path text)
		    RETURNS text
		    IMMUTABLE
		    LANGUAGE plpgsql
		AS $$
		BEGIN
		    IF data IS NULL OR path IS NULL THEN
		        RETURN NULL;
		    END IF;

		    BEGIN
		        RETURN jsonb_path_query_first(data, path::jsonpath) #>> '{}';
		    EXCEPTION WHEN others THEN
		        RETURN NULL;
		    END;
		END;
		$$;

		CREATE OR REPLACE FUNCTION public.pb_json_each(input_data anyelement)
		    RETURNS TABLE(value text)
		    IMMUTABLE
		    LANGUAGE plpgsql
		AS $$
		BEGIN
		    IF input_data IS NULL THEN
		        RETURN;
		    END IF;

		    IF pg_typeof(input_data) = 'jsonb'::regtype THEN
		        RETURN QUERY SELECT * FROM pb_json_each(input_data::jsonb);
		    ELSIF pg_typeof(input_data) = 'json'::regtype THEN
		        RETURN QUERY SELECT * FROM pb_json_each(input_data::jsonb);
		    ELSE
		        BEGIN
		            RETURN QUERY SELECT * FROM pb_json_each(input_data::text::jsonb);
		        EXCEPTION WHEN others THEN
		            RETURN QUERY SELECT input_data::text;
		        END;
		    END IF;
		END;
		$$;

		CREATE OR REPLACE FUNCTION public.pb_json_array_length(input_data anyelement)
		    RETURNS integer
		    IMMUTABLE
		    LANGUAGE plpgsql
		AS $$
		BEGIN
		    IF input_data IS NULL OR input_data::text = '' THEN
		        RETURN 0;
		    END IF;

		    IF pg_typeof(input_data) = 'jsonb'::regtype THEN
		        RETURN pb_json_array_length(input_data::jsonb);
		    ELSIF pg_typeof(input_data) = 'json'::regtype THEN
		        RETURN pb_json_array_length(input_data::jsonb);
		    ELSE
		        BEGIN
		            RETURN pb_json_array_length(input_data::text::jsonb);
		        EXCEPTION WHEN others THEN
		            RETURN 0;
		        END;
		    END IF;
		END;
		$$;

		CREATE OR REPLACE FUNCTION public.pb_json_extract(data anyelement, path text)
		    RETURNS text
		    IMMUTABLE
		    LANGUAGE plpgsql
		AS $$
		BEGIN
		    IF data IS NULL OR path IS NULL THEN
		        RETURN NULL;
		    END IF;

		    IF pg_typeof(data) = 'jsonb'::regtype THEN
		        RETURN pb_json_extract(data::jsonb, path);
		    ELSIF pg_typeof(data) = 'json'::regtype THEN
		        RETURN pb_json_extract(data::jsonb, path);
		    ELSE
		        BEGIN
		            RETURN pb_json_extract(data::text::jsonb, path);
		        EXCEPTION WHEN others THEN
		            RETURN data::text;
		        END;
		    END IF;
		END;
		$$;
	`).Execute()

	return execErr
}

func unregisterPostgresFunctions(txApp core.App) error {
	_, execErr := txApp.DB().NewQuery(`
		drop function if exists public.pb_json_extract(anyelement, text);
		drop function if exists public.pb_json_array_length(anyelement, text);
		drop function if exists public.pb_json_each(anyelement, text);
	`).Execute()

	return execErr
}
