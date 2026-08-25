package store

import (
	"database/sql"
	"fmt"

	"task245-qecorr/internal/model"
)

// CreateRound 开启一个新实验轮次。round_no 必须严格递增（拒绝倒退/重复）。
func (s *Store) CreateRound(latticeID, deviceID string, roundNo int) (*model.MeasurementRound, error) {
	lat, err := s.GetLattice(latticeID)
	if err != nil {
		return nil, err
	}
	if lat.Status == model.LatticeSealed {
		return nil, fmt.Errorf("%w: lattice sealed", model.ErrSealed)
	}
	// 校验轮次严格递增
	const maxQ = `SELECT COALESCE(MAX(round_no),0) FROM rounds WHERE lattice_id=?`
	var maxNo int
	if err := s.DB.QueryRow(maxQ, latticeID).Scan(&maxNo); err != nil {
		return nil, fmt.Errorf("query max round: %w", err)
	}
	if roundNo <= maxNo {
		return nil, fmt.Errorf("%w: round_no %d <= last %d", model.ErrRoundRegression, roundNo, maxNo)
	}
	r := &model.MeasurementRound{
		ID:        newID("rnd"),
		LatticeID: latticeID,
		RoundNo:   roundNo,
		DeviceID:  deviceID,
		Status:    model.RoundReceiving,
		CreatedAt: nowUTC(),
	}
	const q = `INSERT INTO rounds(id, lattice_id, round_no, device_id, status, created_at) VALUES(?,?,?,?,?,?)`
	if _, err := s.DB.Exec(q, r.ID, r.LatticeID, r.RoundNo, r.DeviceID, string(r.Status), r.CreatedAt); err != nil {
		return nil, fmt.Errorf("insert round: %w", err)
	}
	return r, nil
}

// GetRound 按 ID 获取轮次。
func (s *Store) GetRound(id string) (*model.MeasurementRound, error) {
	const q = `SELECT id, lattice_id, round_no, device_id, status, created_at FROM rounds WHERE id=?`
	row := s.DB.QueryRow(q, id)
	var r model.MeasurementRound
	var st string
	if err := row.Scan(&r.ID, &r.LatticeID, &r.RoundNo, &r.DeviceID, &st, &r.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("scan round: %w", err)
	}
	r.Status = model.RoundStatus(st)
	return &r, nil
}

// GetRoundByNo 按 (lattice, round_no) 获取轮次。
func (s *Store) GetRoundByNo(latticeID string, roundNo int) (*model.MeasurementRound, error) {
	const q = `SELECT id, lattice_id, round_no, device_id, status, created_at FROM rounds WHERE lattice_id=? AND round_no=?`
	row := s.DB.QueryRow(q, latticeID, roundNo)
	var r model.MeasurementRound
	var st string
	if err := row.Scan(&r.ID, &r.LatticeID, &r.RoundNo, &r.DeviceID, &st, &r.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("scan round: %w", err)
	}
	r.Status = model.RoundStatus(st)
	return &r, nil
}

// ListRounds 列出某码格的全部轮次（按轮次号升序）。
func (s *Store) ListRounds(latticeID string) ([]*model.MeasurementRound, error) {
	const q = `SELECT id, lattice_id, round_no, device_id, status, created_at FROM rounds WHERE lattice_id=? ORDER BY round_no, id`
	rows, err := s.DB.Query(q, latticeID)
	if err != nil {
		return nil, fmt.Errorf("query rounds: %w", err)
	}
	defer rows.Close()
	var out []*model.MeasurementRound
	for rows.Next() {
		var r model.MeasurementRound
		var st string
		if err := rows.Scan(&r.ID, &r.LatticeID, &r.RoundNo, &r.DeviceID, &st, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan round: %w", err)
		}
		r.Status = model.RoundStatus(st)
		out = append(out, &r)
	}
	return out, rows.Err()
}

// UpdateRoundStatus 更新轮次状态（存储层仅做持久化，状态机校验在上层 service）。
func (s *Store) UpdateRoundStatus(id string, st model.RoundStatus) error {
	const q = `UPDATE rounds SET status=? WHERE id=?`
	res, err := s.DB.Exec(q, string(st), id)
	if err != nil {
		return fmt.Errorf("update round status: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}
