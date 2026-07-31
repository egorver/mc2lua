package minecraft

import "fmt"

type AssetScanner struct {
	fs fsApi
}

func NewAssetScanner(fs fsApi) *AssetScanner {
	return &AssetScanner{fs: fs}
}

func (svc *AssetScanner) Run(root string) (map[string][]string, error) {
	modDirs, err := svc.fs.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", root, err)
	}

	nsToRoots := make(map[string][]string)

	for _, modDir := range modDirs {
		if !modDir.IsDir() {
			continue
		}
		svc.collectNamespaces(root+"/"+modDir.Name(), nsToRoots)
	}

	if len(nsToRoots) == 0 {
		return nil, fmt.Errorf("no Minecraft assets found in %s", root)
	}

	return nsToRoots, nil
}

func (svc *AssetScanner) collectNamespaces(modPath string, nsToRoots map[string][]string) {
	nsDirs, err := svc.fs.ReadDir(modPath)
	if err != nil {
		return
	}
	for _, nsDir := range nsDirs {
		if !nsDir.IsDir() {
			continue
		}
		nsPath := modPath + "/" + nsDir.Name()
		if svc.hasAssets(nsPath) {
			nsToRoots[nsDir.Name()] = append(nsToRoots[nsDir.Name()], nsPath)
		}
	}
}

func (svc *AssetScanner) hasAssets(nsPath string) bool {
	entries, err := svc.fs.ReadDir(nsPath)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() && (e.Name() == BlockstatesDir || e.Name() == ModelsDir) {
			return true
		}
	}
	return false
}
