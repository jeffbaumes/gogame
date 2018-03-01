package server

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

const (
	ChunkSize = 16
)

type planetState struct {
	AltMin, AltDelta, LatMax     float64
	LonCells, LatCells, AltCells int
	Cells                        [][][]*Cell
	Chunks                       map[ChunkKey]*Chunk
}

type Planet struct {
	Universe *Universe
	planetState
}

func NewPlanet(u *Universe, altMin, altDelta, latMax float64, lonCells, latCells, altCells int) *Planet {
	p := Planet{}
	p.Universe = u
	p.AltMin = altMin
	p.AltDelta = altDelta
	p.LatMax = latMax
	p.LonCells = lonCells / ChunkSize * ChunkSize
	p.LatCells = latCells / ChunkSize * ChunkSize
	p.AltCells = altCells / ChunkSize * ChunkSize
	p.Chunks = make(map[ChunkKey]*Chunk)
	return &p
}

type ChunkKey struct {
	Lon, Lat, Alt int
}

func (p *Planet) IndexToChunk(lon, lat, alt float32) *Chunk {
	if lon < 0 || lat < 0 || alt < 0 {
		return nil
	}
	if int(lon) >= p.LonCells || int(lat) >= p.LatCells || int(alt) >= p.AltCells {
		return nil
	}
	clon := int(math.Floor(float64(lon))) / ChunkSize
	clat := int(math.Floor(float64(lat))) / ChunkSize
	calt := int(math.Floor(float64(alt))) / ChunkSize
	key := ChunkKey{clon, clat, calt}
	chunk := p.Chunks[key]
	if chunk == nil {
		chunk = newChunk(clon, clat, calt, p)
		p.Chunks[key] = chunk
	}
	return chunk
}

func (p *Planet) IndexToCellCenterIndex(lon, lat, alt float32) (cLon, cLat, cAlt float32) {
	cLon = float32(math.Floor(float64(lon) + 0.5))
	cLat = float32(math.Floor(float64(lat) + 0.5))
	cAlt = float32(math.Floor(float64(alt) + 0.5))
	return
}

func (p *Planet) IndexToCell(lon, lat, alt float32) *Cell {
	chunk := p.IndexToChunk(lon, lat, alt)
	if chunk == nil {
		return nil
	}
	lonInd := int(lon) % ChunkSize
	latInd := int(lat) % ChunkSize
	altInd := int(alt) % ChunkSize
	return chunk.Cells[lonInd][latInd][altInd]
}

func (p *Planet) SphericalToIndex(r, theta, phi float32) (lon, lat, alt float32) {
	alt = (r - float32(p.AltMin)) / float32(p.AltDelta)
	lat = (180*theta/math.Pi - 90 + float32(p.LatMax)) * float32(p.LatCells) / (2 * float32(p.LatMax))
	if phi < 0 {
		phi += 2 * math.Pi
	}
	lon = phi * float32(p.LonCells) / (2 * math.Pi)
	return
}

func (p *Planet) CartesianToCell(cart mgl32.Vec3) *Cell {
	r, theta, phi := mgl32.CartesianToSpherical(cart)
	lon, lat, alt := p.SphericalToIndex(r, theta, phi)
	return p.IndexToCell(lon, lat, alt)
}

func (p *Planet) CartesianToIndex(cart mgl32.Vec3) (lon, lat, alt float32) {
	r, theta, phi := mgl32.CartesianToSpherical(cart)
	lon, lat, alt = p.SphericalToIndex(r, theta, phi)
	return
}

func (p *Planet) IndexToCartesian(lon, lat, alt float32) mgl32.Vec3 {
	r, theta, phi := p.IndexToSpherical(lon, lat, alt)
	return mgl32.SphericalToCartesian(r, theta, phi)
}

func (p *Planet) IndexToSpherical(lon, lat, alt float32) (r, theta, phi float32) {
	r = alt*float32(p.AltDelta) + float32(p.AltMin)
	theta = (math.Pi / 180) * ((90.0 - float32(p.LatMax)) + (lat/float32(p.LatCells))*(2.0*float32(p.LatMax)))
	phi = 2 * math.Pi * lon / float32(p.LonCells)
	return
}

func (p *Planet) nearestCellNormal(cart mgl32.Vec3) (normal mgl32.Vec3, separation float32) {
	lon, lat, alt := p.CartesianToIndex(cart)
	cLon, cLat, cAlt := p.IndexToCellCenterIndex(lon, lat, alt)
	dLon := float64(lon - cLon)
	dLat := float64(lat - cLat)
	dAlt := float64(alt - cAlt)
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
	nLoc := p.IndexToCartesian(float32(nLon), float32(nLat), float32(nAlt))
	cLoc := p.IndexToCartesian(float32(cLon), float32(cLat), float32(cAlt))
	normal = nLoc.Sub(cLoc).Normalize()
	separation = normal.Dot(cart.Sub(nLoc))
	return
}

type Chunk struct {
	Cells [][][]*Cell
}

func newChunk(lon, lat, alt int, p *Planet) *Chunk {
	chunk := Chunk{}
	chunk.Cells = make([][][]*Cell, ChunkSize, ChunkSize)
	for lonIndex := 0; lonIndex < ChunkSize; lonIndex++ {
		chunk.Cells[lonIndex] = make([][]*Cell, ChunkSize, ChunkSize)
		for latIndex := 0; latIndex < ChunkSize; latIndex++ {
			for altIndex := 0; altIndex < ChunkSize; altIndex++ {
				c := Cell{}
				pos := p.IndexToCartesian(float32(ChunkSize*lon+lonIndex), float32(ChunkSize*lat+latIndex), float32(ChunkSize*alt+altIndex))
				const scale = 0.1
				height := (p.Universe.noise.Eval3(float64(pos[0])*scale, float64(pos[1])*scale, float64(pos[2])*scale) + 1.0) * float64(p.AltCells) / 4.0
				if float64(altIndex) <= height {
					c.Material = Rock
				} else {
					c.Material = Air
				}
				chunk.Cells[lonIndex][latIndex] = append(chunk.Cells[lonIndex][latIndex], &c)
			}
		}
	}
	return &chunk
}

type cellState struct {
	Material int
}

type Cell struct {
	Drawable            uint32
	GraphicsInitialized bool
	cellState
}

const (
	Air  = iota
	Rock = iota
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
