-- +goose Up
-- Accent-insensitive search: "the alla pesca" must find "Tè alla pesca",
-- "caffe" must find "Caffè". unaccent() is STABLE (it reads a dictionary);
-- the IMMUTABLE wrapper pins the dictionary so it can be used in indexes.
CREATE EXTENSION IF NOT EXISTS unaccent;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION pesa_unaccent(text) RETURNS text
  LANGUAGE sql IMMUTABLE PARALLEL SAFE STRICT
  AS $$ SELECT public.unaccent('public.unaccent'::regdictionary, $1) $$;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION IF EXISTS pesa_unaccent(text);
