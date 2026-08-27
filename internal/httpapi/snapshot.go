package httpapi

import (
	"net/http"
	"strings"
)

type snapshotRequest struct {
	BaselineRound int `json:"baseline_round"`
}

func (s *Server) snapshotCollection(w http.ResponseWriter, r *http.Request, latticeID string) {
	if r.Method == http.MethodGet {
		items, err := s.svc.ListSnapshots(latticeID)
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
	var req snapshotRequest
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	item, err := s.svc.DraftSnapshot(latticeID, req.BaselineRound)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}
func (s *Server) snapshotSub(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(idAfter(r.URL.Path, "/api/snapshots/"), "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]
	if len(parts) == 1 {
		if !method(w, r, http.MethodGet) {
			return
		}
		item, err := s.svc.GetSnapshot(id)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
		return
	}
	if !method(w, r, http.MethodPost) {
		return
	}
	switch parts[1] {
	case "publish":
		if err := s.svc.PublishSnapshot(id); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"published": true})
	case "supersede":
		if err := s.svc.SupersedeSnapshot(id); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"superseded": true})
	default:
		http.NotFound(w, r)
	}
}
