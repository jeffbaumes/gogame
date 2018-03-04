package geom

import (
	"log"
	"math"
	"net/rpc"

	"github.com/go-gl/mathgl/mgl32"
	opensimplex "github.com/ojrac/opensimplex-go"
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
	Seed                         int
}

// Planet represents all the cells in a spherical planet
type Planet struct {
	rpc    *rpc.Client
	Chunks map[ChunkIndex]*Chunk
	noise  *opensimplex.Noise
	PlanetState
}

// NewPlanet constructs a Planet instance
func NewPlanet(radius float64, altCells, seed int, crpc *rpc.Client) *Planet {
	p := Planet{}
	p.Seed = seed
	p.noise = opensimplex.NewWithSeed(int64(seed))
	p.AltMin = radius - float64(altCells)
	p.AltDelta = 1.0
	p.LatMax = 60.0
	p.LonCells = int(2.0*math.Pi*3.0/4.0*radius+0.5) / ChunkSize * ChunkSize
	p.LatCells = int(2.0/3.0*math.Pi*radius) / ChunkSize * ChunkSize
	p.AltCells = altCells / ChunkSize * ChunkSize
	p.Chunks = make(map[ChunkIndex]*Chunk)
	p.rpc = crpc
	return &p
}

// ChunkIndex stores the latitude, longitude, and altitude index of a chunk
type ChunkIndex struct {
	Lon, Lat, Alt int
}

// CellLoc stores the latitude, longitude, and altitude of a position in cell coordinates
type CellLoc struct {
	Lon, Lat, Alt float32
}

// GetChunk retrieves the chunk of a planet from chunk indices
func (p *Planet) GetChunk(ind ChunkIndex) *Chunk {
	chunk := p.Chunks[ind]
	if chunk == nil {
		if p.rpc == nil {
			chunk = newChunk(ind, p)
			p.Chunks[ind] = chunk
		} else {
			rchunk := Chunk{}
			e := p.rpc.Call("Server.GetChunk", ind, &rchunk)
			if e != nil {
				log.Fatal("GetChunk error:", e)
			}
			p.Chunks[ind] = &rchunk
		}
	}
	return chunk
}

// CellLocToChunk converts floating-point cell indices to a chunk
func (p *Planet) CellLocToChunk(l CellLoc) *Chunk {
	if l.Lon < 0 || l.Lat < 0 || l.Alt < 0 {
		return nil
	}
	if int(l.Lon) >= p.LonCells || int(l.Lat) >= p.LatCells || int(l.Alt) >= p.AltCells {
		return nil
	}
	clon := int(math.Floor(float64(l.Lon))) / ChunkSize
	clat := int(math.Floor(float64(l.Lat))) / ChunkSize
	calt := int(math.Floor(float64(l.Alt))) / ChunkSize
	return p.GetChunk(ChunkIndex{Lon: clon, Lat: clat, Alt: calt})
}

// CartesianToChunk converts world coordinates to a chunk
func (p *Planet) CartesianToChunk(cart mgl32.Vec3) *Chunk {
	l := p.CartesianToCellLoc(cart)
	return p.CellLocToChunk(l)
}

// CellLocToNearestCellCenter converts floating-point cell indices to the nearest integral indices
func (p *Planet) CellLocToNearestCellCenter(l CellLoc) CellLoc {
	cLon := float32(math.Floor(float64(l.Lon) + 0.5))
	cLat := float32(math.Floor(float64(l.Lat) + 0.5))
	cAlt := float32(math.Floor(float64(l.Alt) + 0.5))
	return CellLoc{Lon: cLon, Lat: cLat, Alt: cAlt}
}

// CellLocToCell converts floating-point chunk indices to a cell
func (p *Planet) CellLocToCell(l CellLoc) *Cell {
	chunk := p.CellLocToChunk(l)
	if chunk == nil {
		return nil
	}
	lonInd := int(l.Lon) % ChunkSize
	latInd := int(l.Lat) % ChunkSize
	altInd := int(l.Alt) % ChunkSize
	return chunk.Cells[lonInd][latInd][altInd]
}

// SphericalToCellLoc converts spherical coordinates to floating-point cell indices
func (p *Planet) SphericalToCellLoc(r, theta, phi float32) CellLoc {
	alt := (r - float32(p.AltMin)) / float32(p.AltDelta)
	lat := (180*theta/math.Pi - 90 + float32(p.LatMax)) * float32(p.LatCells) / (2 * float32(p.LatMax))
	if phi < 0 {
		phi += 2 * math.Pi
	}
	lon := phi * float32(p.LonCells) / (2 * math.Pi)
	return CellLoc{Lon: lon, Lat: lat, Alt: alt}
}

// CartesianToCell returns the cell contianing a set of world coordinates
func (p *Planet) CartesianToCell(cart mgl32.Vec3) *Cell {
	r, theta, phi := mgl32.CartesianToSpherical(cart)
	l := p.SphericalToCellLoc(r, theta, phi)
	return p.CellLocToCell(l)
}

// CartesianToCellLoc converts world coordinates to floating-point cell indices
func (p *Planet) CartesianToCellLoc(cart mgl32.Vec3) CellLoc {
	r, theta, phi := mgl32.CartesianToSpherical(cart)
	return p.SphericalToCellLoc(r, theta, phi)
}

// CellLocToCartesian converts floating-point cell indices to world coordinates
func (p *Planet) CellLocToCartesian(l CellLoc) mgl32.Vec3 {
	r, theta, phi := p.CellLocToSpherical(l)
	return mgl32.SphericalToCartesian(r, theta, phi)
}

// CellLocToSpherical converts floating-point cell indices to spherical coordinates
func (p *Planet) CellLocToSpherical(l CellLoc) (r, theta, phi float32) {
	r = l.Alt*float32(p.AltDelta) + float32(p.AltMin)
	theta = (math.Pi / 180) * ((90.0 - float32(p.LatMax)) + (l.Lat/float32(p.LatCells))*(2.0*float32(p.LatMax)))
	phi = 2 * math.Pi * l.Lon / float32(p.LonCells)
	return
}

// Chunk is a 3D block of planet cells
type Chunk struct {
	Drawable            uint32
	GraphicsInitialized bool
	Cells               [][][]*Cell
}

func newChunk(ind ChunkIndex, p *Planet) *Chunk {
	chunk := Chunk{}
	chunk.Cells = make([][][]*Cell, ChunkSize)
	for lonIndex := 0; lonIndex < ChunkSize; lonIndex++ {
		chunk.Cells[lonIndex] = make([][]*Cell, ChunkSize)
		for latIndex := 0; latIndex < ChunkSize; latIndex++ {
			for altIndex := 0; altIndex < ChunkSize; altIndex++ {
				c := Cell{}
				l := CellLoc{
					Lon: float32(ChunkSize*ind.Lon + lonIndex),
					Lat: float32(ChunkSize*ind.Lat + latIndex),
					Alt: float32(ChunkSize*ind.Alt + altIndex),
				}
				pos := p.CellLocToCartesian(l)
				const scale = 0.1
				height := (p.noise.Eval3(float64(pos[0])*scale, float64(pos[1])*scale, float64(pos[2])*scale) + 1.0) * float64(p.AltCells) / 4.0
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
	Material int
}

// Air is a transparent, empty material
// Rock is an opaque, solid material
const (
	Air  = iota
	Rock = iota
)
