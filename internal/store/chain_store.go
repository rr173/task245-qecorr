package store

import (
	"database/sql"
	"fmt"

	"task245-qecorr/internal/model"
)

// CreateChain 写入一条错误链。involved 为量子比特 ID 列表。
func (s *Store) CreateChain(latticeID string, status model.ChainStatus, firstRound, lastRound int, involved []string, suspectedDevice string, score float64) (*model.ErrorChain, error) {
	c := &model.ErrorChain{
		ID:              newID("chain"),
		LatticeID:       latticeID,
		Status:          status,
		FirstRound:      firstRound,
		LastRound:       lastRound,
		InvolvedQubits:  involved,
		SuspectedDevice: suspectedDevice,
		Score:           score,
		CreatedAt:       nowUTC(),
	}
	qubitsJSON, err := c.QubitsJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal qubits: %w", err)
	}
	const q = `INSERT INTO error_chains(id, lattice_id, status, first_round, last_round, involved_qubits, suspected_device, score, created_at) VALUES(?,?,?,?,?,?,?,?,?)`
	if _, err := s.DB.Exec(q, c.ID, c.LatticeID, string(c.Status), c.FirstRound, c.LastRound, qubitsJSON, c.SuspectedDevice, c.Score, c.CreatedAt); err != nil {
		return nil, fmt.Errorf("insert chain: %w", err)
	}
	return c, nil
}

// GetChain 按 ID 获取错误链（还原 involved_qubits）。
func (s *Store) GetChain(id string) (*model.ErrorChain, error) {
	const q = `SELECT id, lattice_id, status, first_round, last_round, involved_qubits, suspected_device, score, created_at FROM error_chains WHERE id=?`
	row := s.DB.QueryRow(q, id)
	var c model.ErrorChain
	var st, qubitsJSON string
	if err := row.Scan(&c.ID, &c.LatticeID, &st, &c.FirstRound, &c.LastRound, &qubitsJSON, &c.SuspectedDevice, &c.Score, &c.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("scan chain: %w", err)
	}
	c.Status = model.ChainStatus(st)
	c.InvolvedQubits, _ = model.ParseQubitsJSON(qubitsJSON)
	return &c, nil
}

// ListChains 列出某码格的错误链（按跨度降序）。
func (s *Store) ListChains(latticeID string) ([]*model.ErrorChain, error) {
	const q = `SELECT id, lattice_id, status, first_round, last_round, involved_qubits, suspected_device, score, created_at FROM error_chains WHERE lattice_id=? ORDER BY (last_round-first_round) DESC, score DESC, id`
	rows, err := s.DB.Query(q, latticeID)
	if err != nil {
		return nil, fmt.Errorf("query chains: %w", err)
	}
	defer rows.Close()
	var out []*model.ErrorChain
	for rows.Next() {
		var c model.ErrorChain
		var st, qubitsJSON string
		if err := rows.Scan(&c.ID, &c.LatticeID, &st, &c.FirstRound, &c.LastRound, &qubitsJSON, &c.SuspectedDevice, &c.Score, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan chain: %w", err)
		}
		c.Status = model.ChainStatus(st)
		c.InvolvedQubits, _ = model.ParseQubitsJSON(qubitsJSON)
		out = append(out, &c)
	}
	return out, rows.Err()
}

// UpdateChainStatus 更新错误链状态，并可选地改写疑似设备。
func (s *Store) UpdateChainStatus(id string, st model.ChainStatus, suspectedDevice string) error {
	const q = `UPDATE error_chains SET status=?, suspected_device=? WHERE id=?`
	res, err := s.DB.Exec(q, string(st), suspectedDevice, id)
	if err != nil {
		return fmt.Errorf("update chain status: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// DeleteChains 删除某码格的全部错误链（重新关联前调用）。
func (s *Store) DeleteChains(latticeID string) error {
	const q = `DELETE FROM error_chains WHERE lattice_id=?`
	if _, err := s.DB.Exec(q, latticeID); err != nil {
		return fmt.Errorf("delete chains: %w", err)
	}
	return nil
}
