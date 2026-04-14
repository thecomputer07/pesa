package off

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// ImagePersister is a narrow interface so plan-02's imagecache.Store can
// plug into the OFF lookup without creating an import cycle.
type ImagePersister interface {
	// PersistOFFImage downloads originalURL, stores it content-addressed,
	// and returns the local /static/images URL plus license/attribution.
	// ok=false means the image could not be persisted (do not fail lookup).
	PersistOFFImage(ctx context.Context, ean, originalURL string) (localURL, license, attribution string, ok bool)
}

// Service is the cache-through façade used by the HTTP handler.
type Service struct {
	Cache  *Repo
	Client *Client
	Images ImagePersister // optional; may be nil (image rewrite skipped)

	sfMu   sync.Mutex
	sfCh   map[string]chan struct{}
}

// NewService wires the dependencies. Images may be nil.
func NewService(cache *Repo, client *Client, images ImagePersister) *Service {
	return &Service{
		Cache:  cache,
		Client: client,
		Images: images,
		sfCh:   make(map[string]chan struct{}),
	}
}

// LookupByEAN resolves an EAN to a FoodDTO, caching hits/misses/errors.
// Stale hits are served immediately; a background refresh repopulates the row.
func (s *Service) LookupByEAN(ctx context.Context, ean string) (FoodDTO, error) {
	cached, err := s.Cache.Get(ctx, ean)
	if err == nil && cached.Found {
		switch cached.Status {
		case StatusHit:
			if cached.Stale {
				go s.refresh(context.Background(), ean)
			}
			var dto FoodDTO
			if len(cached.Payload) > 0 {
				if err := json.Unmarshal(cached.Payload, &dto); err == nil {
					return dto, nil
				}
			}
			// fallthrough to re-fetch if payload corrupt
		case StatusMiss:
			if !cached.Stale {
				return FoodDTO{}, ErrNotFound
			}
		case StatusError:
			if !cached.Stale {
				return FoodDTO{}, fmt.Errorf("off: cached upstream error")
			}
		}
	}
	return s.fetchAndStore(ctx, ean)
}

// fetchAndStore is guarded by a lightweight singleflight so concurrent
// Lookups for the same EAN issue a single upstream request.
func (s *Service) fetchAndStore(ctx context.Context, ean string) (FoodDTO, error) {
	s.sfMu.Lock()
	if ch, ok := s.sfCh[ean]; ok {
		s.sfMu.Unlock()
		select {
		case <-ch:
		case <-ctx.Done():
			return FoodDTO{}, ctx.Err()
		}
		// Re-read cache after the leader finished.
		cached, err := s.Cache.Get(ctx, ean)
		if err == nil && cached.Found {
			if cached.Status == StatusHit && len(cached.Payload) > 0 {
				var dto FoodDTO
				if err := json.Unmarshal(cached.Payload, &dto); err == nil {
					return dto, nil
				}
			}
			if cached.Status == StatusMiss {
				return FoodDTO{}, ErrNotFound
			}
		}
		return FoodDTO{}, fmt.Errorf("off: concurrent fetch failed")
	}
	ch := make(chan struct{})
	s.sfCh[ean] = ch
	s.sfMu.Unlock()

	defer func() {
		s.sfMu.Lock()
		delete(s.sfCh, ean)
		s.sfMu.Unlock()
		close(ch)
	}()

	body, status, err := s.Client.Fetch(ctx, ean)
	if err != nil {
		_ = s.Cache.Put(ctx, ean, StatusError, nil, err.Error())
		return FoodDTO{}, err
	}
	if status == http.StatusNotFound {
		_ = s.Cache.Put(ctx, ean, StatusMiss, nil, "")
		return FoodDTO{}, ErrNotFound
	}
	dto, err := FromOFF(body, ean)
	if err != nil {
		if err == ErrNotFound {
			_ = s.Cache.Put(ctx, ean, StatusMiss, nil, "")
			return FoodDTO{}, ErrNotFound
		}
		_ = s.Cache.Put(ctx, ean, StatusError, nil, err.Error())
		return FoodDTO{}, err
	}
	// Image substitution happens once at insert time so payload_json holds the
	// stable /static/images URL.
	if s.Images != nil && len(dto.Images) == 1 && dto.Images[0].IsExternal {
		original := dto.Images[0].URL
		if local, lic, attr, ok := s.Images.PersistOFFImage(ctx, ean, original); ok {
			dto.Images[0] = ImageDTO{
				URL:         local,
				License:     strPtr(lic),
				Attribution: strPtr(attr),
				IsExternal:  false,
			}
		}
	}
	payload, err := json.Marshal(dto)
	if err != nil {
		return dto, err // serve result; don't cache garbage
	}
	_ = s.Cache.Put(ctx, ean, StatusHit, payload, "")
	return dto, nil
}

func (s *Service) refresh(ctx context.Context, ean string) {
	defer func() { _ = recover() }()
	_, _ = s.fetchAndStore(ctx, ean)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
