package scene

import (
	"errors"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jeffbaumes/buildorb/pkg/common"
)

type chunkRenderer struct {
	chunk           *common.Chunk
	drawableVAO     uint32
	pointsVBO       uint32
	normalsVBO      uint32
	tcoordsVBO      uint32
	numTriangles    int32
	geometryUpdated bool
}

func newChunkRenderer(chunk *common.Chunk) *chunkRenderer {
	cr := chunkRenderer{}
	cr.chunk = chunk
	cr.pointsVBO = newVBO()
	cr.normalsVBO = newVBO()
	cr.tcoordsVBO = newVBO()
	cr.drawableVAO = newPointsNormalsTcoordsVAO(cr.pointsVBO, cr.normalsVBO, cr.tcoordsVBO)
	return &cr
}

func generateFace(cellIndex common.CellIndex, planet *common.Planet, points []float32, tcoords []float32, lonWidth, latWidth int, material int) (pts []float32, nms []float32, tcs []float32) {
	pts = make([]float32, len(points))
	for i := 0; i < len(points); i += 3 {
		l := common.CellLoc{
			Lon: float32(cellIndex.Lon) + float32(lonWidth-1)/2 + points[i+0]*float32(lonWidth),
			Lat: float32(cellIndex.Lat) + float32(latWidth-1)/2 + points[i+1]*float32(latWidth),
			Alt: float32(cellIndex.Alt) + points[i+2],
		}
		r, theta, phi := planet.CellLocToSpherical(l)
		cart := mgl32.SphericalToCartesian(r, theta, phi)
		pts[i] = cart[0]
		pts[i+1] = cart[1]
		pts[i+2] = cart[2]
	}

	nms = make([]float32, len(points))
	for i := 0; i < len(points); i += 9 {
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

	tcs = make([]float32, len(tcoords))
	for i := 0; i < len(tcoords); i += 2 {
		tcs[i+0] = (tcoords[i+0] + float32(material%4)) / 4
		tcs[i+1] = (tcoords[i+1] + float32(material/4)) / 4
	}

	return
}

func (cr *chunkRenderer) updateGeometry(planet *common.Planet, lonIndex, latIndex, altIndex int) {
	cs := common.ChunkSize
	points := []float32{}
	normals := []float32{}
	tcoords := []float32{}

	lonCells, latCells := planet.LonLatCellsInChunkIndex(common.ChunkIndex{Lon: lonIndex, Lat: latIndex, Alt: altIndex})
	lonWidth := common.ChunkSize / lonCells
	latWidth := common.ChunkSize / latCells

	chunkPosAlt := planet.Chunks[common.ChunkIndex{Lon: lonIndex, Lat: latIndex, Alt: altIndex + 1}]
	maxAltChunk := altIndex >= planet.AltCells/cs-1
	chunkNegAlt := planet.Chunks[common.ChunkIndex{Lon: lonIndex, Lat: latIndex, Alt: altIndex - 1}]
	minAltChunk := altIndex == 0

	chunkPosLat := planet.Chunks[common.ChunkIndex{Lon: lonIndex, Lat: latIndex + 1, Alt: altIndex}]
	chunkNegLat := planet.Chunks[common.ChunkIndex{Lon: lonIndex, Lat: latIndex - 1, Alt: altIndex}]

	lonChunks := planet.LonCells / common.ChunkSize
	lonPos := (lonIndex + 1) % lonChunks
	lonNeg := lonIndex - 1
	if lonNeg < 0 {
		lonNeg = lonChunks - 1
	}

	chunkPosLon := planet.Chunks[common.ChunkIndex{Lon: lonPos, Lat: latIndex, Alt: altIndex}]
	chunkNegLon := planet.Chunks[common.ChunkIndex{Lon: lonNeg, Lat: latIndex, Alt: altIndex}]

	hasAirAlt := func(c *common.Chunk, lon, lat, alt int) bool {
		if len(c.Cells) <= lonCells || len(c.Cells[0]) <= latCells {
			lonFactor := lonCells / len(c.Cells)
			latFactor := latCells / len(c.Cells[0])
			return c.Cells[lon/lonFactor][lat/latFactor][alt].Material == common.Air
		}
		lonFactor := len(c.Cells) / lonCells
		latFactor := len(c.Cells[0]) / latCells
		for olon := lon * lonFactor; olon < (lon+1)*lonFactor; olon++ {
			for olat := lat * latFactor; olat < (lat+1)*latFactor; olat++ {
				if c.Cells[olon][olat][alt].Material == common.Air {
					return true
				}
			}
		}
		return false
	}

	hasAirLat := func(c *common.Chunk, lon, lat, alt int) bool {
		if len(c.Cells[0]) != latCells {
			panic(errors.New("Chunks with same lon and alt should have the same lat cells"))
		}
		if len(c.Cells) <= lonCells {
			lonFactor := lonCells / len(c.Cells)
			return c.Cells[lon/lonFactor][lat][alt].Material == common.Air
		}
		lonFactor := len(c.Cells) / lonCells
		for olon := lon * lonFactor; olon < (lon+1)*lonFactor; olon++ {
			if c.Cells[olon][lat][alt].Material == common.Air {
				return true
			}
		}
		return false
	}

	hasAirLon := func(c *common.Chunk, lon, lat, alt int) bool {
		if len(c.Cells) != lonCells || len(c.Cells[0]) != latCells {
			panic(errors.New("Chunks with same lat and alt should have the same cell dimensions"))
		}
		return c.Cells[lon][lat][alt].Material == common.Air
	}

	for cLon := 0; cLon < lonCells; cLon++ {
		for cLat := 0; cLat < latCells; cLat++ {
			for cAlt := 0; cAlt < cs; cAlt++ {
				cellIndex := common.CellIndex{
					Lon: cs*lonIndex + cLon*lonWidth,
					Lat: cs*latIndex + cLat*latWidth,
					Alt: cs*altIndex + cAlt,
				}
				cell := cr.chunk.Cells[cLon][cLat][cAlt]
				if cell.Material != common.Air {
					if (cAlt+1 >= cs && chunkPosAlt != nil && hasAirAlt(chunkPosAlt, cLon, cLat, 0)) || (cAlt+1 >= cs && maxAltChunk) || (cAlt+1 < cs && cr.chunk.Cells[cLon][cLat][cAlt+1].Material == common.Air) {
						pts, nms, tcs := generateFace(cellIndex, planet, cubePosZ, cubeTcoordPosZ, lonWidth, latWidth, cell.Material)
						points = append(points, pts...)
						normals = append(normals, nms...)
						tcoords = append(tcoords, tcs...)
					}
					if (cAlt-1 < 0 && chunkNegAlt != nil && hasAirAlt(chunkNegAlt, cLon, cLat, cs-1)) || (cAlt-1 < 0 && minAltChunk) || (cAlt-1 >= 0 && cr.chunk.Cells[cLon][cLat][cAlt-1].Material == common.Air) {
						pts, nms, tcs := generateFace(cellIndex, planet, cubeNegZ, cubeTcoordNegZ, lonWidth, latWidth, cell.Material)
						points = append(points, pts...)
						normals = append(normals, nms...)
						tcoords = append(tcoords, tcs...)
					}
					if (cLon+1 >= lonCells && chunkPosLon != nil && hasAirLon(chunkPosLon, 0, cLat, cAlt)) || (cLon+1 < lonCells && cr.chunk.Cells[cLon+1][cLat][cAlt].Material == common.Air) {
						pts, nms, tcs := generateFace(cellIndex, planet, cubePosX, cubeTcoordPosX, lonWidth, latWidth, cell.Material)
						points = append(points, pts...)
						normals = append(normals, nms...)
						tcoords = append(tcoords, tcs...)
					}
					if (cLon-1 < 0 && chunkNegLon != nil && hasAirLon(chunkNegLon, lonCells-1, cLat, cAlt)) || (cLon-1 >= 0 && cr.chunk.Cells[cLon-1][cLat][cAlt].Material == common.Air) {
						pts, nms, tcs := generateFace(cellIndex, planet, cubeNegX, cubeTcoordNegX, lonWidth, latWidth, cell.Material)
						points = append(points, pts...)
						normals = append(normals, nms...)
						tcoords = append(tcoords, tcs...)
					}
					if (cLat+1 >= latCells && chunkPosLat != nil && hasAirLat(chunkPosLat, cLon, 0, cAlt)) || (cLat+1 < latCells && cr.chunk.Cells[cLon][cLat+1][cAlt].Material == common.Air) {
						pts, nms, tcs := generateFace(cellIndex, planet, cubePosY, cubeTcoordPosY, lonWidth, latWidth, cell.Material)
						points = append(points, pts...)
						normals = append(normals, nms...)
						tcoords = append(tcoords, tcs...)
					}
					if (cLat-1 < 0 && chunkNegLat != nil && hasAirLat(chunkNegLat, cLon, latCells-1, cAlt)) || (cLat-1 >= 0 && cr.chunk.Cells[cLon][cLat-1][cAlt].Material == common.Air) {
						pts, nms, tcs := generateFace(cellIndex, planet, cubeNegY, cubeTcoordNegY, lonWidth, latWidth, cell.Material)
						points = append(points, pts...)
						normals = append(normals, nms...)
						tcoords = append(tcoords, tcs...)
					}
				}
			}
		}
	}
	cr.numTriangles = int32(len(points) / 3)
	if cr.numTriangles > 0 {
		fillVBO(cr.pointsVBO, points)
		fillVBO(cr.normalsVBO, normals)
		fillVBO(cr.tcoordsVBO, tcoords)
		cr.geometryUpdated = true
	}
}

func (cr *chunkRenderer) draw() {
	if cr.numTriangles > 0 {
		gl.BindVertexArray(cr.drawableVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, cr.numTriangles)
	}
}
