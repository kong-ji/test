// Package fingerprint implements the banner fingerprint identification
// engine. It consumes a compiled rules.Rules value and maps each input
// (ip, port, banner) to a Result. No fingerprint knowledge lives in code:
// everything is driven by the external rules document.
package fingerprint

import (
	"github.com/kong-ji/test/internal/rules"
)

// Input is a single raw scan sample to identify.
type Input struct {
	IP     string
	Port   int
	Banner string
}

// Result is the identification outcome for one input. IP and Port are passed
// through unchanged; unrecognised fields remain empty strings.
type Result struct {
	IP         string
	Port       int
	Protocol   string
	Product    string
	Version    string
	OSHint     string
	Confidence float64
}

// Engine identifies banners against a compiled rule set.
type Engine struct {
	rules *rules.Rules
}

// New constructs an Engine from a compiled rule set.
func New(r *rules.Rules) *Engine {
	return &Engine{rules: r}
}

// Identify maps a batch of inputs to results, preserving input order.
func (e *Engine) Identify(inputs []Input) []Result {
	results := make([]Result, 0, len(inputs))
	for _, in := range inputs {
		results = append(results, e.identifyOne(in))
	}
	return results
}

func (e *Engine) identifyOne(in Input) Result {
	res := Result{IP: in.IP, Port: in.Port}

	// Phase 1: rules. Take the highest-confidence matching rule.
	best := -1.0
	for _, r := range e.rules.Rules {
		if len(r.Ports) > 0 && !contains(r.Ports, in.Port) {
			continue
		}
		m := r.Pattern.FindStringSubmatch(in.Banner)
		if m == nil {
			continue
		}
		if r.Confidence <= best {
			continue
		}
		best = r.Confidence

		res.Protocol = r.Protocol
		res.Product = r.Product
		if res.Product == "" && r.ProductGroup != "" {
			res.Product = groupValue(m, r.Pattern, r.ProductGroup)
		}
		res.Version = ""
		if r.VersionGroup != "" {
			res.Version = groupValue(m, r.Pattern, r.VersionGroup)
		}
		res.OSHint = matchOSHint(in.Banner, e.rules.OSHints)
		res.Confidence = r.Confidence
	}

	if best >= 0 {
		return res
	}

	// Phase 2: TLS record layer heuristic (banner is a raw TLS handshake).
	if isTLSHandshake(in.Banner) {
		return Result{
			IP:         in.IP,
			Port:       in.Port,
			Protocol:   "TLS",
			Product:    "",
			Version:    "",
			OSHint:     "",
			Confidence: 0.6,
		}
	}

	// Phase 3: port fallbacks (weakest signal).
	for _, pf := range e.rules.PortFallbacks {
		if contains(pf.Ports, in.Port) {
			res.Protocol = pf.Protocol
			res.Product = pf.Product
			res.Confidence = pf.Confidence
			return res
		}
	}

	// Phase 4: nothing matched. Fallback is empty (no guessing).
	res.Protocol = e.rules.Fallback.Protocol
	res.Product = e.rules.Fallback.Product
	res.Version = e.rules.Fallback.Version
	res.OSHint = e.rules.Fallback.OSHint
	res.Confidence = e.rules.Fallback.Confidence
	return res
}

// groupValue extracts a named capture group value from a regexp match.
func groupValue(m []string, re interface{ SubexpNames() []string }, group string) string {
	names := re.SubexpNames()
	for i, n := range names {
		if n == group && i < len(m) {
			return m[i]
		}
	}
	return ""
}

func matchOSHint(banner string, hints []rules.OSHint) string {
	for _, h := range hints {
		if h.Pattern.MatchString(banner) {
			return h.Name
		}
	}
	return ""
}

func contains(list []int, v int) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// isTLSHandshake reports whether banner begins with a TLS record whose
// content type is handshake (0x16) and major version is 0x03.
func isTLSHandshake(banner string) bool {
	b := []byte(banner)
	return len(b) >= 3 && b[0] == 0x16 && b[1] == 0x03
}
