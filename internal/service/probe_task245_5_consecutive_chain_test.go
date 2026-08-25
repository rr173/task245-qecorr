package service_test

import (
	"testing"
	"task245-qecorr/internal/model"
	"task245-qecorr/internal/service"
	"task245-qecorr/internal/store"
)

func TestBug05ChainPersistenceRequiresConsecutiveRounds(t *testing.T) {
	db, err := store.OpenStore(":memory:"); if err != nil { t.Fatal(err) }; defer db.Close(); svc := service.New(db)
	lat, _ := svc.CreateLattice("surface", 3); q, _ := svc.AddQubit(lat.ID, "q", 0, 0)
	for _, no := range []int{1, 3, 5} { r, e := svc.OpenRound(lat.ID, "reader", no); if e != nil { t.Fatal(e) }; if _, e = svc.Ingest(r.ID, q.ID, "x", 1); e != nil { t.Fatal(e) }; if e = svc.CloseRound(r.ID); e != nil { t.Fatal(e) } }
	if _, err := db.CreateDefectEdge(lat.ID, 1, q.ID, 3, q.ID, 1.25); err != nil { t.Fatal(err) }
	if _, err := db.CreateDefectEdge(lat.ID, 3, q.ID, 5, q.ID, 1.25); err != nil { t.Fatal(err) }
	if _, err := svc.RebuildChains(lat.ID); err != nil { t.Fatal(err) }
	chains, err := svc.ListChains(lat.ID); if err != nil { t.Fatal(err) }; if len(chains) != 1 || chains[0].Status != model.ChainTransient { t.Fatalf("chains=%#v", chains) }
}
