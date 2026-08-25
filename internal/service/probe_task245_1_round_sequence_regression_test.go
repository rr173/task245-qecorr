package service_test

import (
	"errors"
	"sync"
	"testing"

	"task245-qecorr/internal/model"
	"task245-qecorr/internal/service"
	"task245-qecorr/internal/store"
)

func TestBug01ConcurrentRoundSequenceIsUnique(t *testing.T) {
	db, err := store.OpenStore(":memory:"); if err != nil { t.Fatal(err) }; defer db.Close()
	svc := service.New(db); lat, err := svc.CreateLattice("surface", 3); if err != nil { t.Fatal(err) }
	if _, err := svc.OpenRound(lat.ID, "reader", 1); err != nil { t.Fatal(err) }
	const participants = 20
	start := make(chan struct{}); var wg sync.WaitGroup; wg.Add(participants)
	var mu sync.Mutex; successes, regressions, other := 0, 0, 0
	for i := 0; i < participants; i++ { go func() { defer wg.Done(); <-start; _, e := svc.OpenRound(lat.ID, "reader", 1); mu.Lock(); defer mu.Unlock(); if e == nil { successes++ } else if errors.Is(e, model.ErrRoundRegression) { regressions++ } else { other++ } }() }
	close(start); wg.Wait()
	if successes != 0 || regressions != participants || other != 0 { t.Fatalf("success=%d regressions=%d other=%d", successes, regressions, other) }
	rounds, err := svc.ListRounds(lat.ID); if err != nil { t.Fatal(err) }; if len(rounds) != 1 { t.Fatalf("round count=%d", len(rounds)) }
}
