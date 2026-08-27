package service

import (
	"path/filepath"
	"testing"

	"task245-qecorr/internal/model"
	"task245-qecorr/internal/store"
)

// TestBatchIngest_PartialSuccess reproduces the device batch-upload scenario:
// three measurements are ingested in one call, the middle one references an
// unknown qubit and fails, while the first and third must still be written and
// the response must report per-position success/error.
func TestBatchIngest_PartialSuccess(t *testing.T) {
	dir := t.TempDir()
	s, err := store.OpenStore(filepath.Join(dir, "batch.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	svc := New(s)
	lat, err := svc.CreateLattice("surface-3", 3)
	if err != nil {
		t.Fatalf("create lattice: %v", err)
	}
	a, err := svc.AddQubit(lat.ID, "q-a", 0, 0)
	if err != nil {
		t.Fatalf("add qubit a: %v", err)
	}
	b, err := svc.AddQubit(lat.ID, "q-b", 1, 0)
	if err != nil {
		t.Fatalf("add qubit b: %v", err)
	}
	r, err := svc.OpenRound(lat.ID, "readout-1", 1)
	if err != nil {
		t.Fatalf("open round: %v", err)
	}

	const unknown = "q-does-not-exist"
	results := svc.BatchIngest(r.ID, []SyndromeInput{
		{QubitID: a.ID, Stabilizer: "x", RawValue: 1},
		{QubitID: unknown, Stabilizer: "x", RawValue: 1}, // fails: unknown qubit
		{QubitID: b.ID, Stabilizer: "z", RawValue: 0},
	})

	if len(results) != 3 {
		t.Fatalf("expected 3 per-position results, got %d", len(results))
	}

	// Position 0: written successfully.
	if results[0].Index != 0 {
		t.Errorf("results[0].Index = %d, want 0", results[0].Index)
	}
	if results[0].Syndrome == nil || results[0].Error != "" {
		t.Fatalf("results[0] should succeed, got syndrome=%v error=%q", results[0].Syndrome, results[0].Error)
	}
	if results[0].Syndrome.Status != model.SynValid {
		t.Errorf("results[0] status = %s, want valid", results[0].Syndrome.Status)
	}

	// Position 1: failed with unknown qubit, no syndrome written.
	if results[1].Index != 1 {
		t.Errorf("results[1].Index = %d, want 1", results[1].Index)
	}
	if results[1].Syndrome != nil || results[1].Error == "" {
		t.Fatalf("results[1] should fail, got syndrome=%v error=%q", results[1].Syndrome, results[1].Error)
	}

	// Position 2: still written despite the preceding failure.
	if results[2].Index != 2 {
		t.Errorf("results[2].Index = %d, want 2", results[2].Index)
	}
	if results[2].Syndrome == nil || results[2].Error != "" {
		t.Fatalf("results[2] should succeed (partial success), got syndrome=%v error=%q", results[2].Syndrome, results[2].Error)
	}
	if results[2].Syndrome.QubitID != b.ID {
		t.Errorf("results[2] qubit = %s, want %s", results[2].Syndrome.QubitID, b.ID)
	}

	// The two valid measurements must actually be persisted; the unknown-qubit
	// one must not be.
	stored, err := svc.ListSyndromes(r.ID)
	if err != nil {
		t.Fatalf("list syndromes: %v", err)
	}
	if len(stored) != 2 {
		t.Fatalf("expected 2 persisted syndromes (positions 0 and 2), got %d", len(stored))
	}
	for _, syn := range stored {
		if syn.QubitID == unknown {
			t.Errorf("unknown-qubit measurement should not be persisted: %#v", syn)
		}
	}
}

// TestBatchIngest_MultipleFailures ensures every failing position reports its
// own error and does not collapse the rest of the batch.
func TestBatchIngest_MultipleFailures(t *testing.T) {
	dir := t.TempDir()
	s, err := store.OpenStore(filepath.Join(dir, "batch2.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	svc := New(s)
	lat, err := svc.CreateLattice("surface-3", 3)
	if err != nil {
		t.Fatalf("create lattice: %v", err)
	}
	a, err := svc.AddQubit(lat.ID, "q-a", 0, 0)
	if err != nil {
		t.Fatalf("add qubit a: %v", err)
	}
	r, err := svc.OpenRound(lat.ID, "readout-1", 1)
	if err != nil {
		t.Fatalf("open round: %v", err)
	}

	results := svc.BatchIngest(r.ID, []SyndromeInput{
		{QubitID: "nope-1", Stabilizer: "x", RawValue: 1},
		{QubitID: a.ID, Stabilizer: "x", RawValue: 1},
		{QubitID: "nope-2", Stabilizer: "z", RawValue: 0},
		{QubitID: a.ID, Stabilizer: "", RawValue: 1}, // bad request: empty stabilizer
	})

	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}
	for i, wantErr := range []bool{true, false, true, true} {
		gotErr := results[i].Error != ""
		if gotErr != wantErr {
			t.Errorf("results[%d] error=%v, want error=%v (err=%q)", i, gotErr, wantErr, results[i].Error)
		}
		if !wantErr && results[i].Syndrome == nil {
			t.Errorf("results[%d] expected syndrome, got nil", i)
		}
		if wantErr && results[i].Syndrome != nil {
			t.Errorf("results[%d] expected no syndrome, got %v", i, results[i].Syndrome)
		}
		if results[i].Index != i {
			t.Errorf("results[%d].Index = %d, want %d", i, results[i].Index, i)
		}
	}

	stored, err := svc.ListSyndromes(r.ID)
	if err != nil {
		t.Fatalf("list syndromes: %v", err)
	}
	if len(stored) != 1 {
		t.Fatalf("expected 1 persisted syndrome (position 1), got %d", len(stored))
	}
}
