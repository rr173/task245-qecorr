package service_test

import (
	"errors"
	"testing"
	"task245-qecorr/internal/model"
	"task245-qecorr/internal/service"
	"task245-qecorr/internal/store"
)

func TestBug08SealedLatticeRejectsNewQubit(t *testing.T) {
	db, err := store.OpenStore(":memory:"); if err != nil { t.Fatal(err) }; defer db.Close(); svc := service.New(db)
	lat, _ := svc.CreateLattice("surface", 3); if err := svc.SealLattice(lat.ID); err != nil { t.Fatal(err) }
	if _, err := svc.AddQubit(lat.ID, "late", 1, 1); !errors.Is(err, model.ErrSealed) { t.Fatalf("expected sealed error, got %v", err) }
	items, err := svc.ListQubits(lat.ID); if err != nil { t.Fatal(err) }; if len(items) != 0 { t.Fatalf("qubits=%#v", items) }
}
