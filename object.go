package tilesre

import "github.com/solarlune/resolv"

// Object represents a tile that can move independently from the background.
type Object struct {
	Physics *resolv.Object
	Sprite  *Sprite
}

// Sprite represents an individual sprite in a scene.
type Sprite struct {
	X, Y       int
	TileID     int
	Attributes []Option
}

// Option represents an attribute of an Object.
type Option struct {
	IsBehind     bool
	XFlip, YFlip bool
}

func (m *Map) createObject(x, y int, tile Tile, layerName string) Object {
	if tile.HasCustomCol {
		tileObject := resolv.NewObject(float64(x)+tile.Shape.X, float64(y)+tile.Shape.Y, tile.Shape.Width, tile.Shape.Height, layerName)
		return Object{
			Physics: tileObject,
			Sprite: &Sprite{
				X:      x,
				Y:      y,
				TileID: tile.ID,
			},
		}
	}

	tileObject := resolv.NewObject(float64(x), float64(y), float64(m.TiledMap.TileWidth), float64(m.TiledMap.TileHeight), layerName)
	return Object{
		Physics: tileObject,
		Sprite: &Sprite{
			X:      x,
			Y:      y,
			TileID: tile.ID,
		},
	}
}

func (m *Map) AddObject(o Object) {
	m.Objects = append(m.Objects, o)
}