// Package snapshot creates and publishes immutable decoding evidence.
package snapshot

import (
	"encoding/json"
	"fmt"
	"sort"

	"task245-qecorr/internal/model"
	"task245-qecorr/internal/store"
)

type evidence struct {
	LatticeID string              `json:"lattice_id"`
	Edges     []*model.DefectEdge `json:"defect_edges"`
	Chains    []*model.ErrorChain `json:"error_chains"`
}

// Draft serializes the current persisted graph and chains into a draft.
func Draft(s *store.Store, latticeID string, baselineRound int) (*model.DecodingSnapshot, error) {
	edges, err := s.ListDefectEdges(latticeID)
	if err != nil {
		return nil, err
	}
	chains, err := s.ListChains(latticeID)
	if err != nil {
		return nil, err
	}
	sort.Slice(chains, func(i, j int) bool { return chains[i].ID < chains[j].ID })
	payload, err := json.Marshal(evidence{LatticeID: latticeID, Edges: edges, Chains: chains})
	if err != nil {
		return nil, fmt.Errorf("marshal evidence: %w", err)
	}
	return s.CreateSnapshot(latticeID, baselineRound, string(payload))
}

// Publish atomically changes a draft to published. A published snapshot is
// immutable; callers must create a new draft for a revised decoder result.
func Publish(s *store.Store, id string) error {
	snap, err := s.GetSnapshot(id)
	if err != nil {
		return err
	}
	if snap.Status != model.SnapDraft {
		return fmt.Errorf("%w: snapshot %s is %s", model.ErrInvalidState, id, snap.Status)
	}
	return s.PublishSnapshot(id)
}

// Supersede marks an older published snapshot as superseded after the caller
// has published a replacement. The operation preserves evidence contents.
func Supersede(s *store.Store, oldID string) error {
	snap, err := s.GetSnapshot(oldID)
	if err != nil {
		return err
	}
	if snap.Status == model.SnapSuperseded {
		return fmt.Errorf("%w: snapshot %s is %s", model.ErrInvalidState, oldID, snap.Status)
	}
	return s.UpdateSnapshotStatus(oldID, model.SnapSuperseded)
}

// Get and List expose immutable snapshot records to service and HTTP layers.
func Get(s *store.Store, id string) (*model.DecodingSnapshot, error) { return s.GetSnapshot(id) }
func List(s *store.Store, latticeID string) ([]*model.DecodingSnapshot, error) {
	return s.ListSnapshots(latticeID)
}
