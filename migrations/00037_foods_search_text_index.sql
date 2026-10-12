-- +goose Up
-- Search spent ~100 ms unaccenting all ~19k names on every query (twice:
-- results and count). Store the unaccented haystack once and index it with
-- trigrams: pg_trgm accelerates regex matches (~), one condition per word.
ALTER TABLE foods
  ADD COLUMN IF NOT EXISTS search_text TEXT
  GENERATED ALWAYS AS (pesa_unaccent(lower(
    coalesce(name_it, '') || ' ' || coalesce(name_en, '') || ' ' || coalesce(synonyms, '')))) STORED;
CREATE INDEX IF NOT EXISTS idx_foods_search_text_trgm ON foods USING gin (search_text gin_trgm_ops);

-- +goose Down
DROP INDEX IF EXISTS idx_foods_search_text_trgm;
ALTER TABLE foods DROP COLUMN IF EXISTS search_text;
