package httpapi

import (
	"net/http"
	"strings"
)

type chainDecision struct {
	DeviceID string `json:"device_id"`
}

func (s *Server) chainSub(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(idAfter(r.URL.Path, "/api/chains/"), "/"), "/")
	if len(parts) != 2 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	if !method(w, r, http.MethodPost) {
		return
	}
	switch parts[1] {
	case "confirm":
		var req chainDecision
		if err := decode(r, &req); err != nil {
			writeError(w, err)
			return
		}
		if err := s.svc.ConfirmChain(parts[0], req.DeviceID); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"confirmed": true})
	case "reject":
		if err := s.svc.RejectChain(parts[0]); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"rejected": true})
	default:
		http.NotFound(w, r)
	}
}
