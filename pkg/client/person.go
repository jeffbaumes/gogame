package client

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jeffbaumes/gogame/pkg/geom"
)

const (
	normal       = iota
	flying       = iota
	numGameModes = iota
)

type person struct {
	upVel        float32
	downVel      float32
	forwardVel   float32
	backVel      float32
	rightVel     float32
	leftVel      float32
	walkVel      float32
	fallVel      float32
	loc          mgl32.Vec3
	lookHeading  mgl32.Vec3
	lookAltitude float64
	height       float64
	radius       float64
	gameMode     int
	holdingJump  bool
	inJump       bool
}

func newPerson() *person {
	p := person{}
	p.walkVel = 5.0
	p.loc = mgl32.Vec3{55, 0, 0}
	p.lookHeading = mgl32.Vec3{0, 1, 0}
	p.height = 2
	p.radius = 0.25
	p.gameMode = normal
	return &p
}

func (player *person) lookDir() mgl32.Vec3 {
	up := player.loc.Normalize()
	player.lookHeading = geom.ProjectToPlane(player.lookHeading, up).Normalize()
	right := player.lookHeading.Cross(up)
	return mgl32.QuatRotate(float32((player.lookAltitude-90.0)*math.Pi/180.0), right).Rotate(up)
}

func (player *person) updatePosition(h float32, planet *geom.Planet) {
	up := player.loc.Normalize()
	right := player.lookHeading.Cross(up)
	if player.gameMode == normal {
		feet := player.loc.Sub(up.Mul(float32(player.height)))
		feetCell := planet.CartesianToCell(feet)

		ind := planet.CartesianToChunkIndex(feet)

		for lon := ind.Lon - renderDistance; lon <= ind.Lon+renderDistance; lon++ {
			validLon := lon
			for validLon < 0 {
				validLon += planet.LonCells / geom.ChunkSize
			}
			for validLon >= planet.LonCells/geom.ChunkSize {
				validLon -= planet.LonCells / geom.ChunkSize
			}
			latMin := geom.Max(ind.Lat-renderDistance, 0)
			latMax := geom.Min(ind.Lat+renderDistance, planet.LatCells/geom.ChunkSize-1)
			for lat := latMin; lat <= latMax; lat++ {
				for alt := 0; alt < planet.AltCells/geom.ChunkSize; alt++ {
					planet.GetChunk(geom.ChunkIndex{Lon: validLon, Lat: lat, Alt: alt})
				}
			}
		}

		falling := feetCell == nil || feetCell.Material == geom.Air
		if falling {
			player.fallVel -= 20 * h
		} else if player.holdingJump && !player.inJump {
			player.fallVel = 7
			player.inJump = true
		} else {
			player.fallVel = 0
			player.inJump = false
		}

		playerVel := mgl32.Vec3{}
		playerVel = playerVel.Add(up.Mul(player.fallVel))
		playerVel = playerVel.Add(player.lookHeading.Mul((player.forwardVel - player.backVel)))
		playerVel = playerVel.Add(right.Mul((player.rightVel - player.leftVel)))

		player.loc = player.loc.Add(playerVel.Mul(h))
		for height := planet.AltDelta / 2; height < player.height; height += planet.AltDelta {
			player.collide(planet, float32(height), geom.CellLoc{Lon: 0, Lat: 0, Alt: -1})
			player.collide(planet, float32(height), geom.CellLoc{Lon: 1, Lat: 0, Alt: 0})
			player.collide(planet, float32(height), geom.CellLoc{Lon: -1, Lat: 0, Alt: 0})
			player.collide(planet, float32(height), geom.CellLoc{Lon: 0, Lat: 1, Alt: 0})
			player.collide(planet, float32(height), geom.CellLoc{Lon: 0, Lat: -1, Alt: 0})
		}
	} else if player.gameMode == flying {
		lookDir := player.lookDir()
		player.loc = player.loc.Add(up.Mul((player.upVel - player.downVel) * h))
		player.loc = player.loc.Add(lookDir.Mul((player.forwardVel - player.backVel) * h))
		player.loc = player.loc.Add(right.Mul((player.rightVel - player.leftVel) * h))
	}
}

func (player *person) collide(p *geom.Planet, height float32, d geom.CellLoc) {
	up := player.loc.Normalize()
	pos := player.loc.Sub(up.Mul(float32(player.height) - height))
	l := p.CartesianToCellLoc(pos)
	c := p.CellLocToNearestCellCenter(l)
	adjCell := p.CellLocToCell(geom.CellLoc{
		Lon: c.Lon + d.Lon,
		Lat: c.Lat + d.Lat,
		Alt: c.Alt + d.Alt,
	})
	if adjCell != nil && adjCell.Material != geom.Air {
		if d.Alt != 0 {
			nLoc := p.CellLocToCartesian(geom.CellLoc{
				Lon: c.Lon + d.Lon/2,
				Lat: c.Lat + d.Lat/2,
				Alt: c.Alt + d.Alt/2,
			})
			distToPlane := up.Dot(pos.Sub(nLoc))
			if distToPlane < 0 {
				move := -distToPlane
				player.loc = player.loc.Add(up.Mul(move))
			}
		} else {
			nLoc := p.CellLocToCartesian(geom.CellLoc{
				Lon: c.Lon + d.Lon/2,
				Lat: c.Lat + d.Lat/2,
				Alt: c.Alt + d.Alt/2,
			})
			aLoc := p.CellLocToCartesian(geom.CellLoc{
				Lon: c.Lon + d.Lon,
				Lat: c.Lat + d.Lat,
				Alt: c.Alt + d.Alt,
			})
			cNorm := nLoc.Sub(aLoc).Normalize()
			cNorm = cNorm.Sub(geom.Project(cNorm, up)).Normalize()
			distToPlane := cNorm.Dot(pos.Sub(nLoc))
			if distToPlane < float32(player.radius) {
				move := float32(player.radius) - distToPlane
				player.loc = player.loc.Add(cNorm.Mul(move))
			}
		}
	}
}
