package minecraft

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type propsKeyBuilder interface {
	Run(props map[string]string) string
}

type BlockstateParser struct {
	fs              fsApi
	propsKeyBuilder propsKeyBuilder
}

func NewBlockstateParser(fs fsApi, pkb propsKeyBuilder) *BlockstateParser {
	return &BlockstateParser{fs: fs, propsKeyBuilder: pkb}
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
		return nil, fmt.Errorf("resolve blockstate %s/%s: %w", ns, blockID, err)
	}
	if len(raw.Variants) > 0 {
		return svc.matchVariant(raw.Variants, props)
	}
	if len(raw.Multipart) > 0 {
		return svc.matchMultipart(raw.Multipart, props)
	}
	return nil, fmt.Errorf("blockstate %s: no variants or multipart data", source)
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

	propStr := svc.propsKeyBuilder.Run(props)

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

type rawMultipartPart struct {
	When  json.RawMessage `json:"when,omitempty"`
	Apply rawApply        `json:"apply"`
}

type rawWhenAND struct {
	AND []map[string]interface{} `json:"AND"`
}

type rawApply struct {
	Model string  `json:"model"`
	X     float64 `json:"x,omitempty"`
	Y     float64 `json:"y,omitempty"`
}

func (svc *BlockstateParser) matchMultipart(parts []json.RawMessage, props map[string]string) ([]blockstateMatch, error) {
	var matches []blockstateMatch
	for _, rawPart := range parts {
		var part rawMultipartPart
		if err := json.Unmarshal(rawPart, &part); err != nil {
			return nil, fmt.Errorf("parse multipart part: %w", err)
		}
		if !svc.matchWhen(part.When, props) {
			continue
		}
		matches = append(matches, blockstateMatch{
			Model: part.Apply.Model,
			RotX:  part.Apply.X,
			RotY:  part.Apply.Y,
		})
	}

	if len(matches) == 0 {
		propStr := svc.propsKeyBuilder.Run(props)
		return nil, fmt.Errorf("no multipart conditions match [%s]", propStr)
	}

	return matches, nil
}

func (svc *BlockstateParser) matchWhen(rawWhen json.RawMessage, props map[string]string) bool {
	if len(rawWhen) == 0 {
		return true
	}

	var andWhen rawWhenAND
	if err := json.Unmarshal(rawWhen, &andWhen); err == nil && len(andWhen.AND) > 0 {
		for _, cond := range andWhen.AND {
			if !svc.matchSimpleWhen(cond, props) {
				return false
			}
		}
		return true
	}

	var simple map[string]interface{}
	if err := json.Unmarshal(rawWhen, &simple); err != nil {
		return false
	}
	return svc.matchSimpleWhen(simple, props)
}

func (svc *BlockstateParser) matchSimpleWhen(cond map[string]interface{}, props map[string]string) bool {
	for k, v := range cond {
		sv := fmt.Sprintf("%v", v)
		if props[k] != sv {
			return false
		}
	}
	return true
}
