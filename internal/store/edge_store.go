package store

import (
	"fmt"

	"task245-qecorr/internal/model"
)

// CreateDefectEdge 写入一条时空缺陷图边。
func (s *Store) CreateDefectEdge(latticeID string, roundA int, qubitA string, roundB int, qubitB string, weight float64) (*model.DefectEdge, error) {
	e := &model.DefectEdge{
		ID:        newID("edge"),
		LatticeID: latticeID,
		RoundA:    roundA,
		QubitA:    qubitA,
		RoundB:    roundB,
		QubitB:    qubitB,
		Weight:    weight,
		CreatedAt: nowUTC(),
	}
	const q = `INSERT INTO defect_edges(id, lattice_id, round_a, qubit_a, round_b, qubit_b, weight, created_at) VALUES(?,?,?,?,?,?,?,?)`
	if _, err := s.DB.Exec(q, e.ID, e.LatticeID, e.RoundA, e.QubitA, e.RoundB, e.QubitB, e.Weight, e.CreatedAt); err != nil {
		return nil, fmt.Errorf("insert defect edge: %w", err)
	}
	return e, nil
}

// ClearDefectEdges 删除某码格的全部缺陷边（重新构图前调用）。
func (s *Store) ClearDefectEdges(latticeID string) error {
	const q = `DELETE FROM defect_edges WHERE lattice_id=?`
	if _, err := s.DB.Exec(q, latticeID); err != nil {
		return fmt.Errorf("clear defect edges: %w", err)
	}
	return nil
}

// ListDefectEdges 列出某码格的缺陷边。
func (s *Store) ListDefectEdges(latticeID string) ([]*model.DefectEdge, error) {
	const q = `SELECT id, lattice_id, round_a, qubit_a, round_b, qubit_b, weight, created_at FROM defect_edges WHERE lattice_id=? ORDER BY round_a, qubit_a, round_b, qubit_b`
	rows, err := s.DB.Query(q, latticeID)
	if err != nil {
		return nil, fmt.Errorf("query defect edges: %w", err)
	}
	defer rows.Close()
	var out []*model.DefectEdge
	for rows.Next() {
		var e model.DefectEdge
		if err := rows.Scan(&e.ID, &e.LatticeID, &e.RoundA, &e.QubitA, &e.RoundB, &e.QubitB, &e.Weight, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan defect edge: %w", err)
		}
		out = append(out, &e)
	}
	return out, rows.Err()
}
