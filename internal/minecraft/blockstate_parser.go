package minecraft

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type BlockstateParser struct {
	fs fsApi
}

func NewBlockstateParser(fs fsApi) *BlockstateParser {
	return &BlockstateParser{fs: fs}
}

type blockstateMatch struct {
	Model string
	RotX  float64
	RotY  float64
}

type rawBlockstate struct {
	Variants  map[string]json.RawMessage `json:"variants"`
	Multipart []json.RawMessage          `json:"multipart,omitempty"`
}

func (svc *BlockstateParser) Run(ns, blockID string, props map[string]string, namespaces map[string][]string) ([]blockstateMatch, error) {
	raw, source, err := svc.readBlockstateFile(ns, blockID, namespaces)
	if err != nil {
		return nil, err
	}
	if len(raw.Variants) == 0 {
		return nil, fmt.Errorf("blockstate %s: no variants or multipart not supported", source)
	}
	return svc.matchVariant(raw.Variants, props)
}

func (svc *BlockstateParser) readBlockstateFile(ns, blockID string, namespaces map[string][]string) (*rawBlockstate, string, error) {
	roots, ok := namespaces[ns]
	if !ok {
		return nil, "", fmt.Errorf("unknown namespace %s", ns)
	}
	for _, root := range roots {
		candidate := root + "/blockstates/" + blockID + ".json"
		data, err := svc.fs.ReadFile(candidate)
		if err != nil {
			continue
		}
		var raw rawBlockstate
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, candidate, fmt.Errorf("parse blockstate %s: %w", candidate, err)
		}
		return &raw, candidate, nil
	}
	return nil, "", fmt.Errorf("blockstate %s/%s: not found in any mod directory", ns, blockID)
}

func (svc *BlockstateParser) matchVariant(variants map[string]json.RawMessage, props map[string]string) ([]blockstateMatch, error) {
	if len(variants) == 0 {
		return nil, fmt.Errorf("no variants defined")
	}

	propStr := svc.propsToKey(props)

	if raw, ok := variants[propStr]; ok {
		return svc.parseVariantValue(raw)
	}

	if raw, ok := variants[""]; ok {
		return svc.parseVariantValue(raw)
	}

	for _, k := range svc.sortedKeys(variants) {
		if svc.matchKey(k, props) {
			return svc.parseVariantValue(variants[k])
		}
	}

	return nil, fmt.Errorf("no matching variant for [%s]", propStr)
}

func (svc *BlockstateParser) matchKey(key string, props map[string]string) bool {
	if key == "" {
		return true
	}
	for _, part := range strings.Split(key, ",") {
		part = strings.TrimSpace(part)
		eq := strings.IndexByte(part, '=')
		if eq == -1 {
			continue
		}
		if props[part[:eq]] != part[eq+1:] {
			return false
		}
	}
	return true
}

func (svc *BlockstateParser) sortedKeys(m map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

type variantEntry struct {
	Model string  `json:"model"`
	X     float64 `json:"x,omitempty"`
	Y     float64 `json:"y,omitempty"`
}

func (svc *BlockstateParser) parseVariantValue(raw json.RawMessage) ([]blockstateMatch, error) {
	text := strings.TrimSpace(string(raw))

	if len(text) == 0 {
		return nil, fmt.Errorf("empty variant value")
	}

	if text[0] == '[' {
		var entries []variantEntry
		if err := json.Unmarshal(raw, &entries); err != nil {
			return nil, fmt.Errorf("parse variant array: %w", err)
		}
		matches := make([]blockstateMatch, len(entries))
		for i, e := range entries {
			matches[i] = blockstateMatch{Model: e.Model, RotX: e.X, RotY: e.Y}
		}
		return matches, nil
	}

	var entry variantEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return nil, fmt.Errorf("parse variant object: %w", err)
	}
	return []blockstateMatch{{Model: entry.Model, RotX: entry.X, RotY: entry.Y}}, nil
}

func (svc *BlockstateParser) propsToKey(props map[string]string) string {
	if len(props) == 0 {
		return ""
	}
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(props[k])
	}
	return sb.String()
}
