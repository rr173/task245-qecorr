package service_test

import (
	"testing"
	"task245-qecorr/internal/model"
	"task245-qecorr/internal/service"
	"task245-qecorr/internal/store"
)

func TestBug03CalibrationAnomalyIsDeviceScoped(t *testing.T) {
	db, err := store.OpenStore(":memory:"); if err != nil { t.Fatal(err) }; defer db.Close(); svc := service.New(db)
	lat, _ := svc.CreateLattice("surface", 3); q, _ := svc.AddQubit(lat.ID, "q", 0, 0); r, _ := svc.OpenRound(lat.ID, "reader-b", 1)
	if _, err := svc.Calibrate(r.ID, "reader-a", "drift", "bad calibration"); err != nil { t.Fatal(err) }
	syn, err := svc.Ingest(r.ID, q.ID, "x", 1); if err != nil { t.Fatal(err) }
	if syn.Status != model.SynValid { t.Fatalf("unrelated device marked measurement bad: %#v", syn) }
}
