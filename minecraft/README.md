# minecraft/

This directory contains unpacked **Minecraft mods and vanilla assets** — models, textures, and other resources used for conversion to Roblox Lua.

## Usage

1. Extract your Minecraft version JAR or mod JARs (e.g. with any archive tool)
2. Place the extracted directories into this folder
3. The project will read models, textures, and block data to generate Roblox counterparts

## Directory structure

```
minecraft/
├── minecraft-<version>-<build>/   # Vanilla Minecraft
│   └── minecraft/
│       ├── blockstates/
│       ├── models/
│       │   ├── block/
│       │   └── item/
│       ├── textures/
│       │   ├── block/
│       │   └── item/
│       └── lang/
├── <mod-name>-<version>/          # Mod (extracted)
│   ├── <mod_id>/
│   │   ├── blockstates/
│   │   ├── models/
│   │   │   ├── block/
│   │   │   └── item/
│   │   ├── textures/
│   │   │   ├── block/
│   │   │   └── item/
│   │   └── lang/
│   └── ...
└── ...
```

## Notes

Only assets (`.json` models, `.png` textures) are used during conversion.  
Mod `.jar` files must be **extracted** before placing them here.
