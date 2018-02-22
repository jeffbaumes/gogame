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
				if float64(altIndex) <= height {
					c.material = rock
				} else {
					c.material = air
				}

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

func (p *planet) indexToCell(lon, lat, alt float32) *cell {
	lonInd := int(math.Floor(float64(lon) + 0.5))
	latInd := int(math.Floor(float64(lat) + 0.5))
	altInd := int(math.Floor(float64(alt) + 0.5))
	if lonInd >= len(p.cells) || lonInd < 0 {
		return nil
	}
	if latInd >= len(p.cells[lonInd]) || latInd < 0 {
		return nil
	}
	if altInd-p.altMin >= len(p.cells[lonInd][latInd]) || altInd-p.altMin < 0 {
		return nil
	}
	return p.cells[lonInd][latInd][altInd-p.altMin]
}

func (p *planet) sphericalToIndex(r, theta, phi float32) (lon, lat, alt float32) {
	alt = r * float32(p.altCells) / float32(p.radius)
	lat = (180*theta/math.Pi - 90 + float32(p.latMax)) * float32(p.latCells) / (2 * float32(p.latMax))
	if phi < 0 {
		phi += 2 * math.Pi
	}
	lon = phi * float32(p.lonCells) / (2 * math.Pi)
	return
}

func (p *planet) cartesianToCell(cart mgl32.Vec3) *cell {
	// log.Println(cart)
	r, theta, phi := mgl32.CartesianToSpherical(cart)
	lon, lat, alt := p.sphericalToIndex(r, theta, phi)
	return p.indexToCell(lon, lat, alt)
}

func (p *planet) cartesianToIndex(cart mgl32.Vec3) (lon, lat, alt float32) {
	r, theta, phi := mgl32.CartesianToSpherical(cart)
	lon, lat, alt = p.sphericalToIndex(r, theta, phi)
	return
}

func (p *planet) indexToCartesian(lon, lat, alt float32) mgl32.Vec3 {
	r, theta, phi := p.indexToSpherical(lon, lat, alt)
	return mgl32.SphericalToCartesian(r, theta, phi)
}

func (p *planet) indexToSpherical(lon, lat, alt float32) (r, theta, phi float32) {
	r = float32(p.radius) * alt / float32(p.altCells)
	theta = (math.Pi / 180) * ((90.0 - float32(p.latMax)) + (lat/float32(p.latCells))*(2.0*float32(p.latMax)))
	phi = 2 * math.Pi * lon / float32(p.lonCells)
	return
}

func (p *planet) nearestCellNormal(cart mgl32.Vec3) (normal mgl32.Vec3, separation float32) {
	lon, lat, alt := p.cartesianToIndex(cart)
	cLon := math.Floor(float64(lon) + 0.5)
	cLat := math.Floor(float64(lat) + 0.5)
	cAlt := math.Floor(float64(alt) + 0.5)
	dLon := float64(lon) - cLon
	dLat := float64(lat) - cLat
	dAlt := float64(alt) - cAlt
	nLon, nLat, nAlt := cLon, cLat, cAlt
	if math.Abs(dLon) > math.Abs(dLat) && math.Abs(dLon) > math.Abs(dAlt) {
		if dLon > 0 {
			nLon = cLon + 0.5
		} else {
			nLon = cLon - 0.5
		}
	} else if math.Abs(dLat) > math.Abs(dAlt) {
		if dLat > 0 {
			nLat = cLat + 0.5
		} else {
			nLat = cLat - 0.5
		}
	} else {
		if dAlt > 0 {
			nAlt = cAlt + 0.5
		} else {
			nAlt = cAlt - 0.5
		}
	}
	// log.Println(lon, lat, alt)
	// log.Println(dLon, dLat, dAlt)
	// log.Println(cLon, cLat, cAlt)
	// log.Println(nLon, nLat, nAlt)
	nLoc := p.indexToCartesian(float32(nLon), float32(nLat), float32(nAlt))
	// log.Println(nLoc)
	cLoc := p.indexToCartesian(float32(cLon), float32(cLat), float32(cAlt))
	// log.Println(cLoc)
	// log.Println(nLoc.Sub(cLoc).Normalize())
	normal = nLoc.Sub(cLoc).Normalize()
	// separation = -project(cart.Sub(nLoc), normal).Len()
	separation = normal.Dot(cart.Sub(nLoc))
	return
}

type cell struct {
	drawable uint32
	material int
}

const (
	air  = iota
	rock = iota
)

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
	if c.material == air {
		return
	}

	gl.BindVertexArray(c.drawable)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
	// gl.DrawArrays(gl.LINES, 0, int32(len(square)/3))
}
