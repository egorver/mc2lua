package minecraft

import "mc2lua/internal/runtime"

const testBlockstateDefault = `{"variants":{"":{"model":"block/cube"}}}`
const testBlockstateVariants = `{"variants":{"facing=north":{"model":"block/furnace"},"facing=south":{"model":"block/furnace","y":180},"facing=east,lit=true":{"model":"block/furnace_on","y":90}}}`
const testBlockstateArray = `{"variants":{"":[{"model":"block/grass"},{"model":"block/grass_alt"}]}}`
const testBlockstateEmptyArray = `{"variants":{"":[ ]}}`
const testBlockstateMultipart = `{"multipart":[{"apply":{"model":"block/block"}},{"when":{"north":"true"},"apply":{"model":"block/side"}},{"when":{"AND":[{"age":"0"},{"rooted":"false"}]},"apply":{"model":"block/plant","x":90}}]}`
const testBlockstateEmptyVariants = `{"variants":{}}`
const testBlockstateNoVariants = `{}`
const testBlockstateInvalidJSON = `{bad json`

const testModelCube = `{"elements":[{"from":[0,0,0],"to":[16,16,16],"shade":true}],"textures":{"particle":"block/particle"}}`
const testModelGrass = `{"parent":"block/cube","textures":{"all":"block/grass_side"}}`
const testModelGrandchild = `{"parent":"block/grass","textures":{"grass":"block/grass_top"}}`
const testModelNoElements = `{"textures":{"particle":"block/stone"}}`
const testModelNoTextures = `{"elements":[{"from":[0,0,0],"to":[16,16,16],"shade":true}]}`
const testModelInvalidJSON = `{bad json`
const testModelImplicitShade = `{"elements":[{"from":[0,0,0],"to":[16,16,16]}]}`

func setupTestFS() (*runtime.FSMock, map[string][]string) {
	fs := runtime.NewFSMock()
	fs.AddDir("assets", 0755)
	fs.AddDir("assets/minecraft", 0755)
	fs.AddDir("assets/minecraft/blockstates", 0755)
	fs.AddDir("assets/minecraft/models", 0755)
	return fs, map[string][]string{"minecraft": {"assets/minecraft"}}
}

func addBlockstate(fs *runtime.FSMock, ns, blockID string, data []byte) {
	fs.AddFile("assets/"+ns+"/blockstates/"+blockID+".json", data, 0644)
}

func addModel(fs *runtime.FSMock, ns, path string, data []byte) {
	fs.AddFile("assets/"+ns+"/models/"+path+".json", data, 0644)
}
