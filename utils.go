package tilesre

import (
	"fmt"
	"image"
	"image/color"
)

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
