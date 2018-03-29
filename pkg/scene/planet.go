package scene

import (
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jeffbaumes/gogame/pkg/common"
)

// Planet draws visible chunks for a planet
type Planet struct {
	Planet            *common.Planet
	chunkRenderers    map[common.ChunkIndex]*chunkRenderer
	program           uint32
	texture           uint32
	textureUnit       int32
	textureUniform    int32
	projectionUniform int32
	sunDirUniform     int32
}

// NewPlanet creates a new planet renderer
func NewPlanet(planet *common.Planet) *Planet {
	const vertexShader = `
		#version 410
		in vec3 vp;
		in vec3 n;
		in vec2 t;
		uniform mat4 proj;
		uniform vec3 sundir;
		out vec3 color;
		out vec3 light;
		out vec2 texcoord;
		void main() {
			color = n;
			texcoord = t;
			gl_Position = proj * vec4(vp, 1.0);

			// Apply lighting effect
			highp vec3 ambientLight = vec3(0, 0, 0);
			highp vec3 vpn = normalize(vp);
			highp vec3 light1Color = vec3(0.9, 0.9, 0.9);
			highp float light1 = max(sqrt(dot(vpn, sundir)), 0.0);
			highp vec3 light2Color = vec3(0.2, 0.2, 0.2);
			highp float light2 = max(sqrt(1 - dot(vpn, sundir)), 0.0);
			highp vec3 light3Color = vec3(1.0, 0.5, 0.1);
			highp float light3 = max(0.4 - sqrt(abs(dot(vpn, sundir))), 0.0);
			light = ambientLight + (light1Color * light1) + (light2Color * light2) + (light3Color * light3);
		}
	`

	const fragmentShader = `
		#version 410
		in vec3 color;
		in vec3 light;
		in vec2 texcoord;
		uniform sampler2D texBase;
		out vec4 frag_color;
		void main() {
			vec4 texel = texture(texBase, texcoord);
			frag_color = texel * vec4(light, 1.0);
		}
	`

	pr := Planet{}
	pr.Planet = planet
	pr.program = createProgram(vertexShader, fragmentShader)
	pr.chunkRenderers = make(map[common.ChunkIndex]*chunkRenderer)
	bindAttribute(pr.program, 0, "vp")
	bindAttribute(pr.program, 1, "n")
	bindAttribute(pr.program, 2, "t")
	pr.projectionUniform = uniformLocation(pr.program, "proj")
	pr.sunDirUniform = uniformLocation(pr.program, "sundir")
	pr.textureUniform = uniformLocation(pr.program, "texBase")

	rgba := LoadTextures()

	pr.textureUnit = 1
	gl.ActiveTexture(uint32(gl.TEXTURE0 + pr.textureUnit))
	gl.GenTextures(1, &pr.texture)
	gl.BindTexture(gl.TEXTURE_2D, pr.texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix),
	)

	gl.GenerateMipmap(gl.TEXTURE_2D)

	return &pr
}

// SetCellMaterial sets the material at a particular cell and marks its chunk for redraw
func (planetRen *Planet) SetCellMaterial(ind common.CellIndex, material int) {
	planetRen.Planet.SetCellMaterial(ind, material)
	chunkInd := planetRen.Planet.CellIndexToChunkIndex(ind)
	chunkRen := planetRen.chunkRenderers[chunkInd]
	if chunkRen == nil {
		return
	}
	chunkRen.geometryUpdated = false
}

// Draw draws the planet's visible chunks
func (planetRen *Planet) Draw(player *common.Player, w *glfw.Window, time float64) {
	_, timeOfDay := math.Modf(time / planetRen.Planet.RotationSeconds)

	gl.UseProgram(planetRen.program)
	lookDir := player.LookDir()
	view := mgl32.LookAtV(player.Loc, player.Loc.Add(lookDir), player.Loc.Normalize())
	width, height := FramebufferSize(w)
	perspective := mgl32.Perspective(45, float32(width)/float32(height), 0.01, 1000)
	proj := perspective.Mul4(view)
	gl.UniformMatrix4fv(planetRen.projectionUniform, 1, false, &proj[0])
	gl.Uniform1i(planetRen.textureUniform, planetRen.textureUnit)
	sunAngle := timeOfDay * math.Pi * 2
	gl.Uniform3f(planetRen.sunDirUniform, float32(math.Sin(sunAngle)), float32(math.Cos(sunAngle)), 0)

	planetRen.Planet.ChunksMutex.Lock()
	for key, chunk := range planetRen.Planet.Chunks {
		if chunk.WaitingForData {
			continue
		}
		cr := planetRen.chunkRenderers[key]
		if cr == nil {
			cr = newChunkRenderer(chunk)
			planetRen.chunkRenderers[key] = cr
		}
		if !cr.geometryUpdated {
			cr.updateGeometry(planetRen.Planet, key.Lon, key.Lat, key.Alt)
		}
		cr.draw()
	}
	planetRen.Planet.ChunksMutex.Unlock()
}
