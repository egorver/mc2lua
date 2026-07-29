package minecraft

import (
	"encoding/json"
	"fmt"
	"mc2lua/internal/model"
	"strings"
)

type ModelParser struct {
	fs fsApi
}

func NewModelParser(fs fsApi) *ModelParser {
	return &ModelParser{fs: fs}
}

type flattenedModel struct {
	Elements []model.ModelElement
	Textures map[string]string
}

func (m *flattenedModel) merge(other *flattenedModel) {
	m.Elements = append(m.Elements, other.Elements...)
	for k, v := range other.Textures {
		m.Textures[k] = v
	}
}

func (svc *ModelParser) Run(modelName string, namespaces map[string][]string) (*flattenedModel, error) {
	return svc.flatten(modelName, namespaces)
}

func (svc *ModelParser) flatten(modelName string, namespaces map[string][]string) (*flattenedModel, error) {
	raw, err := svc.readRaw(modelName, namespaces)
	if err != nil {
		return nil, fmt.Errorf("model %s: %w", modelName, err)
	}

	result := &flattenedModel{
		Elements: nil,
		Textures: make(map[string]string),
	}

	if raw.Parent != "" {
		parent, err := svc.flatten(raw.Parent, namespaces)
		if err != nil {
			return nil, fmt.Errorf("model %s parent %s: %w", modelName, raw.Parent, err)
		}
		result.merge(parent)
	}

	result.merge(&flattenedModel{
		Elements: svc.parseElements(raw.Elements),
		Textures: raw.Textures,
	})

	if len(result.Elements) == 0 && len(result.Textures) > 0 {
		result.Elements = []model.ModelElement{
			{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true},
		}
	}

	return result, nil
}

type rawModel struct {
	Parent   string            `json:"parent,omitempty"`
	Textures map[string]string `json:"textures,omitempty"`
	Elements []json.RawMessage `json:"elements,omitempty"`
}

func (m *rawModel) normalize() {
	if m.Elements == nil {
		m.Elements = []json.RawMessage{}
	}
	if m.Textures == nil {
		m.Textures = make(map[string]string)
	}
}

func (svc *ModelParser) readRaw(modelName string, nsToRoots map[string][]string) (*rawModel, error) {
	ns, path := svc.splitModelName(modelName)
	roots, ok := nsToRoots[ns]
	if !ok {
		return nil, fmt.Errorf("unknown namespace %s", ns)
	}
	var paths []string
	for _, root := range roots {
		paths = append(paths, root+"/models/"+path+".json")
	}
	lastErr := fmt.Errorf("namespace %s: no files found in %v", ns, paths)
	for _, filePath := range paths {
		data, err := svc.fs.ReadFile(filePath)
		if err != nil {
			lastErr = fmt.Errorf("read %s: %w", filePath, err)
			continue
		}

		var raw rawModel
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("parse %s: %w", filePath, err)
		}

		raw.normalize()
		return &raw, nil
	}
	return nil, lastErr
}

func (svc *ModelParser) splitModelName(name string) (namespace, path string) {
	if strings.Contains(name, ":") {
		parts := strings.SplitN(name, ":", 2)
		return parts[0], parts[1]
	}
	return "minecraft", name
}

type rawElement struct {
	Shade    *bool                          `json:"shade"`
	From     model.Vector3                  `json:"from"`
	To       model.Vector3                  `json:"to"`
	Rotation *model.ElementRotation         `json:"rotation,omitempty"`
	Faces    map[string]model.ElementFace   `json:"faces,omitempty"`
}

func (svc *ModelParser) parseElements(raws []json.RawMessage) []model.ModelElement {
	out := make([]model.ModelElement, 0, len(raws))
	for _, rawElem := range raws {
		var re rawElement
		json.Unmarshal(rawElem, &re)
		out = append(out, model.ModelElement{
			From:     re.From,
			To:       re.To,
			Rotation: re.Rotation,
			Shade:    re.Shade == nil || *re.Shade,
			Faces:    re.Faces,
		})
	}
	return out
}
