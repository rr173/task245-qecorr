package service_test

import (
	"testing"
	"task245-qecorr/internal/model"
	"task245-qecorr/internal/service"
	"task245-qecorr/internal/store"
)

func TestBug04NoTemporalEdgeAcrossMissingRound(t *testing.T) {
	db, err := store.OpenStore(":memory:"); if err != nil { t.Fatal(err) }; defer db.Close(); svc := service.New(db)
	lat, _ := svc.CreateLattice("surface", 3); q, _ := svc.AddQubit(lat.ID, "q", 0, 0)
	for _, no := range []int{1, 3} { r, e := svc.OpenRound(lat.ID, "reader", no); if e != nil { t.Fatal(e) }; if _, e = svc.Ingest(r.ID, q.ID, "x", 1); e != nil { t.Fatal(e) }; if e = svc.CloseRound(r.ID); e != nil { t.Fatal(e) } }
	if _, err := svc.Analyze(lat.ID); err != nil { t.Fatal(err) }
	edges, err := svc.ListEdges(lat.ID); if err != nil { t.Fatal(err) }; if len(edges) != 0 { t.Fatalf("unexpected gap edge=%#v", edges) }
	chains, err := svc.ListChains(lat.ID); if err != nil { t.Fatal(err) }; if len(chains) != 2 { t.Fatalf("chains=%#v", chains) }; for _, c := range chains { if c.Status != model.ChainTransient { t.Fatalf("chain=%#v", c) } }
}
