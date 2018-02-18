package main

import "github.com/go-gl/mathgl/mgl32"

type person struct {
	altVel       float32
	forwardVel   float32
	walkVel      float32
	loc          mgl32.Vec3
	lookHeading  mgl32.Vec3
	lookAltitude float64
}

func newPerson() *person {
	p := person{}
	p.walkVel = 5.0
	p.loc = mgl32.Vec3{25, 0, 0}
	p.lookHeading = mgl32.Vec3{0, 1, 0}
	return &p
}
