-- +goose Up
-- Foto sbagliate dall'arricchimento Wikipedia: sui prodotti OFF (senza foto
-- propria) la ricerca per nome+marca dava loghi, stemmi e quadri omonimi;
-- gli SVG sono loghi o diagrammi. Meglio nessuna foto che una sbagliata.
DELETE FROM food_images i
 USING foods f, food_sources s
 WHERE f.id = i.food_id AND s.id = f.source_id
   AND s.code = 'off'
   AND (i.license LIKE 'per-file%' OR i.url LIKE '%wikimedia%');

DELETE FROM food_images WHERE url ILIKE '%.svg%';

-- +goose Down
SELECT 1;
