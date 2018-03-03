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
		for lonInd := range chunk.Cells {
			for latInd := range chunk.Cells[lonInd] {
				for altInd, c := range chunk.Cells[lonInd][latInd] {
					if !c.GraphicsInitialized {
						initCellGraphics(c, p, key.Lon*16+lonInd, key.Lat*16+latInd, key.Alt*16+altInd)
					}
					drawCell(c)
				}
			}
		}
	}
}

func initCellGraphics(c *server.Cell, p *server.Planet, lonIndex, latIndex, altIndex int) {
	points := make([]float32, len(square), len(square))
	copy(points, square)
	n := make([]float32, len(square), len(square))

	for i := 0; i < len(points); i += 3 {
		lonVal := float32(lonIndex) + points[i]
		latVal := float32(latIndex) + points[i+1]
		altVal := float32(altIndex) + points[i+2]
		r, theta, phi := p.IndexToSpherical(lonVal, latVal, altVal)
		cart := mgl32.SphericalToCartesian(r, theta, phi)
		points[i] = cart[0]
		points[i+1] = cart[1]
		points[i+2] = cart[2]
	}

	for i := 0; i < len(points); i += 9 {
		p1 := mgl32.Vec3{points[i], points[i+1], points[i+2]}
		p2 := mgl32.Vec3{points[i+3], points[i+4], points[i+5]}
		p3 := mgl32.Vec3{points[i+6], points[i+7], points[i+8]}
		v1 := p1.Sub(p2)
		v2 := p1.Sub(p3)
		norm := v1.Cross(v2).Normalize()
		for j := 0; j < 3; j++ {
			n[i+3*j+0] = norm[0]
			n[i+3*j+1] = norm[1]
			n[i+3*j+2] = norm[2]
		}
	}

	c.Drawable = makeVao(points, n)
	c.GraphicsInitialized = true
}

func drawCell(c *server.Cell) {
	if c.Material == server.Air {
		return
	}

	gl.BindVertexArray(c.Drawable)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
	// gl.DrawArrays(gl.LINES, 0, int32(len(square)/3))
}
