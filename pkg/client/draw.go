package client

import (
	"image"
	"image/draw"
	"image/png"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jeffbaumes/gogame/pkg/geom"
)

var (
	square = []float32{
		-0.5, 0.5, 0.5,
		-0.5, -0.5, 0.5,
		0.5, -0.5, 0.5,

		-0.5, 0.5, 0.5,
		0.5, -0.5, 0.5,
		0.5, 0.5, 0.5,

		-0.5, 0.5, -0.5,
		-0.5, -0.5, -0.5,
		0.5, -0.5, -0.5,

		-0.5, 0.5, -0.5,
		0.5, -0.5, -0.5,
		0.5, 0.5, -0.5,

		0.5, -0.5, 0.5,
		0.5, -0.5, -0.5,
		0.5, 0.5, -0.5,

		0.5, -0.5, 0.5,
		0.5, 0.5, -0.5,
		0.5, 0.5, 0.5,

		-0.5, -0.5, 0.5,
		-0.5, -0.5, -0.5,
		-0.5, 0.5, -0.5,

		-0.5, -0.5, 0.5,
		-0.5, 0.5, -0.5,
		-0.5, 0.5, 0.5,

		0.5, 0.5, -0.5,
		-0.5, 0.5, -0.5,
		-0.5, 0.5, 0.5,

		0.5, 0.5, -0.5,
		-0.5, 0.5, 0.5,
		0.5, 0.5, 0.5,

		0.5, -0.5, -0.5,
		-0.5, -0.5, -0.5,
		-0.5, -0.5, 0.5,

		0.5, -0.5, -0.5,
		-0.5, -0.5, 0.5,
		0.5, -0.5, 0.5,
	}
	hudDrawable      uint32
	textDrawable     uint32
	textTextureValue uint32
)

func drawPlanet(p *geom.Planet) {
	for key, chunk := range p.Chunks {
		if !chunk.GraphicsInitialized {
			initChunkGraphics(chunk, p, key.Lon, key.Lat, key.Alt)
		}
		drawChunk(chunk)
	}
}

func initChunkGraphics(c *geom.Chunk, p *geom.Planet, lonIndex, latIndex, altIndex int) {
	cs := geom.ChunkSize
	points := []float32{}
	normals := []float32{}

	for cLon := 0; cLon < cs; cLon++ {
		for cLat := 0; cLat < cs; cLat++ {
			for cAlt := 0; cAlt < cs; cAlt++ {
				if c.Cells[cLon][cLat][cAlt].Material != geom.Air {
					pts := make([]float32, len(square))
					for i := 0; i < len(square); i += 3 {
						l := geom.CellLoc{
							Lon: float32(cs*lonIndex+cLon) + square[i+0],
							Lat: float32(cs*latIndex+cLat) + square[i+1],
							Alt: float32(cs*altIndex+cAlt) + square[i+2],
						}
						r, theta, phi := p.CellLocToSpherical(l)
						cart := mgl32.SphericalToCartesian(r, theta, phi)
						pts[i] = cart[0]
						pts[i+1] = cart[1]
						pts[i+2] = cart[2]
					}
					points = append(points, pts...)

					nms := make([]float32, len(square))
					for i := 0; i < len(square); i += 9 {
						p1 := mgl32.Vec3{pts[i+0], pts[i+1], pts[i+2]}
						p2 := mgl32.Vec3{pts[i+3], pts[i+4], pts[i+5]}
						p3 := mgl32.Vec3{pts[i+6], pts[i+7], pts[i+8]}
						v1 := p1.Sub(p2)
						v2 := p1.Sub(p3)
						n := v1.Cross(v2).Normalize()
						for j := 0; j < 3; j++ {
							nms[i+3*j+0] = n[0]
							nms[i+3*j+1] = n[1]
							nms[i+3*j+2] = n[2]
						}
					}
					normals = append(normals, nms...)
				}
			}
		}
	}
	c.Drawable = makeVao(points, normals)
	c.GraphicsInitialized = true
}

func drawChunk(chunk *geom.Chunk) {
	gl.BindVertexArray(chunk.Drawable)
	cs := geom.ChunkSize
	gl.DrawArrays(gl.TRIANGLES, 0, int32(cs*cs*cs*len(square)/3))
}

func initHUD() {
	points := []float32{
		-20.0, 0.0, 0.0,
		19.0, 0.0, 0.0,

		0.0, -20.0, 0.0,
		0.0, 19.0, 0.0,
	}
	hudDrawable = makePointsVao(points, 3)
}

func drawHUD() {
	gl.BindVertexArray(hudDrawable)
	gl.DrawArrays(gl.LINES, 0, 4)
}

func initText() {
	points := []float32{
		0.0, 0.0, 0.0, 0.0,
		0.0, 0.5, 0.0, 1.0,
		0.5, 0.5, 1.0, 1.0,

		0.5, 0.5, 1.0, 1.0,
		0.5, 0.0, 1.0, 0.0,
		0.0, 0.0, 0.0, 0.0,
	}
	textDrawable = makePointsVao(points, 4)

	// Generated from https://evanw.github.io/font-texture-generator/
	// Inconsolata font (installed on system with Google Web Fonts), size 24
	// Power of 2, white with black stroke, thickness 2
	existingImageFile, err := os.Open("font.png")
	if err != nil {
		panic(err)
	}
	defer existingImageFile.Close()
	img, err := png.Decode(existingImageFile)
	if err != nil {
		panic(err)
	}
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Pt(0, 0), draw.Src)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.GenTextures(1, &textTextureValue)
	gl.BindTexture(gl.TEXTURE_2D, textTextureValue)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.SRGB_ALPHA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix),
	)

	gl.GenerateMipmap(gl.TEXTURE_2D)
}

func drawText() {
	gl.BindVertexArray(textDrawable)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
}
