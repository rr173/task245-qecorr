// Package round 负责实验轮次与综合征的接收、去重与状态流转，
// 并在摄入阶段结合校准事件检测疑似坏测量。
package round

import (
	"fmt"

	"task245-qecorr/internal/model"
	"task245-qecorr/internal/store"
)

// Open 开启一个新实验轮次（receiving）。
func Open(s *store.Store, latticeID, deviceID string, roundNo int) (*model.MeasurementRound, error) {
	return s.CreateRound(latticeID, deviceID, roundNo)
}

// Ingest 向某轮次摄入一条综合征。仅 receiving 态允许；重复视为幂等（status=duplicate）；
// 若该设备在该轮次存在校准异常，则标记为疑似坏测量。
func Ingest(s *store.Store, roundID, qubitID, stabilizer string, rawValue int) (*model.Syndrome, error) {
	r, err := s.GetRound(roundID)
	if err != nil {
		return nil, err
	}
	if r.Status != model.RoundReceiving {
		return nil, fmt.Errorf("%w: round not receiving", model.ErrInvalidState)
	}
	syn, err := s.IngestSyndrome(r.ID, r.LatticeID, r.RoundNo, qubitID, stabilizer, rawValue)
	if err != nil {
		return nil, err
	}
	if syn.Status == model.SynDuplicate {
		return syn, nil
	}
	// 坏测量检测：同设备同轮次存在校准异常
	if HasCalibrationAnomaly(s, r.LatticeID, r.DeviceID, r.RoundNo) {
		if err := s.UpdateSyndromeStatus(syn.ID, model.SynSuspected); err == nil {
			syn.Status = model.SynSuspected
		}
	}
	return syn, nil
}

// Calibrate 写入一条校准事件（通常标记某设备某轮次异常）。
func Calibrate(s *store.Store, roundID, deviceID, ctype, detail string) (*model.CalibrationEvent, error) {
	r, err := s.GetRound(roundID)
	if err != nil {
		return nil, err
	}
	if r.Status != model.RoundReceiving {
		return nil, fmt.Errorf("%w: round not receiving", model.ErrInvalidState)
	}
	return s.CreateCalibration(r.LatticeID, deviceID, r.RoundNo, ctype, detail)
}

// Close 标记轮次摄入结束（receiving → pending_correlation）。
func Close(s *store.Store, roundID string) error {
	r, err := s.GetRound(roundID)
	if err != nil {
		return err
	}
	if r.Status != model.RoundReceiving {
		return fmt.Errorf("%w: round %s in %s", model.ErrInvalidState, r.ID, r.Status)
	}
	return s.UpdateRoundStatus(roundID, model.RoundPending)
}

// MarkAnalyzed 关联完成后由 spacetime 调用（pending_correlation → analyzed）。
func MarkAnalyzed(s *store.Store, roundID string) error {
	r, err := s.GetRound(roundID)
	if err != nil {
		return err
	}
	if r.Status != model.RoundPending {
		return fmt.Errorf("%w: round %s not pending", model.ErrInvalidState, r.ID)
	}
	return s.UpdateRoundStatus(roundID, model.RoundAnalyzed)
}

// Seal 封存轮次（analyzed → sealed）。
func Seal(s *store.Store, roundID string) error {
	r, err := s.GetRound(roundID)
	if err != nil {
		return err
	}
	if r.Status != model.RoundAnalyzed {
		return fmt.Errorf("%w: round %s must be analyzed before seal", model.ErrInvalidState, r.ID)
	}
	return s.UpdateRoundStatus(roundID, model.RoundSealed)
}

// Get 获取轮次。
func Get(s *store.Store, id string) (*model.MeasurementRound, error) {
	return s.GetRound(id)
}

// List 列出码格的轮次。
func List(s *store.Store, latticeID string) ([]*model.MeasurementRound, error) {
	return s.ListRounds(latticeID)
}

// ListSyndromes 列出轮次的综合征。
func ListSyndromes(s *store.Store, roundID string) ([]*model.Syndrome, error) {
	return s.ListSyndromesByRound(roundID)
}

// ListCalibrations 列出码格的校准事件。
func ListCalibrations(s *store.Store, latticeID string) ([]*model.CalibrationEvent, error) {
	return s.ListCalibrations(latticeID)
}

// HasCalibrationAnomaly reports whether the device has a calibration event in
// the same lattice and round. An unrelated readout channel must not poison
// otherwise valid measurements.
func HasCalibrationAnomaly(s *store.Store, latticeID, deviceID string, roundNo int) bool {
	events, err := s.ListCalibrations(latticeID)
	if err != nil {
		return false
	}
	for _, event := range events {
		if event.RoundNo == roundNo {
			return true
		}
	}
	return false
}
