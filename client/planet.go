package client

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jeffbaumes/gogame/server"
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
	cs3 := cs * cs * cs
	points := make([]float32, cs3*len(square), cs3*len(square))
	n := make([]float32, cs3*len(square), cs3*len(square))

	cInd := 0
	for cLon := 0; cLon < cs; cLon++ {
		for cLat := 0; cLat < cs; cLat++ {
			for cAlt := 0; cAlt < cs; cAlt, cInd = cAlt+1, cInd+1 {
				if c.Cells[cLon][cLat][cAlt].Material != server.Air {
					for i := 0; i < len(square); i += 3 {
						pInd := cInd*len(square) + i
						lonVal := float32(cs*lonIndex+cLon) + square[i]
						latVal := float32(cs*latIndex+cLat) + square[i+1]
						altVal := float32(cs*altIndex+cAlt) + square[i+2]
						r, theta, phi := p.IndexToSpherical(lonVal, latVal, altVal)
						cart := mgl32.SphericalToCartesian(r, theta, phi)
						points[pInd] = cart[0]
						points[pInd+1] = cart[1]
						points[pInd+2] = cart[2]
					}

					for i := 0; i < len(square); i += 9 {
						pInd := cInd*len(square) + i
						p1 := mgl32.Vec3{points[pInd+0], points[pInd+1], points[pInd+2]}
						p2 := mgl32.Vec3{points[pInd+3], points[pInd+4], points[pInd+5]}
						p3 := mgl32.Vec3{points[pInd+6], points[pInd+7], points[pInd+8]}
						v1 := p1.Sub(p2)
						v2 := p1.Sub(p3)
						norm := v1.Cross(v2).Normalize()
						for j := 0; j < 3; j++ {
							n[pInd+3*j+0] = norm[0]
							n[pInd+3*j+1] = norm[1]
							n[pInd+3*j+2] = norm[2]
						}
					}
				}
			}
		}
	}
	c.Drawable = makeVao(points, n)
	c.GraphicsInitialized = true
}

func drawChunk(chunk *server.Chunk) {
	gl.BindVertexArray(chunk.Drawable)
	cs := server.ChunkSize
	gl.DrawArrays(gl.TRIANGLES, 0, int32(cs*cs*cs*len(square)/3))
	// gl.DrawArrays(gl.LINES, 0, int32(len(square)/3))
}
