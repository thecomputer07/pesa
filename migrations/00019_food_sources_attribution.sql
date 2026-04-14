-- +goose Up
ALTER TABLE food_sources
  ADD COLUMN IF NOT EXISTS scope   VARCHAR(32)  NULL,
  ADD COLUMN IF NOT EXISTS contact VARCHAR(255) NULL,
  ADD COLUMN IF NOT EXISTS notes   TEXT         NULL;

INSERT INTO food_sources (code, name, license, license_url, attribution_text, source_url, scope, contact, notes)
VALUES
  ('crea',
   'CREA — Tabelle di Composizione degli Alimenti',
   'CC BY-NC-SA 4.0',
   'https://creativecommons.org/licenses/by-nc-sa/4.0/',
   'CREA — Consiglio per la ricerca in agricoltura e l''analisi dell''economia agraria',
   'https://www.alimentinutrizione.it/',
   'data',
   'alimentinutrizione@crea.gov.it',
   'Immagini non ospitate da Pesa: se presente, link esterno alla fonte originale.'),
  ('bda-ieo',
   'BDA IEO — Banca Dati di composizione degli Alimenti',
   'citazionale-non-commerciale',
   NULL,
   'BDA IEO — Istituto Europeo di Oncologia, Milano',
   'http://www.bda.ieo.it/',
   'data',
   NULL,
   'Immagini non ospitate da Pesa.'),
  ('off',
   'Open Food Facts',
   'ODbL-1.0',
   'https://opendatacommons.org/licenses/odbl/1-0/',
   'Open Food Facts contributors — ODbL',
   'https://world.openfoodfacts.org/',
   'data+images',
   'contact@openfoodfacts.org',
   'Dati e immagini di prodotti confezionati. Le immagini sono servite da Pesa con banner ODbL.'),
  ('wikimedia',
   'Wikimedia Commons',
   'per-file (CC0/CC-BY/CC-BY-SA)',
   'https://commons.wikimedia.org/wiki/Commons:Reusing_content_outside_Wikimedia',
   'Autori Wikimedia Commons — licenze per file',
   'https://commons.wikimedia.org/',
   'images',
   NULL,
   'Credito e link al file Commons mostrati per ogni immagine.')
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
DELETE FROM food_sources WHERE code IN ('off','wikimedia');
ALTER TABLE food_sources
  DROP COLUMN IF EXISTS notes,
  DROP COLUMN IF EXISTS contact,
  DROP COLUMN IF EXISTS scope;
