package model

// SyndromeStatus 综合征（单个量子比特单次稳定子测量结果）的状态机。
type SyndromeStatus string

const (
	// SynRaw 原始未判定。
	SynRaw SyndromeStatus = "raw"
	// SynValid 有效（合法且非重复）。
	SynValid SyndromeStatus = "valid"
	// SynSuspected 疑似坏测量（设备校准异常或自相矛盾）。
	SynSuspected SyndromeStatus = "suspected_bad_measurement"
	// SynDuplicate 重复（同一 (lattice,round,qubit,stabilizer) 幂等去重）。
	SynDuplicate SyndromeStatus = "duplicate"
)

// Syndrome 单个量子比特在一次轮次中某个稳定子的测量结果。
// RawValue==1 表示出现缺陷（defect），否则为平凡综合征。
type Syndrome struct {
	ID         string         `json:"id"`
	RoundID    string         `json:"round_id"`
	LatticeID  string         `json:"lattice_id"`
	RoundNo    int            `json:"round_no"`
	QubitID    string         `json:"qubit_id"`
	Stabilizer string         `json:"stabilizer"`
	RawValue   int            `json:"raw_value"`
	Status     SyndromeStatus `json:"status"`
	CreatedAt  string         `json:"created_at"`
}

// DefectEdge 时空缺陷图中的一条边：连接相邻轮次中相邻量子比特上的缺陷。
type DefectEdge struct {
	ID        string  `json:"id"`
	LatticeID string  `json:"lattice_id"`
	RoundA    int     `json:"round_a"`
	QubitA    string  `json:"qubit_a"`
	RoundB    int     `json:"round_b"`
	QubitB    string  `json:"qubit_b"`
	Weight    float64 `json:"weight"`
	CreatedAt string  `json:"created_at"`
}
