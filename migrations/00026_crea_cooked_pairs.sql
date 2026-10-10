-- +goose Up
-- Stato crudo/cotto degli alimenti CREA e coppie per il toggle in scheda.
--
-- L'importer CREA scriveva sempre state='raw' (anche "Pasta di semola, cotta,
-- bollita"), quindi il toggle mostrava solo "crudo". Regole:
--   * state='cooked' se il nome CREA contiene un termine di cottura;
--   * toggle (parent_food_id cotto -> crudo) SOLO se il crudo ha UNA sola
--     variante cotta, per bollitura/cottura semplice (come pasta, riso,
--     legumi), e i carboidrati cambiano davvero (>= 1 g e >= 5 g/100g). Piu' metodi (patate, spinaci, pollo:
--     forno/padella/microonde/fritto...) => nessun toggle, voci separate.
-- La funzione e' richiamata anche a fine import CREA (importer.RefreshCREAStates).

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION pesa_refresh_crea_states() RETURNS void
LANGUAGE plpgsql AS $$
DECLARE
  -- un termine di cottura (qualsiasi metodo)
  any_method CONSTANT text := '(cott[oaie]|bollit[oaie]|lessat[oaie]|arrostit[oaie]|arrosto|fritt[oaie]|al forno|in padella|al microonde|allo spiedo|alla griglia|ai ferri|in umido|stufat[oaie]|brasat[oaie]|al vapore)';
  -- cottura semplice in acqua, l'unica che giustifica il toggle
  simple CONSTANT text := '^(cott[oaie]|bollit[oaie]|lessat[oaie])(, (bollit[oaie]|lessat[oaie]))?$';
  crea_id bigint;
BEGIN
  SELECT id INTO crea_id FROM food_sources WHERE code = 'crea';
  IF crea_id IS NULL THEN RETURN; END IF;

  -- 1) stato dal nome (i prodotti "Prosciutto cotto", "Cotechino" e i
  --    "seccati al forno" non sono la versione cotta di un crudo)
  UPDATE foods f SET state = CASE
      WHEN f.name_it ~* ('\m' || any_method || '\M')
       AND f.name_it !~* '^(prosciutto cotto|cotechino)'
       AND f.name_it !~* 'seccat'
        THEN 'cooked'
      WHEN f.state = 'cooked' THEN 'raw'
      ELSE f.state
    END
   WHERE f.source_id = crea_id;

  -- 2) ricalcolo coppie
  UPDATE foods SET parent_food_id = NULL
   WHERE source_id = crea_id AND parent_food_id IS NOT NULL;

  WITH derivs AS (
    -- c = variante cotta di r se il nome e' "<r>, <solo termini di cottura>"
    SELECT r.id AS raw_id, c.id AS cooked_id,
           substr(c.name_it, length(r.name_it) + 3) AS suffix,
           r.carb_g AS raw_carb, c.carb_g AS cooked_carb
      FROM foods r
      JOIN foods c ON c.source_id = r.source_id
                  AND c.state = 'cooked'
                  AND c.superseded_by IS NULL
                  AND lower(c.name_it) LIKE lower(r.name_it) || ', %'
     WHERE r.source_id = crea_id
       AND r.state IN ('raw', 'dried')
       AND r.superseded_by IS NULL
       AND substr(c.name_it, length(r.name_it) + 3)
           ~* ('^' || any_method || '(, ' || any_method || ')*$')
  ), single AS (
    SELECT raw_id, min(cooked_id) AS cooked_id
      FROM derivs GROUP BY raw_id HAVING count(*) = 1
  )
  UPDATE foods c SET parent_food_id = d.raw_id
    FROM single s JOIN derivs d ON d.cooked_id = s.cooked_id AND d.raw_id = s.raw_id
   WHERE c.id = s.cooked_id
     AND d.suffix ~* simple
     -- "non cambia nulla": serve una differenza reale di carboidrati
     AND abs(coalesce(d.raw_carb, 0) - coalesce(d.cooked_carb, 0)) >= 1
     AND greatest(coalesce(d.raw_carb, 0), coalesce(d.cooked_carb, 0)) >= 5;
END;
$$;
-- +goose StatementEnd

-- Fixture iniziali (33 voci inserite a mano, senza record di origine) erano
-- attribuite a CREA ma con valori approssimati e voci non CREA (Nutella,
-- Tiramisu'): duplicano l'import reale. Rimosse se non referenziate.
DELETE FROM foods f
 USING food_sources fs
 WHERE fs.id = f.source_id AND fs.code = 'crea'
   AND f.source_record_id IS NULL
   AND NOT EXISTS (SELECT 1 FROM meal_items m WHERE m.food_id = f.id)
   AND NOT EXISTS (SELECT 1 FROM recipe_items ri WHERE ri.food_id = f.id)
   AND NOT EXISTS (SELECT 1 FROM foods ch WHERE ch.parent_food_id = f.id OR ch.superseded_by = f.id)
   AND f.parent_food_id IS NOT NULL;
DELETE FROM foods f
 USING food_sources fs
 WHERE fs.id = f.source_id AND fs.code = 'crea'
   AND f.source_record_id IS NULL
   AND NOT EXISTS (SELECT 1 FROM meal_items m WHERE m.food_id = f.id)
   AND NOT EXISTS (SELECT 1 FROM recipe_items ri WHERE ri.food_id = f.id)
   AND NOT EXISTS (SELECT 1 FROM foods ch WHERE ch.parent_food_id = f.id OR ch.superseded_by = f.id);

SELECT pesa_refresh_crea_states();

-- +goose Down
DROP FUNCTION IF EXISTS pesa_refresh_crea_states();
