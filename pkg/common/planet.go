package common

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
	ID              int
	Name            string
	GeneratorType   string
	Radius          float64
	AltCells        int
	Seed            int
	OrbitPlanet     int
	OrbitDistance   float64
	OrbitSeconds    float64
	RotationSeconds float64
}

// Planet represents all the cells in a spherical planet
type Planet struct {
	rpc           *rpc.Client
	db            *sql.DB
	Geometry      *PlanetGeometry
	GeometryMutex *sync.Mutex
	Chunks        map[ChunkIndex]*Chunk
	databaseMutex *sync.Mutex
	ChunksMutex   *sync.Mutex
	noise         *opensimplex.Noise
	Generator     func(*Planet, CellLoc) int
	AltMin        float64
	AltDelta      float64
	LatMax        float64
	LonCells      int
	LatCells      int
	PlanetState
}

// NewPlanet constructs a Planet instance
func NewPlanet(state PlanetState, crpc *rpc.Client, db *sql.DB) *Planet {
	p := Planet{}
	p.PlanetState = state
	p.noise = opensimplex.NewWithSeed(int64(p.Seed))
	p.AltCells = p.AltCells / ChunkSize * ChunkSize
	p.AltMin = p.Radius - float64(p.AltCells)
	p.AltDelta = 1.0
	p.LatMax = 90.0
	p.LonCells = int(2.0*math.Pi*3.0/4.0*(0.5*p.Radius)+0.5) / ChunkSize * ChunkSize
	p.LatCells = int(p.LatMax/90.0*math.Pi*(0.5*p.Radius)) / ChunkSize * ChunkSize
	p.Chunks = make(map[ChunkIndex]*Chunk)
	p.rpc = crpc
	p.db = db
	p.databaseMutex = &sync.Mutex{}
	p.ChunksMutex = &sync.Mutex{}
	p.GeometryMutex = &sync.Mutex{}
	p.Generator = generators[p.GeneratorType]
	if p.Generator == nil {
		p.Generator = generators["sphere"]
	}
	return &p
}

// ChunkIndex stores the latitude, longitude, and altitude index of a chunk
type ChunkIndex struct {
	Lon, Lat, Alt int
}

// PlanetChunkIndex stores the planet, latitude, longitude, and altitude index of a chunk
type PlanetChunkIndex struct {
	Planet int
	ChunkIndex
}

// CellIndex stores the latitude, longitude, and altitude index of a cell
type CellIndex struct {
	Lon, Lat, Alt int
}

// CellLoc stores the latitude, longitude, and altitude of a position in cell coordinates
type CellLoc struct {
	Lon, Lat, Alt float32
}

// GetChunk retrieves the chunk of a planet from chunk indices, either synchronously or asynchronously
func (p *Planet) GetChunk(ind ChunkIndex, async bool) *Chunk {
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
				p.databaseMutex.Lock()
				rows, e := p.db.Query("SELECT data FROM chunk WHERE planet = ? AND lon = ? AND lat = ? AND alt = ?", p.ID, ind.Lon, ind.Lat, ind.Alt)
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
				p.databaseMutex.Unlock()
				if chunk == nil {
					chunk = newChunk(ind, p)
					p.databaseMutex.Lock()
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
					_, e = stmt.Exec(p.ID, ind.Lon, ind.Lat, ind.Alt, buf.Bytes())
					if e != nil {
						panic(e)
					}
					p.databaseMutex.Unlock()
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
			pind := PlanetChunkIndex{Planet: p.ID, ChunkIndex: ind}
			if async {
				call := p.rpc.Go("API.GetChunk", pind, &rchunk, nil)
				go func() {
					call = <-call.Done
					p.ChunksMutex.Lock()
					p.Chunks[ind] = &rchunk
					p.ChunksMutex.Unlock()
				}()
				p.ChunksMutex.Lock()
				p.Chunks[ind] = &Chunk{WaitingForData: true}
				p.ChunksMutex.Unlock()
			} else {
				e := p.rpc.Call("API.GetChunk", pind, &rchunk)
				if e != nil {
					panic(e)
				}
				p.ChunksMutex.Lock()
				p.Chunks[ind] = &rchunk
				p.ChunksMutex.Unlock()
			}
		}
	}
	return chunk
}

