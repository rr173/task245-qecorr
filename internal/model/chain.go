package model

import "encoding/json"

// ChainStatus 错误链（跨轮次传播的错误模式）的状态机。
type ChainStatus string

const (
	// ChainCandidate 候选，刚从缺陷图析出。
	ChainCandidate ChainStatus = "candidate"
	// ChainTransient 短暂，仅跨 1~2 轮。
	ChainTransient ChainStatus = "transient"
	// ChainPersistent 持续，跨 ≥3 个连续轮次。
	ChainPersistent ChainStatus = "persistent"
	// ChainConfirmed 确认（研究员裁决为真实错误链）。
	ChainConfirmed ChainStatus = "confirmed"
	// ChainRejected 否决。
	ChainRejected ChainStatus = "rejected"
)

// ErrorChain 由时空缺陷图关联出的候选/持续错误链。
type ErrorChain struct {
	ID              string      `json:"id"`
	LatticeID       string      `json:"lattice_id"`
	Status          ChainStatus `json:"status"`
	FirstRound      int         `json:"first_round"`
	LastRound       int         `json:"last_round"`
	InvolvedQubits  []string    `json:"involved_qubits"`
	SuspectedDevice string      `json:"suspected_device"`
	Score           float64     `json:"score"`
	CreatedAt       string      `json:"created_at"`
}

// QubitsJSON 将量子比特列表序列化为可存储的 JSON 字符串。
func (e ErrorChain) QubitsJSON() (string, error) {
	b, err := json.Marshal(e.InvolvedQubits)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ParseQubitsJSON 从存储的 JSON 字符串还原量子比特列表。
func ParseQubitsJSON(s string) ([]string, error) {
	if s == "" {
		return nil, nil
	}
	var out []string
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, err
	}
	return out, nil
}
