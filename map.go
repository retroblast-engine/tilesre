package tilesre

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
	"github.com/solarlune/resolv"
)

// Map represents the game map with all its components
type Map struct {
	TiledMap               *tiled.Map
	Tileset                *tiled.Tileset
	TilesetImage           *ebiten.Image
	Tiles                  map[int]Tile
	AnimatedTiles          map[int]*Animation
	BackgroundImage        *ebiten.Image
	BackgroundImageOffsetX float64
	BackgroundImageOffsetY float64
	Objects                []Object
	Space                  *resolv.Space
	CameraX, CameraY       float64
}

// createSpace initializes the collision space
func (m *Map) createSpace(spaceCellWidth, spaceCellHeight int) {
	spaceWidth := m.TiledMap.Width * m.TiledMap.TileWidth
	spaceHeight := m.TiledMap.Height * m.TiledMap.TileHeight
	m.Space = resolv.NewSpace(spaceWidth, spaceHeight, spaceCellWidth, spaceCellHeight)
}

// ProcessLayer processes a single layer of the map
func (m *Map) ProcessLayer(index int, layer *tiled.Layer) error {
	if err := m.verifyLayerDimensions(); err != nil {
		return fmt.Errorf("error verifying layer dimensions: %w", err)
	}

	// Iterate over the tiles
	for num, tile := range layer.Tiles {
		myTile := Tile{ID: int(tile.ID)}

		// Add the tile to the space, if it is not nil
		if !tile.IsNil() {
			// If the tile is not already in the map, add it
			if _, ok := m.Tiles[int(tile.ID)]; !ok {
				// Get the image of the tile
				tileImage, err := m.tileToImage(tile)
				if err != nil {
					return fmt.Errorf("error converting tile to image: %w", err)
				}

				// Create the tile
				myTile.Image = tileImage

				// Add the tile to the map
				m.Tiles[int(tile.ID)] = myTile
			}

			// Add the object to the space
			// Get the position of the tile
			x, y := m.tilePosition(num)

			// Create the object
			o := m.createObject(x, y, m.Tiles[int(tile.ID)], layer.Name, m.Tiles[int(tile.ID)].Properties)

			// Add object and tile
			m.AddObject(o)
		}
	}
	return nil
}

// verifyLayerDimensions checks if the layer dimensions match the map dimensions
func (m *Map) verifyLayerDimensions() error {
	rows := m.TiledMap.Height
	columns := m.TiledMap.Width
	if rows != m.TiledMap.Height || columns != m.TiledMap.Width {
		return fmt.Errorf("layer dimensions do not match map dimensions")
	}
	return nil
}

// MapDraw draws the map on the screen
func (m *Map) MapDraw(screen *ebiten.Image, camX, camY float64) {
	// Draw the background image
	if m.BackgroundImage != nil {
		op := &ebiten.DrawImageOptions{}

		// Get the original width and height of the image
		originalWidth := m.BackgroundImage.Bounds().Dx()

		// Calculate the scaling factors for the X and Y axes
		scaleX := 1000.0 / float64(originalWidth)

		// Apply the scaling factors to the X and Y axes
		op.GeoM.Scale(scaleX, 1)

		// Translate the image to the desired position (optional)
		op.GeoM.Translate(0, m.BackgroundImageOffsetY)
		op.GeoM.Translate(camX, 0)

		// Draw the image with the applied scaling
		screen.DrawImage(m.BackgroundImage, op)
	}

	// Draw the game objects
	for _, o := range m.Objects {
		id := o.Sprite.TileID

		if _, ok := m.Tiles[id]; !ok {
			fmt.Println("Error: The tile", id, "is not used in the map, so there is no object associated with it yet")
			os.Exit(1)
		}

		img := m.Tiles[id].Image
		if m.Tiles[id].HasAnimation {
			// Check if it exists in the animation map
			if anim, ok := m.AnimatedTiles[id]; ok {
				// Get the next frame
				tileID := anim.NextFrame()
				img = m.Tiles[tileID].Image
			}
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(o.Sprite.X), float64(o.Sprite.Y))

		// If the object is part of HUD, do not translate it based on the camera
		// so it gets redrawn always at the same position
		if !o.Physics.HasTags("hud") {
			op.GeoM.Translate(camX, 0)
		}
		screen.DrawImage(img, op)
	}
}
