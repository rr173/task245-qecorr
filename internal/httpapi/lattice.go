package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

type latticeRequest struct {
	CodeName string `json:"code_name"`
	Distance int    `json:"distance"`
}
type qubitRequest struct {
	Label string `json:"label"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
}
type adjacencyRequest struct {
	A string `json:"a"`
	B string `json:"b"`
}

func (s *Server) lattices(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		items, err := s.svc.ListLattices()
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
		return
	}
	if !method(w, r, http.MethodPost) {
		return
	}
	var req latticeRequest
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	item, err := s.svc.CreateLattice(req.CodeName, req.Distance)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (s *Server) latticeSub(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(idAfter(r.URL.Path, "/api/lattices/"), "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]
	if len(parts) == 1 {
		if !method(w, r, http.MethodGet) {
			return
		}
		item, err := s.svc.GetLattice(id)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
		return
	}
	switch parts[1] {
	case "qubits":
		s.qubitCollection(w, r, id)
	case "adjacency":
		s.adjacencyCollection(w, r, id)
	case "seal":
		if !method(w, r, http.MethodPost) {
			return
		}
		if err := s.svc.SealLattice(id); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"sealed": true})
	case "analyze":
		if !method(w, r, http.MethodPost) {
			return
		}
		chains, err := s.svc.Analyze(id)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, chains)
	case "edges":
		if !method(w, r, http.MethodGet) {
			return
		}
		edges, err := s.svc.ListEdges(id)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, edges)
	case "chains":
		if !method(w, r, http.MethodGet) {
			return
		}
		chains, err := s.svc.ListChains(id)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, chains)
	case "snapshots":
		s.snapshotCollection(w, r, id)
	default:
		http.NotFound(w, r)
	}
}
func (s *Server) qubitCollection(w http.ResponseWriter, r *http.Request, latticeID string) {
	if r.Method == http.MethodGet {
		items, err := s.svc.ListQubits(latticeID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
		return
	}
	if !method(w, r, http.MethodPost) {
		return
	}
	var req qubitRequest
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	item, err := s.svc.AddQubit(latticeID, req.Label, req.X, req.Y)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}
func (s *Server) adjacencyCollection(w http.ResponseWriter, r *http.Request, latticeID string) {
	if r.Method == http.MethodGet {
		items, err := s.svc.ListAdjacency(latticeID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
		return
	}
	if !method(w, r, http.MethodPost) {
		return
	}
	var req adjacencyRequest
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := s.svc.AddAdjacency(latticeID, req.A, req.B); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, req)
}
func (s *Server) qubitSub(w http.ResponseWriter, r *http.Request) {
	id := idAfter(r.URL.Path, "/api/qubits/")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	if !method(w, r, http.MethodPost) {
		return
	}
	if strings.HasSuffix(id, "/isolate") {
		id = strings.TrimSuffix(id, "/isolate")
		if err := s.svc.IsolateQubit(id); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"isolated": true})
		return
	}
	http.NotFound(w, r)
}
func parseInt(value string) int { n, _ := strconv.Atoi(value); return n }
