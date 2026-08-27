package store

import (
	"database/sql"
	"fmt"

	"task245-qecorr/internal/model"
)

// CreateSnapshot 创建一条解码快照（草稿）。
func (s *Store) CreateSnapshot(latticeID string, baselineRound int, evidenceJSON string) (*model.DecodingSnapshot, error) {
	lat, err := s.GetLattice(latticeID)
	if err != nil {
		return nil, err
	}
	if lat.Status == model.LatticeSealed {
		return nil, fmt.Errorf("%w: lattice sealed", model.ErrSealed)
	}
	if evidenceJSON == "" || baselineRound < 0 {
		return nil, fmt.Errorf("%w: invalid snapshot evidence", model.ErrBadRequest)
	}
	snap := &model.DecodingSnapshot{
		ID:            newID("snap"),
		LatticeID:     latticeID,
		Status:        model.SnapDraft,
		BaselineRound: baselineRound,
		EvidenceJSON:  evidenceJSON,
		CreatedAt:     nowUTC(),
	}
	const q = `INSERT INTO snapshots(id, lattice_id, status, baseline_round, evidence_json, created_at) VALUES(?,?,?,?,?,?)`
	if _, err := s.DB.Exec(q, snap.ID, snap.LatticeID, string(snap.Status), snap.BaselineRound, snap.EvidenceJSON, snap.CreatedAt); err != nil {
		return nil, fmt.Errorf("insert snapshot: %w", err)
	}
	return snap, nil
}

// PublishSnapshot is a compare-and-set transition so two callers cannot both
// publish the same draft.
func (s *Store) PublishSnapshot(id string) error {
	const q = `UPDATE snapshots SET status=? WHERE id=? AND status=?`
	res, err := s.DB.Exec(q, string(model.SnapPublished), id, string(model.SnapDraft))
	if err != nil {
		return fmt.Errorf("publish snapshot: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		if _, err := s.GetSnapshot(id); err != nil {
			return err
		}
		return fmt.Errorf("%w: snapshot is not draft", model.ErrInvalidState)
	}
	return nil
}

// GetSnapshot 按 ID 获取快照。
func (s *Store) GetSnapshot(id string) (*model.DecodingSnapshot, error) {
	const q = `SELECT id, lattice_id, status, baseline_round, evidence_json, created_at FROM snapshots WHERE id=?`
	row := s.DB.QueryRow(q, id)
	var snap model.DecodingSnapshot
	var st string
	if err := row.Scan(&snap.ID, &snap.LatticeID, &st, &snap.BaselineRound, &snap.EvidenceJSON, &snap.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("scan snapshot: %w", err)
	}
	snap.Status = model.SnapshotStatus(st)
	return &snap, nil
}

// ListSnapshots 列出某码格的快照（按创建时间降序）。
func (s *Store) ListSnapshots(latticeID string) ([]*model.DecodingSnapshot, error) {
	const q = `SELECT id, lattice_id, status, baseline_round, evidence_json, created_at FROM snapshots WHERE lattice_id=? ORDER BY created_at DESC, id`
	rows, err := s.DB.Query(q, latticeID)
	if err != nil {
		return nil, fmt.Errorf("query snapshots: %w", err)
	}
	defer rows.Close()
	var out []*model.DecodingSnapshot
	for rows.Next() {
		var snap model.DecodingSnapshot
		var st string
		if err := rows.Scan(&snap.ID, &snap.LatticeID, &st, &snap.BaselineRound, &snap.EvidenceJSON, &snap.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan snapshot: %w", err)
		}
		snap.Status = model.SnapshotStatus(st)
		out = append(out, &snap)
	}
	return out, rows.Err()
}

// UpdateSnapshotStatus 更新快照状态。
func (s *Store) UpdateSnapshotStatus(id string, st model.SnapshotStatus) error {
	const q = `UPDATE snapshots SET status=? WHERE id=?`
	res, err := s.DB.Exec(q, string(st), id)
	if err != nil {
		return fmt.Errorf("update snapshot status: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}
