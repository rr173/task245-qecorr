// Package chain turns a defect graph into deterministic candidate error chains.
package chain

import (
	"fmt"
	"sort"

	"task245-qecorr/internal/model"
	"task245-qecorr/internal/store"
)

type chainNode struct {
	round int
	qubit string
}

// Analyze replaces the derived chains for a lattice with connected components
// from the current persisted defect graph. Components are deterministic and
// contain only valid syndromes; suspected measurements are intentionally not
// used as evidence.
func Analyze(s *store.Store, latticeID string) ([]*model.ErrorChain, error) {
	defects, err := s.ListValidDefects(latticeID)
	if err != nil {
		return nil, err
	}
	edges, err := s.ListDefectEdges(latticeID)
	if err != nil {
		return nil, err
	}
	if err := s.DeleteChains(latticeID); err != nil {
		return nil, err
	}
	if len(defects) == 0 {
		return []*model.ErrorChain{}, nil
	}
	key := func(round int, qubit string) string { return fmt.Sprintf("%d:%s", round, qubit) }
	parent := make(map[string]string)
	find := func(x string) string {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}
	union := func(a, b string) {
		ra, rb := find(a), find(b)
		if ra != rb {
			if ra < rb {
				parent[rb] = ra
			} else {
				parent[ra] = rb
			}
		}
	}
	for _, d := range defects {
		parent[key(d.RoundNo, d.QubitID)] = key(d.RoundNo, d.QubitID)
	}
	for _, e := range edges {
		a, b := key(e.RoundA, e.QubitA), key(e.RoundB, e.QubitB)
		if _, ok := parent[a]; ok {
			if _, ok := parent[b]; ok {
				union(a, b)
			}
		}
	}
	groups := make(map[string][]chainNode)
	for _, d := range defects {
		k := key(d.RoundNo, d.QubitID)
		groups[find(k)] = append(groups[find(k)], chainNode{d.RoundNo, d.QubitID})
	}
	roots := make([]string, 0, len(groups))
	for root := range groups {
		roots = append(roots, root)
	}
	sort.Strings(roots)
	chains := make([]*model.ErrorChain, 0, len(roots))
	for _, root := range roots {
		members := groups[root]
		sort.Slice(members, func(i, j int) bool {
			if members[i].round != members[j].round {
				return members[i].round < members[j].round
			}
			return members[i].qubit < members[j].qubit
		})
		first, last := members[0].round, members[len(members)-1].round
		qubits := uniqueQubits(members)
		span := last - first + 1
		status := model.ChainTransient
		if longestConsecutiveRun(members) >= 3 {
			status = model.ChainPersistent
		}
		c, err := s.CreateChain(latticeID, status, first, last, qubits, "", float64(len(members))+float64(span)/10)
		if err != nil {
			return nil, err
		}
		chains = append(chains, c)
	}
	return chains, nil
}

func longestConsecutiveRun(nodes []chainNode) int {
	seen := make(map[int]bool)
	for _, n := range nodes { seen[n.round] = true }
	best := 0
	for _, n := range nodes {
		if seen[n.round-1] { continue }
		run := 1
		for seen[n.round+run] { run++ }
		if run > best { best = run }
	}
	return best
}

func uniqueQubits(nodes []chainNode) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if !seen[n.qubit] {
			seen[n.qubit] = true
			result = append(result, n.qubit)
		}
	}
	sort.Strings(result)
	return result
}
