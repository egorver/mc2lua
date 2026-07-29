package minecraft

import (
	"slices"
	"strings"
)

type PropsKeyBuilder struct{}

func NewPropsKeyBuilder() *PropsKeyBuilder {
	return &PropsKeyBuilder{}
}

func (svc *PropsKeyBuilder) Run(props map[string]string) string {
	if len(props) == 0 {
		return ""
	}
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(props[k])
	}
	return b.String()
}
