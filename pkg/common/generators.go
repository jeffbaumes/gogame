package common

var (
	generatorsInitialized bool
	generators            map[string](func(*Planet, CellLoc) int)
)

func initializeGenerators() {
	if generatorsInitialized {
		return
	}
	generatorsInitialized = true

	generators = make(map[string](func(*Planet, CellLoc) int))

	generators["sphere"] = func(p *Planet, loc CellLoc) int {
		if float64(loc.Alt)/float64(p.AltCells) < 0.5 {
			return Stone
		}
		return Air
	}

	generators["rings"] = func(p *Planet, loc CellLoc) int {
		scale := 1.0
		n := p.noise.Eval2(float64(loc.Alt)*scale, 0)
		fracHeight := float64(loc.Alt) / float64(p.AltCells)
		if fracHeight < 0.5 {
			return Grass
		}
		if fracHeight > 0.6 && int(loc.Lat) == p.LatCells/2 {
			if n > 0.1 {
				return YellowBlock
			}
			return RedBlock
		}
		return Air
	}

	generators["bumpy"] = func(p *Planet, loc CellLoc) int {
		pos := p.CellLocToCartesian(loc).Normalize().Mul(float32(p.AltCells / 2))
		scale := 0.1
		height := float64(p.AltCells)/2 + p.noise.Eval3(float64(pos[0])*scale, float64(pos[1])*scale, float64(pos[2])*scale)*4
		if float64(loc.Alt) <= height {
			return Stone
		}
		return Air
	}

	generators["caves"] = func(p *Planet, loc CellLoc) int {
		pos := p.CellLocToCartesian(loc)
		const scale = 0.05
		height := (p.noise.Eval3(float64(pos[0])*scale, float64(pos[1])*scale, float64(pos[2])*scale) + 1.0) * float64(p.AltCells) / 2.0
		if height > float64(p.AltCells)/2 {
			return Stone
		}
		return Air
	}

	generators["rocks"] = func(p *Planet, loc CellLoc) int {
		pos := p.CellLocToCartesian(loc)
		const scale = 0.05
		noise := p.noise.Eval3(float64(pos[0])*scale, float64(pos[1])*scale, float64(pos[2])*scale)
		if noise > 0.5 {
			return Stone
		}
		return Air
	}
}
