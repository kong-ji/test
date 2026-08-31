// Package rules loads and parses banner fingerprint rules from an external
// JSON document (see configs/rules.json). No fingerprint is hardcoded here;
// the engine consumes the compiled Rules value.
package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

// Fallback describes the result returned when neither a rule nor a port
// fallback matches. It is intentionally decoupled from the fingerprint
// package to avoid an import cycle.
type Fallback struct {
	Protocol   string  `json:"protocol"`
	Product    string  `json:"product"`
	Version    string  `json:"version"`
	OSHint     string  `json:"os_hint"`
	Confidence float64 `json:"confidence"`
}

// OSHint maps an OS name to a compiled RE2 pattern matched against the banner.
type OSHint struct {
	Name    string
	Pattern *regexp.Regexp
}

// Rule is a single fingerprint rule. Pattern is pre-compiled; Ports may be
// empty meaning the rule applies to any port.
type Rule struct {
	ID           string
	Protocol     string
	Product      string
	VersionGroup string
	ProductGroup string
	Confidence   float64
	Pattern      *regexp.Regexp
	Ports        []int
}

// PortFallback maps a set of ports to a low-confidence identification.
type PortFallback struct {
	Ports      []int
	Protocol   string
	Product    string
	Confidence float64
}

// Rules is the fully parsed and compiled rule set.
type Rules struct {
	Version       int
	Fallback      Fallback
	OSHints       []OSHint
	Rules         []Rule
	PortFallbacks []PortFallback
}

// rawRules mirrors the on-disk JSON shape before compilation.
type rawRules struct {
	Version       int            `json:"version"`
	Comment       string         `json:"comment"`
	Fallback      Fallback       `json:"fallback"`
	OSHints       []rawOSHint    `json:"os_hints"`
	Rules         []rawRule      `json:"rules"`
	PortFallbacks []PortFallback `json:"port_fallbacks"`
}

type rawOSHint struct {
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
}

type rawRule struct {
	ID           string  `json:"id"`
	Protocol     string  `json:"protocol"`
	Product      string  `json:"product"`
	VersionGroup string  `json:"version_group"`
	ProductGroup string  `json:"product_group"`
	Confidence   float64 `json:"confidence"`
	Pattern      string  `json:"pattern"`
	Ports        []int   `json:"ports"`
}

// Load reads rules.json from path and compiles it.
func Load(path string) (*Rules, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("rules: read %s: %w", path, err)
	}
	return LoadBytes(data)
}

// LoadBytes parses and compiles rules from an in-memory JSON document.
func LoadBytes(data []byte) (*Rules, error) {
	var raw rawRules
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("rules: unmarshal: %w", err)
	}

	r := &Rules{
		Version:       raw.Version,
		Fallback:      raw.Fallback,
		OSHints:       make([]OSHint, 0, len(raw.OSHints)),
		Rules:         make([]Rule, 0, len(raw.Rules)),
		PortFallbacks: raw.PortFallbacks,
	}

	for _, h := range raw.OSHints {
		re, err := regexp.Compile(h.Pattern)
		if err != nil {
			return nil, fmt.Errorf("rules: compile os_hint %q pattern %q: %w", h.Name, h.Pattern, err)
		}
		r.OSHints = append(r.OSHints, OSHint{Name: h.Name, Pattern: re})
	}

	for _, rr := range raw.Rules {
		re, err := regexp.Compile(rr.Pattern)
		if err != nil {
			return nil, fmt.Errorf("rules: compile rule %q pattern %q: %w", rr.ID, rr.Pattern, err)
		}
		r.Rules = append(r.Rules, Rule{
			ID:           rr.ID,
			Protocol:     rr.Protocol,
			Product:      rr.Product,
			VersionGroup: rr.VersionGroup,
			ProductGroup: rr.ProductGroup,
			Confidence:   rr.Confidence,
			Pattern:      re,
			Ports:        rr.Ports,
		})
	}

	return r, nil
}
