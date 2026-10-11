-- +goose Up
-- USDA FoodData Central (FNDDS, survey foods): distinct sushi types (maki,
-- nigiri, uramaki) and Japanese sides, which neither CREA nor Ciqual have.
-- US government work: public domain (CC0 1.0). Carbohydrates are stored as
-- available carbohydrate (by difference minus fibre), like CREA and Ciqual.
INSERT INTO food_sources (code, name, license, license_url, attribution_text, source_url, scope, contact, notes)
VALUES
  ('usda',
   'USDA FoodData Central — FNDDS (Stati Uniti)',
   'CC0 1.0 (pubblico dominio)',
   'https://creativecommons.org/publicdomain/zero/1.0/',
   'U.S. Department of Agriculture, Agricultural Research Service. FoodData Central, FNDDS 2021-2023 (rilascio 2024-10-31)',
   'https://fdc.nal.usda.gov/',
   'data',
   NULL,
   'Nomi tradotti in italiano da Pesa. Carboidrati disponibili = carboidrati per differenza meno fibre; sale = sodio x 2,5. Solo sushi e cucina giapponese.')
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
 WHERE s.code = 'usda'
   AND NOT EXISTS (SELECT 1 FROM foods f WHERE f.source_id = s.id);
