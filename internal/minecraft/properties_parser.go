package minecraft

import (
	"encoding/json"
	"fmt"
)

type PropertiesParser struct{}

func NewPropertiesParser() *PropertiesParser {
	return &PropertiesParser{}
}

func (svc *PropertiesParser) Run(jsonStr string) map[string]string {
	props := make(map[string]string)
	if jsonStr == "" {
		return props
	}
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return props
	}
	for k, v := range raw {
		props[k] = fmt.Sprintf("%v", v)
	}
	return props
}
