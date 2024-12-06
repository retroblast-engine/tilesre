package tilesre

import (
	"fmt"
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
)

// Tile represents an 8x8 image used to build backgrounds or moving objects.
type Tile struct {
	ID           int
	Image        *ebiten.Image
	HasAnimation bool
	HasCustomCol bool
	Shape        CustomShape
	Properties   map[string]Property
}

// Property represents a single property with its name, value, and type.
type Property struct {
	Name  string
	Value string
	Type  string
}

// CustomShape represents a custom collision shape for a tile.
type CustomShape struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// tileToImage converts a tiled.LayerTile to an ebiten.Image.
func (m *Map) tileToImage(tile *tiled.LayerTile) (*ebiten.Image, error) {
	if m.TilesetImage == nil {
		return nil, nil
	}

	tileX := (int(tile.ID)%m.Tileset.Columns)*(m.Tileset.TileWidth+m.Tileset.Spacing) + m.Tileset.Margin
	tileY := (int(tile.ID)/m.Tileset.Columns)*(m.Tileset.TileHeight+m.Tileset.Spacing) + m.Tileset.Margin

	// Extract the tile image from the tileset image
	tileImage := m.TilesetImage.SubImage(image.Rect(tileX, tileY, tileX+m.Tileset.TileWidth, tileY+m.Tileset.TileHeight)).(*ebiten.Image)

	return tileImage, nil
}

// tilePosition calculates the position of a tile in the map.
func (m *Map) tilePosition(num int) (int, int) {
	x := (num % m.TiledMap.Width) * m.TiledMap.TileWidth
	y := (num / m.TiledMap.Width) * m.TiledMap.TileHeight

	return x, y
}

func parseProperties(properties tiled.Properties) map[string]Property {
	result := make(map[string]Property)
	for _, prop := range properties {
		result[prop.Name] = Property{
			Name:  prop.Name,
			Value: prop.Value,
			Type:  prop.Type,
		}
	}
	return result
}

// processTiles processes custom collision shapes and animations for tiles.
func (m *Map) processTiles() error {
	// Parsing tiles from TMX that have Properties
	for _, customTile := range m.Tileset.Tiles {
		tile := Tile{
			ID:           int(customTile.ID),
			HasCustomCol: false,
			Properties:   parseProperties(customTile.Properties),
		}

		tileImage, err := m.tileToImage(&tiled.LayerTile{ID: uint32(tile.ID)})
		if err != nil {
			return fmt.Errorf("error converting tile to image: %w", err)
		}
		tile.Image = tileImage

		m.Tiles[int(customTile.ID)] = tile

	}

	// Parsing tiles from TMX that have custom collision data
	for _, customTile := range m.Tileset.Tiles {
		if len(customTile.ObjectGroups) > 0 {
			for _, obj := range customTile.ObjectGroups {
				object0 := obj.Objects[0]
				customShape := CustomShape{
					X:      float64(object0.X),
					Y:      float64(object0.Y),
					Width:  float64(object0.Width),
					Height: float64(object0.Height),
				}

				tile := Tile{
					ID:           int(customTile.ID),
					HasCustomCol: true,
					Shape:        customShape,
					Properties:   parseProperties(customTile.Properties),
				}

				tileImage, err := m.tileToImage(&tiled.LayerTile{ID: uint32(tile.ID)})
				if err != nil {
					return fmt.Errorf("error converting tile to image: %w", err)
				}
				tile.Image = tileImage

				m.Tiles[int(customTile.ID)] = tile
			}
		}
	}

	// Parsing tiles from TMX that have Animation
	for _, animatedTile := range m.Tileset.Tiles {
		if len(animatedTile.Animation) > 0 {
			tile := Tile{
				ID:           int(animatedTile.ID),
				HasAnimation: true,
				Properties:   parseProperties(animatedTile.Properties), // Parse properties
			}

			tileImage, err := m.tileToImage(&tiled.LayerTile{ID: uint32(tile.ID)})
			if err != nil {
				return fmt.Errorf("error converting animated tile to image: %w", err)
			}
			tile.Image = tileImage

			m.Tiles[int(animatedTile.ID)] = tile

			animation := &Animation{
				Frames:     make([]int, len(animatedTile.Animation)),
				Index:      0,
				Duration:   make([]time.Duration, len(animatedTile.Animation)),
				LastChange: time.Now(),
			}

			for i, v := range animatedTile.Animation {
				if _, ok := m.Tiles[int(v.TileID)]; !ok {
					tile := Tile{
						ID:           int(v.TileID),
						HasAnimation: false,
					}

					tileImage, err := m.tileToImage(&tiled.LayerTile{ID: v.TileID})
					if err != nil {
						return fmt.Errorf("error converting animation frame to image: %w", err)
					}
					tile.Image = tileImage

					if tile.Image == nil {
						return fmt.Errorf("tile image is nil for tile ID: %d", v.TileID)
					}

					m.Tiles[int(v.TileID)] = tile
				}

				animation.Frames[i] = int(v.TileID)
				animation.Duration[i] = time.Duration(v.Duration) * time.Millisecond
			}

			m.AnimatedTiles[int(animatedTile.ID)] = animation
		}
	}

	return nil
}
