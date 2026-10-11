-- +goose Up
-- "Estathé" is one word: a search for "the alla pesca" could not reach it.
-- Give Estathé products the plain word as a synonym. Idempotent.
UPDATE foods f
   SET synonyms = trim(both ' ' FROM coalesce(f.synonyms, '') || ' the')
 WHERE pesa_unaccent(lower(f.name_it)) ~ 'estath?e'
   AND coalesce(f.synonyms, '') !~* '\mthe\M';

-- +goose Down
SELECT 1;