// RPCSetCellMaterialArgs contains the arguments for the SetCellMaterial RPC call
type RPCSetCellMaterialArgs struct {
	Planet   int
	Index    CellIndex
	Material int
}

// SetCellMaterial sets the material for a cell
func (p *Planet) SetCellMaterial(ind CellIndex, material int, updateServer bool) bool {
	cell := p.CellIndexToCell(ind)
	if cell == nil {
		return false
	}
	if cell.Material == material {
		return false
	}
	cell.Material = material
	if p.rpc != nil && updateServer {
		var ret bool
		p.rpc.Go("API.SetCellMaterial", RPCSetCellMaterialArgs{
			Planet:   p.ID,
			Index:    ind,
			Material: material,
		}, &ret, nil)
	}
	if p.db != nil {
		chunkInd := p.CellIndexToChunkIndex(ind)
		chunk := p.CellIndexToChunk(ind)
		p.databaseMutex.Lock()
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
		p.databaseMutex.Unlock()
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
	return p.GetChunk(ind, true)
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

// CellIndexToCellLoc converts a cell index to floating-point cell indices
func (p *Planet) CellIndexToCellLoc(l CellIndex) CellLoc {
	return CellLoc{Lon: float32(l.Lon), Lat: float32(l.Lat), Alt: float32(l.Alt)}
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
	lonCells, latCells := p.LonLatCellsInChunkIndex(chunkIndex)
	lonWidth := ChunkSize / lonCells
	latWidth := ChunkSize / latCells
	chunk := p.CellIndexToChunk(cellIndex)
	if chunk == nil {
		return nil
	}
	lonInd := (cellIndex.Lon % ChunkSize) / lonWidth
	latInd := (cellIndex.Lat % ChunkSize) / latWidth
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

// CellIndexToCartesian converts a cell index to world coordinates
func (p *Planet) CellIndexToCartesian(ind CellIndex) mgl32.Vec3 {
	loc := p.CellIndexToCellLoc(ind)
	return p.CellLocToCartesian(loc)
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

// LonLatCellsInChunkIndex returns the number of longitude and latitude cells in a chunk, which changes based on latitude and altitude
func (p *Planet) LonLatCellsInChunkIndex(ind ChunkIndex) (lonCells, latCells int) {
	lonCells = ChunkSize
	latCells = ChunkSize

	// If chunk is too close to the poles, lower the longitude cells per chunk
	theta := (90.0 - float32(p.LatMax) + (float32(ind.Lat)+0.5)*float32(ChunkSize)/float32(p.LatCells)) * (2.0 * float32(p.LatMax))
	if math.Abs(float64(theta-90)) >= 60 {
		lonCells /= 2
	}
	if math.Abs(float64(theta-90)) >= 80 {
		lonCells /= 2
	}

	// If chunk is too close to the center of the planet, lower both lon and lat cells per chunk
	if (float64(ind.Alt)+0.5)*ChunkSize < p.Radius/4 {
		lonCells /= 2
		latCells /= 2
	}
	if (float64(ind.Alt)+0.5)*ChunkSize < p.Radius/8 {
		lonCells /= 2
		latCells /= 2
	}

	return
}

// Chunk is a 3D block of planet cells
type Chunk struct {
	WaitingForData bool
	Cells          [][][]*Cell
}

func newChunk(ind ChunkIndex, p *Planet) *Chunk {
	chunk := Chunk{}
	lonCells, latCells := p.LonLatCellsInChunkIndex(ind)
	lonWidth := ChunkSize / lonCells
	latWidth := ChunkSize / latCells
	chunk.Cells = make([][][]*Cell, lonCells)
	for lonIndex := 0; lonIndex < lonCells; lonIndex++ {
		chunk.Cells[lonIndex] = make([][]*Cell, latCells)
		for latIndex := 0; latIndex < latCells; latIndex++ {
			for altIndex := 0; altIndex < ChunkSize; altIndex++ {
				c := Cell{}
				l := CellLoc{
					Lon: float32(ChunkSize*ind.Lon + lonIndex*lonWidth),
					Lat: float32(ChunkSize*ind.Lat + latIndex*latWidth),
					Alt: float32(ChunkSize*ind.Alt + altIndex),
				}

				c.Material = p.Generator(p, l)

				// Always give the planet a solid core
				if l.Alt < 2 {
					c.Material = Stone
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

type stringSlice []string

func (slice stringSlice) pos(value string) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}

// List of materials
var (
	Materials = stringSlice{
		"air",
		"grass",
		"dirt",
		"stone",
		"moon",
		"asteroid",
		"sun",
		"blue_block",
		"blue_sand",
		"purple_block",
		"purple_sand",
		"red_block",
		"red_sand",
		"yellow_block",
		"yellow_sand",
	}
	MaterialColors = []mgl32.Vec3{
		{0.0, 0.0, 0.0},
		{0.5, 1.0, 0.5},
		{0.5, 0.3, 0.0},
		{0.5, 0.5, 0.5},
		{0.7, 0.7, 0.7},
		{0.4, 0.4, 0.4},
		{1.0, 0.9, 0.5},
		{0.5, 0.5, 1.0},
		{0.5, 0.5, 1.0},
		{1.0, 0.0, 1.0},
		{1.0, 0.0, 1.0},
		{1.0, 0.5, 0.5},
		{1.0, 0.5, 0.5},
		{1.0, 1.0, 0.0},
		{1.0, 1.0, 0.0},
	}
	Air         = Materials.pos("air")
	Grass       = Materials.pos("grass")
	Dirt        = Materials.pos("dirt")
	Stone       = Materials.pos("stone")
	Moon        = Materials.pos("moon")
	Asteroid    = Materials.pos("asteroid")
	Sun         = Materials.pos("sun")
	BlueBlock   = Materials.pos("blue_block")
	BlueSand    = Materials.pos("blue_sand")
	PurpleBlock = Materials.pos("purple_block")
	PurpleSand  = Materials.pos("purple_sand")
	RedBlock    = Materials.pos("red_block")
	RedSand     = Materials.pos("red_sand")
	YellowBlock = Materials.pos("yellow_block")
	YellowSand  = Materials.pos("yellow_sand")
)

// PlanetGeometry holds the low-resolution geometry for a planet.
type PlanetGeometry struct {
	Altitude  [][]int
	Material  [][]int
	IsLoading bool
}

func (p *Planet) generateGeometry() *PlanetGeometry {
	geom := PlanetGeometry{}
	lonCells := 64
	latCells := 32 + 1
	geom.Material = make([][]int, lonCells)
	geom.Altitude = make([][]int, lonCells)
	for lon := 0; lon < lonCells; lon++ {
		for lat := 0; lat < latCells; lat++ {
			lonInd := math.Floor(float64(p.LonCells) * float64(lon) / float64(lonCells))

			// Make sure latitude hits both poles, hence the need for division by (latCells - 1)
			latInd := math.Floor(float64(p.LatCells) * float64(lat) / float64(latCells-1))

			loc := CellLoc{Lon: float32(lonInd), Lat: float32(latInd), Alt: float32(p.AltCells - 1)}
			m := p.Generator(p, loc)
			for m == Air && loc.Alt > 0 {
				loc.Alt--
				m = p.Generator(p, loc)
			}
			geom.Material[lon] = append(geom.Material[lon], m)
			geom.Altitude[lon] = append(geom.Altitude[lon], int(loc.Alt))
		}
	}
	return &geom
}

// GetGeometry returns the low-resultion geometry for the planet.
func (p *Planet) GetGeometry(async bool) *PlanetGeometry {
	if p.Geometry != nil && p.Geometry.IsLoading {
		return nil
	}
	if p.Geometry != nil {
		return p.Geometry
	}
	if p.rpc != nil {
		if async {
			geom := PlanetGeometry{}
			call := p.rpc.Go("API.GetPlanetGeometry", &p.ID, &geom, nil)
			go func() {
				call = <-call.Done
				p.GeometryMutex.Lock()
				p.Geometry = &geom
				p.GeometryMutex.Unlock()
			}()
			p.GeometryMutex.Lock()
			p.Geometry = &PlanetGeometry{IsLoading: true}
			p.GeometryMutex.Unlock()
			return p.Geometry
		}
		geom := PlanetGeometry{}
		e := p.rpc.Call("API.GetPlanetGeometry", &p.ID, &geom)
		if e != nil {
			panic(e)
		}
		p.GeometryMutex.Lock()
		p.Geometry = &geom
		p.GeometryMutex.Unlock()
		return p.Geometry
	}
	p.Geometry = p.generateGeometry()
	return p.Geometry
}
