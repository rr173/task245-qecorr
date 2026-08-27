package model

// RoundStatus 实验轮次（一次稳定子测量采样）的状态机。
type RoundStatus string

const (
	// RoundReceiving 接收中，仍可写入综合征。
	RoundReceiving RoundStatus = "receiving"
	// RoundPending 待关联，综合征已齐备、等待构图。
	RoundPending RoundStatus = "pending_correlation"
	// RoundAnalyzed 已分析，缺陷图与错误链已产出。
	RoundAnalyzed RoundStatus = "analyzed"
	// RoundSealed 封存，拒绝任何修改。
	RoundSealed RoundStatus = "sealed"
)

// MeasurementRound 一次稳定子测量的实验轮次。
type MeasurementRound struct {
	ID        string      `json:"id"`
	LatticeID string      `json:"lattice_id"`
	RoundNo   int         `json:"round_no"`
	DeviceID  string      `json:"device_id"`
	Status    RoundStatus `json:"status"`
	CreatedAt string      `json:"created_at"`
}

// CalibrationEvent 校准事件，用于标记某设备在某轮次的异常（坏测量器）。
type CalibrationEvent struct {
	ID        string `json:"id"`
	LatticeID string `json:"lattice_id"`
	DeviceID  string `json:"device_id"`
	RoundNo   int    `json:"round_no"`
	Type      string `json:"type"`
	Detail    string `json:"detail"`
	CreatedAt string `json:"created_at"`
}
