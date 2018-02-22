package main

import "github.com/go-gl/mathgl/mgl32"

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
