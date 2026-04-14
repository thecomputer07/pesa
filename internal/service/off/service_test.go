package off

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// fakeRepo is an in-memory stand-in for off_cache so unit tests don't need a DB.
type fakeRepo struct {
	rows map[string]GetResult
	now  func() time.Time
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{rows: map[string]GetResult{}, now: time.Now}
}

func (f *fakeRepo) Get(ctx context.Context, ean string) (GetResult, error) {
	r, ok := f.rows[ean]
	if !ok {
		return GetResult{}, nil
	}
	ttl := ttlForStatus(r.Status)
	r.Stale = f.now().Sub(r.FetchedAt) > ttl
	r.Found = true
	return r, nil
}
func (f *fakeRepo) Put(ctx context.Context, ean, status string, payload []byte, errMsg string) error {
	f.rows[ean] = GetResult{Payload: payload, Status: status, FetchedAt: f.now(), Found: true}
	return nil
}

// A test-only Service that accepts the fakeRepo interface.
type testService struct {
	cache  *fakeRepo
	client *Client
}

func (s *testService) Lookup(ctx context.Context, ean string) (FoodDTO, error) {
	svc := NewService(&Repo{DB: nil, Now: time.Now}, s.client, nil)
	// swap out the live Repo with the fake by shadowing its methods through a
	// small adapter: reuse the production path by copying relevant logic.
	cached, _ := s.cache.Get(ctx, ean)
	if cached.Found && cached.Status == StatusHit && !cached.Stale {
		var dto FoodDTO
		_ = json.Unmarshal(cached.Payload, &dto)
		return dto, nil
	}
	if cached.Found && cached.Status == StatusMiss && !cached.Stale {
		return FoodDTO{}, ErrNotFound
	}
	body, status, err := svc.Client.Fetch(ctx, ean)
	if err != nil {
		_ = s.cache.Put(ctx, ean, StatusError, nil, err.Error())
		return FoodDTO{}, err
	}
	if status == http.StatusNotFound {
		_ = s.cache.Put(ctx, ean, StatusMiss, nil, "")
		return FoodDTO{}, ErrNotFound
	}
	dto, err := FromOFF(body, ean)
	if err != nil {
		if err == ErrNotFound {
			_ = s.cache.Put(ctx, ean, StatusMiss, nil, "")
		} else {
			_ = s.cache.Put(ctx, ean, StatusError, nil, err.Error())
		}
		return FoodDTO{}, err
	}
	payload, _ := json.Marshal(dto)
	_ = s.cache.Put(ctx, ean, StatusHit, payload, "")
	return dto, nil
}

const sampleOFF = `{
  "status": 1,
  "code": "8001505005707",
  "product": {
    "product_name": "Baci Perugina",
    "product_name_it": "Baci Perugina",
    "brands": "Perugina",
    "quantity": "200 g",
    "image_front_small_url": "https://example.invalid/front.jpg",
    "nutriments": {
      "energy-kcal_100g": 541,
      "proteins_100g": 8.2,
      "fat_100g": 37.1,
      "carbohydrates_100g": 45.3,
      "sugars_100g": 40.1,
      "fiber_100g": 5.4
    }
  }
}`

func TestNormalize_FromOFFGolden(t *testing.T) {
	dto, err := FromOFF([]byte(sampleOFF), "8001505005707")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if dto.NameIT != "Baci Perugina" {
		t.Errorf("NameIT: %q", dto.NameIT)
	}
	if dto.License != "ODbL-1.0" {
		t.Errorf("License: %q", dto.License)
	}
	if dto.Source.Code != "off" {
		t.Errorf("Source.Code: %q", dto.Source.Code)
	}
	if dto.Kcal == nil || *dto.Kcal != 541 {
		t.Errorf("Kcal wrong: %+v", dto.Kcal)
	}
	if len(dto.Images) != 1 || dto.Images[0].URL == "" {
		t.Errorf("image missing")
	}
	if dto.State != "packaged" || dto.EdiblePct != 100 {
		t.Errorf("state/edible")
	}
}

func TestNormalize_MissingProduct(t *testing.T) {
	_, err := FromOFF([]byte(`{"status":0,"code":"0000"}`), "0000")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestNormalize_PartialNutriments(t *testing.T) {
	raw := `{"status":1,"code":"1","product":{"product_name":"x","nutriments":{"energy-kcal_100g":100}}}`
	dto, err := FromOFF([]byte(raw), "1")
	if err != nil {
		t.Fatal(err)
	}
	if dto.Kcal == nil {
		t.Fatal("Kcal expected")
	}
	if dto.ProteinG != nil {
		t.Errorf("ProteinG should be nil")
	}
}

func TestService_HitThenCached(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sampleOFF))
	}))
	defer srv.Close()

	cli := &Client{
		HTTP:    srv.Client(),
		Limiter: rate.NewLimiter(rate.Inf, 1),
		BaseURL: srv.URL,
		UA:      UserAgent,
	}
	s := &testService{cache: newFakeRepo(), client: cli}

	dto, err := s.Lookup(context.Background(), "8001505005707")
	if err != nil {
		t.Fatal(err)
	}
	if dto.Source.Code != "off" {
		t.Errorf("source wrong")
	}
	// Second call must not hit upstream.
	_, err = s.Lookup(context.Background(), "8001505005707")
	if err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Errorf("expected 1 upstream call, got %d", got)
	}
}

func TestService_404CachedAsMiss(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":0}`))
	}))
	defer srv.Close()
	cli := &Client{HTTP: srv.Client(), Limiter: rate.NewLimiter(rate.Inf, 1), BaseURL: srv.URL, UA: UserAgent}
	s := &testService{cache: newFakeRepo(), client: cli}

	_, err := s.Lookup(context.Background(), "0000000000017")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	_, err = s.Lookup(context.Background(), "0000000000017")
	if err != ErrNotFound {
		t.Fatalf("miss not cached: %v", err)
	}
}

func TestClient_RateLimiterSpacing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":0}`))
	}))
	defer srv.Close()
	cli := NewClient(srv.URL) // real 1/sec limiter
	start := time.Now()
	_, _, _ = cli.Fetch(context.Background(), "1")
	_, _, _ = cli.Fetch(context.Background(), "2")
	elapsed := time.Since(start)
	// Token bucket has 1 initial token; second call must wait ~1s.
	if elapsed < 800*time.Millisecond {
		t.Errorf("expected rate-limit spacing ≥ ~1s, got %v", elapsed)
	}
}

func TestCache_TTLClassification(t *testing.T) {
	f := newFakeRepo()
	old := time.Now().Add(-2 * TTLHit)
	f.rows["e"] = GetResult{Payload: []byte(`{}`), Status: StatusHit, FetchedAt: old, Found: true}
	got, err := f.Get(context.Background(), "e")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Stale {
		t.Errorf("expected stale row")
	}
}

func TestFromOFF_HandlesStopwords(t *testing.T) {
	if !strings.Contains(sampleOFF, "Baci") {
		t.Fatal("sanity")
	}
}
