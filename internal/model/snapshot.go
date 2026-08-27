package model

// SnapshotStatus 解码快照（不可变错误链判定证据）的状态机。
type SnapshotStatus string

const (
	// SnapDraft 草稿，可继续修订。
	SnapDraft SnapshotStatus = "draft"
	// SnapPublished 已发布，内容不可变。
	SnapPublished SnapshotStatus = "published"
	// SnapSuperseded 已被替代（新快照发布后旧快照转为该状态）。
	SnapSuperseded SnapshotStatus = "superseded"
)

// DecodingSnapshot 研究员发布的可追溯解码证据快照。
type DecodingSnapshot struct {
	ID            string         `json:"id"`
	LatticeID     string         `json:"lattice_id"`
	Status        SnapshotStatus `json:"status"`
	BaselineRound int            `json:"baseline_round"`
	EvidenceJSON  string         `json:"evidence_json"`
	CreatedAt     string         `json:"created_at"`
}
