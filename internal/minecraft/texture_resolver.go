package minecraft

import "strings"

type TextureResolver struct{}

func NewTextureResolver() *TextureResolver {
	return &TextureResolver{}
}

func (svc *TextureResolver) Run(textures map[string]string) map[string]string {
	result := make(map[string]string, len(textures))

	for k, v := range textures {
		resolved := v
		visited := map[string]bool{k: true}

		for strings.HasPrefix(resolved, "#") {
			refKey := resolved[1:]
			if visited[refKey] {
				resolved = v
				break
			}

			visited[refKey] = true
			next, ok := textures[refKey]
			if !ok {
				resolved = v
				break
			}

			resolved = next
		}

		result[k] = resolved
	}

	return result
}
