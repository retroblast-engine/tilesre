package maploader

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
	"github.com/solarlune/resolv"
)

type Map struct {
	// Tiled Variables
	TiledMap               *tiled.Map
	Tileset                *tiled.Tileset
	TilesetImage           *ebiten.Image
	Tiles                  map[int]Tile
	AnimatedTiles          map[int]*Animation
	BackgroundImage        *ebiten.Image
	BackgroundImageOffsetX float64
	BackgroundImageOffsetY float64

	// Space Resolv Variables
	Objects []Object
	Space   *resolv.Space

	// Draw the map variables
	CameraX, CameraY float64
}

func Load(path, assetsTiledPath, assetsAsepritePath string) (*Map, error) {
	// Step 1: Initialize the map and tileset variables
	var m = &Map{}
	var err error

	m.Tiles = make(map[int]Tile)
	m.AnimatedTiles = make(map[int]*Animation)
	m.TiledMap, err = tiled.LoadFile(path)
	if err != nil {
		fmt.Printf("error parsing map: %s", err.Error())
		os.Exit(2)
	}

	// Step 2. Create space resolv for the map
	m.createSpace()

	// Step 3. Get the tileset (you must have saved it separately in the Tiled editor, and not embedded)
	// it uses only the the first tileset for now.
	if len(m.TiledMap.Tilesets) == 0 {
		fmt.Printf("no tilesets found in the map")
		os.Exit(2)
	} else {
		// Load the tileset image
		m.Tileset = m.TiledMap.Tilesets[0]
		file, err := os.Open(assetsTiledPath + m.Tileset.Image.Source)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		var tmpTilesetImage image.Image
		tmpTilesetImage, err = png.Decode(file)
		if err != nil {
			fmt.Println("Error decoding tileset image, file:", m.Tileset.Image.Source)
			os.Exit(2)
		} else {
			// Treat the transparent color chosen by Tiled, as a transparent color.
			transparencyColorinHex := m.Tileset.Image.Trans.String()
			transColor, err := hexToRGBA(transparencyColorinHex)
			if err != nil {
				fmt.Printf("Error converting hex to RGBA: %v", err)
				os.Exit(2)
			} else {
				tmpTilesetImage = replaceColor(tmpTilesetImage, transColor)
			}
		}

		m.TilesetImage = ebiten.NewImageFromImage(tmpTilesetImage)
	}

	// Step 4. Take care of animated tiles, if any exist
	if len(m.Tileset.Tiles) > 0 {
		// Which tiles have custom collision shapes?
		for _, customTile := range m.Tileset.Tiles {
			if len(customTile.ObjectGroups) > 0 {
				for _, obj := range customTile.ObjectGroups {
					object0 := obj.Objects[0]
					x0 := float64(object0.X)
					y0 := float64(object0.Y)
					w := float64(object0.Width)
					h := float64(object0.Height)
					// name := object0.Name
					customShape := CustomShape{
						X:      x0,
						Y:      y0,
						Width:  w,
						Height: h,
					}

					tile := Tile{
						ID:           int(customTile.ID),
						HasCustomCol: true,
						Shape:        customShape,
					}

					tileImage, err := m.tileToImage(&tiled.LayerTile{ID: uint32(tile.ID)})
					if err != nil {
						panic(err)
					}

					tile.Image = tileImage

					m.Tiles[int(customTile.ID)] = tile
				}
			}
		}

		// Which tiles have animations?
		for _, animatedTile := range m.Tileset.Tiles {
			if len(animatedTile.Animation) > 0 {

				// ------ Create and add the tile ------ //
				tile := Tile{
					ID:           int(animatedTile.ID),
					HasAnimation: true,
				}

				tileImage, err := m.tileToImage(&tiled.LayerTile{ID: uint32(tile.ID)})
				if err != nil {
					panic(err)
				}

				tile.Image = tileImage

				m.Tiles[int(animatedTile.ID)] = tile

				// ------ Create and add the animation ------ //
				// This tile has other tiles as animations, lets create those tiles:

				animation := &Animation{
					Frames:     make([]int, len(animatedTile.Animation)),
					Index:      0,
					Duration:   make([]time.Duration, len(animatedTile.Animation)),
					LastChange: time.Now(),
				}

				for i, v := range animatedTile.Animation {

					// Create the tile
					if _, ok := m.Tiles[int(v.TileID)]; !ok {
						tile := Tile{
							ID:           int(v.TileID),
							HasAnimation: false,
						}

						tileImage, err := m.tileToImage(&tiled.LayerTile{ID: v.TileID})
						if err != nil {
							panic(err)
						}

						tile.Image = tileImage

						// check if tile image is nill
						if tile.Image == nil {
							panic("tile image is nil")
						}

						m.Tiles[int(v.TileID)] = tile
					}

					animation.Frames[i] = int(v.TileID)
					animation.Duration[i] = time.Duration(v.Duration) * time.Millisecond

				}

				// Add to the animation map
				m.AnimatedTiles[int(animatedTile.ID)] = animation
			}
		}
	}

	// Step 5 and 6 are in the Init() function of the World1x1 scene

	return m, nil
}

func (m *Map) Draw(screen *ebiten.Image) {
	// Map draw logic
}
