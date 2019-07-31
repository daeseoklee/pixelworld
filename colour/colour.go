package colour

import (
	"image"
	"image/color"
	"math/rand"

	"github.com/daeseoklee/pixelworld/imutil"
)

//Colour : color used in the world
type Colour struct {
	M float64
	R float64
	G float64
	B float64
}

//RandColour : random colour
func RandColour() Colour {
	return Colour{M: rand.Float64(), R: rand.Float64(), G: rand.Float64(), B: rand.Float64()}
}

//RGBA : Colour to color.RGBA
func (c Colour) RGBA() color.RGBA {
	return color.RGBA{uint8(255 * c.R), uint8(255 * c.G), uint8(255 * c.B), 255}
}

//ToImage : colour slice to RGBA image
func ToImage(d [][]Colour, k int) *image.RGBA {
	m, n := len(d), len(d[0])
	im := image.NewRGBA(image.Rect(0, 0, m*k, n*k))
	var c Colour
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			c = d[i][j]
			for a := 0; a < k; a++ {
				for b := 0; b < k; b++ {
					imutil.SetPix(im, k*i+a, k*j+b, uint8(255*c.R), uint8(255*c.G), uint8(255*c.B), 255)
				}
			}
		}
	}
	return im
}
