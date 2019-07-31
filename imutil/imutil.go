package imutil

import (
	"image"
	"image/color"
)

func Width(im image.Image) int {
	return im.Bounds().Max.X - im.Bounds().Min.X
}
func Height(im image.Image) int {
	return im.Bounds().Max.Y - im.Bounds().Min.Y
}

func Pix(im image.Image, i, j int) (uint8, uint8, uint8, uint8) {
	c := im.At(i, j).(color.RGBA)
	return c.R, c.G, c.B, c.A
}

func SetPix(im image.Image, i, j int, r, g, b, a uint8) {
	img := im.(*image.RGBA)
	w := Width(im)
	img.Pix[4*(w*j+i)] = r
	img.Pix[4*(w*j+i)+1] = g
	img.Pix[4*(w*j+i)+2] = b
	img.Pix[4*(w*j+i)+3] = a
}
