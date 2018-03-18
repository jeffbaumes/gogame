package geom

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"math"
	"net/rpc"
	"sync"

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
	Seed                         int
}

// Planet represents all the cells in a spherical planet
type Planet struct {
	rpc         *rpc.Client
	db          *sql.DB
	Chunks      map[ChunkIndex]*Chunk
	ChunksMutex *sync.Mutex
	noise       *opensimplex.Noise
	PlanetState
}

// NewPlanet constructs a Planet instance
func NewPlanet(radius float64, altCells, seed int, crpc *rpc.Client, db *sql.DB) *Planet {
	p := Planet{}
	p.Seed = seed
	p.noise = opensimplex.NewWithSeed(int64(seed))
	p.AltMin = radius - float64(altCells)
	p.AltDelta = 1.0
	p.LatMax = 90.0
	p.LonCells = int(2.0*math.Pi*3.0/4.0*radius+0.5) / ChunkSize * ChunkSize
	p.LatCells = int(p.LatMax/90.0*math.Pi*radius) / ChunkSize * ChunkSize
	p.AltCells = altCells / ChunkSize * ChunkSize
	p.Chunks = make(map[ChunkIndex]*Chunk)
	p.rpc = crpc
	p.db = db
	p.ChunksMutex = &sync.Mutex{}
	return &p
}

// ChunkIndex stores the latitude, longitude, and altitude index of a chunk
type ChunkIndex struct {
	Lon, Lat, Alt int
}

// CellIndex stores the latitude, longitude, and altitude index of a cell
type CellIndex struct {
	Lon, Lat, Alt int
}

// CellLoc stores the latitude, longitude, and altitude of a position in cell coordinates
type CellLoc struct {
	Lon, Lat, Alt float32
}

// GetChunk retrieves the chunk of a planet from chunk indices
func (p *Planet) GetChunk(ind ChunkIndex) *Chunk {
	cs := ChunkSize
	if ind.Lon < 0 || ind.Lon >= p.LonCells/cs {
		return nil
	}
	if ind.Lat < 0 || ind.Lat >= p.LatCells/cs {
		return nil
	}
	if ind.Alt < 0 || ind.Alt >= p.AltCells/cs {
		return nil
	}

	p.ChunksMutex.Lock()
	chunk := p.Chunks[ind]
	p.ChunksMutex.Unlock()

	if chunk != nil && chunk.WaitingForData {
		return nil
	}
	if chunk == nil {
		if p.rpc == nil {
			if p.db != nil {
				rows, e := p.db.Query("SELECT data FROM chunk WHERE planet = 0 AND lon = ? AND lat = ? AND alt = ?", ind.Lon, ind.Lat, ind.Alt)
				if e != nil {
					panic(e)
				}
				if rows.Next() {
					var data []byte
					e = rows.Scan(&data)
					if e != nil {
						panic(e)
					}
					var dbuf bytes.Buffer
					dbuf.Write(data)
					dec := gob.NewDecoder(&dbuf)
					var ch Chunk
					e = dec.Decode(&ch)
					if e != nil {
						panic(e)
					}
					chunk = &ch
				}
				rows.Close()
				if chunk == nil {
					chunk = newChunk(ind, p)
					stmt, e := p.db.Prepare("INSERT INTO chunk VALUES (?, ?, ?, ?, ?)")
					if e != nil {
						panic(e)
					}
					var buf bytes.Buffer
					enc := gob.NewEncoder(&buf)
					e = enc.Encode(chunk)
					if e != nil {
						panic(e)
					}
					_, e = stmt.Exec(0, ind.Lon, ind.Lat, ind.Alt, buf.Bytes())
					if e != nil {
						panic(e)
					}
				}
				p.ChunksMutex.Lock()
				p.Chunks[ind] = chunk
				p.ChunksMutex.Unlock()
			} else {
				chunk = newChunk(ind, p)
				p.ChunksMutex.Lock()
				p.Chunks[ind] = chunk
				p.ChunksMutex.Unlock()
			}
		} else {
			rchunk := Chunk{}
			call := p.rpc.Go("API.GetChunk", ind, &rchunk, nil)
			go func() {
				call = <-call.Done
				p.ChunksMutex.Lock()
				p.Chunks[ind] = &rchunk
				p.ChunksMutex.Unlock()
			}()
			p.ChunksMutex.Lock()
			p.Chunks[ind] = &Chunk{WaitingForData: true}
			p.ChunksMutex.Unlock()
		}
	}
	return chunk
}

// RPCSetCellMaterialArgs contains the arguments for the SetCellMaterial RPC call
type RPCSetCellMaterialArgs struct {
	Index    CellIndex
	Material int
}

