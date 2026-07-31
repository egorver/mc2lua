package minecraft

const (
	BlocksPerChunkSection = 16
	BlocksPerSection      = 4096
	ChunkSectionXMask     = 15
	ChunkSectionZShift    = 4
	ChunkSectionYShift    = 8
	RegionSize            = 512
	ChunksPerRegion       = 32
	DefaultNamespace      = "minecraft"
	NamespaceSeparator    = ":"
	MinecraftNamespacePrefix = "minecraft:"
	RegionFilePattern     = "r.*.*.mca"
	RegionFileFormat      = "r.%d.%d.mca"
	BlockstatesDir        = "blockstates"
	ModelsDir             = "models"
	BlockstatesDirPath    = "/blockstates/"
	ModelsDirPath         = "/models/"
	TextureDirPath        = "/textures/"
	PNGExtension          = ".png"
	JSONExtension         = ".json"
)
