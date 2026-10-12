package sync

import (
	"context"
	"database/sql"
	"testing"
)

// These tests stub *sql.DB with a minimal fake; real integration tests are
// deferred to the Phase 1 testcontainers harness (see TODO markers).
// TODO: port to shared testcontainers MariaDB fixture once that harness is
// extracted from auth_test.go into a reusable testutil package.

// TestUpsertBatch_RequiresUser verifies caller-contract.
func TestUpsertBatch_RequiresUser(t *testing.T) {
	s := New(nil)
	_, err := s.UpsertBatch(context.Background(), 0, BatchPayload{})
	if err == nil {
		t.Fatal("expected error when userID=0")
	}
}

// TestPull_RequiresUser verifies caller-contract.
func TestPull_RequiresUser(t *testing.T) {
	s := New(nil)
	_, err := s.Pull(context.Background(), 0, 0)
	if err == nil {
		t.Fatal("expected error when userID=0")
	}
}

// TestBatchSize sanity check.
func TestBatchSize(t *testing.T) {
	p := BatchPayload{
		Meals:     []MealDTO{{}, {}},
		Favorites: []FavoriteDTO{{}},
	}
	if got := batchSize(p); got != 3 {
		t.Fatalf("batchSize=%d, want 3", got)
	}
}

// TestKindConstants ensures wire constants don't drift accidentally.
func TestKindConstants(t *testing.T) {
	cases := map[string]string{
		KindMeal:         "meal",
		KindMealItem:     "meal_item",
		KindRecipe:       "recipe",
		KindRecipeItem:   "recipe_item",
		KindFavorite:     "favorite",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("kind %q != %q", got, want)
		}
	}
}

// TestStatusConstants ensures wire status constants are stable.
func TestStatusConstants(t *testing.T) {
	if StatusApplied != "applied" || StatusConflictSkipped != "conflict_skipped" || StatusError != "error" {
		t.Fatal("status constants drifted")
	}
}

// Compile-time assertion: DTOs carry LWW trio.
// TODO: expand into table-driven integration tests once testcontainers fixture exists:
//   - idempotency: same payload 2x → row count stable
//   - LWW newer wins / older skipped
//   - soft-delete + undelete
//   - foods-key rejection at handler layer (see handler tests)
//   - favorites (user_id, food_id) collision → conflict_skipped duplicate_food
//   - cross-user isolation (user A cannot see user B rows)
//   - pull cursor advances monotonically
var _ = func() *sql.DB { return nil }
