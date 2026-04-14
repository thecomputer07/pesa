-- +goose Up
CREATE TABLE food_sources (
  id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  code             VARCHAR(32) NOT NULL,
  name             VARCHAR(128) NOT NULL,
  license          VARCHAR(64) NOT NULL,
  license_url      VARCHAR(512) NULL,
  attribution_text TEXT NOT NULL,
  source_url       VARCHAR(512) NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT uniq_food_sources_code UNIQUE (code)
);

INSERT INTO food_sources (code, name, license, license_url, attribution_text, source_url) VALUES
  ('crea',    'CREA — Tabelle di Composizione degli Alimenti',
              'CC BY-NC-SA 4.0', 'https://creativecommons.org/licenses/by-nc-sa/4.0/',
              'Fonte: CREA — Centro di Ricerca Alimenti e Nutrizione',
              'https://www.alimentinutrizione.it/'),
  ('bda-ieo', 'BDA IEO — Banca Dati di composizione degli Alimenti',
              'citazionale-non-commerciale', NULL,
              'Fonte: BDA IEO — Istituto Europeo di Oncologia, Milano',
              'http://www.bda.ieo.it/'),
  ('manual',  'Inserimento manuale Pesa',
              'CC0', 'https://creativecommons.org/publicdomain/zero/1.0/',
              'Inserimento manuale del titolare del progetto Pesa',
              NULL);

-- +goose Down
DROP TABLE food_sources;
