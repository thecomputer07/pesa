-- Dev fixture: ~30 alimenti italiani comuni per testing locale.
-- NON usare in produzione (i valori sono indicativi, non da fonte certificata).
-- Esecuzione: mariadb -upesa -ppesa --skip-ssl pesa < scripts/seed-foods-dev.sql

-- source_id 1 = manual (seeded da migration 00007 / 00019)
-- Usiamo manual per chiarezza che è dev fixture, non vero CREA/BDA.

INSERT INTO foods
  (source_id, name_it, name_en, category, state, edible_pct,
   kcal, protein_g, fat_g, carb_g, sugar_g, fiber_g, water_g, salt_g,
   last_imported_at)
VALUES
  -- Cereali e pasta
  (1, 'Pasta di semola di grano duro',  'Durum wheat pasta', 'Cereali',   'raw',    100, 357, 12.5, 1.4,  73.6, 3.2, 3.0, 10.5, 0.01, NOW()),
  (1, 'Pasta di semola, cotta',         'Durum wheat pasta, cooked', 'Cereali', 'cooked', 100, 158, 5.5,  0.6,  31.0, 1.4, 1.3, 61.5, 0.01, NOW()),
  (1, 'Riso bianco brillato',            'White rice', 'Cereali',          'raw',    100, 332, 6.7,  0.4,  80.4, 0.2, 1.0, 12.0, 0.01, NOW()),
  (1, 'Riso bianco, bollito',            'White rice, boiled', 'Cereali',  'cooked', 100, 130, 2.4,  0.2,  29.0, 0.1, 0.4, 68.0, 0.01, NOW()),
  (1, 'Pane comune di frumento',         'Common wheat bread', 'Cereali',  'raw',    100, 289, 8.6,  0.5,  63.5, 2.1, 3.8, 23.5, 1.2,  NOW()),
  (1, 'Pane integrale',                  'Whole-wheat bread', 'Cereali',   'raw',    100, 224, 7.5,  1.3,  48.5, 2.5, 6.5, 36.0, 1.0,  NOW()),
  (1, 'Pizza margherita',                'Margherita pizza', 'Cereali',    'cooked', 100, 271, 11.4, 11.0, 33.0, 3.5, 2.2, 41.0, 1.3,  NOW()),
  (1, 'Focaccia genovese',               'Focaccia',         'Cereali',    'cooked', 100, 290, 7.0,  10.5, 42.0, 1.8, 2.0, 36.0, 1.5,  NOW()),
  -- Frutta
  (1, 'Mela',                            'Apple',            'Frutta',     'raw',    85,  52,  0.3,  0.2,  14.0, 10.4,2.4, 85.0, 0.0,  NOW()),
  (1, 'Pera',                            'Pear',             'Frutta',     'raw',    90,  57,  0.4,  0.1,  15.2, 9.8, 3.1, 84.0, 0.0,  NOW()),
  (1, 'Banana',                          'Banana',           'Frutta',     'raw',    65,  89,  1.1,  0.3,  22.8, 12.2,2.6, 74.5, 0.0,  NOW()),
  (1, 'Arancia',                         'Orange',           'Frutta',     'raw',    72,  47,  0.9,  0.1,  11.7, 9.4, 2.4, 86.7, 0.0,  NOW()),
  (1, 'Fragole',                         'Strawberries',     'Frutta',     'raw',    97,  32,  0.7,  0.3,  7.7,  4.9, 2.0, 90.9, 0.0,  NOW()),
  -- Verdura
  (1, 'Pomodoro',                        'Tomato',           'Verdura',    'raw',    100, 18,  0.9,  0.2,  3.9,  2.6, 1.2, 94.5, 0.0,  NOW()),
  (1, 'Zucchine',                        'Zucchini',         'Verdura',    'raw',    100, 17,  1.2,  0.3,  3.1,  2.5, 1.0, 94.8, 0.0,  NOW()),
  (1, 'Insalata mista',                  'Mixed salad',      'Verdura',    'raw',    100, 16,  1.4,  0.2,  2.9,  1.4, 1.5, 94.5, 0.0,  NOW()),
  -- Latticini
  (1, 'Latte vaccino intero',            'Whole cow milk',   'Latticini',  'raw',    100, 64,  3.3,  3.6,  4.9,  4.9, 0.0, 87.5, 0.1,  NOW()),
  (1, 'Latte vaccino parzialmente scremato','Semi-skimmed cow milk','Latticini','raw',100,46, 3.5,  1.5,  5.0,  5.0, 0.0, 89.5, 0.1,  NOW()),
  (1, 'Yogurt intero',                   'Whole yogurt',     'Latticini',  'raw',    100, 66,  3.8,  3.9,  4.3,  4.3, 0.0, 87.7, 0.1,  NOW()),
  (1, 'Mozzarella di bufala campana',    'Buffalo mozzarella','Latticini', 'raw',    100, 288, 16.7, 24.4, 0.4,  0.4, 0.0, 55.5, 1.0,  NOW()),
  (1, 'Parmigiano Reggiano',             'Parmigiano Reggiano','Latticini','raw',    100, 392, 33.0, 28.0, 0.0,  0.0, 0.0, 30.8, 1.6,  NOW()),
  (1, 'Grana Padano',                    'Grana Padano',     'Latticini',  'raw',    100, 384, 32.7, 27.5, 0.0,  0.0, 0.0, 32.0, 1.7,  NOW()),
  -- Salumi e carni
  (1, 'Prosciutto crudo di Parma',       'Parma ham',        'Salumi',     'raw',    100, 224, 25.9, 13.6, 0.0,  0.0, 0.0, 56.0, 4.6,  NOW()),
  (1, 'Bresaola della Valtellina',       'Bresaola',         'Salumi',     'raw',    100, 151, 32.0, 2.6,  0.0,  0.0, 0.0, 60.0, 4.4,  NOW()),
  (1, 'Petto di pollo',                  'Chicken breast',   'Carni',      'cooked', 100, 165, 31.0, 3.6,  0.0,  0.0, 0.0, 65.0, 0.1,  NOW()),
  -- Pesce
  (1, 'Salmone, filetto',                'Salmon fillet',    'Pesce',      'raw',    100, 185, 19.9, 12.0, 0.0,  0.0, 0.0, 67.0, 0.1,  NOW()),
  (1, 'Tonno al naturale, sgocciolato',  'Tuna in water',    'Pesce',      'cooked', 100, 116, 26.0, 1.0,  0.0,  0.0, 0.0, 70.0, 0.5,  NOW()),
  -- Legumi
  (1, 'Fagioli borlotti, secchi',        'Borlotti beans, dried','Legumi', 'dried',  100, 291, 23.1, 2.0,  47.5, 4.5, 17.3, 10.5, 0.01, NOW()),
  (1, 'Ceci, secchi',                    'Chickpeas, dried', 'Legumi',     'dried',  100, 316, 20.9, 6.3,  46.9, 5.0, 13.6, 9.0,  0.02, NOW()),
  (1, 'Lenticchie, secche',              'Lentils, dried',   'Legumi',     'dried',  100, 291, 22.7, 1.0,  51.1, 1.8, 13.8, 11.2, 0.02, NOW()),
  -- Olio e condimenti
  (1, 'Olio extravergine di oliva',      'Extra virgin olive oil','Condimenti','raw',100, 884, 0.0,  100.0, 0.0, 0.0, 0.0, 0.0,  0.0,  NOW()),
  -- Dolci
  (1, 'Nutella',                         'Nutella',          'Dolci',      'raw',    100, 539, 6.3,  30.9, 57.5, 56.3,3.4, 1.0,  0.1,  NOW()),
  (1, 'Tiramisù',                        'Tiramisu',         'Dolci',      'cooked', 100, 305, 5.7,  18.6, 30.0, 22.0,0.5, 43.0, 0.2,  NOW());

SELECT COUNT(*) AS total_foods FROM foods;
