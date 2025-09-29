--
-- PostgreSQL database dump
--

\restrict DF9FExrGxX28ulqgCixzacrQ7cMeRgfhhncVVTe3nMnYDCQnGraFW582JjwY1e5

-- Dumped from database version 17.6 (Debian 17.6-1.pgdg13+1)
-- Dumped by pg_dump version 17.6 (Debian 17.6-1.pgdg13+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: nocase; Type: COLLATION; Schema: public; Owner: postgres
--

CREATE COLLATION public.nocase (provider = icu, deterministic = false, locale = 'en-US-u-ks-level2');


ALTER COLLATION public.nocase OWNER TO postgres;

--
-- Name: pb_json_array_length(anyelement); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.pb_json_array_length(input_data anyelement) RETURNS integer
    LANGUAGE plpgsql IMMUTABLE
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


ALTER FUNCTION public.pb_json_array_length(input_data anyelement) OWNER TO postgres;

--
-- Name: FUNCTION pb_json_array_length(input_data anyelement); Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON FUNCTION public.pb_json_array_length(input_data anyelement) IS 'Returns the length of a JSON array. Returns 0 for non-arrays, NULL values, empty strings, or invalid JSON.

Handles various input formats including native JSON/JSONB types and JSON string literals.';


--
-- Name: pb_json_array_length(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.pb_json_array_length(input_data jsonb) RETURNS integer
    LANGUAGE sql IMMUTABLE
    AS $$
    SELECT CASE
        WHEN input_data IS NULL THEN 0
        WHEN jsonb_typeof(input_data) = 'array' THEN jsonb_array_length(input_data)
        ELSE 0
    END;
$$;


ALTER FUNCTION public.pb_json_array_length(input_data jsonb) OWNER TO postgres;

--
-- Name: pb_json_each(anyelement); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.pb_json_each(input_data anyelement) RETURNS TABLE(value text)
    LANGUAGE plpgsql IMMUTABLE
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


ALTER FUNCTION public.pb_json_each(input_data anyelement) OWNER TO postgres;

--
-- Name: pb_json_each(jsonb); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.pb_json_each(input_data jsonb) RETURNS TABLE(value text)
    LANGUAGE plpgsql IMMUTABLE
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
        RETURN QUERY SELECT v::text FROM jsonb_each(input_data) AS t(k, v);
    ELSE
        RETURN QUERY SELECT trim(both '"' from input_data::text);
    END IF;
END;
$$;


ALTER FUNCTION public.pb_json_each(input_data jsonb) OWNER TO postgres;

--
-- Name: pb_json_extract(anyelement, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.pb_json_extract(data anyelement, path text) RETURNS text
    LANGUAGE plpgsql IMMUTABLE
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


ALTER FUNCTION public.pb_json_extract(data anyelement, path text) OWNER TO postgres;

--
-- Name: FUNCTION pb_json_extract(data anyelement, path text); Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON FUNCTION public.pb_json_extract(data anyelement, path text) IS 'Extracts values from JSON data as text using a path in the format $.x.y (similar to SQLite''s JSON_EXTRACT).

		Uses PostgreSQL''s jsonb_path_query_first with proper path formatting.

		Returns NULL for NULL inputs or when the path doesn''t exist.

		For non-JSON inputs, attempts to convert to JSON or returns the input as text.';


--
-- Name: pb_json_extract(jsonb, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.pb_json_extract(data jsonb, path text) RETURNS text
    LANGUAGE plpgsql IMMUTABLE
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


ALTER FUNCTION public.pb_json_extract(data jsonb, path text) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: _authOrigins; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."_authOrigins" (
    "collectionRef" text DEFAULT ''::text NOT NULL,
    created timestamp with time zone DEFAULT now() NOT NULL,
    fingerprint text DEFAULT ''::text NOT NULL,
    id text DEFAULT length(substr(md5((random())::text), 1, 15)) NOT NULL,
    "recordRef" text DEFAULT ''::text NOT NULL,
    updated timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public."_authOrigins" OWNER TO postgres;

--
-- Name: _collections; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public._collections (
    id character varying(15) DEFAULT length(substr(md5((random())::text), 1, 15)) NOT NULL,
    system boolean DEFAULT false NOT NULL,
    type text DEFAULT 'base'::text NOT NULL,
    name text NOT NULL,
    fields jsonb DEFAULT '[]'::jsonb NOT NULL,
    indexes jsonb DEFAULT '[]'::jsonb NOT NULL,
    "listRule" text,
    "viewRule" text,
    "createRule" text,
    "updateRule" text,
    "deleteRule" text,
    options jsonb DEFAULT '{}'::jsonb NOT NULL,
    created timestamp with time zone DEFAULT now() NOT NULL,
    updated timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public._collections OWNER TO postgres;

--
-- Name: _externalAuths; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."_externalAuths" (
    "collectionRef" text DEFAULT ''::text NOT NULL,
    created timestamp with time zone DEFAULT now() NOT NULL,
    id text DEFAULT length(substr(md5((random())::text), 1, 15)) NOT NULL,
    provider text DEFAULT ''::text NOT NULL,
    "providerId" text DEFAULT ''::text NOT NULL,
    "recordRef" text DEFAULT ''::text NOT NULL,
    updated timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public."_externalAuths" OWNER TO postgres;

--
-- Name: _mfas; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public._mfas (
    "collectionRef" text DEFAULT ''::text NOT NULL,
    created timestamp with time zone DEFAULT now() NOT NULL,
    id text DEFAULT length(substr(md5((random())::text), 1, 15)) NOT NULL,
    method text DEFAULT ''::text NOT NULL,
    "recordRef" text DEFAULT ''::text NOT NULL,
    updated timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public._mfas OWNER TO postgres;

--
-- Name: _migrations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public._migrations (
    file character varying(255) NOT NULL,
    applied bigint NOT NULL
);


ALTER TABLE public._migrations OWNER TO postgres;

--
-- Name: _otps; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public._otps (
    "collectionRef" text DEFAULT ''::text NOT NULL,
    created timestamp with time zone DEFAULT now() NOT NULL,
    id text DEFAULT length(substr(md5((random())::text), 1, 15)) NOT NULL,
    password text DEFAULT ''::text NOT NULL,
    "recordRef" text DEFAULT ''::text NOT NULL,
    "sentTo" text DEFAULT ''::text NOT NULL,
    updated timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public._otps OWNER TO postgres;

--
-- Name: _params; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public._params (
    id character varying(15) DEFAULT length(substr(md5((random())::text), 1, 15)) NOT NULL,
    value text,
    created timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public._params OWNER TO postgres;

--
-- Name: _superusers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public._superusers (
    created timestamp with time zone DEFAULT now() NOT NULL,
    email text DEFAULT ''::text NOT NULL,
    "emailVisibility" boolean DEFAULT false NOT NULL,
    id text DEFAULT length(substr(md5((random())::text), 1, 15)) NOT NULL,
    password text DEFAULT ''::text NOT NULL,
    "tokenKey" text DEFAULT ''::text NOT NULL,
    updated timestamp with time zone DEFAULT now() NOT NULL,
    verified boolean DEFAULT false NOT NULL
);


ALTER TABLE public._superusers OWNER TO postgres;

--
-- Name: clients; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.clients (
    created timestamp with time zone DEFAULT now() NOT NULL,
    email text DEFAULT ''::text NOT NULL,
    "emailVisibility" boolean DEFAULT false NOT NULL,
    id text DEFAULT length(substr(md5((random())::text), 1, 15)) NOT NULL,
    name text DEFAULT ''::text NOT NULL,
    password text DEFAULT ''::text NOT NULL,
    "tokenKey" text DEFAULT ''::text NOT NULL,
    updated timestamp with time zone DEFAULT now() NOT NULL,
    username text DEFAULT ''::text NOT NULL,
    verified boolean DEFAULT false NOT NULL
);


ALTER TABLE public.clients OWNER TO postgres;

--
-- Name: demo1; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.demo1 (
    bool boolean DEFAULT false NOT NULL,
    created timestamp with time zone DEFAULT now() NOT NULL,
    datetime text DEFAULT ''::text NOT NULL,
    email text DEFAULT ''::text NOT NULL,
    file_many jsonb DEFAULT '[]'::jsonb NOT NULL,
    file_one text DEFAULT ''::text NOT NULL,
    id text DEFAULT length(substr(md5((random())::text), 1, 15)) NOT NULL,
    "json" jsonb,
    number numeric DEFAULT 0 NOT NULL,
    point jsonb DEFAULT '{"lat": 0, "lon": 0}'::jsonb NOT NULL,
    rel_many jsonb DEFAULT '[]'::jsonb NOT NULL,
    rel_one text DEFAULT ''::text NOT NULL,
    select_many jsonb DEFAULT '[]'::jsonb NOT NULL,
    select_one text DEFAULT ''::text NOT NULL,
    text text DEFAULT ''::text NOT NULL,
    updated timestamp with time zone DEFAULT now() NOT NULL,
    url text DEFAULT ''::text NOT NULL
);


ALTER TABLE public.demo1 OWNER TO postgres;

--
-- Name: demo2; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.demo2 (
    active boolean DEFAULT false NOT NULL,
    created timestamp with time zone DEFAULT now() NOT NULL,
    id text DEFAULT length(substr(md5((random())::text), 1, 15)) NOT NULL,
    title text DEFAULT ''::text NOT NULL,
    updated timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.demo2 OWNER TO postgres;

--
-- Name: demo3; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.demo3 (
    created timestamp with time zone DEFAULT now() NOT NULL,
    files jsonb DEFAULT '[]'::jsonb NOT NULL,
    id text DEFAULT length(substr(md5((random())::text), 1, 15)) NOT NULL,
    title text DEFAULT ''::text NOT NULL,
    updated timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.demo3 OWNER TO postgres;

--
-- Name: demo4; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.demo4 (
    created timestamp with time zone DEFAULT now() NOT NULL,
    id text DEFAULT length(substr(md5((random())::text), 1, 15)) NOT NULL,
    "json_array" jsonb,
    "json_object" jsonb,
    rel_many_cascade jsonb DEFAULT '[]'::jsonb NOT NULL,
    rel_many_no_cascade jsonb DEFAULT '[]'::jsonb NOT NULL,
    rel_many_no_cascade_required jsonb DEFAULT '[]'::jsonb NOT NULL,
    rel_many_unique jsonb DEFAULT '[]'::jsonb NOT NULL,
    rel_one_cascade text DEFAULT ''::text NOT NULL,
    rel_one_no_cascade text DEFAULT ''::text NOT NULL,
    rel_one_no_cascade_required text DEFAULT ''::text NOT NULL,
    rel_one_unique text DEFAULT ''::text NOT NULL,
    self_rel_many jsonb DEFAULT '[]'::jsonb NOT NULL,
    self_rel_one text DEFAULT ''::text NOT NULL,
    title text DEFAULT ''::text NOT NULL,
    updated timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.demo4 OWNER TO postgres;

--
-- Name: demo5; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.demo5 (
    created timestamp with time zone DEFAULT now() NOT NULL,
    file text DEFAULT ''::text NOT NULL,
    id text DEFAULT length(substr(md5((random())::text), 1, 15)) NOT NULL,
    rel_many jsonb DEFAULT '[]'::jsonb NOT NULL,
    rel_one text DEFAULT ''::text NOT NULL,
    select_many jsonb DEFAULT '[]'::jsonb NOT NULL,
    select_one text DEFAULT ''::text NOT NULL,
    total numeric DEFAULT 0 NOT NULL,
    updated timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.demo5 OWNER TO postgres;

--
-- Name: nologin; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.nologin (
    created timestamp with time zone DEFAULT now() NOT NULL,
    email text DEFAULT ''::text NOT NULL,
    "emailVisibility" boolean DEFAULT false NOT NULL,
    id text DEFAULT length(substr(md5((random())::text), 1, 15)) NOT NULL,
    name text DEFAULT ''::text NOT NULL,
    password text DEFAULT ''::text NOT NULL,
    "tokenKey" text DEFAULT ''::text NOT NULL,
    updated timestamp with time zone DEFAULT now() NOT NULL,
    username text DEFAULT ''::text NOT NULL,
    verified boolean DEFAULT false NOT NULL
);


ALTER TABLE public.nologin OWNER TO postgres;

--
-- Name: numeric_id_view; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.numeric_id_view AS
 SELECT id,
    email
   FROM ( SELECT (unnamed_subquery_1.id)::text AS id,
            unnamed_subquery_1.email
           FROM ( SELECT row_number() OVER () AS id,
                    clients.email
                   FROM public.clients) unnamed_subquery_1) unnamed_subquery;


ALTER VIEW public.numeric_id_view OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    email text DEFAULT ''::text NOT NULL,
    "emailVisibility" boolean DEFAULT false NOT NULL,
    id text DEFAULT length(substr(md5((random())::text), 1, 15)) NOT NULL,
    password text DEFAULT ''::text NOT NULL,
    "tokenKey" text DEFAULT ''::text NOT NULL,
    verified boolean DEFAULT false NOT NULL,
    username text DEFAULT ''::text NOT NULL,
    name text DEFAULT ''::text NOT NULL,
    avatar text DEFAULT ''::text NOT NULL,
    rel text DEFAULT ''::text NOT NULL,
    created timestamp with time zone DEFAULT now() NOT NULL,
    updated timestamp with time zone DEFAULT now() NOT NULL,
    file jsonb DEFAULT '[]'::jsonb NOT NULL
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: view1; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.view1 AS
 SELECT id,
    text,
    bool,
    url,
    select_one,
    select_many,
    file_one,
    file_many,
    number,
    email,
    datetime,
    "json",
    rel_one,
    rel_many,
    created
   FROM ( SELECT demo1.id,
            demo1.text,
            demo1.bool,
            demo1.url,
            demo1.select_one,
            demo1.select_many,
            demo1.file_one,
            demo1.file_many,
            demo1.number,
            demo1.email,
            demo1.datetime,
            demo1."json",
            demo1.rel_one,
            demo1.rel_many,
            demo1.created
           FROM public.demo1) unnamed_subquery;


ALTER VIEW public.view1 OWNER TO postgres;

--
-- Name: view2; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.view2 AS
 SELECT id,
    state,
    file_many,
    rel_many
   FROM ( SELECT view1.id,
            view1.bool AS state,
            view1.file_many,
            view1.rel_many
           FROM public.view1) unnamed_subquery;


ALTER VIEW public.view2 OWNER TO postgres;

--
-- Data for Name: _authOrigins; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public."_authOrigins" ("collectionRef", created, fingerprint, id, "recordRef", updated) FROM stdin;
pbc_3142635823	2024-07-13 09:24:15.442+00	6afbfe481c31c08c55a746cccb88ece0	5f29jy38bf5zm3f	sywbhecnh46rhm0	2024-11-04 16:41:03.658+00
pbc_3142635823	2024-07-26 12:10:47.102+00	22bbbcbed36e25321f384ccf99f60057	dmy260k6ksjr4ib	sbmbsdb40jyxf7h	2024-11-04 16:41:03.656+00
pbc_3142635823	2024-07-26 12:11:38.697+00	6afbfe481c31c08c55a746cccb88ece0	ic55o70g4f8pcl4	sbmbsdb40jyxf7h	2024-11-04 16:41:03.654+00
pbc_3142635823	2024-07-26 12:12:17.972+00	dc879cfc889d0f1c1f3258d6f3a828fe	5798yh833k6w6w0	sbmbsdb40jyxf7h	2024-11-04 16:41:03.65+00
v851q4r790rhknl	2024-07-26 12:22:37.681+00	22bbbcbed36e25321f384ccf99f60057	9r2j0m74260ur8i	gk390qegs4y47wn	2024-07-26 12:22:37.681+00
\.


--
-- Data for Name: _collections; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public._collections (id, system, type, name, fields, indexes, "listRule", "viewRule", "createRule", "updateRule", "deleteRule", options, created, updated) FROM stdin;
_pb_users_auth_	f	auth	users	[{"id": "_pbf_text_id_", "max": 15, "min": 15, "name": "id", "type": "text", "hidden": false, "system": true, "pattern": "^[a-z0-9]+$", "required": true, "primaryKey": true, "presentable": false, "autogeneratePattern": "[a-z0-9]{15}"}, {"id": "_pbf_auth_password_", "max": 0, "min": 8, "cost": 10, "name": "password", "type": "password", "hidden": true, "system": true, "pattern": "", "required": true, "presentable": false}, {"id": "_pbf_auth_tokenKey_", "max": 60, "min": 30, "name": "tokenKey", "type": "text", "hidden": true, "system": true, "pattern": "", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": "[a-zA-Z0-9_]{50}"}, {"id": "_pbf_auth_email_", "name": "email", "type": "email", "hidden": false, "system": true, "required": false, "onlyDomains": null, "presentable": false, "exceptDomains": null}, {"id": "_pbf_auth_emailVisibility_", "name": "emailVisibility", "type": "bool", "hidden": false, "system": true, "required": false, "presentable": false}, {"id": "_pbf_auth_verified_", "name": "verified", "type": "bool", "hidden": false, "system": true, "required": false, "presentable": false}, {"id": "_pbf_auth_username_", "max": 150, "min": 3, "name": "username", "type": "text", "hidden": false, "system": false, "pattern": "^[\\\\w][\\\\w\\\\.\\\\-]*$", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": "users[0-9]{5}"}, {"id": "users_name", "max": 0, "min": 0, "name": "name", "type": "text", "hidden": false, "system": false, "pattern": "", "required": false, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "users_avatar", "name": "avatar", "type": "file", "hidden": false, "system": false, "thumbs": ["70x50", "70x50t", "70x50b", "70x50f", "0x50", "70x0"], "maxSize": 5242880, "required": false, "maxSelect": 1, "mimeTypes": ["image/jpg", "image/jpeg", "image/png", "image/svg+xml", "image/gif"], "protected": false, "presentable": false}, {"id": "xtecur3m", "name": "file", "type": "file", "hidden": false, "system": false, "thumbs": null, "maxSize": 5242880, "required": false, "maxSelect": 5, "mimeTypes": null, "protected": false, "presentable": false}, {"id": "lkeigvv3", "name": "rel", "type": "relation", "hidden": false, "system": false, "required": false, "maxSelect": 1, "minSelect": 0, "presentable": false, "collectionId": "sz5l5z67tg7gku0", "cascadeDelete": false}, {"id": "_pbf_autodate_created_", "name": "created", "type": "autodate", "hidden": false, "system": false, "onCreate": true, "onUpdate": false, "presentable": false}, {"id": "_pbf_autodate_updated_", "name": "updated", "type": "autodate", "hidden": false, "system": false, "onCreate": true, "onUpdate": true, "presentable": false}]	["CREATE UNIQUE INDEX `__pb_users_auth__username_idx` ON `users` (username)", "CREATE UNIQUE INDEX `__pb_users_auth__email_idx` ON `users` (email) WHERE email != ''", "CREATE UNIQUE INDEX `__pb_users_auth__tokenKey_idx` ON `users` (tokenKey)", "CREATE INDEX `__pb_users_auth__created_idx` ON `users` (`created`)"]	\N	id = @request.auth.id		id = @request.auth.id	id = @request.auth.id	{"mfa": {"rule": "", "enabled": true, "duration": 1800}, "otp": {"length": 8, "enabled": true, "duration": 300, "emailTemplate": {"body": "<p>Hello {RECORD:name}{RECORD:tokenKey},</p>\\n<p>Your one-time password is: <strong>{OTP}</strong></p>\\n<p><i>If you didn't ask for the one-time password, you can ignore this email.</i></p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "OTP for {APP_NAME}"}}, "oauth2": {"enabled": true, "providers": [{"name": "gitlab", "pkce": null, "authURL": "", "clientId": "test1", "tokenURL": "", "displayName": "", "userInfoURL": "", "clientSecret": "test2"}, {"name": "google", "pkce": null, "authURL": "", "clientId": "test", "tokenURL": "", "displayName": "", "userInfoURL": "", "clientSecret": "test2"}], "mappedFields": {"id": "", "name": "", "username": "username", "avatarURL": ""}}, "authRule": "", "authAlert": {"enabled": true, "emailTemplate": {"body": "<p>Hello {RECORD:name}{RECORD:tokenKey},</p>\\n<p>We noticed a login to your {APP_NAME} account from a new location.</p>\\n<p>If this was you, you may disregard this email.</p>\\n<p><strong>If this wasn't you, you should immediately change your {APP_NAME} account password to revoke access from all other locations.</strong></p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "Login from a new location"}}, "authToken": {"secret": "PjVU4hAV7CZIWbCByJHkDcMUlSEWCLI6M5aWSZOpEq0a3rYxKT", "duration": 1209600}, "fileToken": {"secret": "4Ax9zDm2Rwtny81dGaGQrJQBnIx5wVOuNe89X6v7NbNzrAZhvn", "duration": 180}, "manageRule": null, "passwordAuth": {"enabled": true, "identityFields": ["email", "username"]}, "emailChangeToken": {"secret": "eON2TTJZiGCEi7mvUvwMLADj8CMHQzwZN3gmyMjQb24EY08ATP", "duration": 1800}, "verificationToken": {"secret": "dgGGHlzzdCJ2C5MjXGoondllwSXkJHyL50FuvLvXGHNmBhvGKO", "duration": 604800}, "passwordResetToken": {"secret": "BC6jYPe4JXpQGGNzu6VXtYw0yhKoH2mh2ezIJClOJQuZYrd4Ol", "duration": 1800}, "verificationTemplate": {"body": "<p>Hello {RECORD:name}{RECORD:tokenKey},</p>\\n<p>Thank you for joining us at {APP_NAME}.</p>\\n<p>Click on the button below to verify your email address.</p>\\n<p>\\n  <a class=\\"btn\\" href=\\"{APP_URL}/_/#/auth/confirm-verification/{TOKEN}\\" target=\\"_blank\\" rel=\\"noopener\\">Verify</a>\\n</p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "Verify your {APP_NAME} email"}, "resetPasswordTemplate": {"body": "<p>Hello {RECORD:name}{RECORD:tokenKey},</p>\\n<p>Click on the button below to reset your password.</p>\\n<p>\\n  <a class=\\"btn\\" href=\\"{APP_URL}/_/#/auth/confirm-password-reset/{TOKEN}\\" target=\\"_blank\\" rel=\\"noopener\\">Reset password</a>\\n</p>\\n<p><i>If you didn't ask to reset your password, you can ignore this email.</i></p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "Reset your {APP_NAME} password"}, "confirmEmailChangeTemplate": {"body": "<p>Hello {RECORD:name}{RECORD:tokenKey},</p>\\n<p>Click on the button below to confirm your new email address.</p>\\n<p>\\n  <a class=\\"btn\\" href=\\"{APP_URL}/_/#/auth/confirm-email-change/{TOKEN}\\" target=\\"_blank\\" rel=\\"noopener\\">Confirm new email</a>\\n</p>\\n<p><i>If you didn't ask to change your email address, you can ignore this email.</i></p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "Confirm your {APP_NAME} new email address"}}	2022-10-10 09:49:46.145+00	2024-09-13 10:48:13.365+00
wsmn24bux7wo113	f	base	demo1	[{"id": "_pbf_text_id_", "max": 15, "min": 15, "name": "id", "type": "text", "hidden": false, "system": true, "pattern": "^[a-z0-9]+$", "required": true, "primaryKey": true, "presentable": false, "autogeneratePattern": "[a-z0-9]{15}"}, {"id": "u7spsiph", "max": 0, "min": 0, "name": "text", "type": "text", "hidden": false, "system": false, "pattern": "", "required": false, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "puk534el", "name": "bool", "type": "bool", "hidden": false, "system": false, "required": false, "presentable": false}, {"id": "ktas5n7b", "name": "url", "type": "url", "hidden": false, "system": false, "required": false, "onlyDomains": null, "presentable": false, "exceptDomains": null}, {"id": "dc4abz4i", "name": "select_one", "type": "select", "hidden": false, "system": false, "values": ["optionA", "optionB", "optionC"], "required": false, "maxSelect": 1, "presentable": false}, {"id": "owtlq7zl", "name": "select_many", "type": "select", "hidden": false, "system": false, "values": ["optionA", "optionB", "optionC"], "required": false, "maxSelect": 3, "presentable": false}, {"id": "4ulkdevf", "name": "file_one", "type": "file", "hidden": false, "system": false, "thumbs": null, "maxSize": 5242880, "required": false, "maxSelect": 1, "mimeTypes": null, "protected": true, "presentable": false}, {"id": "fjzhrsvq", "name": "file_many", "type": "file", "hidden": false, "system": false, "thumbs": null, "maxSize": 5242880, "required": false, "maxSelect": 99, "mimeTypes": null, "protected": false, "presentable": false}, {"id": "1z1ld0i5", "max": null, "min": null, "name": "number", "type": "number", "hidden": false, "system": false, "onlyInt": false, "required": false, "presentable": false}, {"id": "khvhpwgj", "name": "email", "type": "email", "hidden": false, "system": false, "required": false, "onlyDomains": null, "presentable": false, "exceptDomains": null}, {"id": "ro6p02gk", "max": "", "min": "", "name": "datetime", "type": "date", "hidden": false, "system": false, "required": false, "presentable": false}, {"id": "ei2fg4v1", "name": "json", "type": "json", "hidden": false, "system": false, "maxSize": 5242880, "required": false, "presentable": false}, {"id": "zaedritp", "name": "rel_one", "type": "relation", "hidden": false, "system": false, "required": false, "maxSelect": 1, "minSelect": 0, "presentable": false, "collectionId": "wsmn24bux7wo113", "cascadeDelete": false}, {"id": "t9bpk2ug", "name": "rel_many", "type": "relation", "hidden": false, "system": false, "required": false, "maxSelect": 9999, "minSelect": 0, "presentable": false, "collectionId": "_pb_users_auth_", "cascadeDelete": true}, {"id": "geoPoint3081106212", "name": "point", "type": "geoPoint", "hidden": false, "system": false, "required": false, "presentable": false}, {"id": "_pbf_autodate_created_", "name": "created", "type": "autodate", "hidden": false, "system": false, "onCreate": true, "onUpdate": false, "presentable": false}, {"id": "_pbf_autodate_updated_", "name": "updated", "type": "autodate", "hidden": false, "system": false, "onCreate": true, "onUpdate": true, "presentable": false}]	["CREATE INDEX `_wsmn24bux7wo113_created_idx` ON `demo1` (`created`)"]	\N	\N	\N	\N	\N	{}	2022-10-10 09:51:19.868+00	2025-04-27 07:49:04.021+00
sz5l5z67tg7gku0	f	base	demo2	[{"id": "_pbf_text_id_", "max": 15, "min": 15, "name": "id", "type": "text", "hidden": false, "system": true, "pattern": "^[a-z0-9]+$", "required": true, "primaryKey": true, "presentable": false, "autogeneratePattern": "[a-z0-9]{15}"}, {"id": "mkrguaaf", "max": 0, "min": 2, "name": "title", "type": "text", "hidden": false, "system": false, "pattern": "", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "izkl5z2s", "name": "active", "type": "bool", "hidden": false, "system": false, "required": false, "presentable": false}, {"id": "_pbf_autodate_created_", "name": "created", "type": "autodate", "hidden": false, "system": false, "onCreate": true, "onUpdate": false, "presentable": false}, {"id": "_pbf_autodate_updated_", "name": "updated", "type": "autodate", "hidden": false, "system": false, "onCreate": true, "onUpdate": true, "presentable": false}]	["CREATE INDEX `idx_demo2_created` ON `demo2` (`created`)", "CREATE UNIQUE INDEX \\"idx_unique_demo2_title\\" on \\"demo2\\" (\\"title\\")", "CREATE INDEX \\"idx_demo2_active\\" ON \\"demo2\\" (\\n\\"active\\"\\n)"]						{}	2022-10-10 09:51:28.452+00	2023-03-20 16:39:25.211+00
wzlqyes4orhoygb	f	base	demo3	[{"id": "_pbf_text_id_", "max": 15, "min": 15, "name": "id", "type": "text", "hidden": false, "system": true, "pattern": "^[a-z0-9]+$", "required": true, "primaryKey": true, "presentable": false, "autogeneratePattern": "[a-z0-9]{15}"}, {"id": "w5z2x0nq", "max": 0, "min": 0, "name": "title", "type": "text", "hidden": false, "system": false, "pattern": "", "required": false, "primaryKey": false, "presentable": true, "autogeneratePattern": ""}, {"id": "tgqrbwio", "name": "files", "type": "file", "hidden": false, "system": false, "thumbs": null, "maxSize": 5242880, "required": false, "maxSelect": 99, "mimeTypes": null, "protected": false, "presentable": false}, {"id": "_pbf_autodate_created_", "name": "created", "type": "autodate", "hidden": false, "system": false, "onCreate": true, "onUpdate": false, "presentable": false}, {"id": "_pbf_autodate_updated_", "name": "updated", "type": "autodate", "hidden": false, "system": false, "onCreate": true, "onUpdate": true, "presentable": false}]	["CREATE INDEX `_wzlqyes4orhoygb_created_idx` ON `demo3` (`created`)"]	@request.auth.id != "" && @request.auth.collectionName != "users"	@request.auth.id != "" && @request.auth.collectionName != "users"	@request.auth.id != "" && @request.auth.collectionName != "users"	@request.auth.id != "" && @request.auth.collectionName != "users"	@request.auth.id != "" && @request.auth.collectionName != "users"	{}	2022-10-10 09:51:36.853+00	2023-11-20 18:26:53.176+00
4d1blo5cuycfaca	f	base	demo4	[{"id": "_pbf_text_id_", "max": 15, "min": 15, "name": "id", "type": "text", "hidden": false, "system": true, "pattern": "^[a-z0-9]+$", "required": true, "primaryKey": true, "presentable": false, "autogeneratePattern": "[a-z0-9]{15}"}, {"id": "erkxnabw", "max": 0, "min": 0, "name": "title", "type": "text", "hidden": false, "system": false, "pattern": "", "required": false, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "t5jskeyz", "name": "rel_one_no_cascade", "type": "relation", "hidden": false, "system": false, "required": false, "maxSelect": 1, "minSelect": 0, "presentable": false, "collectionId": "wzlqyes4orhoygb", "cascadeDelete": false}, {"id": "ldrcjhk8", "name": "rel_one_no_cascade_required", "type": "relation", "hidden": false, "system": false, "required": true, "maxSelect": 1, "minSelect": 0, "presentable": false, "collectionId": "wzlqyes4orhoygb", "cascadeDelete": false}, {"id": "pl5lcd4y", "name": "rel_one_cascade", "type": "relation", "hidden": false, "system": false, "required": false, "maxSelect": 1, "minSelect": 0, "presentable": false, "collectionId": "wzlqyes4orhoygb", "cascadeDelete": true}, {"id": "jz0oue3z", "name": "rel_many_no_cascade", "type": "relation", "hidden": false, "system": false, "required": false, "maxSelect": 999, "minSelect": 0, "presentable": false, "collectionId": "wzlqyes4orhoygb", "cascadeDelete": false}, {"id": "bsxyqrhb", "name": "rel_many_no_cascade_required", "type": "relation", "hidden": false, "system": false, "required": true, "maxSelect": 999, "minSelect": 0, "presentable": false, "collectionId": "wzlqyes4orhoygb", "cascadeDelete": false}, {"id": "kwmchnf7", "name": "rel_many_cascade", "type": "relation", "hidden": false, "system": false, "required": false, "maxSelect": 999, "minSelect": 0, "presentable": false, "collectionId": "wzlqyes4orhoygb", "cascadeDelete": true}, {"id": "pmynkqk5", "name": "rel_one_unique", "type": "relation", "hidden": false, "system": false, "required": false, "maxSelect": 1, "minSelect": 0, "presentable": false, "collectionId": "wzlqyes4orhoygb", "cascadeDelete": false}, {"id": "mjzyk9vb", "name": "rel_many_unique", "type": "relation", "hidden": false, "system": false, "required": false, "maxSelect": 999, "minSelect": 0, "presentable": false, "collectionId": "wzlqyes4orhoygb", "cascadeDelete": false}, {"id": "dagiyxj4", "name": "self_rel_one", "type": "relation", "hidden": false, "system": false, "required": false, "maxSelect": 1, "minSelect": 0, "presentable": false, "collectionId": "4d1blo5cuycfaca", "cascadeDelete": false}, {"id": "tsrki8kc", "name": "self_rel_many", "type": "relation", "hidden": false, "system": false, "required": false, "maxSelect": 999, "minSelect": 0, "presentable": false, "collectionId": "4d1blo5cuycfaca", "cascadeDelete": false}, {"id": "4wpx0hhx", "name": "json_array", "type": "json", "hidden": false, "system": false, "maxSize": 5242880, "required": false, "presentable": false}, {"id": "ufpwiqnx", "name": "json_object", "type": "json", "hidden": false, "system": false, "maxSize": 5242880, "required": false, "presentable": false}, {"id": "_pbf_autodate_created_", "name": "created", "type": "autodate", "hidden": false, "system": false, "onCreate": true, "onUpdate": false, "presentable": false}, {"id": "_pbf_autodate_updated_", "name": "updated", "type": "autodate", "hidden": false, "system": false, "onCreate": true, "onUpdate": true, "presentable": false}]	["CREATE INDEX `_4d1blo5cuycfaca_created_idx` ON `demo4` (`created`)", "CREATE UNIQUE INDEX `idx_luoQV2A` ON `demo4` (`rel_one_unique`) WHERE rel_one_unique != ''"]			@request.auth.collectionName = 'users'	@request.auth.collectionName = 'users'		{}	2022-10-10 09:51:47.77+00	2025-04-27 07:49:03.999+00
v851q4r790rhknl	f	auth	clients	[{"id": "_pbf_text_id_", "max": 15, "min": 15, "name": "id", "type": "text", "hidden": false, "system": true, "pattern": "^[a-z0-9]+$", "required": true, "primaryKey": true, "presentable": false, "autogeneratePattern": "[a-z0-9]{15}"}, {"id": "_pbf_auth_password_", "max": 0, "min": 8, "cost": 10, "name": "password", "type": "password", "hidden": true, "system": true, "pattern": "", "required": true, "presentable": false}, {"id": "_pbf_auth_tokenKey_", "max": 60, "min": 30, "name": "tokenKey", "type": "text", "hidden": true, "system": true, "pattern": "", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": "[a-zA-Z0-9_]{50}"}, {"id": "_pbf_auth_email_", "name": "email", "type": "email", "hidden": false, "system": true, "required": true, "onlyDomains": null, "presentable": false, "exceptDomains": null}, {"id": "_pbf_auth_emailVisibility_", "name": "emailVisibility", "type": "bool", "hidden": false, "system": true, "required": false, "presentable": false}, {"id": "_pbf_auth_verified_", "name": "verified", "type": "bool", "hidden": false, "system": true, "required": false, "presentable": false}, {"id": "_pbf_auth_username_", "max": 150, "min": 3, "name": "username", "type": "text", "hidden": false, "system": false, "pattern": "^[\\\\w][\\\\w\\\\.\\\\-]*$", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": "users[0-9]{5}"}, {"id": "lacorw19", "max": 0, "min": 0, "name": "name", "type": "text", "hidden": false, "system": false, "pattern": "", "required": false, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "_pbf_autodate_created_", "name": "created", "type": "autodate", "hidden": false, "system": false, "onCreate": true, "onUpdate": false, "presentable": false}, {"id": "_pbf_autodate_updated_", "name": "updated", "type": "autodate", "hidden": false, "system": false, "onCreate": true, "onUpdate": true, "presentable": false}]	["CREATE UNIQUE INDEX `_v851q4r790rhknl_username_idx` ON `clients` (username)", "CREATE UNIQUE INDEX `_v851q4r790rhknl_email_idx` ON `clients` (email) WHERE email != ''", "CREATE UNIQUE INDEX `_v851q4r790rhknl_tokenKey_idx` ON `clients` (tokenKey)", "CREATE INDEX `_v851q4r790rhknl_created_idx` ON `clients` (`created`)"]	\N	\N	\N	\N	\N	{"mfa": {"enabled": false, "duration": 1800}, "otp": {"length": 8, "enabled": false, "duration": 300, "emailTemplate": {"body": "<p>Hello,</p>\\n<p>Your one-time password is: <strong>{OTP}</strong></p>\\n<p><i>If you didn't ask for the one-time password, you can ignore this email.</i></p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "OTP for {APP_NAME}"}}, "oauth2": {"enabled": false, "providers": [], "mappedFields": {"id": "", "name": "", "username": "username", "avatarUrl": ""}}, "authRule": "verified=true", "authAlert": {"enabled": true, "emailTemplate": {"body": "<p>Hello,</p>\\n<p>We noticed a login to your {APP_NAME} account from a new location.</p>\\n<p>If this was you, you may disregard this email.</p>\\n<p><strong>If this wasn't you, you should immediately change your {APP_NAME} account password to revoke access from all other locations.</strong></p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "Login from a new location"}}, "authToken": {"secret": "PjVU4hAV7CZIWbCByJHkDcMUlSEWCLI6M5aWSZOpEq0a3rYxKT", "duration": 1209600}, "fileToken": {"secret": "4Ax9zDm2Rwtny81dGaGQrJQBnIx5wVOuNe89X6v7NbNzrAZhvn", "duration": 180}, "manageRule": null, "passwordAuth": {"enabled": true, "identityFields": ["email", "username"]}, "emailChangeToken": {"secret": "eON2TTJZiGCEi7mvUvwMLADj8CMHQzwZN3gmyMjQb24EY08ATP", "duration": 1800}, "verificationToken": {"secret": "dgGGHlzzdCJ2C5MjXGoondllwSXkJHyL50FuvLvXGHNmBhvGKO", "duration": 604800}, "passwordResetToken": {"secret": "BC6jYPe4JXpQGGNzu6VXtYw0yhKoH2mh2ezIJClOJQuZYrd4Ol", "duration": 1800}, "verificationTemplate": {"body": "<p>Hello,</p>\\n<p>Thank you for joining us at {APP_NAME}.</p>\\n<p>Click on the button below to verify your email address.</p>\\n<p>\\n  <a class=\\"btn\\" href=\\"{APP_URL}/_/#/auth/confirm-verification/{TOKEN}\\" target=\\"_blank\\" rel=\\"noopener\\">Verify</a>\\n</p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "Verify your {APP_NAME} email"}, "resetPasswordTemplate": {"body": "<p>Hello,</p>\\n<p>Click on the button below to reset your password.</p>\\n<p>\\n  <a class=\\"btn\\" href=\\"{APP_URL}/_/#/auth/confirm-password-reset/{TOKEN}\\" target=\\"_blank\\" rel=\\"noopener\\">Reset password</a>\\n</p>\\n<p><i>If you didn't ask to reset your password, you can ignore this email.</i></p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "Reset your {APP_NAME} password"}, "confirmEmailChangeTemplate": {"body": "<p>Hello,</p>\\n<p>Click on the button below to confirm your new email address.</p>\\n<p>\\n  <a class=\\"btn\\" href=\\"{APP_URL}/_/#/auth/confirm-email-change/{TOKEN}\\" target=\\"_blank\\" rel=\\"noopener\\">Confirm new email</a>\\n</p>\\n<p><i>If you didn't ask to change your email address, you can ignore this email.</i></p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "Confirm your {APP_NAME} new email address"}}	2022-10-11 09:41:36.712+00	2024-06-28 15:59:08.446+00
kpv709sk2lqbqk8	t	auth	nologin	[{"id": "_pbf_text_id_", "max": 15, "min": 15, "name": "id", "type": "text", "hidden": false, "system": true, "pattern": "^[a-z0-9]+$", "required": true, "primaryKey": true, "presentable": false, "autogeneratePattern": "[a-z0-9]{15}"}, {"id": "_pbf_auth_password_", "max": 0, "min": 8, "cost": 10, "name": "password", "type": "password", "hidden": true, "system": true, "pattern": "", "required": true, "presentable": false}, {"id": "_pbf_auth_tokenKey_", "max": 60, "min": 30, "name": "tokenKey", "type": "text", "hidden": true, "system": true, "pattern": "", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": "[a-zA-Z0-9_]{50}"}, {"id": "_pbf_auth_email_", "name": "email", "type": "email", "hidden": false, "system": true, "required": true, "onlyDomains": null, "presentable": false, "exceptDomains": null}, {"id": "_pbf_auth_emailVisibility_", "name": "emailVisibility", "type": "bool", "hidden": false, "system": true, "required": false, "presentable": false}, {"id": "_pbf_auth_verified_", "name": "verified", "type": "bool", "hidden": false, "system": true, "required": false, "presentable": false}, {"id": "_pbf_auth_username_", "max": 150, "min": 3, "name": "username", "type": "text", "hidden": false, "system": false, "pattern": "^[\\\\w][\\\\w\\\\.\\\\-]*$", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": "users[0-9]{5}"}, {"id": "x8zzktwe", "max": 0, "min": 0, "name": "name", "type": "text", "hidden": false, "system": false, "pattern": "", "required": false, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "_pbf_autodate_created_", "name": "created", "type": "autodate", "hidden": false, "system": false, "onCreate": true, "onUpdate": false, "presentable": false}, {"id": "_pbf_autodate_updated_", "name": "updated", "type": "autodate", "hidden": false, "system": false, "onCreate": true, "onUpdate": true, "presentable": false}]	["CREATE UNIQUE INDEX `_kpv709sk2lqbqk8_username_idx` ON `nologin` (username)", "CREATE UNIQUE INDEX `_kpv709sk2lqbqk8_email_idx` ON `nologin` (email) WHERE email != ''", "CREATE UNIQUE INDEX `_kpv709sk2lqbqk8_tokenKey_idx` ON `nologin` (tokenKey)", "CREATE INDEX `_kpv709sk2lqbqk8_created_idx` ON \\"nologin\\" (`created`)"]						{"mfa": {"enabled": false, "duration": 1800}, "otp": {"length": 8, "enabled": false, "duration": 300, "emailTemplate": {"body": "<p>Hello,</p>\\n<p>Your one-time password is: <strong>{OTP}</strong></p>\\n<p><i>If you didn't ask for the one-time password, you can ignore this email.</i></p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "OTP for {APP_NAME}"}}, "oauth2": {"enabled": false, "providers": [{"name": "gitlab", "pkce": null, "authUrl": "", "clientId": "test", "tokenUrl": "", "userApiUrl": "", "displayName": "", "clientSecret": "test"}], "mappedFields": {"id": "", "name": "", "username": "username", "avatarUrl": ""}}, "authRule": "", "authAlert": {"enabled": true, "emailTemplate": {"body": "<p>Hello,</p>\\n<p>We noticed a login to your {APP_NAME} account from a new location.</p>\\n<p>If this was you, you may disregard this email.</p>\\n<p><strong>If this wasn't you, you should immediately change your {APP_NAME} account password to revoke access from all other locations.</strong></p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "Login from a new location"}}, "authToken": {"secret": "PjVU4hAV7CZIWbCByJHkDcMUlSEWCLI6M5aWSZOpEq0a3rYxKT", "duration": 1209600}, "fileToken": {"secret": "4Ax9zDm2Rwtny81dGaGQrJQBnIx5wVOuNe89X6v7NbNzrAZhvn", "duration": 180}, "manageRule": "@request.auth.collectionName = 'users'", "passwordAuth": {"enabled": false, "identityFields": ["email"]}, "emailChangeToken": {"secret": "eON2TTJZiGCEi7mvUvwMLADj8CMHQzwZN3gmyMjQb24EY08ATP", "duration": 1800}, "verificationToken": {"secret": "dgGGHlzzdCJ2C5MjXGoondllwSXkJHyL50FuvLvXGHNmBhvGKO", "duration": 604800}, "passwordResetToken": {"secret": "BC6jYPe4JXpQGGNzu6VXtYw0yhKoH2mh2ezIJClOJQuZYrd4Ol", "duration": 1800}, "verificationTemplate": {"body": "<p>Hello,</p>\\n<p>Thank you for joining us at {APP_NAME}.</p>\\n<p>Click on the button below to verify your email address.</p>\\n<p>\\n  <a class=\\"btn\\" href=\\"{APP_URL}/_/#/auth/confirm-verification/{TOKEN}\\" target=\\"_blank\\" rel=\\"noopener\\">Verify</a>\\n</p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "Verify your {APP_NAME} email"}, "resetPasswordTemplate": {"body": "<p>Hello,</p>\\n<p>Click on the button below to reset your password.</p>\\n<p>\\n  <a class=\\"btn\\" href=\\"{APP_URL}/_/#/auth/confirm-password-reset/{TOKEN}\\" target=\\"_blank\\" rel=\\"noopener\\">Reset password</a>\\n</p>\\n<p><i>If you didn't ask to reset your password, you can ignore this email.</i></p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "Reset your {APP_NAME} password"}, "confirmEmailChangeTemplate": {"body": "<p>Hello,</p>\\n<p>Click on the button below to confirm your new email address.</p>\\n<p>\\n  <a class=\\"btn\\" href=\\"{APP_URL}/_/#/auth/confirm-email-change/{TOKEN}\\" target=\\"_blank\\" rel=\\"noopener\\">Confirm new email</a>\\n</p>\\n<p><i>If you didn't ask to change your email address, you can ignore this email.</i></p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "Confirm your {APP_NAME} new email address"}}	2022-10-12 10:39:21.294+00	2024-06-26 18:37:35.353+00
9n89pl5vkct6330	f	base	demo5	[{"id": "_pbf_text_id_", "max": 15, "min": 15, "name": "id", "type": "text", "hidden": false, "system": true, "pattern": "^[a-z0-9]+$", "required": true, "primaryKey": true, "presentable": false, "autogeneratePattern": "[a-z0-9]{15}"}, {"id": "sozvpexq", "name": "select_one", "type": "select", "hidden": false, "system": false, "values": ["a", "b", "c", "d"], "required": false, "maxSelect": 1, "presentable": false}, {"id": "qlq1nxlc", "name": "select_many", "type": "select", "hidden": false, "system": false, "values": ["a", "b", "c", "d", "e"], "required": false, "maxSelect": 5, "presentable": false}, {"id": "ajrrsq1a", "name": "rel_one", "type": "relation", "hidden": false, "system": false, "required": false, "maxSelect": 1, "minSelect": 0, "presentable": false, "collectionId": "4d1blo5cuycfaca", "cascadeDelete": false}, {"id": "soxhs0ou", "name": "rel_many", "type": "relation", "hidden": false, "system": false, "required": false, "maxSelect": 5, "minSelect": 0, "presentable": false, "collectionId": "4d1blo5cuycfaca", "cascadeDelete": false}, {"id": "kvbyzuqj", "max": null, "min": null, "name": "total", "type": "number", "hidden": false, "system": false, "onlyInt": false, "required": false, "presentable": false}, {"id": "ob7dsrcl", "name": "file", "type": "file", "hidden": false, "system": false, "thumbs": null, "maxSize": 5242880, "required": false, "maxSelect": 1, "mimeTypes": null, "protected": false, "presentable": false}, {"id": "_pbf_autodate_created_", "name": "created", "type": "autodate", "hidden": false, "system": false, "onCreate": true, "onUpdate": false, "presentable": false}, {"id": "_pbf_autodate_updated_", "name": "updated", "type": "autodate", "hidden": false, "system": false, "onCreate": true, "onUpdate": true, "presentable": false}]	["CREATE INDEX `_9n89pl5vkct6330_created_idx` ON `demo5` (`created`)"]	select_many:length = 3	rel_many.self_rel_many.rel_many_cascade.files:length = 1	@request.body.total = 3	@request.body.total = 3	@request.query.test:isset = true	{}	2023-01-07 13:13:08.733+00	2023-04-04 13:10:52.723+00
v9gwnfh02gjq1q0	f	view	view1	[{"id": "text3208210256", "max": 0, "min": 0, "name": "id", "type": "text", "hidden": false, "system": true, "pattern": "^[a-z0-9]+$", "required": true, "primaryKey": true, "presentable": false, "autogeneratePattern": ""}, {"id": "_clone_F4wf", "max": 0, "min": 0, "name": "text", "type": "text", "hidden": false, "system": false, "pattern": "", "required": false, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "_clone_F2ir", "name": "bool", "type": "bool", "hidden": false, "system": false, "required": false, "presentable": false}, {"id": "_clone_X6zE", "name": "url", "type": "url", "hidden": false, "system": false, "required": false, "onlyDomains": null, "presentable": false, "exceptDomains": null}, {"id": "_clone_bZA2", "name": "select_one", "type": "select", "hidden": false, "system": false, "values": ["optionA", "optionB", "optionC"], "required": false, "maxSelect": 1, "presentable": false}, {"id": "_clone_B5De", "name": "select_many", "type": "select", "hidden": false, "system": false, "values": ["optionA", "optionB", "optionC"], "required": false, "maxSelect": 3, "presentable": false}, {"id": "_clone_gPsK", "name": "file_one", "type": "file", "hidden": false, "system": false, "thumbs": null, "maxSize": 5242880, "required": false, "maxSelect": 1, "mimeTypes": null, "protected": true, "presentable": false}, {"id": "_clone_sZpk", "name": "file_many", "type": "file", "hidden": false, "system": false, "thumbs": null, "maxSize": 5242880, "required": false, "maxSelect": 99, "mimeTypes": null, "protected": false, "presentable": false}, {"id": "_clone_Goev", "max": null, "min": null, "name": "number", "type": "number", "hidden": false, "system": false, "onlyInt": false, "required": false, "presentable": false}, {"id": "_clone_Das5", "name": "email", "type": "email", "hidden": false, "system": false, "required": false, "onlyDomains": null, "presentable": false, "exceptDomains": null}, {"id": "_clone_psUd", "max": "", "min": "", "name": "datetime", "type": "date", "hidden": false, "system": false, "required": false, "presentable": false}, {"id": "_clone_WfVC", "name": "json", "type": "json", "hidden": false, "system": false, "maxSize": 5242880, "required": false, "presentable": false}, {"id": "_clone_OqCt", "name": "rel_one", "type": "relation", "hidden": false, "system": false, "required": false, "maxSelect": 1, "minSelect": 0, "presentable": false, "collectionId": "wsmn24bux7wo113", "cascadeDelete": false}, {"id": "_clone_9UR2", "name": "rel_many", "type": "relation", "hidden": false, "system": false, "required": false, "maxSelect": 9999, "minSelect": 0, "presentable": false, "collectionId": "_pb_users_auth_", "cascadeDelete": true}, {"id": "_clone_Y06D", "name": "created", "type": "autodate", "hidden": false, "system": false, "onCreate": true, "onUpdate": false, "presentable": false}]	[]	@request.auth.id != "" && bool = true	@request.auth.id != "" && bool = true	\N	\N	\N	{"viewQuery": "select id, text, bool, url, select_one, select_many, file_one, file_many, number, email, datetime, json, rel_one, rel_many, created from demo1"}	2023-02-12 18:58:12.315+00	2024-11-19 15:29:43.683+00
ib3m2700k5hlsjz	f	view	view2	[{"id": "text3208210256", "max": 0, "min": 0, "name": "id", "type": "text", "hidden": false, "system": true, "pattern": "^[a-z0-9]+$", "required": true, "primaryKey": true, "presentable": false, "autogeneratePattern": ""}, {"id": "_clone_77Ik", "name": "state", "type": "bool", "hidden": false, "system": false, "required": false, "presentable": false}, {"id": "_clone_IbSu", "name": "file_many", "type": "file", "hidden": false, "system": false, "thumbs": null, "maxSize": 5242880, "required": false, "maxSelect": 99, "mimeTypes": null, "protected": false, "presentable": false}, {"id": "_clone_Asz0", "name": "rel_many", "type": "relation", "hidden": false, "system": false, "required": false, "maxSelect": 9999, "minSelect": 0, "presentable": false, "collectionId": "_pb_users_auth_", "cascadeDelete": true}]	[]			\N	\N	\N	{"viewQuery": "SELECT view1.id, view1.bool as state, view1.file_many, view1.rel_many from view1\\n"}	2023-02-17 19:42:54.278+00	2024-11-19 15:29:43.857+00
zahsr9d2mix2fvk	f	view	numeric_id_view	[{"id": "text3208210256", "max": 0, "min": 0, "name": "id", "type": "text", "hidden": false, "system": true, "pattern": "^[a-z0-9]+$", "required": true, "primaryKey": true, "presentable": false, "autogeneratePattern": ""}, {"id": "_clone_jO9O", "name": "email", "type": "email", "hidden": false, "system": true, "required": true, "onlyDomains": null, "presentable": false, "exceptDomains": null}]	[]			\N	\N	\N	{"viewQuery": "select (ROW_NUMBER() OVER()) as id, email from clients"}	2023-08-11 09:41:00.997+00	2024-11-19 15:29:43.901+00
pbc_3142635823	t	auth	_superusers	[{"id": "text3208210256", "max": 15, "min": 15, "name": "id", "type": "text", "hidden": false, "system": true, "pattern": "^[a-z0-9]+$", "required": true, "primaryKey": true, "presentable": false, "autogeneratePattern": "[a-z0-9]{15}"}, {"id": "password901924565", "max": 0, "min": 8, "cost": 0, "name": "password", "type": "password", "hidden": true, "system": true, "pattern": "", "required": true, "presentable": false}, {"id": "text2504183744", "max": 60, "min": 30, "name": "tokenKey", "type": "text", "hidden": true, "system": true, "pattern": "", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": "[a-zA-Z0-9_]{50}"}, {"id": "email3885137012", "name": "email", "type": "email", "hidden": false, "system": true, "required": true, "onlyDomains": null, "presentable": false, "exceptDomains": null}, {"id": "bool1547992806", "name": "emailVisibility", "type": "bool", "hidden": false, "system": true, "required": false, "presentable": false}, {"id": "bool256245529", "name": "verified", "type": "bool", "hidden": false, "system": true, "required": false, "presentable": false}, {"id": "autodate2990389176", "name": "created", "type": "autodate", "hidden": false, "system": true, "onCreate": true, "onUpdate": false, "presentable": false}, {"id": "autodate3332085495", "name": "updated", "type": "autodate", "hidden": false, "system": true, "onCreate": true, "onUpdate": true, "presentable": false}]	["CREATE UNIQUE INDEX `idx_tokenKey__pbc_3323866339` ON `_superusers` (`tokenKey`)", "CREATE UNIQUE INDEX `idx_email__pbc_3323866339` ON `_superusers` (`email`) WHERE `email` != ''"]	\N	\N	\N	\N	\N	{"mfa": {"rule": "", "enabled": false, "duration": 1800}, "otp": {"length": 8, "enabled": false, "duration": 300, "emailTemplate": {"body": "<p>Hello,</p>\\n<p>Your one-time password is: <strong>{OTP}</strong></p>\\n<p><i>If you didn't ask for the one-time password, you can ignore this email.</i></p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "OTP for {APP_NAME}"}}, "oauth2": {"enabled": false, "providers": [], "mappedFields": {"id": "", "name": "", "username": "", "avatarURL": ""}}, "authRule": "", "authAlert": {"enabled": true, "emailTemplate": {"body": "<p>Hello,</p>\\n<p>We noticed a login to your {APP_NAME} account from a new location.</p>\\n<p>If this was you, you may disregard this email.</p>\\n<p><strong>If this wasn't you, you should immediately change your {APP_NAME} account password to revoke access from all other locations.</strong></p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "Login from a new location"}}, "authToken": {"secret": "MyN3nDlzmHnuCjd35vb6cyIdqNr7Os0PmgiPVDMxmbFToSpBvS", "duration": 1209600}, "fileToken": {"secret": "sjJAjTNPrOcRDmnIKwQm7qY9FyjuXTG5KNcaqw4U1TSDVfu4r9", "duration": 180}, "manageRule": null, "passwordAuth": {"enabled": true, "identityFields": ["email"]}, "emailChangeToken": {"secret": "unYNiYeuIxH7BCV09NIb81abe2bkPgaexMYdDQ6uOOIFh74urD", "duration": 1800}, "verificationToken": {"secret": "uhr68rXLVjPBWALFtw8uEHeQwDdN4t0MiTLr2pBWVkEQnNICe1", "duration": 259200}, "passwordResetToken": {"secret": "fPSpFm9rxjj4mdeWYfyQ5OZQ4UWpyainTO0dqrJe3LHEYEDduq", "duration": 1800}, "verificationTemplate": {"body": "<p>Hello,</p>\\n<p>Thank you for joining us at {APP_NAME}.</p>\\n<p>Click on the button below to verify your email address.</p>\\n<p>\\n  <a class=\\"btn\\" href=\\"{APP_URL}/_/#/auth/confirm-verification/{TOKEN}\\" target=\\"_blank\\" rel=\\"noopener\\">Verify</a>\\n</p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "Verify your {APP_NAME} email"}, "resetPasswordTemplate": {"body": "<p>Hello,</p>\\n<p>Click on the button below to reset your password.</p>\\n<p>\\n  <a class=\\"btn\\" href=\\"{APP_URL}/_/#/auth/confirm-password-reset/{TOKEN}\\" target=\\"_blank\\" rel=\\"noopener\\">Reset password</a>\\n</p>\\n<p><i>If you didn't ask to reset your password, you can ignore this email.</i></p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "Reset your {APP_NAME} password"}, "confirmEmailChangeTemplate": {"body": "<p>Hello,</p>\\n<p>Click on the button below to confirm your new email address.</p>\\n<p>\\n  <a class=\\"btn\\" href=\\"{APP_URL}/_/#/auth/confirm-email-change/{TOKEN}\\" target=\\"_blank\\" rel=\\"noopener\\">Confirm new email</a>\\n</p>\\n<p><i>If you didn't ask to change your email address, you can ignore this email.</i></p>\\n<p>\\n  Thanks,<br/>\\n  {APP_NAME} team\\n</p>", "subject": "Confirm your {APP_NAME} new email address"}}	2024-06-20 09:29:22.826+00	2024-09-08 10:49:38.496+00
pbc_2281828961	t	base	_externalAuths	[{"id": "text3208210256", "max": 15, "min": 15, "name": "id", "type": "text", "hidden": false, "system": true, "pattern": "^[a-z0-9]+$", "required": true, "primaryKey": true, "presentable": false, "autogeneratePattern": "[a-z0-9]{15}"}, {"id": "text455797646", "max": 0, "min": 0, "name": "collectionRef", "type": "text", "hidden": false, "system": true, "pattern": "", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "text127846527", "max": 0, "min": 0, "name": "recordRef", "type": "text", "hidden": false, "system": true, "pattern": "", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "text2462348188", "max": 0, "min": 0, "name": "provider", "type": "text", "hidden": false, "system": true, "pattern": "", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "text1044722854", "max": 0, "min": 0, "name": "providerId", "type": "text", "hidden": false, "system": true, "pattern": "", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "autodate2990389176", "name": "created", "type": "autodate", "hidden": false, "system": true, "onCreate": true, "onUpdate": false, "presentable": false}, {"id": "autodate3332085495", "name": "updated", "type": "autodate", "hidden": false, "system": true, "onCreate": true, "onUpdate": true, "presentable": false}]	["CREATE UNIQUE INDEX `idx_externalAuths_record_provider` ON `externalAuths` (collectionRef, recordRef, provider)", "CREATE UNIQUE INDEX `idx_externalAuths_collection_provider` ON `externalAuths` (collectionRef, provider, providerId)"]	@request.auth.id != '' && recordRef = @request.auth.id && collectionRef = @request.auth.collectionId	@request.auth.id != '' && recordRef = @request.auth.id && collectionRef = @request.auth.collectionId	\N	\N	@request.auth.id != '' && recordRef = @request.auth.id && collectionRef = @request.auth.collectionId	{}	2024-06-20 09:29:22.862+00	2024-07-31 16:33:51.596+00
pbc_2279338944	t	base	_mfas	[{"id": "text3208210256", "max": 15, "min": 15, "name": "id", "type": "text", "hidden": false, "system": true, "pattern": "^[a-z0-9]+$", "required": true, "primaryKey": true, "presentable": false, "autogeneratePattern": "[a-z0-9]{15}"}, {"id": "text455797646", "max": 0, "min": 0, "name": "collectionRef", "type": "text", "hidden": false, "system": true, "pattern": "", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "text127846527", "max": 0, "min": 0, "name": "recordRef", "type": "text", "hidden": false, "system": true, "pattern": "", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "text1582905952", "max": 0, "min": 0, "name": "method", "type": "text", "hidden": false, "system": true, "pattern": "", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "autodate2990389176", "name": "created", "type": "autodate", "hidden": false, "system": true, "onCreate": true, "onUpdate": false, "presentable": false}, {"id": "autodate3332085495", "name": "updated", "type": "autodate", "hidden": false, "system": true, "onCreate": true, "onUpdate": true, "presentable": false}]	["CREATE INDEX `idx_mfas_collectionRef_recordRef` ON `mfas` (\\n  `collectionRef`,\\n  `recordRef`\\n)"]	@request.auth.id != '' && recordRef = @request.auth.id && collectionRef = @request.auth.collectionId	@request.auth.id != '' && recordRef = @request.auth.id && collectionRef = @request.auth.collectionId	\N	\N	\N	{}	2024-06-20 09:29:22.894+00	2024-10-24 18:34:35.352+00
pbc_1638494021	t	base	_otps	[{"id": "text3208210256", "max": 15, "min": 15, "name": "id", "type": "text", "hidden": false, "system": true, "pattern": "^[a-z0-9]+$", "required": true, "primaryKey": true, "presentable": false, "autogeneratePattern": "[a-z0-9]{15}"}, {"id": "text455797646", "max": 0, "min": 0, "name": "collectionRef", "type": "text", "hidden": false, "system": true, "pattern": "", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "text127846527", "max": 0, "min": 0, "name": "recordRef", "type": "text", "hidden": false, "system": true, "pattern": "", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "password901924565", "max": 0, "min": 0, "cost": 8, "name": "password", "type": "password", "hidden": true, "system": true, "pattern": "", "required": true, "presentable": false}, {"id": "autodate2990389176", "name": "created", "type": "autodate", "hidden": false, "system": true, "onCreate": true, "onUpdate": false, "presentable": false}, {"id": "autodate3332085495", "name": "updated", "type": "autodate", "hidden": false, "system": true, "onCreate": true, "onUpdate": true, "presentable": false}, {"id": "text3866985172", "max": 0, "min": 0, "name": "sentTo", "type": "text", "hidden": true, "system": true, "pattern": "", "required": false, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}]	["CREATE INDEX `idx_otps_collectionRef_recordRef` ON `otps` (\\n  `collectionRef`,\\n  `recordRef`\\n)"]	@request.auth.id != '' && recordRef = @request.auth.id && collectionRef = @request.auth.collectionId	@request.auth.id != '' && recordRef = @request.auth.id && collectionRef = @request.auth.collectionId	\N	\N	\N	{}	2024-06-20 09:29:22.901+00	2024-11-19 15:29:43.633+00
pbc_4275539003	t	base	_authOrigins	[{"id": "text3208210256", "max": 15, "min": 15, "name": "id", "type": "text", "hidden": false, "system": true, "pattern": "^[a-z0-9]+$", "required": true, "primaryKey": true, "presentable": false, "autogeneratePattern": "[a-z0-9]{15}"}, {"id": "text455797646", "max": 0, "min": 0, "name": "collectionRef", "type": "text", "hidden": false, "system": true, "pattern": "", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "text127846527", "max": 0, "min": 0, "name": "recordRef", "type": "text", "hidden": false, "system": true, "pattern": "", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "text4228609354", "max": 0, "min": 0, "name": "fingerprint", "type": "text", "hidden": false, "system": true, "pattern": "", "required": true, "primaryKey": false, "presentable": false, "autogeneratePattern": ""}, {"id": "autodate2990389176", "name": "created", "type": "autodate", "hidden": false, "system": true, "onCreate": true, "onUpdate": false, "presentable": false}, {"id": "autodate3332085495", "name": "updated", "type": "autodate", "hidden": false, "system": true, "onCreate": true, "onUpdate": true, "presentable": false}]	["CREATE UNIQUE INDEX `idx_authOrigins_unique_pairs` ON `authDevices` (collectionRef, recordRef, fingerprint)"]	@request.auth.id != '' && recordRef = @request.auth.id && collectionRef = @request.auth.collectionId	@request.auth.id != '' && recordRef = @request.auth.id && collectionRef = @request.auth.collectionId	\N	\N	@request.auth.id != '' && recordRef = @request.auth.id && collectionRef = @request.auth.collectionId	{}	2024-06-20 12:10:52.542+00	2024-07-31 16:32:24.722+00
\.


--
-- Data for Name: _externalAuths; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public."_externalAuths" ("collectionRef", created, id, provider, "providerId", "recordRef", updated) FROM stdin;
_pb_users_auth_	2024-07-29 11:57:28.598+00	5eto7nmys833164	github	test456	bgs820n361vj1qd	2024-07-29 11:57:28.598+00
_pb_users_auth_	2023-01-01 01:02:03.456+00	clmflokuq1xl341	google	test123	4q1xlclmfloku33	2022-01-01 01:01:01.123+00
_pb_users_auth_	2022-01-01 01:01:01.123+00	dlmflokuq1xl342	gitlab	test123	4q1xlclmfloku33	2022-01-01 01:01:01.123+00
v851q4r790rhknl	2024-07-29 11:56:28.439+00	f1z5b3843pzc964	google	test456	gk390qegs4y47wn	2024-07-29 11:56:35.151+00
\.


--
-- Data for Name: _mfas; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public._mfas ("collectionRef", created, id, method, "recordRef", updated) FROM stdin;
\.


--
-- Data for Name: _migrations; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public._migrations (file, applied) FROM stdin;
1640988000_aux_init.go	1745478534887500
1640988000_init.go	1745478534947825
1640987000_pg.go	1745478534947825
\.


--
-- Data for Name: _otps; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public._otps ("collectionRef", created, id, password, "recordRef", "sentTo", updated) FROM stdin;
\.


--
-- Data for Name: _params; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public._params (id, value, created, updated) FROM stdin;
settings	{"smtp":{"enabled":false,"port":587,"host":"smtp.example.com","username":"","authMethod":"","tls":false,"localName":""},"backups":{"cron":"","cronMaxKeep":0,"s3":{"enabled":false,"bucket":"","region":"","endpoint":"","accessKey":"","forcePathStyle":false}},"s3":{"enabled":false,"bucket":"","region":"","endpoint":"","accessKey":"","forcePathStyle":false},"meta":{"appName":"acme_test","appURL":"http://localhost:8090","senderName":"Support","senderAddress":"support@example.com","hideControls":false},"logs":{"maxDays":5,"minLevel":0,"logIP":false},"batch":{"enabled":true,"maxRequests":50,"timeout":5,"maxBodySize":0},"rateLimits":{"rules":[{"label":"*:auth","maxRequests":2,"duration":3},{"label":"*:create","maxRequests":20,"duration":5},{"label":"/api/batch","maxRequests":3,"duration":1},{"label":"/api/","maxRequests":300,"duration":10}],"enabled":false},"trustedProxy":{"headers":[],"useLeftmostIP":false}}	2025-04-24 07:08:54.951+00	2025-04-24 08:25:31.571+00
\.


--
-- Data for Name: _superusers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public._superusers (created, email, "emailVisibility", id, password, "tokenKey", updated, verified) FROM stdin;
2022-10-10 09:50:06.449+00	test@example.com	f	sywbhecnh46rhm0	$2a$13$6PP1fpShEsQHpZElw4.BN.vMl/2CMIKc2kpumJM//T4qV4RGUgb6i	O4rvW9FSUyTA3xUuQmXR3wHF2db9bHs19nBHeSgVTxerOsTAl4	2022-10-10 09:50:06.449+00	t
2022-10-10 11:22:27.982+00	test2@example.com	f	sbmbsdb40jyxf7h	$2a$13$1U32ttl.v1VtQkIH.a.z2.VVrU2IFwkW41TnY2OgyMcjADv34ynMK	cvg1nk1dKRFlazQH8nCKuFYwczdReQx6ZJimxXvei0uDyTkgEb	2022-10-10 11:22:31.096+00	t
2022-10-10 11:22:50.693+00	test3@example.com	f	9q2trqumvlyr3bd	$2a$13$gqXkfs0WjqTtUNRjRDxAHuLul1sxmA.elEfYuL0KT0ef5cMDt9Fjm	ezLvEu7DRFtUp9BI6nxtXCpgtp7qWaNQLdD6dDwjIVB0mA0uUr	2022-10-10 11:22:54.907+00	t
2024-07-26 12:17:43.787+00	test4@example.com	f	q911776rrfy658l	$2a$10$0CUWH81tPWUhxKqksZnO/eDfJiWTTG/xg7v5Yg8RhcLIGWY3oeCHe	B4LS_3ZiIlL_TQ97_W_u2_82__W1_l64037w_x211_F_BM_J4_	2024-07-26 12:17:43.787+00	t
\.


--
-- Data for Name: clients; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.clients (created, email, "emailVisibility", id, name, password, "tokenKey", updated, username, verified) FROM stdin;
2022-10-11 09:42:15.113+00	test@example.com	f	gk390qegs4y47wn		$2a$13$utsLtiNVf226R7hhXXyOM.kHTyTcVLRNqTIrT44Fw8JZbxMhIK0uu	rMb1gUpn27s53t66gOGscSHfYsa272cgOgn4nhTZIl4fIC8XP8	2022-10-11 09:42:15.113+00	clients57772	t
2022-10-11 09:42:25.984+00	test2@example.com	f	o1y0dd0spd786md	test_name	$2a$13$D/8j2Q7NzN5g/INiVn8qPOa2O3qkyZj7U82CXOyqzBFV0B9OdWhvC	RunSD73nFfH3sNScreizPcZYNkiPls2YjmFYPNo73cWKsDnZVm	2022-10-14 11:44:33.15+00	clients43362	f
\.


--
-- Data for Name: demo1; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.demo1 (bool, created, datetime, email, file_many, file_one, id, "json", number, point, rel_many, rel_one, select_many, select_one, text, updated, url) FROM stdin;
t	2022-10-14 10:13:12.397+00	2022-10-01 12:00:00.000Z	test@example.com	["test_QZFjKjXchk.txt", "300_WlbFWSGmW9.png", "logo_vcfJJG5TAh.svg", "test_MaWC6mWyrP.txt", "test_tC1Yc87DfC.txt"]	test_d61b33QdDU.txt	84nmscqy84lsi1t	[1, 2, 3]	123456	{"lat": 0, "lon": 0}	["oap640cot4yru2s"]		["optionB", "optionC"]	optionB	test	2023-01-04 20:11:20.732+00	https://example.copm
f	2022-10-14 10:14:04.685+00		test2@example.com	[]	300_Jsjq7RdBgA.png	al1h9ijdeojtsjy	null	456	{"lat": 42.654318, "lon": 23.333157}	["bgs820n361vj1qd", "4q1xlclmfloku33", "oap640cot4yru2s"]	84nmscqy84lsi1t	["optionB"]	optionB	test2	2025-04-02 11:05:26.292+00	
f	2022-10-14 10:36:21.012+00			[]		imy661ixudk5izi	null	0	{"lat": 40.712728, "lon": -74.006015}	[]		[]		lorem ipsum	2025-04-02 11:05:03.433+00	
\.


--
-- Data for Name: demo2; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.demo2 (active, created, id, title, updated) FROM stdin;
t	2022-10-12 11:42:58.215+00	0yxhwia2amd8gec	test3	2022-10-14 10:52:49.596+00
f	2022-10-12 11:42:51.509+00	llvuca81nly1qls	test1	2022-10-12 11:42:51.509+00
t	2022-10-12 11:42:55.076+00	achvryl401bhse3	test2	2022-10-14 10:52:46.726+00
\.


--
-- Data for Name: demo3; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.demo3 (created, files, id, title, updated) FROM stdin;
2022-10-14 11:41:09.381+00	[]	1tmknxy2868d869	test1	2022-10-14 11:41:09.381+00
2022-10-14 11:41:12.548+00	["test_FLurQTgrY8.txt", "300_UhLKX91HVb.png"]	lcl9d87w22ml6jy	test2	2022-10-14 14:08:29.095+00
2022-10-14 11:41:15.689+00	["test_JnXeKEwgwr.txt"]	7nwo8tuiatetxdm	test3	2022-10-14 14:08:16.072+00
2022-10-14 11:41:18.877+00	["300_JdfBOieXAW.png"]	mk5fmymtx4wsprk	test4	2022-10-14 14:08:06.446+00
\.


--
-- Data for Name: demo4; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.demo4 (created, id, "json_array", "json_object", rel_many_cascade, rel_many_no_cascade, rel_many_no_cascade_required, rel_many_unique, rel_one_cascade, rel_one_no_cascade, rel_one_no_cascade_required, rel_one_unique, self_rel_many, self_rel_one, title, updated) FROM stdin;
2022-10-14 14:15:43.771+00	qzaqccwrmva4o1n	[1]	{"a": 123}	["mk5fmymtx4wsprk"]	["1tmknxy2868d869"]	["lcl9d87w22ml6jy", "1tmknxy2868d869"]	[]	7nwo8tuiatetxdm	mk5fmymtx4wsprk	lcl9d87w22ml6jy		["i9naidtvr6qsgb4", "qzaqccwrmva4o1n"]	i9naidtvr6qsgb4	test1	2022-10-20 19:23:59.427+00
2022-10-14 17:35:18.647+00	i9naidtvr6qsgb4	[1, 2, 3]	{"a": {"b": "test"}}	[]	[]	["7nwo8tuiatetxdm"]	[]			lcl9d87w22ml6jy		[]	qzaqccwrmva4o1n	test2	2022-10-20 19:23:39.645+00
\.


--
-- Data for Name: demo5; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.demo5 (created, file, id, rel_many, rel_one, select_many, select_one, total, updated) FROM stdin;
2023-01-07 13:13:24.249+00	logo_vcf_jjg5_tah_9MtIHytOmZ.svg	qjeql998mtp1azp	["qzaqccwrmva4o1n", "i9naidtvr6qsgb4"]	i9naidtvr6qsgb4	["b", "c", "a"]	b	0	2023-02-18 11:51:08.128+00
2023-01-07 13:13:28.337+00	300_uh_lkx91_hvb_Da8K5pl069.png	la4y2w4o98acwuj	[]		[]		2	2023-02-18 11:50:46.943+00
\.


--
-- Data for Name: nologin; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.nologin (created, email, "emailVisibility", id, name, password, "tokenKey", updated, username, verified) FROM stdin;
2022-10-12 10:39:47.633+00	test@example.com	f	dc49k6jgejn40h3	test	$2a$13$395zl/w/nUmeo.vov5t6SO0/RsoE4suJnKLN2SzP41sJCe6PBAePq	6mi4JvnX8uIxS7JiO8LWl150Af8mnAGWkWRImHL2YB4XhOYf9c	2022-10-12 10:39:47.633+00	test_username	f
2022-10-12 10:39:59.058+00	test2@example.com	t	phhq3wr65cap535		$2a$13$ldkBOjQbXIXP3.xvJ8AXJep5I3kMzmXWu7wC75mP5RN4qK7mwQIK.	9Bsj2ogZ5b0Q3daAKZ5ZZrUuVb7lWHO0YDD4d5TbPCLfYE3COY	2022-10-14 12:03:49.335+00	viewers74618	f
2022-10-14 12:16:52.515+00	test3@example.com	f	oos036e9xvqeexy		$2a$13$oXNjS0xVi4aYHSaQ10bTOOMRvVFBY2t474S3UjM8.5BNnL5WB1Gce	uYafEN2cDgImHKMW4SfZ6tayrENAJVVWOTDxFXatOs2AO0KEMY	2022-10-14 12:16:52.515+00	nologin84738	t
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (email, "emailVisibility", id, password, "tokenKey", verified, username, name, avatar, rel, created, updated, file) FROM stdin;
test@example.com	f	4q1xlclmfloku33	$2a$13$uULN32dHTkWQIALJP1N54.22HW/K9/qkXqcmsfz2hOA1wGeuUDbfG	tfYe7rCTX4D2KuWQY3pJjBifgsrMbecyXBatEPjrSfGEGS2jh6	f	users75657	test1	300_1SEi6Q6U72.png	llvuca81nly1qls	2022-10-10 10:36:26.879+00	2022-10-12 11:46:10.49+00	[]
test3@example.com	t	bgs820n361vj1qd	$2a$13$qzF1J0ePG5.fvBrm0fVrtez5RRBjPOUoezvYRbTQGAfsT85d4XH2K	x6vHUi00LvM5bFeGIpwXN9xuol8k1BknfTmlySQ7YQWoTLKOa7	t	users69238	test3		0yxhwia2amd8gec	2022-10-10 10:37:33.119+00	2022-10-12 11:46:02.462+00	[]
test2@example.com	f	oap640cot4yru2s	$2a$13$RC6/uXsHWM1ZV1v0cPJLRuWPXxyNINDmUDIHTq1x1dM.K.TBgWzFK	AQbE30CNb8Ncwr6Sg0sfvDJGJuepriTJN24EHZqO5DsEBTk1kA	t	test2_username				2022-10-10 10:36:59.438+00	2022-10-11 18:37:20.744+00	["test_kfd2wYLxkz.txt"]
\.


--
-- Name: _authOrigins _authOrigins_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."_authOrigins"
    ADD CONSTRAINT "_authOrigins_pkey" PRIMARY KEY (id);


--
-- Name: _collections _collections_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public._collections
    ADD CONSTRAINT _collections_name_key UNIQUE (name);


--
-- Name: _collections _collections_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public._collections
    ADD CONSTRAINT _collections_pkey PRIMARY KEY (id);


--
-- Name: _externalAuths _externalAuths_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."_externalAuths"
    ADD CONSTRAINT "_externalAuths_pkey" PRIMARY KEY (id);


--
-- Name: _mfas _mfas_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public._mfas
    ADD CONSTRAINT _mfas_pkey PRIMARY KEY (id);


--
-- Name: _migrations _migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public._migrations
    ADD CONSTRAINT _migrations_pkey PRIMARY KEY (file);


--
-- Name: _otps _otps_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public._otps
    ADD CONSTRAINT _otps_pkey PRIMARY KEY (id);


--
-- Name: _params _params_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public._params
    ADD CONSTRAINT _params_pkey PRIMARY KEY (id);


--
-- Name: _superusers _superusers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public._superusers
    ADD CONSTRAINT _superusers_pkey PRIMARY KEY (id);


--
-- Name: clients clients_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.clients
    ADD CONSTRAINT clients_pkey PRIMARY KEY (id);


--
-- Name: demo1 demo1_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.demo1
    ADD CONSTRAINT demo1_pkey PRIMARY KEY (id);


--
-- Name: demo2 demo2_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.demo2
    ADD CONSTRAINT demo2_pkey PRIMARY KEY (id);


--
-- Name: demo3 demo3_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.demo3
    ADD CONSTRAINT demo3_pkey PRIMARY KEY (id);


--
-- Name: demo4 demo4_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.demo4
    ADD CONSTRAINT demo4_pkey PRIMARY KEY (id);


--
-- Name: demo5 demo5_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.demo5
    ADD CONSTRAINT demo5_pkey PRIMARY KEY (id);


--
-- Name: nologin nologin_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.nologin
    ADD CONSTRAINT nologin_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: _4d1blo5cuycfaca_created_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX _4d1blo5cuycfaca_created_idx ON public.demo4 USING btree (created);


--
-- Name: _9n89pl5vkct6330_created_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX _9n89pl5vkct6330_created_idx ON public.demo5 USING btree (created);


--
-- Name: __pb_users_auth__created_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX __pb_users_auth__created_idx ON public.users USING btree (created);


--
-- Name: __pb_users_auth__email_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX __pb_users_auth__email_idx ON public.users USING btree (email) WHERE (email <> ''::text);


--
-- Name: __pb_users_auth__tokenKey_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "__pb_users_auth__tokenKey_idx" ON public.users USING btree ("tokenKey");


--
-- Name: __pb_users_auth__username_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX __pb_users_auth__username_idx ON public.users USING btree (username);


--
-- Name: _kpv709sk2lqbqk8_created_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX _kpv709sk2lqbqk8_created_idx ON public.nologin USING btree (created);


--
-- Name: _kpv709sk2lqbqk8_email_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX _kpv709sk2lqbqk8_email_idx ON public.nologin USING btree (email) WHERE (email <> ''::text);


--
-- Name: _kpv709sk2lqbqk8_tokenKey_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "_kpv709sk2lqbqk8_tokenKey_idx" ON public.nologin USING btree ("tokenKey");


--
-- Name: _kpv709sk2lqbqk8_username_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX _kpv709sk2lqbqk8_username_idx ON public.nologin USING btree (username);


--
-- Name: _v851q4r790rhknl_created_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX _v851q4r790rhknl_created_idx ON public.clients USING btree (created);


--
-- Name: _v851q4r790rhknl_email_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX _v851q4r790rhknl_email_idx ON public.clients USING btree (email) WHERE (email <> ''::text);


--
-- Name: _v851q4r790rhknl_tokenKey_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "_v851q4r790rhknl_tokenKey_idx" ON public.clients USING btree ("tokenKey");


--
-- Name: _v851q4r790rhknl_username_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX _v851q4r790rhknl_username_idx ON public.clients USING btree (username);


--
-- Name: _wsmn24bux7wo113_created_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX _wsmn24bux7wo113_created_idx ON public.demo1 USING btree (created);


--
-- Name: _wzlqyes4orhoygb_created_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX _wzlqyes4orhoygb_created_idx ON public.demo3 USING btree (created);


--
-- Name: idx__collections_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx__collections_type ON public._collections USING btree (type);


--
-- Name: idx_authOrigins_unique_pairs; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "idx_authOrigins_unique_pairs" ON public."_authOrigins" USING btree ("collectionRef", "recordRef", fingerprint);


--
-- Name: idx_demo2_active; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_demo2_active ON public.demo2 USING btree (active);


--
-- Name: idx_demo2_created; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_demo2_created ON public.demo2 USING btree (created);


--
-- Name: idx_email_pbc_3142635823; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_email_pbc_3142635823 ON public._superusers USING btree (email) WHERE (email <> ''::text);


--
-- Name: idx_externalAuths_collection_provider; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "idx_externalAuths_collection_provider" ON public."_externalAuths" USING btree ("collectionRef", provider, "providerId");


--
-- Name: idx_externalAuths_record_provider; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "idx_externalAuths_record_provider" ON public."_externalAuths" USING btree ("collectionRef", "recordRef", provider);


--
-- Name: idx_luoQV2A; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "idx_luoQV2A" ON public.demo4 USING btree (rel_one_unique) WHERE (rel_one_unique <> ''::text);


--
-- Name: idx_mfas_collectionRef_recordRef; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "idx_mfas_collectionRef_recordRef" ON public._mfas USING btree ("collectionRef", "recordRef");


--
-- Name: idx_otps_collectionRef_recordRef; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "idx_otps_collectionRef_recordRef" ON public._otps USING btree ("collectionRef", "recordRef");


--
-- Name: idx_tokenKey_pbc_3142635823; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "idx_tokenKey_pbc_3142635823" ON public._superusers USING btree ("tokenKey");


--
-- Name: idx_unique_demo2_title; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_unique_demo2_title ON public.demo2 USING btree (title);


--
-- PostgreSQL database dump complete
--

\unrestrict DF9FExrGxX28ulqgCixzacrQ7cMeRgfhhncVVTe3nMnYDCQnGraFW582JjwY1e5

