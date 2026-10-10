-- +goose Up
-- Coppie crudo/cotto anche quando il crudo CREA si chiama "<X>, crudo"
-- (es. "Riso Venere, crudo" <-> "Riso Venere, cotto, bollito").
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
           substr(c.name_it, length(b.base) + 3) AS suffix,
           r.carb_g AS raw_carb, c.carb_g AS cooked_carb
      FROM foods r
      CROSS JOIN LATERAL (SELECT regexp_replace(r.name_it, ', crud[oaie]$', '', 'i') AS base) b
      JOIN foods c ON c.source_id = r.source_id
                  AND c.state = 'cooked'
                  AND c.superseded_by IS NULL
                  AND lower(c.name_it) LIKE lower(b.base) || ', %'
     WHERE r.source_id = crea_id
       AND r.state IN ('raw', 'dried')
       AND r.superseded_by IS NULL
       AND substr(c.name_it, length(b.base) + 3)
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

SELECT pesa_refresh_crea_states();

-- +goose Down
-- La funzione 00026 resta compatibile: nessun rollback dei dati.
SELECT 1;
