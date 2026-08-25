// Package httpapi exposes the qecorr service as a small JSON HTTP API.
package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"task245-qecorr/internal/model"
	"task245-qecorr/internal/service"
)

type Server struct{ svc *service.Service }

func NewServer(svc *service.Service) *Server { return &Server{svc: svc} }

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.health)
	mux.HandleFunc("/api/selfcheck", s.selfCheck)
	mux.HandleFunc("/api/lattices", s.lattices)
	mux.HandleFunc("/api/lattices/", s.latticeSub)
	mux.HandleFunc("/api/rounds", s.rounds)
	mux.HandleFunc("/api/rounds/", s.roundSub)
	mux.HandleFunc("/api/qubits/", s.qubitSub)
	mux.HandleFunc("/api/chains/", s.chainSub)
	mux.HandleFunc("/api/snapshots/", s.snapshotSub)
	mux.HandleFunc("/", s.root)
	return mux
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, err error) {
	code := http.StatusBadRequest
	switch {
	case errors.Is(err, model.ErrNotFound):
		code = http.StatusNotFound
	case errors.Is(err, model.ErrDuplicate), errors.Is(err, model.ErrSealed), errors.Is(err, model.ErrInvalidState):
		code = http.StatusConflict
	case errors.Is(err, model.ErrUnknownQubit):
		code = http.StatusUnprocessableEntity
	}
	writeJSON(w, code, map[string]any{"error": err.Error()})
}
func decode(r *http.Request, dst any) error { return json.NewDecoder(r.Body).Decode(dst) }
func method(w http.ResponseWriter, r *http.Request, allowed ...string) bool {
	for _, m := range allowed {
		if r.Method == m {
			return true
		}
	}
	w.Header().Set("Allow", strings.Join(allowed, ", "))
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	return false
}
func idAfter(path, prefix string) string {
	value := strings.TrimPrefix(path, prefix)
	return strings.Trim(value, "/")
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if !method(w, r, http.MethodGet) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "qecorr"})
}
func (s *Server) selfCheck(w http.ResponseWriter, r *http.Request) {
	if !method(w, r, http.MethodGet) {
		return
	}
	path, err := service.TemporaryDB()
	if err != nil {
		writeError(w, err)
		return
	}
	if err := service.RunSelfCheck(path); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"service": "qecorr", "resources": []string{"lattices", "rounds", "chains", "snapshots"}})
}
