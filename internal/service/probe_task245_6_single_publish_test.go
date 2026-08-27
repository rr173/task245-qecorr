package service_test

import (
	"errors"
	"sync"
	"testing"
	"task245-qecorr/internal/model"
	"task245-qecorr/internal/service"
	"task245-qecorr/internal/store"
)

func TestBug06ConcurrentPublishIsSingleUse(t *testing.T) {
	db, err := store.OpenStore(":memory:"); if err != nil { t.Fatal(err) }; defer db.Close(); svc := service.New(db)
	lat, _ := svc.CreateLattice("surface", 3); snap, err := svc.DraftSnapshot(lat.ID, 0); if err != nil { t.Fatal(err) }
	const participants = 20; start := make(chan struct{}); var wg sync.WaitGroup; wg.Add(participants); var mu sync.Mutex; success, invalid, other := 0, 0, 0
	for i := 0; i < participants; i++ { go func() { defer wg.Done(); <-start; e := svc.PublishSnapshot(snap.ID); mu.Lock(); defer mu.Unlock(); if e == nil { success++ } else if errors.Is(e, model.ErrInvalidState) { invalid++ } else { other++ } }() }
	close(start); wg.Wait()
	if success != 1 || invalid != participants-1 || other != 0 { t.Fatalf("success=%d invalid=%d other=%d", success, invalid, other) }
	got, err := svc.GetSnapshot(snap.ID); if err != nil { t.Fatal(err) }; if got.Status != model.SnapPublished { t.Fatalf("snapshot=%#v", got) }
}
