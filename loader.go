package tilesre

import (
	"fmt"
	"image/png"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
)

// Load initializes and loads the map from the given paths
func Load(path, assetsTiledPath, assetsAsepritePath string, cellWidth, cellHeight int) (*Map, error) {
	m := &Map{
		Tiles:         make(map[int]Tile),
		AnimatedTiles: make(map[int]*Animation),
	}

	var err error
	m.TiledMap, err = tiled.LoadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error parsing map: %w", err)
	}

	m.createSpace(cellWidth, cellHeight)

	if len(m.TiledMap.Tilesets) == 0 {
		return nil, fmt.Errorf("no tilesets found in the map")
	}

	m.Tileset = m.TiledMap.Tilesets[0]
	file, err := os.Open(assetsTiledPath + m.Tileset.Image.Source)
	if err != nil {
		return nil, fmt.Errorf("error opening tileset image: %w", err)
	}
	defer file.Close()

	tmpTilesetImage, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("error decoding tileset image: %w", err)
	}

	transColor, err := hexToRGBA(m.Tileset.Image.Trans.String())
	if err != nil {
		return nil, fmt.Errorf("error converting hex to RGBA: %w", err)
	}
	tmpTilesetImage = replaceColor(tmpTilesetImage, transColor)

	m.TilesetImage = ebiten.NewImageFromImage(tmpTilesetImage)

	if err := m.processTiles(); err != nil {
		return nil, err
	}

	return m, nil
}
