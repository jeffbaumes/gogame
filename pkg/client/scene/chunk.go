package scene

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jeffbaumes/gogame/pkg/common"
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

func generateFace(cellIndex common.CellIndex, planet *common.Planet, points []float32, tcoords []float32, lonWidth int, material int) (pts []float32, nms []float32, tcs []float32) {
	pts = make([]float32, len(points))
	for i := 0; i < len(points); i += 3 {
		l := common.CellLoc{
			Lon: float32(cellIndex.Lon) + float32(lonWidth-1)/2 + points[i+0]*float32(lonWidth),
			Lat: float32(cellIndex.Lat) + points[i+1],
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

	lonCells := planet.LonCellsInChunkIndex(common.ChunkIndex{Lon: lonIndex, Lat: latIndex, Alt: altIndex})
	lonWidth := common.ChunkSize / lonCells

	for cLon := 0; cLon < lonCells; cLon++ {
		for cLat := 0; cLat < cs; cLat++ {
			for cAlt := 0; cAlt < cs; cAlt++ {
				cellIndex := common.CellIndex{
					Lon: cs*lonIndex + cLon*lonWidth,
					Lat: cs*latIndex + cLat,
					Alt: cs*altIndex + cAlt,
				}
				cell := cr.chunk.Cells[cLon][cLat][cAlt]
				if cell.Material != common.Air {
					if cAlt+1 >= cs || cr.chunk.Cells[cLon][cLat][cAlt+1].Material == common.Air {
						pts, nms, tcs := generateFace(cellIndex, planet, cubePosZ, cubeTcoordPosZ, lonWidth, cell.Material)
						points = append(points, pts...)
						normals = append(normals, nms...)
						tcoords = append(tcoords, tcs...)
					}
					if cAlt-1 < 0 || cr.chunk.Cells[cLon][cLat][cAlt-1].Material == common.Air {
						pts, nms, tcs := generateFace(cellIndex, planet, cubeNegZ, cubeTcoordNegZ, lonWidth, cell.Material)
						points = append(points, pts...)
						normals = append(normals, nms...)
						tcoords = append(tcoords, tcs...)
					}
					if cLon+1 >= lonCells || cr.chunk.Cells[cLon+1][cLat][cAlt].Material == common.Air {
						pts, nms, tcs := generateFace(cellIndex, planet, cubePosX, cubeTcoordPosX, lonWidth, cell.Material)
						points = append(points, pts...)
						normals = append(normals, nms...)
						tcoords = append(tcoords, tcs...)
					}
					if cLon-1 < 0 || cr.chunk.Cells[cLon-1][cLat][cAlt].Material == common.Air {
						pts, nms, tcs := generateFace(cellIndex, planet, cubeNegX, cubeTcoordNegX, lonWidth, cell.Material)
						points = append(points, pts...)
						normals = append(normals, nms...)
						tcoords = append(tcoords, tcs...)
					}
					if cLat+1 >= cs || cr.chunk.Cells[cLon][cLat+1][cAlt].Material == common.Air {
						pts, nms, tcs := generateFace(cellIndex, planet, cubePosY, cubeTcoordPosY, lonWidth, cell.Material)
						points = append(points, pts...)
						normals = append(normals, nms...)
						tcoords = append(tcoords, tcs...)
					}
					if cLat-1 < 0 || cr.chunk.Cells[cLon][cLat-1][cAlt].Material == common.Air {
						pts, nms, tcs := generateFace(cellIndex, planet, cubeNegY, cubeTcoordNegY, lonWidth, cell.Material)
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
