package store

import (
	"database/sql"
	"fmt"

	"task245-qecorr/internal/model"
)

// CreateLattice 创建码格。distance 必须 ≥ 2。
func (s *Store) CreateLattice(codeName string, distance int) (*model.Lattice, error) {
	if distance < 2 {
		return nil, fmt.Errorf("%w: distance must be >= 2", model.ErrBadRequest)
	}
	lat := &model.Lattice{
		ID:        newID("lat"),
		CodeName:  codeName,
		Distance:  distance,
		Status:    model.LatticeActive,
		CreatedAt: nowUTC(),
	}
	const q = `INSERT INTO lattices(id, code_name, distance, status, created_at) VALUES(?,?,?,?,?)`
	if _, err := s.DB.Exec(q, lat.ID, lat.CodeName, lat.Distance, string(lat.Status), lat.CreatedAt); err != nil {
		return nil, fmt.Errorf("insert lattice: %w", err)
	}
	return lat, nil
}

// GetLattice 按 ID 获取码格。
func (s *Store) GetLattice(id string) (*model.Lattice, error) {
	const q = `SELECT id, code_name, distance, status, created_at FROM lattices WHERE id=?`
	row := s.DB.QueryRow(q, id)
	var lat model.Lattice
	var st string
	if err := row.Scan(&lat.ID, &lat.CodeName, &lat.Distance, &st, &lat.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("scan lattice: %w", err)
	}
	lat.Status = model.LatticeStatus(st)
	return &lat, nil
}

// ListLattices 列出全部码格。
func (s *Store) ListLattices() ([]*model.Lattice, error) {
	const q = `SELECT id, code_name, distance, status, created_at FROM lattices ORDER BY created_at, id`
	rows, err := s.DB.Query(q)
	if err != nil {
		return nil, fmt.Errorf("query lattices: %w", err)
	}
	defer rows.Close()
	var out []*model.Lattice
	for rows.Next() {
		var lat model.Lattice
		var st string
		if err := rows.Scan(&lat.ID, &lat.CodeName, &lat.Distance, &st, &lat.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan lattice: %w", err)
		}
		lat.Status = model.LatticeStatus(st)
		out = append(out, &lat)
	}
	return out, rows.Err()
}

// UpdateLatticeStatus 更新码格状态。
func (s *Store) UpdateLatticeStatus(id string, st model.LatticeStatus) error {
	const q = `UPDATE lattices SET status=? WHERE id=?`
	res, err := s.DB.Exec(q, string(st), id)
	if err != nil {
		return fmt.Errorf("update lattice status: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// CreateQubit 在码格中新增量子比特。
func (s *Store) CreateQubit(latticeID, label string, posX, posY int) (*model.Qubit, error) {
	lat, err := s.GetLattice(latticeID)
	if err != nil {
		return nil, err
	}
	if lat.Status == model.LatticeSealed {
		return nil, fmt.Errorf("%w: lattice sealed", model.ErrSealed)
	}
	qb := &model.Qubit{
		ID:        newID("qb"),
		LatticeID: latticeID,
		Label:     label,
		PosX:      posX,
		PosY:      posY,
		Status:    model.QubitActive,
	}
	const q = `INSERT INTO qubits(id, lattice_id, label, pos_x, pos_y, status) VALUES(?,?,?,?,?,?)`
	if _, err := s.DB.Exec(q, qb.ID, qb.LatticeID, qb.Label, qb.PosX, qb.PosY, string(qb.Status)); err != nil {
		return nil, fmt.Errorf("insert qubit: %w", err)
	}
	return qb, nil
}

// GetQubit 按 ID 获取量子比特。
func (s *Store) GetQubit(id string) (*model.Qubit, error) {
	const q = `SELECT id, lattice_id, label, pos_x, pos_y, status FROM qubits WHERE id=?`
	row := s.DB.QueryRow(q, id)
	var qb model.Qubit
	var st string
	if err := row.Scan(&qb.ID, &qb.LatticeID, &qb.Label, &qb.PosX, &qb.PosY, &st); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("scan qubit: %w", err)
	}
	qb.Status = model.QubitStatus(st)
	return &qb, nil
}

// ListQubits 列出某码格的全部量子比特。
func (s *Store) ListQubits(latticeID string) ([]*model.Qubit, error) {
	const q = `SELECT id, lattice_id, label, pos_x, pos_y, status FROM qubits WHERE lattice_id=? ORDER BY pos_y, pos_x, id`
	rows, err := s.DB.Query(q, latticeID)
	if err != nil {
		return nil, fmt.Errorf("query qubits: %w", err)
	}
	defer rows.Close()
	var out []*model.Qubit
	for rows.Next() {
		var qb model.Qubit
		var st string
		if err := rows.Scan(&qb.ID, &qb.LatticeID, &qb.Label, &qb.PosX, &qb.PosY, &st); err != nil {
			return nil, fmt.Errorf("scan qubit: %w", err)
		}
		qb.Status = model.QubitStatus(st)
		out = append(out, &qb)
	}
	return out, rows.Err()
}

// UpdateQubitStatus 更新量子比特状态。
func (s *Store) UpdateQubitStatus(id string, st model.QubitStatus) error {
	const q = `UPDATE qubits SET status=? WHERE id=?`
	res, err := s.DB.Exec(q, string(st), id)
	if err != nil {
		return fmt.Errorf("update qubit status: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// AddAdjacency 写入一条无向邻接边（对称双写）。
func (s *Store) AddAdjacency(latticeID, a, b string) error {
	if a == b {
		return fmt.Errorf("%w: self adjacency", model.ErrBadRequest)
	}
	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()
	const ins = `INSERT INTO adjacency(lattice_id, qubit_a, qubit_b) VALUES(?,?,?)`
	if _, err := tx.Exec(ins, latticeID, a, b); err != nil {
		return fmt.Errorf("insert adjacency a-b: %w", err)
	}
	if _, err := tx.Exec(ins, latticeID, b, a); err != nil {
		return fmt.Errorf("insert adjacency b-a: %w", err)
	}
	return tx.Commit()
}

// GetAdjacency 返回某码格全部邻接边（仅 a< b 主记录，避免重复）。
func (s *Store) GetAdjacency(latticeID string) ([]*model.Adjacency, error) {
	const q = `SELECT lattice_id, qubit_a, qubit_b FROM adjacency WHERE lattice_id=? AND qubit_a < qubit_b ORDER BY qubit_a, qubit_b`
	rows, err := s.DB.Query(q, latticeID)
	if err != nil {
		return nil, fmt.Errorf("query adjacency: %w", err)
	}
	defer rows.Close()
	var out []*model.Adjacency
	for rows.Next() {
		var a model.Adjacency
		if err := rows.Scan(&a.LatticeID, &a.QubitA, &a.QubitB); err != nil {
			return nil, fmt.Errorf("scan adjacency: %w", err)
		}
		out = append(out, &a)
	}
	return out, rows.Err()
}
