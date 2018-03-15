package client

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
)

// LoadTextures loads textures from the textures directory into a single texture image
func LoadTextures() *image.RGBA {
	files := []string{
		"grass-side",
		"grass-top",
		"dirt",
		"stone",
		"asteroid",
		"moon",
		"sun",
	}
	rgba := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for x := 0; x < len(files); x++ {
		ImageFile, err := os.Open(fmt.Sprintf("textures/%s.png", files[x]))
		if err != nil {
			panic(err)
		}
		img, err := png.Decode(ImageFile)
		if err != nil {
			panic(err)
		}
		sx := (x % 4) * 16
		sy := (x / 4) * 16
		draw.Draw(rgba, image.Rect(sx, sy, sx+16, sy+16), img, image.Pt(0, 0), draw.Src)
		ImageFile.Close()
	}
	return rgba

}
