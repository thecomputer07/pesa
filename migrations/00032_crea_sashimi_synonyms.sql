-- +goose Up
-- Sashimi is raw sliced fish: its values are the CREA raw-fish ones. No new
-- foods, just a search synonym so "sashimi" finds them. Matched by CREA name
-- (ids differ between environments); idempotent.
UPDATE foods f
   SET synonyms = trim(both ' ' FROM coalesce(f.synonyms, '') || ' sashimi')
  FROM food_sources s
 WHERE s.id = f.source_id
   AND s.code = 'crea'
   AND f.name_it IN ('Salmone', 'Tonno', 'Orata, filetti', 'Spigola', 'Calamaro',
                     'Polpo', 'Gamberi', 'Sgombro o maccarello', 'Seppia')
   AND coalesce(f.synonyms, '') NOT ILIKE '%sashimi%';

-- +goose Down
UPDATE foods SET synonyms = nullif(trim(both ' ' FROM replace(synonyms, 'sashimi', '')), '')
 WHERE synonyms ILIKE '%sashimi%';
