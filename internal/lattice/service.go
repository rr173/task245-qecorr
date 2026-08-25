// Package lattice 负责码格（量子纠错码布局）及其量子比特、邻接关系的管理，
// 包含邻接对称性校验、坏测量器隔离与封存等状态约束。
package lattice

import (
	"fmt"

	"task245-qecorr/internal/model"
	"task245-qecorr/internal/store"
)

// Create 创建码格。
func Create(s *store.Store, codeName string, distance int) (*model.Lattice, error) {
	lat, err := s.CreateLattice(codeName, distance)
	if err != nil {
		return nil, err
	}
	return lat, nil
}

// AddQubit 向码格新增一个量子比特。
func AddQubit(s *store.Store, latticeID, label string, posX, posY int) (*model.Qubit, error) {
	if _, err := s.GetLattice(latticeID); err != nil {
		return nil, err
	}
	return s.CreateQubit(latticeID, label, posX, posY)
}

// AddAdjacency 声明一条无向邻接边，校验两端量子比特均属该码格且对称。
func AddAdjacency(s *store.Store, latticeID, a, b string) error {
	qa, err := s.GetQubit(a)
	if err != nil {
		if err == model.ErrNotFound {
			return model.ErrUnknownQubit
		}
		return err
	}
	qb, err := s.GetQubit(b)
	if err != nil {
		if err == model.ErrNotFound {
			return model.ErrUnknownQubit
		}
		return err
	}
	if qa.LatticeID != latticeID || qb.LatticeID != latticeID {
		return model.ErrUnknownQubit
	}
	if a == b {
		return fmt.Errorf("%w: self edge", model.ErrBadRequest)
	}
	return s.AddAdjacency(latticeID, a, b)
}

// IsolateQubit 隔离量子比特（疑似坏测量器时停用）。
func IsolateQubit(s *store.Store, qubitID string) error {
	qb, err := s.GetQubit(qubitID)
	if err != nil {
		return err
	}
	if qb.Status == model.QubitIsolated {
		return nil
	}
	return s.UpdateQubitStatus(qubitID, model.QubitIsolated)
}

// Seal 封存码格，禁止后续写入。
func Seal(s *store.Store, latticeID string) error {
	if _, err := s.GetLattice(latticeID); err != nil {
		return err
	}
	return s.UpdateLatticeStatus(latticeID, model.LatticeSealed)
}

// Get 获取码格。
func Get(s *store.Store, id string) (*model.Lattice, error) {
	return s.GetLattice(id)
}

// List 列出全部码格。
func List(s *store.Store) ([]*model.Lattice, error) {
	return s.ListLattices()
}

// ListQubits 列出码格的量子比特。
func ListQubits(s *store.Store, latticeID string) ([]*model.Qubit, error) {
	return s.ListQubits(latticeID)
}

// Adjacency 返回码格邻接关系。
func Adjacency(s *store.Store, latticeID string) ([]*model.Adjacency, error) {
	return s.GetAdjacency(latticeID)
}
