package store

import (
	"database/sql"
	"fmt"

	"task245-qecorr/internal/model"
)

// IngestSyndrome 写入一条综合征（按 (lattice,round,qubit,stabilizer) 幂等去重）。
// 若重复，返回 status=duplicate 的记录且 err=nil（视为幂等成功）。
func (s *Store) IngestSyndrome(roundID, latticeID string, roundNo int, qubitID, stabilizer string, rawValue int) (*model.Syndrome, error) {
	// 量子比特须属于该码格且在线
	qb, err := s.GetQubit(qubitID)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, model.ErrUnknownQubit
		}
		return nil, err
	}
	if qb.LatticeID != latticeID {
		return nil, model.ErrUnknownQubit
	}
	if qb.Status == model.QubitIsolated {
		return nil, fmt.Errorf("%w: qubit isolated", model.ErrBadRequest)
	}
	const dupQ = `SELECT id, round_id, lattice_id, round_no, qubit_id, stabilizer, raw_value, status, created_at
		FROM syndromes WHERE lattice_id=? AND round_no=? AND qubit_id=? AND stabilizer=?`
	row := s.DB.QueryRow(dupQ, latticeID, roundNo, qubitID, stabilizer)
	var existing model.Syndrome
	var st string
	err = row.Scan(&existing.ID, &existing.RoundID, &existing.LatticeID, &existing.RoundNo, &existing.QubitID, &existing.Stabilizer, &existing.RawValue, &st, &existing.CreatedAt)
	if err == nil {
		existing.Status = model.SyndromeStatus(st)
		// 标记为重复，保持幂等
		_, _ = s.DB.Exec(`UPDATE syndromes SET status=? WHERE id=?`, string(model.SynDuplicate), existing.ID)
		existing.Status = model.SynDuplicate
		return &existing, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("dup scan syndrome: %w", err)
	}
	syn := &model.Syndrome{
		ID:         newID("syn"),
		RoundID:    roundID,
		LatticeID:  latticeID,
		RoundNo:    roundNo,
		QubitID:    qubitID,
		Stabilizer: stabilizer,
		RawValue:   rawValue,
		Status:     model.SynValid,
		CreatedAt:  nowUTC(),
	}
	const q = `INSERT INTO syndromes(id, round_id, lattice_id, round_no, qubit_id, stabilizer, raw_value, status, created_at) VALUES(?,?,?,?,?,?,?,?,?)`
	if _, err := s.DB.Exec(q, syn.ID, syn.RoundID, syn.LatticeID, syn.RoundNo, syn.QubitID, syn.Stabilizer, syn.RawValue, string(syn.Status), syn.CreatedAt); err != nil {
		return nil, fmt.Errorf("insert syndrome: %w", err)
	}
	return syn, nil
}

// ListSyndromesByRound 列出某轮次的综合征。
func (s *Store) ListSyndromesByRound(roundID string) ([]*model.Syndrome, error) {
	const q = `SELECT id, round_id, lattice_id, round_no, qubit_id, stabilizer, raw_value, status, created_at FROM syndromes WHERE round_id=? ORDER BY round_no, qubit_id, stabilizer`
	rows, err := s.DB.Query(q, roundID)
	if err != nil {
		return nil, fmt.Errorf("query syndromes: %w", err)
	}
	defer rows.Close()
	var out []*model.Syndrome
	for rows.Next() {
		var syn model.Syndrome
		var st string
		if err := rows.Scan(&syn.ID, &syn.RoundID, &syn.LatticeID, &syn.RoundNo, &syn.QubitID, &syn.Stabilizer, &syn.RawValue, &st, &syn.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan syndrome: %w", err)
		}
		syn.Status = model.SyndromeStatus(st)
		out = append(out, &syn)
	}
	return out, rows.Err()
}

// ListValidDefects 列出某码格 all 轮次中 raw_value=1 的有效缺陷（用于构图）。
func (s *Store) ListValidDefects(latticeID string) ([]*model.Syndrome, error) {
	const q = `SELECT id, round_id, lattice_id, round_no, qubit_id, stabilizer, raw_value, status, created_at
		FROM syndromes WHERE lattice_id=? AND raw_value=1 ORDER BY round_no, qubit_id`
	rows, err := s.DB.Query(q, latticeID)
	if err != nil {
		return nil, fmt.Errorf("query defects: %w", err)
	}
	defer rows.Close()
	var out []*model.Syndrome
	for rows.Next() {
		var syn model.Syndrome
		var st string
		if err := rows.Scan(&syn.ID, &syn.RoundID, &syn.LatticeID, &syn.RoundNo, &syn.QubitID, &syn.Stabilizer, &syn.RawValue, &st, &syn.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan syndrome: %w", err)
		}
		syn.Status = model.SyndromeStatus(st)
		out = append(out, &syn)
	}
	return out, rows.Err()
}

// UpdateSyndromeStatus 更新综合征状态（如标记疑似坏测量）。
func (s *Store) UpdateSyndromeStatus(id string, st model.SyndromeStatus) error {
	const q = `UPDATE syndromes SET status=? WHERE id=?`
	res, err := s.DB.Exec(q, string(st), id)
	if err != nil {
		return fmt.Errorf("update syndrome status: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// CreateCalibration 写入一条校准事件。
func (s *Store) CreateCalibration(latticeID, deviceID string, roundNo int, ctype, detail string) (*model.CalibrationEvent, error) {
	c := &model.CalibrationEvent{
		ID:        newID("cal"),
		LatticeID: latticeID,
		DeviceID:  deviceID,
		RoundNo:   roundNo,
		Type:      ctype,
		Detail:    detail,
		CreatedAt: nowUTC(),
	}
	const q = `INSERT INTO calibrations(id, lattice_id, device_id, round_no, type, detail, created_at) VALUES(?,?,?,?,?,?,?)`
	if _, err := s.DB.Exec(q, c.ID, c.LatticeID, c.DeviceID, c.RoundNo, c.Type, c.Detail, c.CreatedAt); err != nil {
		return nil, fmt.Errorf("insert calibration: %w", err)
	}
	return c, nil
}

// ListCalibrations 列出某码格的校准事件。
func (s *Store) ListCalibrations(latticeID string) ([]*model.CalibrationEvent, error) {
	const q = `SELECT id, lattice_id, device_id, round_no, type, detail, created_at FROM calibrations WHERE lattice_id=? ORDER BY round_no, id`
	rows, err := s.DB.Query(q, latticeID)
	if err != nil {
		return nil, fmt.Errorf("query calibrations: %w", err)
	}
	defer rows.Close()
	var out []*model.CalibrationEvent
	for rows.Next() {
		var c model.CalibrationEvent
		if err := rows.Scan(&c.ID, &c.LatticeID, &c.DeviceID, &c.RoundNo, &c.Type, &c.Detail, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan calibration: %w", err)
		}
		out = append(out, &c)
	}
	return out, rows.Err()
}
