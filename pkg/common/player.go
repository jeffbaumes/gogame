package common

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

const (
	Normal       = iota
	Flying       = iota
	NumGameModes = iota
)

// Player represents a player of the game
type Player struct {
	UpVel            float32
	DownVel          float32
	ForwardVel       float32
	BackVel          float32
	RightVel         float32
	LeftVel          float32
	WalkVel          float32
	FallVel          float32
	Loc              mgl32.Vec3
	lookHeading      mgl32.Vec3
	lookAltitude     float64
	height           float64
	radius           float64
	GameMode         int
	HoldingJump      bool
	inJump           bool
	Name             string
	ActiveHotBarSlot int
	FocusCellIndex   CellIndex
	InInventory      bool
	HotbarOn         bool
	Hotbar           [12]int
	renderDistance   int
}

func NewPlayer(name string) *Player {
	p := Player{}
	p.WalkVel = 5.0
	p.Loc = mgl32.Vec3{60, 0, 0}
	p.lookHeading = mgl32.Vec3{0, 1, 0}
	p.height = 2
	p.radius = 0.25
	p.GameMode = Normal
	p.Name = name
	p.ActiveHotBarSlot = 0
	p.HotbarOn = true
	p.renderDistance = 4
	return &p
}

func (player *Player) LookDir() mgl32.Vec3 {
	up := player.Loc.Normalize()
	player.lookHeading = ProjectToPlane(player.lookHeading, up).Normalize()
	right := player.lookHeading.Cross(up)
	return mgl32.QuatRotate(float32((player.lookAltitude-90.0)*math.Pi/180.0), right).Rotate(up)
}

func (player *Player) Swivel(deltaX float64, deltaY float64) {
	lookHeadingDelta := -0.1 * deltaX
	normalDir := player.Loc.Normalize()
	player.lookHeading = mgl32.QuatRotate(float32(lookHeadingDelta*math.Pi/180.0), normalDir).Rotate(player.lookHeading)
	player.lookAltitude = player.lookAltitude - 0.1*deltaY
	player.lookAltitude = math.Max(math.Min(player.lookAltitude, 89.9), -89.9)
}

func (player *Player) UpdatePosition(h float32, planet *Planet) {
	up := player.Loc.Normalize()
	right := player.lookHeading.Cross(up)
	if player.GameMode == Normal {
		feet := player.Loc.Sub(up.Mul(float32(player.height)))
		feetCell := planet.CartesianToCell(feet)

		ind := planet.CartesianToChunkIndex(feet)

		for lon := ind.Lon - player.renderDistance; lon <= ind.Lon+player.renderDistance; lon++ {
			validLon := lon
			for validLon < 0 {
				validLon += planet.LonCells / ChunkSize
			}
			for validLon >= planet.LonCells/ChunkSize {
				validLon -= planet.LonCells / ChunkSize
			}
			latMin := Max(ind.Lat-player.renderDistance, 0)
			latMax := Min(ind.Lat+player.renderDistance, planet.LatCells/ChunkSize-1)
			for lat := latMin; lat <= latMax; lat++ {
				for alt := 0; alt < planet.AltCells/ChunkSize; alt++ {
					planet.GetChunk(ChunkIndex{Lon: validLon, Lat: lat, Alt: alt})
				}
			}
		}

		falling := feetCell == nil || feetCell.Material == Air
		if falling {
			player.FallVel -= 20 * h
		} else if player.HoldingJump && !player.inJump {
			player.FallVel = 7
			player.inJump = true
		} else {
			player.FallVel = 0
			player.inJump = false
		}

		playerVel := mgl32.Vec3{}
		playerVel = playerVel.Add(up.Mul(player.FallVel))
		playerVel = playerVel.Add(player.lookHeading.Mul((player.ForwardVel - player.BackVel)))
		playerVel = playerVel.Add(right.Mul((player.RightVel - player.LeftVel)))

		player.Loc = player.Loc.Add(playerVel.Mul(h))
		for height := planet.AltDelta / 2; height < player.height; height += planet.AltDelta {
			player.collide(planet, float32(height), CellLoc{Lon: 0, Lat: 0, Alt: -1})
			player.collide(planet, float32(height), CellLoc{Lon: 1, Lat: 0, Alt: 0})
			player.collide(planet, float32(height), CellLoc{Lon: -1, Lat: 0, Alt: 0})
			player.collide(planet, float32(height), CellLoc{Lon: 0, Lat: 1, Alt: 0})
			player.collide(planet, float32(height), CellLoc{Lon: 0, Lat: -1, Alt: 0})
		}
	} else if player.GameMode == Flying {
		LookDir := player.LookDir()
		player.Loc = player.Loc.Add(up.Mul((player.UpVel - player.DownVel) * h))
		player.Loc = player.Loc.Add(LookDir.Mul((player.ForwardVel - player.BackVel) * h))
		player.Loc = player.Loc.Add(right.Mul((player.RightVel - player.LeftVel) * h))
	}

	// Update focused cell
	increment := player.LookDir().Mul(0.05)
	pos := player.Loc
	player.FocusCellIndex = CellIndex{Lat: 0, Lon: 0, Alt: 0}
	for i := 0; i < 100; i++ {
		pos = pos.Add(increment)
		cell := planet.CartesianToCell(pos)
		if cell != nil && cell.Material != Air {
			cellIndex := planet.CartesianToCellIndex(pos)
			player.FocusCellIndex = cellIndex
			break
		}
	}
}

func (player *Player) collide(p *Planet, height float32, d CellLoc) {
	up := player.Loc.Normalize()
	pos := player.Loc.Sub(up.Mul(float32(player.height) - height))
	l := p.CartesianToCellLoc(pos)
	c := p.CellLocToNearestCellCenter(l)
	adjCell := p.CellLocToCell(CellLoc{
		Lon: c.Lon + d.Lon,
		Lat: c.Lat + d.Lat,
		Alt: c.Alt + d.Alt,
	})
	if adjCell != nil && adjCell.Material != Air {
		if d.Alt != 0 {
			nLoc := p.CellLocToCartesian(CellLoc{
				Lon: c.Lon + d.Lon/2,
				Lat: c.Lat + d.Lat/2,
				Alt: c.Alt + d.Alt/2,
			})
			distToPlane := up.Dot(pos.Sub(nLoc))
			if distToPlane < 0 {
				move := -distToPlane
				player.Loc = player.Loc.Add(up.Mul(move))
			}
		} else {
			nLoc := p.CellLocToCartesian(CellLoc{
				Lon: c.Lon + d.Lon/2,
				Lat: c.Lat + d.Lat/2,
				Alt: c.Alt + d.Alt/2,
			})
			aLoc := p.CellLocToCartesian(CellLoc{
				Lon: c.Lon + d.Lon,
				Lat: c.Lat + d.Lat,
				Alt: c.Alt + d.Alt,
			})
			cNorm := nLoc.Sub(aLoc).Normalize()
			cNorm = cNorm.Sub(Project(cNorm, up)).Normalize()
			distToPlane := cNorm.Dot(pos.Sub(nLoc))
			if distToPlane < float32(player.radius) {
				move := float32(player.radius) - distToPlane
				player.Loc = player.Loc.Add(cNorm.Mul(move))
			}
		}
	}
}
