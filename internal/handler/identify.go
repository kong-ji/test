package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kong-ji/test/internal/fingerprint"
	"github.com/kong-ji/test/internal/response"
)

// maxIdentifyBodySize limits the request body size for the identify endpoint.
const maxIdentifyBodySize = 1 << 20 // 1 MiB

// identifyRequestItem is a single element of the identify request payload.
type identifyRequestItem struct {
	IP     string `json:"ip"`
	Port   int    `json:"port"`
	Banner string `json:"banner"`
}

// identifyResponseItem is a single element of the identify response payload.
type identifyResponseItem struct {
	IP         string  `json:"ip"`
	Port       int     `json:"port"`
	Protocol   string  `json:"protocol"`
	Product    string  `json:"product"`
	Version    string  `json:"version"`
	OSHint     string  `json:"os_hint"`
	Confidence float64 `json:"confidence"`
}

// IdentifyHandler serves POST /fingerprint.
type IdentifyHandler struct {
	engine *fingerprint.Engine
}

// NewIdentifyHandler constructs an IdentifyHandler backed by the given engine.
func NewIdentifyHandler(e *fingerprint.Engine) *IdentifyHandler {
	return &IdentifyHandler{engine: e}
}

// ServeHTTP implements http.Handler.
func (h *IdentifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed: "+r.Method+" "+r.URL.Path)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxIdentifyBodySize)

	var req []identifyRequestItem
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	inputs := make([]fingerprint.Input, 0, len(req))
	for _, item := range req {
		inputs = append(inputs, fingerprint.Input{
			IP:     item.IP,
			Port:   item.Port,
			Banner: item.Banner,
		})
	}

	results := h.engine.Identify(inputs)

	out := make([]identifyResponseItem, 0, len(results))
	for _, res := range results {
		out = append(out, identifyResponseItem{
			IP:         res.IP,
			Port:       res.Port,
			Protocol:   res.Protocol,
			Product:    res.Product,
			Version:    res.Version,
			OSHint:     res.OSHint,
			Confidence: res.Confidence,
		})
	}

	response.JSON(w, http.StatusOK, out)
}
