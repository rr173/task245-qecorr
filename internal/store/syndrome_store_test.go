package store

import (
	"path/filepath"
	"testing"

	"task245-qecorr/internal/model"
)

// TestIngestSyndromeDuplicateKeepsFirstEvidence 断言：当同一 (lattice, round,
// qubit, stabilizer) 被重试摄入时，记录被标记为重复，但首次采集的 raw_value
// 不被后续重试值覆盖（保留首条证据）。
func TestIngestSyndromeDuplicateKeepsFirstEvidence(t *testing.T) {
	s, err := OpenStore(filepath.Join(t.TempDir(), "qecorr_dup.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	lat, err := s.CreateLattice("surface", 3)
	if err != nil {
		t.Fatalf("create lattice: %v", err)
	}
	qb, err := s.CreateQubit(lat.ID, "q0", 0, 0)
	if err != nil {
		t.Fatalf("create qubit: %v", err)
	}

	// 首次摄入：缺陷（raw_value=1）。
	first, err := s.IngestSyndrome("round-1", lat.ID, 1, qb.ID, "X", 1)
	if err != nil {
		t.Fatalf("first ingest: %v", err)
	}
	if first.Status != model.SynValid || first.RawValue != 1 {
		t.Fatalf("first = status=%s raw=%d, want valid/1", first.Status, first.RawValue)
	}

	// 重试：设备上报平凡值 0，不得覆盖首条缺陷证据。
	retry, err := s.IngestSyndrome("round-1", lat.ID, 1, qb.ID, "X", 0)
	if err != nil {
		t.Fatalf("retry ingest: %v", err)
	}
	if retry.Status != model.SynDuplicate {
		t.Fatalf("retry status = %s, want duplicate", retry.Status)
	}
	if retry.RawValue != 1 {
		t.Fatalf("retry raw_value = %d, want 1 (first evidence must be preserved)", retry.RawValue)
	}

	// 落库的记录同样须保留首条证据。
	persisted, err := s.ListSyndromesByRound("round-1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(persisted) != 1 {
		t.Fatalf("want 1 persisted syndrome, got %d", len(persisted))
	}
	if persisted[0].RawValue != 1 || persisted[0].Status != model.SynDuplicate {
		t.Fatalf("persisted = raw=%d status=%s, want raw=1 status=duplicate",
			persisted[0].RawValue, persisted[0].Status)
	}
}
