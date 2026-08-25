package service_test

import (
	"testing"
	"task245-qecorr/internal/service"
	"task245-qecorr/internal/store"
)

func TestBug10SuspectedMeasurementIsExcludedFromChains(t *testing.T) {
	db, err := store.OpenStore(":memory:"); if err != nil { t.Fatal(err) }; defer db.Close(); svc := service.New(db)
	lat, _ := svc.CreateLattice("surface", 3); q, _ := svc.AddQubit(lat.ID, "q", 0, 0); r, _ := svc.OpenRound(lat.ID, "reader", 1)
	if _, err := svc.Calibrate(r.ID, "reader", "drift", "bad"); err != nil { t.Fatal(err) }
	syn, err := svc.Ingest(r.ID, q.ID, "x", 1); if err != nil { t.Fatal(err) }; if syn.Status == "valid" { t.Fatal("test setup did not mark suspect") }
	if err := svc.CloseRound(r.ID); err != nil { t.Fatal(err) }; if _, err := svc.Analyze(lat.ID); err != nil { t.Fatal(err) }
	chains, err := svc.ListChains(lat.ID); if err != nil { t.Fatal(err) }; edges, err := svc.ListEdges(lat.ID); if err != nil { t.Fatal(err) }; if len(chains) != 0 || len(edges) != 0 { t.Fatalf("chains=%#v edges=%#v", chains, edges) }
}
