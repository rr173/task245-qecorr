package service_test

import (
	"errors"
	"testing"
	"task245-qecorr/internal/model"
	"task245-qecorr/internal/service"
	"task245-qecorr/internal/store"
)

func TestBug07DraftCannotBeSuperseded(t *testing.T) {
	db, err := store.OpenStore(":memory:"); if err != nil { t.Fatal(err) }; defer db.Close(); svc := service.New(db)
	lat, _ := svc.CreateLattice("surface", 3); snap, err := svc.DraftSnapshot(lat.ID, 0); if err != nil { t.Fatal(err) }
	if err := svc.SupersedeSnapshot(snap.ID); !errors.Is(err, model.ErrInvalidState) { t.Fatalf("expected invalid state, got %v", err) }
	got, err := svc.GetSnapshot(snap.ID); if err != nil { t.Fatal(err) }; if got.Status != model.SnapDraft { t.Fatalf("snapshot=%#v", got) }
}
