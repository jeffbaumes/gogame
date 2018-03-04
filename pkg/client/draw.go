package client

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jeffbaumes/gogame/pkg/server"
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
)

func drawPlanet(p *server.Planet) {
	for key, chunk := range p.Chunks {
		if !chunk.GraphicsInitialized {
			initChunkGraphics(chunk, p, key.Lon, key.Lat, key.Alt)
		}
		drawChunk(chunk)
	}
}

func initChunkGraphics(c *server.Chunk, p *server.Planet, lonIndex, latIndex, altIndex int) {
	cs := server.ChunkSize
	points := []float32{}
	normals := []float32{}

	for cLon := 0; cLon < cs; cLon++ {
		for cLat := 0; cLat < cs; cLat++ {
			for cAlt := 0; cAlt < cs; cAlt++ {
				if c.Cells[cLon][cLat][cAlt].Material != server.Air {
					pts := make([]float32, len(square))
					for i := 0; i < len(square); i += 3 {
						lonVal := float32(cs*lonIndex+cLon) + square[i]
						latVal := float32(cs*latIndex+cLat) + square[i+1]
						altVal := float32(cs*altIndex+cAlt) + square[i+2]
						r, theta, phi := p.IndexToSpherical(lonVal, latVal, altVal)
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

func drawChunk(chunk *server.Chunk) {
	gl.BindVertexArray(chunk.Drawable)
	cs := server.ChunkSize
	gl.DrawArrays(gl.TRIANGLES, 0, int32(cs*cs*cs*len(square)/3))
}
