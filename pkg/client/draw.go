package client

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jeffbaumes/gogame/pkg/geom"
)

var (
	squareTcoords = []float32{
		0, 0,
		0, 1,
		1, 0,

		1, 0,
		0, 1,
		1, 1,

		0, 1,
		0, 0,
		1, 0,

		0, 1,
		1, 0,
		1, 1,

		0, 0,
		0, 1,
		1, 0,

		1, 0,
		0, 1,
		1, 1,

		0, 1,
		0, 0,
		1, 0,

		0, 1,
		1, 0,
		1, 1,

		0, 0,
		0, 1,
		1, 0,

		1, 0,
		0, 1,
		1, 1,

		0, 1,
		0, 0,
		1, 0,

		0, 1,
		1, 0,
		1, 1,
	}

	square = []float32{
		-0.5, -0.5, 0.5,
		-0.5, 0.5, 0.5,
		0.5, -0.5, 0.5,

		0.5, -0.5, 0.5,
		-0.5, 0.5, 0.5,
		0.5, 0.5, 0.5,

		-0.5, 0.5, -0.5,
		-0.5, -0.5, -0.5,
		0.5, -0.5, -0.5,

		-0.5, 0.5, -0.5,
		0.5, -0.5, -0.5,
		0.5, 0.5, -0.5,

		0.5, -0.5, -0.5,
		0.5, -0.5, 0.5,
		0.5, 0.5, -0.5,

		0.5, 0.5, -0.5,
		0.5, -0.5, 0.5,
		0.5, 0.5, 0.5,

		-0.5, -0.5, 0.5,
		-0.5, -0.5, -0.5,
		-0.5, 0.5, -0.5,

		-0.5, -0.5, 0.5,
		-0.5, 0.5, -0.5,
		-0.5, 0.5, 0.5,

		-0.5, 0.5, -0.5,
		0.5, 0.5, -0.5,
		-0.5, 0.5, 0.5,

		-0.5, 0.5, 0.5,
		0.5, 0.5, -0.5,
		0.5, 0.5, 0.5,

		0.5, -0.5, -0.5,
		-0.5, -0.5, -0.5,
		-0.5, -0.5, 0.5,

		0.5, -0.5, -0.5,
		-0.5, -0.5, 0.5,
		0.5, -0.5, 0.5,
	}

	lw  float32 = 0.1
	box         = []float32{
		0.5, 0.5, 0.5,
		-0.5, 0.5, 0.5,
		0.5, 0.5 - lw, 0.5,
		0.5, 0.5 - lw, 0.5,
		-0.5, 0.5, 0.5,
		-0.5, 0.5 - lw, 0.5,
		0.5, 0.5, 0.5,
		-0.5, 0.5, 0.5,
		0.5, 0.5, 0.5 - lw,
		0.5, 0.5, 0.5 - lw,
		-0.5, 0.5, 0.5,
		-0.5, 0.5, 0.5 - lw,

		0.5, 0.5, -0.5,
		-0.5, 0.5, -0.5,
		0.5, 0.5 - lw, -0.5,
		0.5, 0.5 - lw, -0.5,
		-0.5, 0.5, -0.5,
		-0.5, 0.5 - lw, -0.5,
		0.5, 0.5, -0.5,
		-0.5, 0.5, -0.5,
		0.5, 0.5, -0.5 + lw,
		0.5, 0.5, -0.5 + lw,
		-0.5, 0.5, -0.5,
		-0.5, 0.5, -0.5 + lw,

		0.5, -0.5, 0.5,
		-0.5, -0.5, 0.5,
		0.5, -0.5 + lw, 0.5,
		0.5, -0.5 + lw, 0.5,
		-0.5, -0.5, 0.5,
		-0.5, -0.5 + lw, 0.5,
		0.5, -0.5, 0.5,
		-0.5, -0.5, 0.5,
		0.5, -0.5, 0.5 - lw,
		0.5, -0.5, 0.5 - lw,
		-0.5, -0.5, 0.5,
		-0.5, -0.5, 0.5 - lw,

		0.5, -0.5, -0.5,
		-0.5, -0.5, -0.5,
		0.5, -0.5 + lw, -0.5,
		0.5, -0.5 + lw, -0.5,
		-0.5, -0.5, -0.5,
		-0.5, -0.5 + lw, -0.5,
		0.5, -0.5, -0.5,
		-0.5, -0.5, -0.5,
		0.5, -0.5, -0.5 + lw,
		0.5, -0.5, -0.5 + lw,
		-0.5, -0.5, -0.5,
		-0.5, -0.5, -0.5 + lw,

		0.5, 0.5, 0.5,
		0.5, -0.5, 0.5,
		0.5, 0.5, 0.5 - lw,
		0.5, 0.5, 0.5 - lw,
		0.5, -0.5, 0.5,
		0.5, -0.5, 0.5 - lw,
		0.5, 0.5, 0.5,
		0.5, -0.5, 0.5,
		0.5 - lw, 0.5, 0.5,
		0.5 - lw, 0.5, 0.5,
		0.5, -0.5, 0.5,
		0.5 - lw, -0.5, 0.5,

		-0.5, 0.5, 0.5,
		-0.5, -0.5, 0.5,
		-0.5, 0.5, 0.5 - lw,
		-0.5, 0.5, 0.5 - lw,
		-0.5, -0.5, 0.5,
		-0.5, -0.5, 0.5 - lw,
		-0.5, 0.5, 0.5,
		-0.5, -0.5, 0.5,
		-0.5 + lw, 0.5, 0.5,
		-0.5 + lw, 0.5, 0.5,
		-0.5, -0.5, 0.5,
		-0.5 + lw, -0.5, 0.5,

		0.5, 0.5, -0.5,
		0.5, -0.5, -0.5,
		0.5, 0.5, -0.5 + lw,
		0.5, 0.5, -0.5 + lw,
		0.5, -0.5, -0.5,
		0.5, -0.5, -0.5 + lw,
		0.5, 0.5, -0.5,
		0.5, -0.5, -0.5,
		0.5 - lw, 0.5, -0.5,
		0.5 - lw, 0.5, -0.5,
		0.5, -0.5, -0.5,
		0.5 - lw, -0.5, -0.5,

		-0.5, 0.5, -0.5,
		-0.5, -0.5, -0.5,
		-0.5, 0.5, -0.5 + lw,
		-0.5, 0.5, -0.5 + lw,
		-0.5, -0.5, -0.5,
		-0.5, -0.5, -0.5 + lw,
		-0.5, 0.5, -0.5,
		-0.5, -0.5, -0.5,
		-0.5 + lw, 0.5, -0.5,
		-0.5 + lw, 0.5, -0.5,
		-0.5, -0.5, -0.5,
		-0.5 + lw, -0.5, -0.5,

		0.5, 0.5, 0.5,
		0.5, 0.5, -0.5,
		0.5 - lw, 0.5, 0.5,
		0.5 - lw, 0.5, 0.5,
		0.5, 0.5, -0.5,
		0.5 - lw, 0.5, -0.5,
		0.5, 0.5, 0.5,
		0.5, 0.5, -0.5,
		0.5, 0.5 - lw, 0.5,
		0.5, 0.5 - lw, 0.5,
		0.5, 0.5, -0.5,
		0.5, 0.5 - lw, -0.5,

		-0.5, 0.5, 0.5,
		-0.5, 0.5, -0.5,
		-0.5 + lw, 0.5, 0.5,
		-0.5 + lw, 0.5, 0.5,
		-0.5, 0.5, -0.5,
		-0.5 + lw, 0.5, -0.5,
		-0.5, 0.5, 0.5,
		-0.5, 0.5, -0.5,
		-0.5, 0.5 - lw, 0.5,
		-0.5, 0.5 - lw, 0.5,
		-0.5, 0.5, -0.5,
		-0.5, 0.5 - lw, -0.5,

		0.5, -0.5, 0.5,
		0.5, -0.5, -0.5,
		0.5 - lw, -0.5, 0.5,
		0.5 - lw, -0.5, 0.5,
		0.5, -0.5, -0.5,
		0.5 - lw, -0.5, -0.5,
		0.5, -0.5, 0.5,
		0.5, -0.5, -0.5,
		0.5, -0.5 + lw, 0.5,
		0.5, -0.5 + lw, 0.5,
		0.5, -0.5, -0.5,
		0.5, -0.5 + lw, -0.5,

		-0.5, -0.5, 0.5,
		-0.5, -0.5, -0.5,
		-0.5 + lw, -0.5, 0.5,
		-0.5 + lw, -0.5, 0.5,
		-0.5, -0.5, -0.5,
		-0.5 + lw, -0.5, -0.5,
		-0.5, -0.5, 0.5,
		-0.5, -0.5, -0.5,
		-0.5, -0.5 + lw, 0.5,
		-0.5, -0.5 + lw, 0.5,
		-0.5, -0.5, -0.5,
		-0.5, -0.5 + lw, -0.5,
	}
)

