# mc2lua

Converts Minecraft Java Edition Anvil region files (`.mca`) into Lua scripts for
Roblox Studio. The generated script reconstructs the source world as a hierarchy
of Parts, preserving block geometry, color, and material of the original blocks.

## Overview

mc2lua is a command-line utility that processes Minecraft Java Edition world data
stored in the Anvil format. The application reads region files, resolves each block
against the corresponding Minecraft assets (blockstates, models, textures), and
produces a standalone Lua script. When executed in Roblox Studio, the script
reconstructs the world as a set of Parts with accurate dimensions, colors, and
materials.

The conversion pipeline supports:

- Full-resolution block reconstruction, including complex geometry (slabs, stairs,
  walls, multi-element models)
- Automated merging of identical adjacent blocks into single Parts, reducing the
  total number of instances
- Modded blocks, provided the corresponding assets are available
- Selective conversion of a bounded subregion of the world

## System Requirements

| Requirement | Value |
|---|---|
| Operating system | Windows, macOS, Linux |
| Go | 1.26 or later (required only for building) |
| Minecraft | Java Edition, Anvil format (`.mca`), approximately 1.13 and later |
| Assets | Extracted contents of the Minecraft version JAR; optional mod JARs |
| Roblox Studio | Any version capable of executing Luau in the Command Bar |

## Installation

Build the executable from source:

```bash
git clone https://github.com/yourname/mc2lua.git
cd mc2lua
go build -o mc2lua ./cmd/mc2lua
```

Alternatively, the application may be executed without compilation:

```bash
go run ./cmd/mc2lua [options]
```

## Quick Start

### 1. Prepare world data

Copy the region files (`.mca`) from the Minecraft save directory into the
`region/` folder:

```
%APPDATA%\.minecraft\saves\<world-name>\region\  →  region\
```

### 2. Prepare assets

Extract the contents of the Minecraft version JAR (and any mod JARs) into the
`assets/` directory. The JAR contents, not the archive itself, must be placed
in the directory.

### 3. Run the conversion

```bash
go run ./cmd/mc2lua
```

The output script is written to `output.lua` by default.

### 4. Import into Roblox Studio

1. Open a new empty place in Roblox Studio.
2. Open the Output window (View → Output).
3. Open the Command Bar (View → Command Bar).
4. Paste the contents of `output.lua` into the Command Bar and execute it.

The world is constructed progressively; progress is reported in the Output
window. The resulting hierarchy is placed under a Folder named `Imported` in
the Workspace.

## CLI Reference

| Option | Default | Description |
|---|---|---|
| `-input` | `region` | Path to the directory containing `.mca` region files |
| `-assets` | `assets` | Path to the extracted Minecraft assets directory |
| `-output` | `output.lua` | Output Lua script path |
| `-scale` | `4` | Block scale factor in studs |
| `-config` | `config` | Path to the configuration directory |
| `-no-offset` | `false` | Disables automatic vertical offset to y = 0 |
| `-xmin`, `-xmax` | unbounded | Inclusive bounds on the X axis |
| `-ymin`, `-ymax` | unbounded | Inclusive bounds on the Y axis |
| `-zmin`, `-zmax` | unbounded | Inclusive bounds on the Z axis |
| `-h`, `-help` | — | Displays usage information |

### Example

```bash
mc2lua \
  -input region \
  -assets assets \
  -output output.lua \
  -scale 4 \
  -xmin -2201 -xmax -2125 \
  -ymin 60 -ymax 110 \
  -zmin 2669 -zmax 2722
```

The example above converts the specified subregion of the world; coordinates
outside the bounds are omitted from the output.

## Configuration

### Parts (`config/parts.yaml`)

Per-block styling: surfaces (textures and tint colors) and, alongside
`transparency`, an optional top-level `color` field that sets an explicit base
color for the block. It is used for blocks whose color cannot be derived from a
texture (for example, transparent liquids) and is scaled by the brightness
factor of the part's final material from `materials.yaml`:

```yaml
parts:
  minecraft:glass:
    transparency: 0.5
    all:
      texture: rbxassetid://232395521
      color: [191, 191, 191]
  minecraft:tuff_bricks:
    all:
      texture: rbxassetid://187626492
      studs_per_tile: 12
  minecraft:water:
    color: [38, 94, 173]
```

The optional `studs_per_tile` field lives on a surface alongside its `texture`
and overrides the default texture tile size (the block `-scale` value) for that
surface in the generated Lua script.

### Materials (`config/materials.yaml`)

The file contains four sections controlling the Roblox `Enum.Material`
assignment:

- `mappings` — associates block identifiers with materials (for example,
  `planks: Wood`, `glass: Glass`)
- `suffixes` — lists block name suffixes (for example, `_slab`, `_stairs`,
  `_door`) whose material is derived from the base name
- `overrides` — explicit per-block material assignments
- `brightness` — per-material brightness factors applied to extracted colors

### Tints (`config/tints.yaml`)

Block faces whose model declares a `tintindex` are multiplied by a tint color.
`grass` and `foliage` colors are sampled from the vanilla biome colormaps
(`colormap/grass.png` and `colormap/foliage.png`) at a fixed plains climate;
`water` and `redstone` use fixed colors. Add mod blocks to this file so they
receive the same tinting:

```yaml
tints:
  minecraft:grass_block: grass
  minecraft:oak_leaves: foliage
```

## Output Format

The generated script consists of:

- A `createPart` helper function defining a Part with size, position, color,
  and material
- One invocation per merged cuboid or per model element for complex blocks
- Progress reporting at regular intervals during construction

Each Part is assigned two attributes:

| Attribute | Description |
|---|---|
| `original_block_id` | Identifier of the source block (for example, `minecraft:stone`) |
| `original_properties` | Block state properties of the source block |

## Limitations

- Bedrock Edition world data is not supported.
- Liquids (water, lava) are rendered as solid blocks of a single color.
- Entities, items, and tile entities are not converted.
- Biome data is read but does not affect the output.

## License

MIT
