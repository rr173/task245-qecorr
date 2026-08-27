// Package store 负责 SQLite 持久化：连接、建表迁移与各实体的增删查改。
// 使用纯 Go 驱动 modernc.org/sqlite（CGO 无关，可离线构建）。
package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Store 封装数据库连接，供各业务包直接使用。
type Store struct {
	DB *sql.DB
}

// OpenStore 打开（必要时创建）SQLite 数据库并执行建表迁移。
func OpenStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1) // SQLite 单写者，避免跨连接写冲突
	for _, p := range []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA busy_timeout=5000;",
	} {
		if _, err := db.Exec(p); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("set pragma %q: %w", p, err)
		}
	}
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &Store{DB: db}, nil
}

// Close 关闭数据库连接。
func (s *Store) Close() error {
	return s.DB.Close()
}

func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS lattices (
			id TEXT PRIMARY KEY,
			code_name TEXT NOT NULL,
			distance INTEGER NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS qubits (
			id TEXT PRIMARY KEY,
			lattice_id TEXT NOT NULL,
			label TEXT NOT NULL,
			pos_x INTEGER NOT NULL,
			pos_y INTEGER NOT NULL,
			status TEXT NOT NULL,
			FOREIGN KEY(lattice_id) REFERENCES lattices(id)
		);`,
		`CREATE TABLE IF NOT EXISTS adjacency (
			lattice_id TEXT NOT NULL,
			qubit_a TEXT NOT NULL,
			qubit_b TEXT NOT NULL,
			PRIMARY KEY(lattice_id, qubit_a, qubit_b)
		);`,
		`CREATE TABLE IF NOT EXISTS rounds (
			id TEXT PRIMARY KEY,
			lattice_id TEXT NOT NULL,
			round_no INTEGER NOT NULL,
			device_id TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL,
			UNIQUE(lattice_id, round_no)
		);`,
		`CREATE TABLE IF NOT EXISTS syndromes (
			id TEXT PRIMARY KEY,
			round_id TEXT NOT NULL,
			lattice_id TEXT NOT NULL,
			round_no INTEGER NOT NULL,
			qubit_id TEXT NOT NULL,
			stabilizer TEXT NOT NULL,
			raw_value INTEGER NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL,
			UNIQUE(lattice_id, round_no, qubit_id, stabilizer)
		);`,
		`CREATE TABLE IF NOT EXISTS calibrations (
			id TEXT PRIMARY KEY,
			lattice_id TEXT NOT NULL,
			device_id TEXT NOT NULL,
			round_no INTEGER NOT NULL,
			type TEXT NOT NULL,
			detail TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS defect_edges (
			id TEXT PRIMARY KEY,
			lattice_id TEXT NOT NULL,
			round_a INTEGER NOT NULL,
			qubit_a TEXT NOT NULL,
			round_b INTEGER NOT NULL,
			qubit_b TEXT NOT NULL,
			weight REAL NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS error_chains (
			id TEXT PRIMARY KEY,
			lattice_id TEXT NOT NULL,
			status TEXT NOT NULL,
			first_round INTEGER NOT NULL,
			last_round INTEGER NOT NULL,
			involved_qubits TEXT NOT NULL,
			suspected_device TEXT NOT NULL,
			score REAL NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS snapshots (
			id TEXT PRIMARY KEY,
			lattice_id TEXT NOT NULL,
			status TEXT NOT NULL,
			baseline_round INTEGER NOT NULL,
			evidence_json TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_syndromes_round ON syndromes(round_id);`,
		`CREATE INDEX IF NOT EXISTS idx_defects_lattice ON defect_edges(lattice_id);`,
		`CREATE INDEX IF NOT EXISTS idx_chains_lattice ON error_chains(lattice_id);`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("exec migration %q: %w", s, err)
		}
	}
	return nil
}

// newID 生成短随机 ID（无外部依赖）。
func newID(prefix string) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// 极不可能失败；退化为时间熵
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(b)
}

// nowUTC 返回 RFC3339 时间戳。
func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}
