package main

import (
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
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

type planet struct {
	cells                                        [][][]*cell
	radius, latMax                               float64
	lonCells, latCells, altCells, altMin, altMax int
}

func newPlanet(radius, latMax float64, lonCells, latCells, altCells, altMin, altMax int) *planet {
	p := planet{nil, radius, latMax, lonCells, latCells, altCells, altMin, altMax}
	p.generateCells()
	return &p
}

func (p *planet) generateCells() {
	p.cells = make([][][]*cell, p.lonCells, p.lonCells)
	for lonIndex := 0; lonIndex < p.lonCells; lonIndex++ {
		p.cells[lonIndex] = make([][]*cell, p.latCells, p.latCells)
		for latIndex := 0; latIndex < p.latCells; latIndex++ {
			for altIndex := p.altMin; altIndex <= p.altMax; altIndex++ {
				c := newCell(p, lonIndex, latIndex, altIndex)

				height := (noise.Eval2(float64(lonIndex)/20.0, float64(latIndex)/20.0)+1.0)*float64(p.altMax-p.altMin)/2.0 + float64(p.altMin)
				c.alive = float64(altIndex) <= height

				p.cells[lonIndex][latIndex] = append(p.cells[lonIndex][latIndex], c)
			}
		}
	}
}

func (p *planet) draw() {
	for x := range p.cells {
		for y := range p.cells[x] {
			for _, c := range p.cells[x][y] {
				c.draw()
			}
		}
	}
}

func (p *planet) sphericalToIndex(r, theta, phi float32) (lonInd, latInd, altInd float32) {
	altInd = r * float32(p.altCells) / float32(p.radius)
	latInd = (180*theta/math.Pi - 90 + float32(p.latMax)) * float32(p.latCells) / (2 * float32(p.latMax))
	lonInd = phi * float32(p.lonCells) / (2 * math.Pi)
	return
}

func (p *planet) cartesianToCell(cart mgl32.Vec3) *cell {
	r, theta, phi := mgl32.CartesianToSpherical(cart)
	lonInd, latInd, altInd := p.sphericalToIndex(r, theta, phi)
	return p.cells[int(lonInd)][int(latInd)][int(altInd)-p.altMin]
}

func (p *planet) indexToSpherical(lonInd, latInd, altInd float32) (r, theta, phi float32) {
	r = float32(p.radius) * altInd / float32(p.altCells)
	theta = (math.Pi / 180) * ((90.0 - float32(p.latMax)) + (latInd/float32(p.latCells))*(2.0*float32(p.latMax)))
	phi = 2 * math.Pi * lonInd / float32(p.lonCells)
	return
}

type cell struct {
	drawable uint32
	alive    bool
}

func newCell(p *planet, lonIndex, latIndex, altIndex int) *cell {
	points := make([]float32, len(square), len(square))
	copy(points, square)
	n := make([]float32, len(square), len(square))

	for i := 0; i < len(points); i += 3 {
		lonVal := float32(lonIndex) + points[i]
		latVal := float32(latIndex) + points[i+1]
		altVal := float32(altIndex) + points[i+2]
		r, theta, phi := p.indexToSpherical(lonVal, latVal, altVal)
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

	return &cell{
		drawable: MakeVao(points, n),
	}
}

func (c *cell) draw() {
	if !c.alive {
		return
	}

	gl.BindVertexArray(c.drawable)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
}
