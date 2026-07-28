# region/

This directory contains **Minecraft Anvil region files** (`.mca`) — fragments of a Minecraft world map.

## Usage

1. Locate your Minecraft world saves (typically in `~/.minecraft/saves/<world>/region/`)
2. Copy the desired `.mca` region files into this directory
3. The project will read these files and convert the world data into Roblox Lua code

## Format

Each `.mca` file stores a 32×32 chunk region of the Minecraft world.
Files follow the standard naming convention: `r.<regionX>.<regionZ>.mca`