const (
	vertexShaderSource = `
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

	fragmentShaderSource = `
		#version 410
		in vec3 color;
		in vec3 light;
		in vec2 texcoord;
		uniform sampler2D texBase;
		out vec4 frag_color;
		void main() {
			vec4 texel = texture(texBase, texcoord);
			// frag_color = vec4(1,1,1,1);
			// frag_color = vec4(light, 1.0);
			frag_color = texel * vec4(light, 1.0);
			// frag_color = vec4(light, 1.0);
		}
	`

	vertexShaderSourcePeople = `
		#version 410
		in vec3 vp;
		in vec3 n;
		uniform mat4 proj;
		out vec3 color;
		out vec3 light;
		void main() {
			color = n;
			gl_Position = proj * vec4(vp, 1.0);

			// Apply lighting effect
			highp vec3 ambientLight = vec3(0.0, 0.0, 0.2);
			highp vec3 light1Color = vec3(0.5, 0.5, 0.4);
			highp vec3 light1Dir = normalize(vec3(0.85, 0.8, 0.75));
			highp float light1 = max(dot(n, light1Dir), 0.0);
			highp vec3 light2Color = vec3(0.1, 0.1, 0.2);
			highp vec3 light2Dir = normalize(vec3(-0.85, -0.8, -0.75));
			highp float light2 = max(dot(n, light2Dir), 0.0);
			light = ambientLight + (light1Color * light1) + (light2Color * light2);
		}
	`

	fragmentShaderSourcePeople = `
		#version 410
		in vec3 color;
		in vec3 light;
		out vec4 frag_color;
		void main() {
			frag_color = vec4(light, 1.0);
		}
	`

	vertexShaderSourceHUD = `
		#version 410
		in vec3 vp;
		uniform mat4 proj;
		void main() {
			gl_Position = proj * vec4(vp, 1.0);
		}
	`

	fragmentShaderSourceHUD = `
		#version 410
		out vec4 frag_color;
		void main() {
			frag_color = vec4(1.0, 1.0, 1.0, 1.0);
		}
	`

	vertexShaderSourceText = `
		#version 410

		in vec4 coord;
		out vec2 texcoord;

		void main(void) {
			gl_Position = vec4(coord.xy, 0, 1);
			texcoord = coord.zw;
		}
	`

	fragmentShaderSourceText = `
		#version 410

		in vec2 texcoord;
		uniform sampler2D texFont;
		out vec4 frag_color;

		void main(void) {
			vec4 texel = texture(texFont, texcoord);
			if (texel.a < 0.5) {
				discard;
		  }
			frag_color = texel;
		}
	`

	vertexShaderSourceHotbar = `
		#version 410

		in vec4 coord;
		out vec2 tcoord;

		void main(void) {
			gl_Position = vec4(coord.xy, 0, 1);
			tcoord = coord.zw;
		}
	`

	fragmentShaderSourceHotbar = `
		#version 410

		in vec2 tcoord;
		uniform sampler2D tex;
		out vec4 frag_color;

		void main(void) {
			vec4 texel = texture(tex, tcoord);
			frag_color = texel;
		}
	`

	vertexShaderSourceFocus = `
		#version 410
		uniform mat4 proj;
		in vec3 position;
		void main() {
			gl_Position = proj * vec4(position, 1.0);
		}
	`

	fragmentShaderSourceFocus = `
		#version 410
		out vec4 frag_color;
		void main() {
			frag_color = vec4(0,0,0,1);
		}
	`
)

func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(500, 500, "World Blocks", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

func initOpenGL() {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	gl.Enable(gl.DEPTH_TEST)
	// gl.Enable(gl.POLYGON_OFFSET_FILL)
	// gl.PolygonOffset(2, 0)

	gl.LineWidth(5)
	gl.Enable(gl.LINE_SMOOTH)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

func drawFrame(h float32, player *person, text *screenText, over *overlay, planetRen *planetRenderer, peopleRen *peopleRenderer, focusRen *focusRenderer, bar *hotbar, window *glfw.Window, timeOfDay float64) {
	sunAngle := timeOfDay * math.Pi * 2
	sunDir := mgl32.Vec3{float32(math.Sin(sunAngle)), float32(math.Cos(sunAngle)), 0}
	vpnDotSun := float64(player.loc.Normalize().Dot(sunDir))
	light1Color := mgl32.Vec3{0.5, 0.7, 1.0}
	light1 := math.Max(math.Sqrt(vpnDotSun), 0)
	if math.IsNaN(light1) {
		light1 = 0
	}
	light2Color := mgl32.Vec3{0, 0, 0}
	light2 := math.Max(math.Sqrt(1-vpnDotSun), 0)
	if math.IsNaN(light2) {
		light2 = 0
	}
	light3Color := mgl32.Vec3{0.7, 0.5, 0.4}
	light3 := math.Max(0.6-math.Sqrt(math.Abs(vpnDotSun)), 0)
	if math.IsNaN(light3) {
		light3 = 0
	}
	light := light1Color.Mul(float32(light1)).Add(light2Color.Mul(float32(light2))).Add(light3Color.Mul(float32(light3)))

	gl.ClearColor(light.X(), light.Y(), light.Z(), 1)
	planetRen.planet.CartesianToCellIndex(player.loc)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	planetRen.draw(player, window, timeOfDay)
	peopleRen.draw(player, window)
	focusRen.draw(player, planetRen.planet, window)
	over.draw(window)
	text.draw(player, window)
	bar.draw(player, window)

	glfw.PollEvents()
	window.SwapBuffers()
}

type peopleRenderer struct {
	api               *API
	program           uint32
	drawableVAO       uint32
	pointsVBO         uint32
	normalsVBO        uint32
	projectionUniform int32
}

func newPeopleRenderer(api *API) *peopleRenderer {
	peopleRen := peopleRenderer{}
	peopleRen.api = api
	peopleRen.program = createProgram(vertexShaderSourcePeople, fragmentShaderSourcePeople)
	bindAttribute(peopleRen.program, 0, "vp")
	bindAttribute(peopleRen.program, 1, "n")
	peopleRen.projectionUniform = uniformLocation(peopleRen.program, "proj")
	peopleRen.pointsVBO = newVBO()
	fillVBO(peopleRen.pointsVBO, square)
	peopleRen.normalsVBO = newVBO()
	peopleRen.drawableVAO = newPointsNormalsVAO(peopleRen.pointsVBO, peopleRen.normalsVBO)
	return &peopleRen
}

func (peopleRen *peopleRenderer) draw(player *person, w *glfw.Window) {
	gl.UseProgram(peopleRen.program)
	for _, p := range peopleRen.api.connectedPeople {
		pts := make([]float32, len(square))
		for i := 0; i < len(square); i += 3 {
			pts[i] = p.Position[0] + square[i]
			pts[i+1] = p.Position[1] + square[i+1]
			pts[i+2] = p.Position[2] + square[i+2]
		}

		nms := make([]float32, len(square))
		for i := 0; i < len(square); i += 9 {
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

		fillVBO(peopleRen.normalsVBO, nms)

		lookDir := player.lookDir()
		view := mgl32.LookAtV(player.loc, player.loc.Add(lookDir), player.loc.Normalize())
		width, height := framebufferSize(w)
		perspective := mgl32.Perspective(45, float32(width)/float32(height), 0.01, 1000)
		proj := perspective.Mul4(view)
		proj = proj.Mul4(mgl32.Translate3D(p.Position[0], p.Position[1], p.Position[2]))
		right := p.LookDir.Cross(p.Position.Normalize()).Normalize()
		up := right.Cross(p.LookDir).Normalize()
		proj = proj.Mul4(mgl32.Mat4FromCols(p.LookDir.Vec4(0), up.Vec4(0), right.Vec4(0), mgl32.Vec4{0, 0, 0, 1}))
		gl.UniformMatrix4fv(peopleRen.projectionUniform, 1, false, &proj[0])

		gl.BindVertexArray(peopleRen.drawableVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
	}
}

type planetRenderer struct {
	planet            *geom.Planet
	chunkRenderers    map[geom.ChunkIndex]*chunkRenderer
	program           uint32
	texture           uint32
	textureUnit       int32
	textureUniform    int32
	projectionUniform int32
	sunDirUniform     int32
}

func newPlanetRenderer(planet *geom.Planet) *planetRenderer {
	pr := planetRenderer{}
	pr.planet = planet
	pr.program = createProgram(vertexShaderSource, fragmentShaderSource)
	pr.chunkRenderers = make(map[geom.ChunkIndex]*chunkRenderer)
	bindAttribute(pr.program, 0, "vp")
	bindAttribute(pr.program, 1, "n")
	bindAttribute(pr.program, 2, "t")
	pr.projectionUniform = uniformLocation(pr.program, "proj")
	pr.sunDirUniform = uniformLocation(pr.program, "sundir")
	pr.textureUniform = uniformLocation(pr.program, "texBase")

	// existingImageFile, err := os.Open("textures.png")
	// if err != nil {
	// 	panic(err)
	// }
	// defer existingImageFile.Close()
	// img, err := png.Decode(existingImageFile)
	// if err != nil {
	// 	panic(err)
	// }
	rgba := LoadTextures()
	// draw.Draw(rgba, rgba.Bounds(), img, image.Pt(0, 0), draw.Src)

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

func (planetRen *planetRenderer) setCellMaterial(ind geom.CellIndex, material int) {
	planetRen.planet.SetCellMaterial(ind, material)
	chunkInd := planetRen.planet.CellIndexToChunkIndex(ind)
	chunkRen := planetRen.chunkRenderers[chunkInd]
	if chunkRen == nil {
		return
	}
	chunkRen.geometryUpdated = false
}

func (planetRen *planetRenderer) draw(player *person, w *glfw.Window, timeOfDay float64) {
	gl.UseProgram(planetRen.program)
	lookDir := player.lookDir()
	view := mgl32.LookAtV(player.loc, player.loc.Add(lookDir), player.loc.Normalize())
	width, height := framebufferSize(w)
	perspective := mgl32.Perspective(45, float32(width)/float32(height), 0.01, 1000)
	proj := perspective.Mul4(view)
	gl.UniformMatrix4fv(planetRen.projectionUniform, 1, false, &proj[0])
	gl.Uniform1i(planetRen.textureUniform, planetRen.textureUnit)
	sunAngle := timeOfDay * math.Pi * 2
	gl.Uniform3f(planetRen.sunDirUniform, float32(math.Sin(sunAngle)), float32(math.Cos(sunAngle)), 0)

	planetRen.planet.ChunksMutex.Lock()
	for key, chunk := range planetRen.planet.Chunks {
		if chunk.WaitingForData {
			continue
		}
		cr := planetRen.chunkRenderers[key]
		if cr == nil {
			cr = newChunkRenderer(chunk)
			planetRen.chunkRenderers[key] = cr
		}
		if !cr.geometryUpdated {
			cr.updateGeometry(planetRen.planet, key.Lon, key.Lat, key.Alt)
		}
		cr.draw()
	}
	planetRen.planet.ChunksMutex.Unlock()
}

type chunkRenderer struct {
	chunk           *geom.Chunk
	drawableVAO     uint32
	pointsVBO       uint32
	normalsVBO      uint32
	tcoordsVBO      uint32
	numTriangles    int32
	geometryUpdated bool
}

func newChunkRenderer(chunk *geom.Chunk) *chunkRenderer {
	cr := chunkRenderer{}
	cr.chunk = chunk
	cr.pointsVBO = newVBO()
	cr.normalsVBO = newVBO()
	cr.tcoordsVBO = newVBO()
	cr.drawableVAO = newPointsNormalsTcoordsVAO(cr.pointsVBO, cr.normalsVBO, cr.tcoordsVBO)
	return &cr
}

func (cr *chunkRenderer) updateGeometry(planet *geom.Planet, lonIndex, latIndex, altIndex int) {
	cs := geom.ChunkSize
	points := []float32{}
	normals := []float32{}
	tcoords := []float32{}

	lonCells := planet.LonCellsInChunkIndex(geom.ChunkIndex{Lon: lonIndex, Lat: latIndex, Alt: altIndex})
	lonWidth := geom.ChunkSize / lonCells

	for cLon := 0; cLon < lonCells; cLon++ {
		for cLat := 0; cLat < cs; cLat++ {
			for cAlt := 0; cAlt < cs; cAlt++ {
				cellIndex := geom.CellIndex{
					Lon: cs*lonIndex + cLon*lonWidth,
					Lat: cs*latIndex + cLat,
					Alt: cs*altIndex + cAlt,
				}
				cell := cr.chunk.Cells[cLon][cLat][cAlt]
				if cell.Material != geom.Air {
					pts := make([]float32, len(square))
					for i := 0; i < len(square); i += 3 {
						l := geom.CellLoc{
							Lon: float32(cellIndex.Lon) + float32(lonWidth-1)/2 + square[i+0]*float32(lonWidth),
							Lat: float32(cellIndex.Lat) + square[i+1],
							Alt: float32(cellIndex.Alt) + square[i+2],
						}
						r, theta, phi := planet.CellLocToSpherical(l)
						cart := mgl32.SphericalToCartesian(r, theta, phi)
						pts[i] = cart[0]
						pts[i+1] = cart[1]
						pts[i+2] = cart[2]
					}
					points = append(points, pts...)

					nms := make([]float32, len(square))
					for i := 0; i < len(square); i += 9 {
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
					normals = append(normals, nms...)

					tcs := make([]float32, len(squareTcoords))
					for i := 0; i < len(squareTcoords); i += 2 {
						material := cell.Material
						// material := (cLat + cLon) % 7
						tcs[i+0] = (squareTcoords[i+0] + float32(material%4)) / 4
						tcs[i+1] = (squareTcoords[i+1] + float32(material/4)) / 4
					}
					tcoords = append(tcoords, tcs...)
				}
			}
		}
	}
	fillVBO(cr.pointsVBO, points)
	fillVBO(cr.normalsVBO, normals)
	fillVBO(cr.tcoordsVBO, tcoords)
	cr.numTriangles = int32(len(points) / 3)
	cr.geometryUpdated = true
}

func (cr *chunkRenderer) draw() {
	gl.BindVertexArray(cr.drawableVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, cr.numTriangles)
}

type overlay struct {
	program           uint32
	drawableVAO       uint32
	pointsVBO         uint32
	projectionUniform int32
}

func newOverlay() *overlay {
	over := overlay{}
	points := []float32{
		-20.0, 0.0, 0.0,
		19.0, 0.0, 0.0,

		0.0, -20.0, 0.0,
		0.0, 19.0, 0.0,
	}
	over.pointsVBO = newVBO()
	fillVBO(over.pointsVBO, points)
	over.drawableVAO = newPointsVAO(over.pointsVBO, 3)

	over.program = createProgram(vertexShaderSourceHUD, fragmentShaderSourceHUD)
	bindAttribute(over.program, 0, "vp")

	over.projectionUniform = uniformLocation(over.program, "proj")
	return &over
}

func (over *overlay) draw(w *glfw.Window) {
	gl.UseProgram(over.program)
	width, height := w.GetSize()
	proj := mgl32.Scale3D(1/float32(width), 1/float32(height), 1.0)
	gl.UniformMatrix4fv(over.projectionUniform, 1, false, &proj[0])

	gl.BindVertexArray(over.drawableVAO)
	gl.DrawArrays(gl.LINES, 0, 4)
}

type screenText struct {
	charInfo       map[string]charInfo
	statusLine     textLine
	charCount      int
	drawableVAO    uint32
	pointsVBO      uint32
	numTriangles   int32
	program        uint32
	texture        uint32
	textureUnit    int32
	textureUniform int32
}

type textLine struct {
	str  string
	x, y int
}

type charInfo struct {
	x, y, width, height, originX, originY, advance int
}

func newScreenText() *screenText {
	text := screenText{}

	text.charInfo = make(map[string]charInfo)
	text.statusLine.x = 1
	text.statusLine.y = 1

	text.program = createProgram(vertexShaderSourceText, fragmentShaderSourceText)
	bindAttribute(text.program, 0, "coord")

	text.textureUniform = uniformLocation(text.program, "texFont")

	text.pointsVBO = newVBO()
	text.drawableVAO = newPointsVAO(text.pointsVBO, 4)

	// Load up texture info
	var textMeta map[string]interface{}
	textMetaBytes, e := ioutil.ReadFile("font.json")
	if e != nil {
		panic(e)
	}
	json.Unmarshal(textMetaBytes, &textMeta)
	characters := textMeta["characters"].(map[string]interface{})
	for ch, props := range characters {
		propMap := props.(map[string]interface{})
		text.charInfo[ch] = charInfo{
			x:       int(propMap["x"].(float64)),
			y:       int(propMap["y"].(float64)),
			width:   int(propMap["width"].(float64)),
			height:  int(propMap["height"].(float64)),
			originX: int(propMap["originX"].(float64)),
			originY: int(propMap["originY"].(float64)),
			advance: int(propMap["advance"].(float64)),
		}
	}

	// Generated from https://evanw.github.io/font-texture-generator/
	// Inconsolata font (installed on system with Google Web Fonts), size 24
	// Power of 2, white with black stroke, thickness 2
	existingImageFile, err := os.Open("font.png")
	if err != nil {
		panic(err)
	}
	defer existingImageFile.Close()
	img, err := png.Decode(existingImageFile)
	if err != nil {
		panic(err)
	}
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Pt(0, 0), draw.Src)

	text.textureUnit = 0
	gl.ActiveTexture(uint32(gl.TEXTURE0 + text.textureUnit))
	gl.GenTextures(1, &text.texture)
	gl.BindTexture(gl.TEXTURE_2D, text.texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		// gl.SRGB_ALPHA,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix),
	)

	gl.GenerateMipmap(gl.TEXTURE_2D)
	return &text
}

func (text *screenText) computeGeometry(width, height int) {
	sx := 2.0 / float32(width) * 12
	sy := 2.0 / float32(height) * 12
	points := []float32{}
	text.charCount = 0
	line := text.statusLine
	for i, ch := range line.str + " " {
		aInfo := text.charInfo[string(ch)]
		ax1 := 1.0 / 512.0 * float32(aInfo.x-1)
		ax2 := 1.0 / 512.0 * float32(aInfo.x-1+aInfo.width)
		ay1 := 1.0 / 128.0 * float32(aInfo.y)
		ay2 := 1.0 / 128.0 * float32(aInfo.y+aInfo.height)
		x1 := -1 + sx*float32(line.x) + float32(i)*sx - sx*float32(aInfo.originX)/12
		x2 := -1 + sx*float32(line.x) + float32(i)*sx + sx*float32(aInfo.width)/12 - sx*float32(aInfo.originX)/12
		y1 := 1 - sy*float32(line.y) + -sy*float32(aInfo.originY)/12
		y2 := 1 - sy*float32(line.y) + sy*float32(aInfo.height)/12 - sy*float32(aInfo.originY)/12
		points = append(points, []float32{
			x1, -y1, ax1, ay1,
			x1, -y2, ax1, ay2,
			x2, -y2, ax2, ay2,

			x2, -y2, ax2, ay2,
			x2, -y1, ax2, ay1,
			x1, -y1, ax1, ay1,
		}...)
		text.charCount++
	}
	fillVBO(text.pointsVBO, points)
}

func (text *screenText) draw(player *person, w *glfw.Window) {
	r, theta, phi := mgl32.CartesianToSpherical(player.loc)
	text.statusLine.str = fmt.Sprintf("LAT %v, LON %v, ALT %v", int(theta/math.Pi*180-90+0.5), int(phi/math.Pi*180+0.5), int(r+0.5))
	text.computeGeometry(framebufferSize(w))

	gl.UseProgram(text.program)
	gl.Uniform1i(text.textureUniform, text.textureUnit)
	gl.BindVertexArray(text.drawableVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6*int32(text.charCount))
}

type hotbar struct {
	drawableVAO    uint32
	pointsVBO      uint32
	numPoints      int32
	program        uint32
	texture        uint32
	textureUnit    int32
	textureUniform int32
}

func newHotbar() *hotbar {
	h := hotbar{}

	h.program = createProgram(vertexShaderSourceHotbar, fragmentShaderSourceHotbar)
	bindAttribute(h.program, 0, "coord")

	h.textureUniform = uniformLocation(h.program, "tex")
	fmt.Print(uniformLocation(h.program, "tex"))
	h.pointsVBO = newVBO()
	h.drawableVAO = newPointsVAO(h.pointsVBO, 4)

	// existingImageFile, err := os.Open("textures.png")
	// if err != nil {
	// 	panic(err)
	// }
	// defer existingImageFile.Close()
	// img, err := png.Decode(existingImageFile)
	// if err != nil {
	// 	panic(err)
	// }
	// rgba := image.NewRGBA(img.Bounds())
	// draw.Draw(rgba, rgba.Bounds(), img, image.Pt(0, 0), draw.Src)
	rgba := LoadTextures()
	h.textureUnit = 3
	gl.ActiveTexture(uint32(gl.TEXTURE0 + h.textureUnit))
	gl.GenTextures(1, &h.texture)
	gl.BindTexture(gl.TEXTURE_2D, h.texture)

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
	return &h
}

func (h *hotbar) computeGeometry(player *person, width, height int) {
	aspect := float32(width) / float32(height)
	sq := []float32{-1, -1, -1, 1, 1, 1, 1, 1, 1, -1, -1, -1}
	points := []float32{}
	sz := float32(0.05)
	for m, mat := range player.hotBar {
		mx := float32(mat % 4)
		my := float32(mat / 4)
		px := 1.25 * 2 * sz * (float32(m+1) - float32(len(geom.Materials))/2)
		py := 1 - 0.1*aspect
		scale := sz
		if m == player.activeHotBarSlot {
			scale = 1.5 * sz
		}
		pts := make([]float32, 2*len(sq))
		for i := 0; i < len(sq); i += 2 {
			pts = append(pts, []float32{
				px + sq[i+0]*scale,
				py + sq[i+1]*scale*aspect,
				(mx + (sq[i+0]+1)/2) / 4,
				(my + (sq[i+1]+1)/2) / 4,
			}...)
		}
		points = append(points, pts...)
	}
	if player.inInventory {
		for m := 1; m < len(geom.Materials); m++ {
			mx := float32(m % 4)
			my := float32(m / 4)
			px := 1.25 * 2 * sz * (float32(m) - float32(len(geom.Materials))/2)
			py := 1 - 0.25*aspect
			scale := sz
			pts := make([]float32, 2*len(sq))
			for i := 0; i < len(sq); i += 2 {
				pts = append(pts, []float32{
					px + sq[i+0]*scale,
					py + sq[i+1]*scale*aspect,
					(mx + (sq[i+0]+1)/2) / 4,
					(my + (sq[i+1]+1)/2) / 4,
				}...)
			}
			points = append(points, pts...)
		}
	}
	h.numPoints = int32(len(points) / 4)
	fillVBO(h.pointsVBO, points)
}

func (h *hotbar) draw(player *person, w *glfw.Window) {
	if player.hotBarOn {
		width, height := framebufferSize(w)
		h.textureUniform = 0
		h.computeGeometry(player, width, height)
		gl.UseProgram(h.program)
		gl.Uniform1i(h.textureUniform, h.textureUnit)
		gl.BindVertexArray(h.drawableVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, h.numPoints)
	}
}

type focusRenderer struct {
	program           uint32
	drawableVAO       uint32
	pointsVBO         uint32
	projectionUniform int32
}

func newFocusRenderer() *focusRenderer {
	focusRen := focusRenderer{}
	focusRen.program = createProgram(vertexShaderSourceFocus, fragmentShaderSourceFocus)
	bindAttribute(focusRen.program, 0, "position")
	focusRen.projectionUniform = uniformLocation(focusRen.program, "proj")
	if focusRen.projectionUniform < 0 {
		panic("Could not find projection uniform")
	}
	focusRen.pointsVBO = newVBO()
	focusRen.drawableVAO = newPointsVAO(focusRen.pointsVBO, 3)
	return &focusRen
}

func (focusRen *focusRenderer) draw(player *person, planet *geom.Planet, w *glfw.Window) {
	gl.UseProgram(focusRen.program)
	lonCells := planet.LonCellsInChunkIndex(planet.CellIndexToChunkIndex(player.focusCellIndex))
	lonWidth := geom.ChunkSize / lonCells

	pts := make([]float32, len(box))
	for i := 0; i < len(box); i += 3 {
		ind := geom.CellLoc{
			Lon: float32(player.focusCellIndex.Lon/lonWidth*lonWidth) + float32(lonWidth-1)/2 + float32(lonWidth)*(box[i+0]*1.01),
			Lat: float32(player.focusCellIndex.Lat) + (box[i+1] * 1.01),
			Alt: float32(player.focusCellIndex.Alt) + (box[i+2] * 1.01),
		}
		pt := planet.CellLocToCartesian(ind)
		pts[i+0] = pt.X()
		pts[i+1] = pt.Y()
		pts[i+2] = pt.Z()
	}

	fillVBO(focusRen.pointsVBO, pts)

	lookDir := player.lookDir()
	view := mgl32.LookAtV(player.loc, player.loc.Add(lookDir), player.loc.Normalize())
	width, height := framebufferSize(w)
	perspective := mgl32.Perspective(45, float32(width)/float32(height), 0.01, 1000)
	proj := perspective.Mul4(view)
	gl.UniformMatrix4fv(focusRen.projectionUniform, 1, false, &proj[0])

	gl.BindVertexArray(focusRen.drawableVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(pts)/3))
}
