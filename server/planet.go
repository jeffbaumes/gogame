package server

import (
	"log"
	"math"
	"net/rpc"

	"github.com/go-gl/mathgl/mgl32"
)

// ChunkSize is the number of cells per side of a chunk
const (
	ChunkSize = 16
)

// PlanetState is the serializable portion of a Planet
type PlanetState struct {
	AltMin, AltDelta, LatMax     float64
	LonCells, LatCells, AltCells int
	Cells                        [][][]*Cell
	Chunks                       map[ChunkKey]*Chunk
	RPC                          *rpc.Client
}

// Planet represents all the cells in a spherical planet
type Planet struct {
	Universe *Universe
	PlanetState
}

// NewPlanet constructs a Planet instance
func NewPlanet(u *Universe, altMin, altDelta, latMax float64, lonCells, latCells, altCells int, crpc *rpc.Client) *Planet {
	p := Planet{}
	p.Universe = u
	p.AltMin = altMin
	p.AltDelta = altDelta
	p.LatMax = latMax
	p.LonCells = lonCells / ChunkSize * ChunkSize
	p.LatCells = latCells / ChunkSize * ChunkSize
	p.AltCells = altCells / ChunkSize * ChunkSize
	p.Chunks = make(map[ChunkKey]*Chunk)
	p.RPC = crpc
	return &p
}

// ChunkKey stores the latitude, longitude, and altitude index of a chunk
type ChunkKey struct {
	Lon, Lat, Alt int
}

// GetChunk retrieves the chunk of a planet from chunk indices
func (p *Planet) GetChunk(lon, lat, alt int) *Chunk {
	key := ChunkKey{lon, lat, alt}
	chunk := p.Chunks[key]
	if chunk == nil {
		if p.RPC == nil {
			chunk = newChunk(lon, lat, alt, p)
			p.Chunks[key] = chunk
		} else {
			args := GetChunkArgs{Lon: lon, Lat: lat, Alt: alt}
			rchunk := Chunk{}
			e := p.RPC.Call("Server.GetChunk", args, &rchunk)
			if e != nil {
				log.Fatal("GetChunk error:", e)
			}
			p.Chunks[key] = &rchunk
		}
	}
	return chunk
}

// IndexToChunk converts floating-point cell indices to a chunk
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
	return p.GetChunk(clon, clat, calt)
}

// IndexToCellCenterIndex converts floating-point cell indices to the nearest integral indices
func (p *Planet) IndexToCellCenterIndex(lon, lat, alt float32) (cLon, cLat, cAlt float32) {
	cLon = float32(math.Floor(float64(lon) + 0.5))
	cLat = float32(math.Floor(float64(lat) + 0.5))
	cAlt = float32(math.Floor(float64(alt) + 0.5))
	return
}

// IndexToCell converts floating-point chunk indices to a cell
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

// SphericalToIndex converts spherical coordinates to floating-point cell indices
func (p *Planet) SphericalToIndex(r, theta, phi float32) (lon, lat, alt float32) {
	alt = (r - float32(p.AltMin)) / float32(p.AltDelta)
	lat = (180*theta/math.Pi - 90 + float32(p.LatMax)) * float32(p.LatCells) / (2 * float32(p.LatMax))
	if phi < 0 {
		phi += 2 * math.Pi
	}
	lon = phi * float32(p.LonCells) / (2 * math.Pi)
	return
}

// CartesianToCell returns the cell contianing a set of world coordinates
func (p *Planet) CartesianToCell(cart mgl32.Vec3) *Cell {
	r, theta, phi := mgl32.CartesianToSpherical(cart)
	lon, lat, alt := p.SphericalToIndex(r, theta, phi)
	return p.IndexToCell(lon, lat, alt)
}

// CartesianToIndex converts world coordinates to floating-point cell indices
func (p *Planet) CartesianToIndex(cart mgl32.Vec3) (lon, lat, alt float32) {
	r, theta, phi := mgl32.CartesianToSpherical(cart)
	lon, lat, alt = p.SphericalToIndex(r, theta, phi)
	return
}

// IndexToCartesian converts floating-point cell indices to world coordinates
func (p *Planet) IndexToCartesian(lon, lat, alt float32) mgl32.Vec3 {
	r, theta, phi := p.IndexToSpherical(lon, lat, alt)
	return mgl32.SphericalToCartesian(r, theta, phi)
}

// IndexToSpherical converts floating-point cell indices to spherical coordinates
func (p *Planet) IndexToSpherical(lon, lat, alt float32) (r, theta, phi float32) {
	r = alt*float32(p.AltDelta) + float32(p.AltMin)
	theta = (math.Pi / 180) * ((90.0 - float32(p.LatMax)) + (lat/float32(p.LatCells))*(2.0*float32(p.LatMax)))
	phi = 2 * math.Pi * lon / float32(p.LonCells)
	return
}

// Chunk is a 3D block of planet cells
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

// Cell is a single block on a planet
type Cell struct {
	Drawable            uint32
	GraphicsInitialized bool
	Material            int
}

// Air is a transparent, empty material
// Rock is an opaque, solid material
const (
	Air  = iota
	Rock = iota
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
