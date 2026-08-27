package snapshot_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"task245-qecorr/internal/model"
	"task245-qecorr/internal/service"
	"task245-qecorr/internal/store"
)

// TestPublishOnceConcurrent reproduces the requirement that only one of many
// concurrent publish requests on the same draft may succeed. Each loser must
// observe the snapshot as already published, and the published state must not
// be overwritten by a straggler.
func TestPublishOnceConcurrent(t *testing.T) {
	dbPath := t.TempDir() + "/qecorr.db"
	st, err := store.OpenStore(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	svc := service.New(st)
	lat, err := svc.CreateLattice("surface-d3", 3)
	if err != nil {
		t.Fatalf("create lattice: %v", err)
	}
	// Draft a snapshot directly (the lattice/round pipeline is incidental to
	// the publish-once guarantee).
	snap, err := st.CreateSnapshot(lat.ID, 0, `{"lattice_id":"x"}`)
	if err != nil {
		t.Fatalf("create snapshot: %v", err)
	}

	const n = 32
	var wg sync.WaitGroup
	var ok, fail int64
	start := make(chan struct{})
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			<-start
			if err := svc.PublishSnapshot(snap.ID); err == nil {
				atomic.AddInt64(&ok, 1)
			} else {
				atomic.AddInt64(&fail, 1)
			}
		}()
	}
	close(start)
	wg.Wait()

	if got := atomic.LoadInt64(&ok); got != 1 {
		t.Fatalf("expected exactly one successful publish, got %d", got)
	}
	if got := atomic.LoadInt64(&fail); got != n-1 {
		t.Fatalf("expected %d failures, got %d", n-1, got)
	}

	got, err := st.GetSnapshot(snap.ID)
	if err != nil {
		t.Fatalf("get snapshot: %v", err)
	}
	if got.Status != model.SnapPublished {
		t.Fatalf("snapshot status = %s, want published", got.Status)
	}
}

// TestPublishPublishedImmutable ensures a draft that is already published
// cannot be republished (which would mutate the immutable evidence).
func TestPublishPublishedImmutable(t *testing.T) {
	dbPath := t.TempDir() + "/qecorr.db"
	st, err := store.OpenStore(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	svc := service.New(st)
	lat, err := svc.CreateLattice("surface-d3", 3)
	if err != nil {
		t.Fatalf("create lattice: %v", err)
	}
	snap, err := st.CreateSnapshot(lat.ID, 0, `{"lattice_id":"x"}`)
	if err != nil {
		t.Fatalf("create snapshot: %v", err)
	}
	if err := svc.PublishSnapshot(snap.ID); err != nil {
		t.Fatalf("first publish: %v", err)
	}
	if err := svc.PublishSnapshot(snap.ID); err == nil {
		t.Fatalf("second publish should fail")
	} else if !errors.Is(err, model.ErrInvalidState) {
		t.Fatalf("second publish error = %v, want ErrInvalidState", err)
	}

	got, err := st.GetSnapshot(snap.ID)
	if err != nil {
		t.Fatalf("get snapshot: %v", err)
	}
	if got.Status != model.SnapPublished {
		t.Fatalf("snapshot status = %s, want published", got.Status)
	}
}
