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
