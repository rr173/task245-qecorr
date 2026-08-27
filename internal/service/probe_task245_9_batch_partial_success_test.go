package service_test

import (
	"testing"
	"task245-qecorr/internal/service"
	"task245-qecorr/internal/store"
)

func TestBug09BatchIngestKeepsPositionsAndContinues(t *testing.T) {
	db, err := store.OpenStore(":memory:"); if err != nil { t.Fatal(err) }; defer db.Close(); svc := service.New(db)
	lat, _ := svc.CreateLattice("surface", 3); a, _ := svc.AddQubit(lat.ID, "a", 0, 0); b, _ := svc.AddQubit(lat.ID, "b", 1, 0); r, _ := svc.OpenRound(lat.ID, "reader", 1)
	results := svc.BatchIngest(r.ID, []service.SyndromeInput{{QubitID: a.ID, Stabilizer: "x", RawValue: 1}, {QubitID: "missing", Stabilizer: "x", RawValue: 1}, {QubitID: b.ID, Stabilizer: "x", RawValue: 0}})
	if len(results) != 3 || results[0].Syndrome == nil || results[1].Error == "" || results[2].Syndrome == nil { t.Fatalf("results=%#v", results) }
	items, err := svc.ListSyndromes(r.ID); if err != nil { t.Fatal(err) }; if len(items) != 2 { t.Fatalf("stored=%#v", items) }
}
