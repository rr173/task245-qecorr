package service_test

import (
	"testing"

	"task245-qecorr/internal/model"
	"task245-qecorr/internal/service"
	"task245-qecorr/internal/store"
)

func TestBug02DuplicateSyndromePreservesFirstEvidence(t *testing.T) {
	db, err := store.OpenStore(":memory:"); if err != nil { t.Fatal(err) }; defer db.Close(); svc := service.New(db)
	lat, _ := svc.CreateLattice("surface", 3); q, _ := svc.AddQubit(lat.ID, "q", 0, 0); r, _ := svc.OpenRound(lat.ID, "reader", 1)
	first, err := svc.Ingest(r.ID, q.ID, "x", 1); if err != nil { t.Fatal(err) }
	dup, err := svc.Ingest(r.ID, q.ID, "x", 0); if err != nil { t.Fatal(err) }
	if dup.Status != model.SynDuplicate || dup.RawValue != first.RawValue || dup.RawValue != 1 { t.Fatalf("duplicate=%#v first=%#v", dup, first) }
	items, err := svc.ListSyndromes(r.ID); if err != nil { t.Fatal(err) }; if len(items) != 1 || items[0].RawValue != 1 { t.Fatalf("stored=%#v", items) }
}
