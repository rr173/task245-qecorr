package httpapi

import (
	"net/http"
	"strings"

	"task245-qecorr/internal/service"
)

type roundRequest struct {
	LatticeID string `json:"lattice_id"`
	DeviceID  string `json:"device_id"`
	RoundNo   int    `json:"round_no"`
}
type syndromeRequest struct {
	QubitID    string `json:"qubit_id"`
	Stabilizer string `json:"stabilizer"`
	RawValue   int    `json:"raw_value"`
}
type calibrationRequest struct {
	DeviceID string `json:"device_id"`
	Type     string `json:"type"`
	Detail   string `json:"detail"`
}

func (s *Server) rounds(w http.ResponseWriter, r *http.Request) {
	if !method(w, r, http.MethodPost) {
		return
	}
	var req roundRequest
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	item, err := s.svc.OpenRound(req.LatticeID, req.DeviceID, req.RoundNo)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}
func (s *Server) roundSub(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(idAfter(r.URL.Path, "/api/rounds/"), "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]
	if len(parts) == 1 {
		if !method(w, r, http.MethodGet) {
			return
		}
		item, err := s.svc.GetRound(id)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
		return
	}
	switch parts[1] {
	case "syndromes":
		if len(parts) == 3 && parts[2] == "batch" {
			s.batchSyndromes(w, r, id)
			return
		}
		s.syndromeCollection(w, r, id)
	case "calibrations":
		if !method(w, r, http.MethodPost) {
			return
		}
		var req calibrationRequest
		if err := decode(r, &req); err != nil {
			writeError(w, err)
			return
		}
		item, err := s.svc.Calibrate(id, req.DeviceID, req.Type, req.Detail)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, item)
	case "close":
		if !method(w, r, http.MethodPost) {
			return
		}
		if err := s.svc.CloseRound(id); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"closed": true})
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) batchSyndromes(w http.ResponseWriter, r *http.Request, roundID string) {
	if !method(w, r, http.MethodPost) {
		return
	}
	var inputs []service.SyndromeInput
	if err := decode(r, &inputs); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s.svc.BatchIngest(roundID, inputs))
}
func (s *Server) syndromeCollection(w http.ResponseWriter, r *http.Request, roundID string) {
	if r.Method == http.MethodGet {
		items, err := s.svc.ListSyndromes(roundID)
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
	var req syndromeRequest
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	item, err := s.svc.Ingest(roundID, req.QubitID, req.Stabilizer, req.RawValue)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}
