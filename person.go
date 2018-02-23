package main

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
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
	p.loc = mgl32.Vec3{45, 0, 0}
	p.lookHeading = mgl32.Vec3{0, 1, 0}
	p.height = 2
	p.radius = 0.5
	p.gameMode = normal
	return &p
}

func (player *person) lookDir() mgl32.Vec3 {
	up := player.loc.Normalize()
	player.lookHeading = projectToPlane(player.lookHeading, up).Normalize()
	right := player.lookHeading.Cross(up)
	return mgl32.QuatRotate(float32((player.lookAltitude-90.0)*math.Pi/180.0), right).Rotate(up)
}

func (player *person) collide(p *planet, height, dLon, dLat, dAlt float32) {
	up := player.loc.Normalize()
	pos := player.loc.Sub(up.Mul(float32(player.height) - height))
	lon, lat, alt := p.cartesianToIndex(pos)
	cLon, cLat, cAlt := p.indexToCellCenterIndex(lon, lat, alt)
	adjCell := p.indexToCell(cLon+dLon, cLat+dLat, cAlt+dAlt)
	if adjCell != nil && adjCell.material != air {
		if dAlt != 0 {
			nLoc := p.indexToCartesian(cLon+dLon/2, cLat+dLat/2, cAlt+dAlt/2)
			distToPlane := up.Dot(pos.Sub(nLoc))
			if distToPlane < 0 {
				move := -distToPlane
				player.loc = player.loc.Add(up.Mul(move))
			}
		} else {
			nLoc := p.indexToCartesian(cLon+dLon/2, cLat+dLat/2, cAlt+dAlt/2)
			aLoc := p.indexToCartesian(cLon+dLon, cLat+dLat, cAlt+dAlt)
			cNorm := nLoc.Sub(aLoc).Normalize()
			cNorm = cNorm.Sub(project(cNorm, up)).Normalize()
			distToPlane := cNorm.Dot(pos.Sub(nLoc))
			if distToPlane < float32(player.radius) {
				move := float32(player.radius) - distToPlane
				player.loc = player.loc.Add(cNorm.Mul(move))
			}
		}
	}
}
