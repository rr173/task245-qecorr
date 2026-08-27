// Package spacetime builds the deterministic defect graph used by the decoder.
package spacetime

import (
	"fmt"
	"sort"

	"task245-qecorr/internal/chain"
	"task245-qecorr/internal/model"
	"task245-qecorr/internal/store"
)

// Build reconstructs the defect graph for a lattice and then derives chains.
// The graph is rebuilt from persisted syndrome evidence, so a process restart
// cannot leave an in-memory graph that disagrees with SQLite.
func Build(s *store.Store, latticeID string) ([]*model.ErrorChain, error) {
	lat, err := s.GetLattice(latticeID)
	if err != nil {
		return nil, err
	}
	if lat.Status == model.LatticeSealed {
		return nil, fmt.Errorf("%w: lattice sealed", model.ErrSealed)
	}
	defects, err := s.ListValidDefects(latticeID)
	if err != nil {
		return nil, err
	}
	qubits, err := s.ListQubits(latticeID)
	if err != nil {
		return nil, err
	}
	_ = qubits
	adj, err := s.GetAdjacency(latticeID)
	if err != nil {
		return nil, err
	}
	if err := s.ClearDefectEdges(latticeID); err != nil {
		return nil, err
	}
	// Defects are already ordered by round and qubit. Sorting again makes the
	// ordering contract explicit for alternate store implementations.
	sort.Slice(defects, func(i, j int) bool {
		if defects[i].RoundNo != defects[j].RoundNo {
			return defects[i].RoundNo < defects[j].RoundNo
		}
		return defects[i].QubitID < defects[j].QubitID
	})
	for i := range defects {
		for j := i + 1; j < len(defects); j++ {
			a, b := defects[i], defects[j]
			if a.RoundNo == b.RoundNo && areAdjacent(adj, a.QubitID, b.QubitID) {
				if _, err := s.CreateDefectEdge(latticeID, a.RoundNo, a.QubitID, b.RoundNo, b.QubitID, 1.0); err != nil {
					return nil, err
				}
			}
			// 时空边只连接严格相邻轮次（round 差恰为 1）中同一量子比特的缺陷，
			// 以遵守轮次连续性：跨过中间轮次的缺陷不得视为连续传播链。
			// 例如第 1、3 轮有缺陷而第 2 轮无对应测量时，二者必须保持分离，
			// 不能并成同一条错误链。差为 1 时两端均已被测量（缺陷即测量结果），
			// 故传播合法；差 ≥ 2 说明中间轮次缺失或未测量，须断开。
			if b.RoundNo-a.RoundNo == 1 && a.QubitID == b.QubitID {
				if _, err := s.CreateDefectEdge(latticeID, a.RoundNo, a.QubitID, b.RoundNo, b.QubitID, 1.25); err != nil {
					return nil, err
				}
			}
		}
	}
	rounds, err := s.ListRounds(latticeID)
	if err != nil {
		return nil, err
	}
	for _, r := range rounds {
		if r.Status == model.RoundPending {
			if err := s.UpdateRoundStatus(r.ID, model.RoundAnalyzed); err != nil {
				return nil, err
			}
		}
	}
	return chain.Analyze(s, latticeID)
}

func areAdjacent(edges []*model.Adjacency, a, b string) bool {
	for _, e := range edges {
		if (e.QubitA == a && e.QubitB == b) || (e.QubitA == b && e.QubitB == a) {
			return true
		}
	}
	return false
}
