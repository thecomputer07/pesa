-- +goose Up
-- A recipe is also a food: one private custom food per recipe (same owner),
-- values per 100 g computed from the ingredients, kept in step on every sync.
-- Search, the detail calculator, favourites and the diary food picker then
-- work for recipes with no special case.
ALTER TABLE foods
  ADD COLUMN IF NOT EXISTS recipe_id BIGINT NULL
    REFERENCES recipes(id) ON DELETE SET NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_foods_recipe ON foods (recipe_id) WHERE recipe_id IS NOT NULL;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION pesa_refresh_recipe_food(rid BIGINT) RETURNS void
LANGUAGE plpgsql AS $$
DECLARE
  r          recipes%ROWTYPE;
  manual_id  BIGINT;
  sum_g      NUMERIC;
  t_kcal     NUMERIC; t_carb NUMERIC; t_prot NUMERIC;
  t_fat      NUMERIC; t_sugar NUMERIC; t_fiber NUMERIC;
  g          NUMERIC;
  alive      BOOLEAN := false;
BEGIN
  SELECT * INTO r FROM recipes WHERE id = rid;
  alive := FOUND AND r.deleted_at_ms IS NULL;
  IF alive THEN
    -- snapshot_json holds each ingredient's values per 100 g, frozen at save.
    SELECT sum(i.grams),
           sum(i.grams * (i.snapshot_json->>'kcal')::numeric      / 100),
           sum(i.grams * (i.snapshot_json->>'carb_g')::numeric    / 100),
           sum(i.grams * (i.snapshot_json->>'protein_g')::numeric / 100),
           sum(i.grams * (i.snapshot_json->>'fat_g')::numeric     / 100),
           sum(i.grams * (i.snapshot_json->>'sugar_g')::numeric   / 100),
           sum(i.grams * (i.snapshot_json->>'fiber_g')::numeric   / 100)
      INTO sum_g, t_kcal, t_carb, t_prot, t_fat, t_sugar, t_fiber
      FROM recipe_items i
     WHERE i.recipe_id = rid AND i.deleted_at_ms IS NULL;
    -- Cooked weight if the user gave it (water lost or gained), else the
    -- sum of the ingredients.
    g := coalesce(nullif(r.total_grams, 0), sum_g);
  END IF;

  IF NOT alive OR g IS NULL OR g <= 0 THEN
    -- Recipe gone or empty: drop its food, unless the diary or another
    -- recipe still points at it (then it stays as it was).
    DELETE FROM foods f
     WHERE f.recipe_id = rid
       AND NOT EXISTS (SELECT 1 FROM meal_items m WHERE m.food_id = f.id)
       AND NOT EXISTS (SELECT 1 FROM recipe_items ri WHERE ri.food_id = f.id);
    RETURN;
  END IF;

  SELECT id INTO manual_id FROM food_sources WHERE code = 'manual';
  INSERT INTO foods (source_id, user_id, recipe_id, name_it, category, state, edible_pct,
                     kcal, carb_g, protein_g, fat_g, sugar_g, fiber_g)
  VALUES (manual_id, r.user_id, rid, r.name, 'Ricette', 'raw', 100,
          round(t_kcal * 100 / g, 2), round(t_carb * 100 / g, 2), round(t_prot * 100 / g, 2),
          round(t_fat * 100 / g, 2), round(t_sugar * 100 / g, 2), round(t_fiber * 100 / g, 2))
  ON CONFLICT (recipe_id) WHERE recipe_id IS NOT NULL DO UPDATE SET
    name_it   = EXCLUDED.name_it,
    kcal      = EXCLUDED.kcal,
    carb_g    = EXCLUDED.carb_g,
    protein_g = EXCLUDED.protein_g,
    fat_g     = EXCLUDED.fat_g,
    sugar_g   = EXCLUDED.sugar_g,
    fiber_g   = EXCLUDED.fiber_g;
END $$;
-- +goose StatementEnd

-- Existing recipes become foods too.
SELECT pesa_refresh_recipe_food(id) FROM recipes;

-- +goose Down
DROP FUNCTION IF EXISTS pesa_refresh_recipe_food(BIGINT);
DELETE FROM foods f
 WHERE f.recipe_id IS NOT NULL
   AND NOT EXISTS (SELECT 1 FROM meal_items m WHERE m.food_id = f.id)
   AND NOT EXISTS (SELECT 1 FROM recipe_items ri WHERE ri.food_id = f.id);
DROP INDEX IF EXISTS uq_foods_recipe;
ALTER TABLE foods DROP COLUMN IF EXISTS recipe_id;
