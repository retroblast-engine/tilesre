package maploader

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
	"github.com/solarlune/resolv"
)

type CustomShape struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// Tile represents an 8x8 image used to build backgrounds or moving objects.
type Tile struct {
	ID           int
	Image        *ebiten.Image
	HasAnimation bool
	HasCustomCol bool
	Shape        CustomShape
}

// Animation represents a series of tiles that make up an animation.
type Animation struct {
	Frames     []int
	Index      int
	Duration   []time.Duration // how long the current frame should be displayed
	LastChange time.Time       // is updated to the current time each time the frame changes
}

// --- Needed for Space Resolv --- //
// ------------------------------------------------------------------------------- //
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

/// ------------------------------------------------------------------------------- //

func hexToRGBA(hex string) (color.RGBA, error) {
	var r, g, b uint8
	if hex[0] == '#' {
		hex = hex[1:]
	}
	_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return color.RGBA{}, err
	}
	return color.RGBA{R: r, G: g, B: b, A: 0xFF}, nil
}

func replaceColor(img image.Image, oldColor color.Color) *image.RGBA {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	oldR, oldG, oldB, oldA := oldColor.RGBA()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if r == oldR && g == oldG && b == oldB && a == oldA {
				newImg.Set(x, y, color.Transparent)
			} else {
				newImg.Set(x, y, img.At(x, y))
			}
		}
	}

	return newImg
}

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

func (m *Map) tilePosition(num int) (int, int) {
	x := (num % m.TiledMap.Width) * m.TiledMap.TileWidth
	y := (num / m.TiledMap.Width) * m.TiledMap.TileHeight

	return x, y
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

func (m *Map) createSpace() {
	spaceWidth := m.TiledMap.Width * m.TiledMap.TileWidth
	spaceHeight := m.TiledMap.Height * m.TiledMap.TileHeight
	spaceCellWidth := 16
	spaceCellHeight := 16
	m.Space = resolv.NewSpace(spaceWidth, spaceHeight, spaceCellWidth, spaceCellHeight)
}

// MapDraw draws the map on the screen.
func (m *Map) MapDraw(screen *ebiten.Image, camX, camY float64) {
	// Draw the background image
	// ---------------------------------------------------------- //
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

// NextFrame returns the next frame of the animation and resets to the first frame if it's the last frame.
func (s *Animation) NextFrame() int {
	timePassed := time.Since(s.LastChange)
	if timePassed >= s.Duration[s.Index] {
		s.Index++
		if s.Index >= len(s.Frames) {
			s.Index = 0
		}
		s.LastChange = time.Now()
	}

	return s.Frames[s.Index]
}
