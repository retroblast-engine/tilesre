package tilesre

import (
	"fmt"

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

func (m *Map) createSpace(spaceCellWidth, spaceCellHeight int) {
	spaceWidth := m.TiledMap.Width * m.TiledMap.TileWidth
	spaceHeight := m.TiledMap.Height * m.TiledMap.TileHeight
	m.Space = resolv.NewSpace(spaceWidth, spaceHeight, spaceCellWidth, spaceCellHeight)
}

func (m *Map) ProcessLayer(index int, layer *tiled.Layer) {
	fmt.Printf("Layer %d: %s\n", index, layer.Name)
	countTiles := len(layer.Tiles)
	fmt.Printf("Layer %d has %d tiles\n", index, countTiles)
	m.verifyLayerDimensions(index)

	// Iterate over the tiles
	for num, tile := range layer.Tiles {

		myTile := Tile{ID: int(tile.ID)}

		// Add the tile to the space, if it is not nil
		if !tile.IsNil() {

			// 1. If the tile is not already in the map, add it
			if _, ok := m.Tiles[int(tile.ID)]; !ok {

				// Get the image of the tile
				tileImage, err := m.tileToImage(tile)
				if err != nil {
					panic(err)
				}

				// Create the tile
				myTile.Image = tileImage

				// Add the tile to the map
				m.Tiles[int(tile.ID)] = myTile
			}

			// 2. Add the object to the space

			// Get the position of the tile
			x, y := m.tilePosition(num)

			// Create the object
			o := m.createObject(x, y, m.Tiles[int(tile.ID)], layer.Name)

			// Add object and tile
			m.AddObject(o)
		}
	}
}

func (m *Map) verifyLayerDimensions(index int) {
	rows := m.TiledMap.Height
	columns := m.TiledMap.Width
	if rows != m.TiledMap.Height || columns != m.TiledMap.Width {
		panic("Layer dimensions do not match map dimensions")
	}

	fmt.Printf("Layer %d has %d rows and %d columns\n", index, rows, columns)
}

// MapDraw draws the map on the screen.
func (m *Map) MapDraw(screen *ebiten.Image, camX, camY float64) {
	// Draw the background image
	// ---------------------------------------------------------- //

	// Draw the background image, if it exists
	if m.BackgroundImage != nil {
		op := &ebiten.DrawImageOptions{}

		// Get the original width and height of the image
		originalWidth := m.BackgroundImage.Bounds().Dx()
		// originalHeight := backgroundImage.Bounds().Dy()

		// Calculate the scaling factors for the X and Y axes
		scaleX := 1000.0 / float64(originalWidth)
		// scaleY := 0 / float64(originalHeight)

		// Apply the scaling factors to the X and Y axes
		op.GeoM.Scale(scaleX, 1)

		// Translate the image to the desired position (optional)
		op.GeoM.Translate(0, m.BackgroundImageOffsetY)
		op.GeoM.Translate(camX, 0)

		// Draw the image with the applied scaling
		screen.DrawImage(m.BackgroundImage, op)
	}
	// End Draw the background image
	// ---------------------------------------------------------- //

	// Start Draw the game objects
	// ---------------------------------------------------------- //
	// Loop through all of the game objects and draw their sprites
	for _, o := range m.Objects {

		id := o.Sprite.TileID
		img := m.Tiles[id].Image
		if m.Tiles[id].HasAnimation {
			// Check if it exists in the animation map
			if _, ok := m.AnimatedTiles[id]; ok {
				// Get the next frame
				anim := m.AnimatedTiles[id]
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

	// End Draw the game objects
	// ---------------------------------------------------------- //
}
