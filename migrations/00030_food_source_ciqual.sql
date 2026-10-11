-- +goose Up
-- Ciqual (ANSES) copre cio' che CREA non ha: piatti composti (sushi, pizze,
-- panini, zuppe), dolci confezionati, cereali da colazione, gelati, salse.
-- Licenza Ouverte / Etalab 2.0: riuso libero anche commerciale citando la
-- fonte e la data dell'ultimo aggiornamento. I nomi sono tradotti da Pesa:
-- va detto, perche' e' una modifica del dato.
INSERT INTO food_sources (code, name, license, license_url, attribution_text, source_url, scope, contact, notes)
VALUES
  ('ciqual',
   'Ciqual — Tabella di composizione nutrizionale degli alimenti (ANSES)',
   'Licence Ouverte / Etalab 2.0',
   'https://www.etalab.gouv.fr/licence-ouverte-open-licence/',
   'Anses. 2025. Table de composition nutritionnelle des aliments Ciqual (aggiornamento 2025-11-19)',
   'https://ciqual.anses.fr/',
   'data',
   NULL,
   'Nomi degli alimenti tradotti in italiano da Pesa; valori nutrizionali invariati. Solo piatti composti, dolci, gelati e salse.')
ON CONFLICT (code) DO UPDATE SET
  name             = EXCLUDED.name,
  license          = EXCLUDED.license,
  license_url      = EXCLUDED.license_url,
  attribution_text = EXCLUDED.attribution_text,
  source_url       = EXCLUDED.source_url,
  scope            = EXCLUDED.scope,
  contact          = EXCLUDED.contact,
  notes            = EXCLUDED.notes;

-- +goose Down
DELETE FROM food_sources s
 WHERE s.code = 'ciqual'
   AND NOT EXISTS (SELECT 1 FROM foods f WHERE f.source_id = s.id);
