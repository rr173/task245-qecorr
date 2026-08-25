package model

// LatticeStatus 码格（量子纠错码布局）的生命周期状态。
type LatticeStatus string

const (
	// LatticeActive 活跃，可接收测量。
	LatticeActive LatticeStatus = "active"
	// LatticeIsolated 已隔离（坏测量器导致局部停用）。
	LatticeIsolated LatticeStatus = "isolated"
	// LatticeSealed 封存，禁止任何写入。
	LatticeSealed LatticeStatus = "sealed"
)

// Lattice 表示一个量子纠错码布局（code lattice），承载量子比特与邻接关系。
type Lattice struct {
	ID        string        `json:"id"`
	CodeName  string        `json:"code_name"`
	Distance  int           `json:"distance"`
	Status    LatticeStatus `json:"status"`
	CreatedAt string        `json:"created_at"`
}

// QubitStatus 量子比特状态。
type QubitStatus string

const (
	// QubitActive 在线可用。
	QubitActive QubitStatus = "active"
	// QubitIsolated 被隔离（疑似坏测量器）。
	QubitIsolated QubitStatus = "isolated"
)

// Qubit 码格中的一个物理/逻辑量子比特节点。
type Qubit struct {
	ID        string      `json:"id"`
	LatticeID string      `json:"lattice_id"`
	Label     string      `json:"label"`
	PosX      int         `json:"pos_x"`
	PosY      int         `json:"pos_y"`
	Status    QubitStatus `json:"status"`
}

// Adjacency 码格内两个量子比特的邻接边（无向，存储时对称双写）。
type Adjacency struct {
	LatticeID string `json:"lattice_id"`
	QubitA    string `json:"qubit_a"`
	QubitB    string `json:"qubit_b"`
}
