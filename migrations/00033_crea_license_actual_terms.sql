-- +goose Up
-- The CREA licence was recorded as CC BY-NC-SA 4.0 at project start, an
-- assumption. What CREA actually publishes (checked 2026-10-11):
--  * alimentinutrizione.it/tabelle-nutrizionali: data and texts may not be
--    reproduced "senza una chiara indicazione della fonte originale"; the
--    reference to give is "CREA Centro di ricerca Alimenti e Nutrizione" with
--    crea.gov.it/alimenti-e-nutrizione and alimentinutrizione.it;
--  * crea.gov.it/note-legali: data published under CC BY 3.0 unless stated
--    otherwise.
-- So: reuse with attribution, no non-commercial clause. The attribution now
-- uses the exact reference CREA asks for.
UPDATE food_sources SET
  name             = 'CREA — Tabelle di composizione degli alimenti',
  license          = 'Riuso con citazione della fonte (CC BY 3.0)',
  license_url      = 'https://www.crea.gov.it/note-legali',
  attribution_text = 'CREA Centro di ricerca Alimenti e Nutrizione — https://www.crea.gov.it/alimenti-e-nutrizione e https://www.alimentinutrizione.it',
  notes            = 'Condizioni: riproduzione consentita con chiara indicazione della fonte originale (avviso su alimentinutrizione.it/tabelle-nutrizionali); dati CREA in CC BY 3.0 salvo diversa indicazione (crea.gov.it/note-legali). Immagini non fornite da CREA.'
 WHERE code = 'crea';

-- +goose Down
UPDATE food_sources SET
  name             = 'CREA — Tabelle di Composizione degli Alimenti',
  license          = 'CC BY-NC-SA 4.0',
  license_url      = 'https://creativecommons.org/licenses/by-nc-sa/4.0/',
  attribution_text = 'CREA — Consiglio per la ricerca in agricoltura e l''analisi dell''economia agraria',
  notes            = 'Immagini non ospitate da Pesa: se presente, link esterno alla fonte originale.'
 WHERE code = 'crea';