// SetCellMaterial sets the material for a cell
func (p *Planet) SetCellMaterial(ind CellIndex, material int) bool {
	cell := p.CellIndexToCell(ind)
	if cell == nil {
		return false
	}
	if cell.Material == material {
		return false
	}
	cell.Material = material
	if p.rpc != nil {
		var ret bool
		p.rpc.Go("API.SetCellMaterial", RPCSetCellMaterialArgs{
			Index:    ind,
			Material: material,
		}, &ret, nil)
	}
	if p.db != nil {
		chunkInd := p.CellIndexToChunkIndex(ind)
		chunk := p.CellIndexToChunk(ind)
		stmt, e := p.db.Prepare("UPDATE chunk SET data = ? WHERE planet = 0 AND lon = ? AND lat = ? AND alt = ?")
		if e != nil {
			panic(e)
		}
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		e = enc.Encode(chunk)
		if e != nil {
			panic(e)
		}
		_, e = stmt.Exec(buf.Bytes(), chunkInd.Lon, chunkInd.Lat, chunkInd.Alt)
		if e != nil {
			panic(e)
		}
	}

	return true
}

func (p *Planet) validateCellLoc(l CellLoc) CellLoc {
	if l.Lon < 0 {
		l.Lon += float32(p.LonCells)
	}
	for l.Lon >= float32(p.LonCells) {
		l.Lon -= float32(p.LonCells)
	}
	return l
}

// CellLocToChunk converts floating-point cell indices to a chunk
func (p *Planet) CellLocToChunk(l CellLoc) *Chunk {
	l = p.validateCellLoc(l)
	return p.CellIndexToChunk(p.CellLocToCellIndex(l))
}

// CellIndexToChunk converts a cell index to its containing chunk
func (p *Planet) CellIndexToChunk(cellIndex CellIndex) *Chunk {
	ind := p.CellIndexToChunkIndex(cellIndex)
	if ind.Lon < 0 || ind.Lon >= p.LonCells/ChunkSize {
		return nil
	}
	if ind.Lat < 0 || ind.Lat >= p.LatCells/ChunkSize {
		return nil
	}
	if ind.Alt < 0 || ind.Alt >= p.AltCells/ChunkSize {
		return nil
	}
	return p.GetChunk(ind)
}

// CellLocToChunkIndex converts floating-point cell indices to a chunk index
func (p *Planet) CellLocToChunkIndex(l CellLoc) ChunkIndex {
	l = p.validateCellLoc(l)
	return p.CellIndexToChunkIndex(p.CellLocToCellIndex(l))
}

// CellIndexToChunkIndex converts a cell index to its containing chunk index
func (p *Planet) CellIndexToChunkIndex(cellInd CellIndex) ChunkIndex {
	cs := float64(ChunkSize)
	return ChunkIndex{
		Lon: int(math.Floor(float64(cellInd.Lon) / cs)),
		Lat: int(math.Floor(float64(cellInd.Lat) / cs)),
		Alt: int(math.Floor(float64(cellInd.Alt) / cs)),
	}
}

// CellLocToCellIndex converts floating-point cell indices to a cell index
func (p *Planet) CellLocToCellIndex(l CellLoc) CellIndex {
	l = p.validateCellLoc(l)
	l = p.CellLocToNearestCellCenter(l)
	l = p.validateCellLoc(l)
	return CellIndex{Lon: int(l.Lon), Lat: int(l.Lat), Alt: int(l.Alt)}
}

// CartesianToChunkIndex converts world coordinates to a chunk index
func (p *Planet) CartesianToChunkIndex(cart mgl32.Vec3) ChunkIndex {
	l := p.CartesianToCellLoc(cart)
	return p.CellLocToChunkIndex(l)
}

// CartesianToCellIndex converts world coordinates to a cell index
func (p *Planet) CartesianToCellIndex(cart mgl32.Vec3) CellIndex {
	return p.CellLocToCellIndex(p.CartesianToCellLoc(cart))
}

// CartesianToChunk converts world coordinates to a chunk
func (p *Planet) CartesianToChunk(cart mgl32.Vec3) *Chunk {
	return p.CellLocToChunk(p.CartesianToCellLoc(cart))
}

// CellLocToNearestCellCenter converts floating-point cell indices to the nearest integral indices
func (p *Planet) CellLocToNearestCellCenter(l CellLoc) CellLoc {
	l = p.validateCellLoc(l)
	return CellLoc{
		Lon: float32(math.Floor(float64(l.Lon) + 0.5)),
		Lat: float32(math.Floor(float64(l.Lat) + 0.5)),
		Alt: float32(math.Floor(float64(l.Alt) + 0.5)),
	}
}

