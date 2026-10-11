-- +goose Up
-- Foto Wikipedia sbagliate su carne fresca, frattaglie e pesce CREA: la voce
-- illustra l'animale vivo o uno schema anatomico, non il cibo. Meglio l'icona
-- di default. I salumi restano (la foto mostra il prodotto).
DELETE FROM food_images i
 USING foods f
 WHERE f.id = i.food_id
   AND f.category IN ('Carni fresche', 'Frattaglie', 'Prodotti della pesca')
   AND (i.license LIKE 'per-file%' OR i.url LIKE '%wikimedia%');

-- +goose Down
SELECT 1;