// CellLocToCell converts floating-point chunk indices to a cell
func (p *Planet) CellLocToCell(l CellLoc) *Cell {
	l = p.validateCellLoc(l)
	return p.CellIndexToCell(p.CellLocToCellIndex(l))
}

// CellIndexToCell converts a cell index to a cell
func (p *Planet) CellIndexToCell(cellIndex CellIndex) *Cell {
	chunkIndex := p.CellIndexToChunkIndex(cellIndex)
	lonCells := p.LonCellsInChunkIndex(chunkIndex)
	lonWidth := ChunkSize / lonCells
	chunk := p.CellIndexToChunk(cellIndex)
	if chunk == nil {
		return nil
	}
	lonInd := (cellIndex.Lon % ChunkSize) / lonWidth
	latInd := cellIndex.Lat % ChunkSize
	altInd := cellIndex.Alt % ChunkSize
	return chunk.Cells[lonInd][latInd][altInd]
}

// SphericalToCellLoc converts spherical coordinates to floating-point cell indices
func (p *Planet) SphericalToCellLoc(r, theta, phi float32) CellLoc {
	alt := (r - float32(p.AltMin)) / float32(p.AltDelta)
	lat := (180*theta/math.Pi-90+float32(p.LatMax))*float32(p.LatCells)/(2*float32(p.LatMax)) - 0.5
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
	l = p.validateCellLoc(l)
	r, theta, phi := p.CellLocToSpherical(l)
	return mgl32.SphericalToCartesian(r, theta, phi)
}

// CellLocToSpherical converts floating-point cell indices to spherical coordinates
func (p *Planet) CellLocToSpherical(l CellLoc) (r, theta, phi float32) {
	l = p.validateCellLoc(l)
	r = l.Alt*float32(p.AltDelta) + float32(p.AltMin)
	theta = (math.Pi / 180) * ((90.0 - float32(p.LatMax)) + ((l.Lat+0.5)/float32(p.LatCells))*(2.0*float32(p.LatMax)))
	phi = 2 * math.Pi * l.Lon / float32(p.LonCells)
	return
}

// LonCellsInChunkIndex returns the number of longitude cells in a chunk, which changes based on latitude
func (p *Planet) LonCellsInChunkIndex(ind ChunkIndex) int {
	theta := (90.0 - float32(p.LatMax) + (float32(ind.Lat)+0.5)*float32(ChunkSize)/float32(p.LatCells)) * (2.0 * float32(p.LatMax))
	if math.Abs(float64(theta-90)) >= 80 {
		return ChunkSize / 4
	}
	if math.Abs(float64(theta-90)) >= 60 {
		return ChunkSize / 2
	}
	return ChunkSize
}

// Chunk is a 3D block of planet cells
type Chunk struct {
	WaitingForData bool
	Cells          [][][]*Cell
}

func newChunk(ind ChunkIndex, p *Planet) *Chunk {
	chunk := Chunk{}
	lonCells := p.LonCellsInChunkIndex(ind)
	lonWidth := ChunkSize / lonCells
	chunk.Cells = make([][][]*Cell, lonCells)
	for lonIndex := 0; lonIndex < lonCells; lonIndex++ {
		chunk.Cells[lonIndex] = make([][]*Cell, ChunkSize)
		for latIndex := 0; latIndex < ChunkSize; latIndex++ {
			for altIndex := 0; altIndex < ChunkSize; altIndex++ {
				c := Cell{}
				l := CellLoc{
					Lon: float32(ChunkSize*ind.Lon + lonIndex*lonWidth),
					Lat: float32(ChunkSize*ind.Lat + latIndex),
					Alt: float32(ChunkSize*ind.Alt + altIndex),
				}
				pos := p.CellLocToCartesian(l)
				const scale = 0.1
				height := (p.noise.Eval3(float64(pos[0])*scale, float64(pos[1])*scale, float64(pos[2])*scale) + 1.0) * float64(p.AltCells) / 2.0
				if float64(l.Alt) <= height {
					if l.Alt > 8 {
						c.Material = Asteroid
					} else {
						c.Material = Moon
					}
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

const (
	Air      = 0
	Grass    = 1
	Dirt     = 2
	Stone    = 3
	Moon     = 4
	Asteroid = 5
	Sun      = 6
)

var (
	Materials []string = []string{
		"air",
		"grass",
		"dirt",
		"stone",
		"moon",
		"asteroid",
		"sun",
	}
)
