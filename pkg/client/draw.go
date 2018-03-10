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
	square = []float32{
		-0.5, 0.5, 0.5,
		-0.5, -0.5, 0.5,
		0.5, -0.5, 0.5,

		-0.5, 0.5, 0.5,
		0.5, -0.5, 0.5,
		0.5, 0.5, 0.5,

		-0.5, 0.5, -0.5,
		-0.5, -0.5, -0.5,
		0.5, -0.5, -0.5,

		-0.5, 0.5, -0.5,
		0.5, -0.5, -0.5,
		0.5, 0.5, -0.5,

		0.5, -0.5, 0.5,
		0.5, -0.5, -0.5,
		0.5, 0.5, -0.5,

		0.5, -0.5, 0.5,
		0.5, 0.5, -0.5,
		0.5, 0.5, 0.5,

		-0.5, -0.5, 0.5,
		-0.5, -0.5, -0.5,
		-0.5, 0.5, -0.5,

		-0.5, -0.5, 0.5,
		-0.5, 0.5, -0.5,
		-0.5, 0.5, 0.5,

		0.5, 0.5, -0.5,
		-0.5, 0.5, -0.5,
		-0.5, 0.5, 0.5,

		0.5, 0.5, -0.5,
		-0.5, 0.5, 0.5,
		0.5, 0.5, 0.5,

		0.5, -0.5, -0.5,
		-0.5, -0.5, -0.5,
		-0.5, -0.5, 0.5,

		0.5, -0.5, -0.5,
		-0.5, -0.5, 0.5,
		0.5, -0.5, 0.5,
	}
)

const (
	vertexShaderSource = `
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
			highp vec3 ambientLight = vec3(0.1, 0.2, 0.1);
			highp vec3 light1Color = vec3(0.5, 0.5, 0.4);
			highp vec3 light1Dir = normalize(vec3(0.85, 0.8, 0.75));
			highp float light1 = max(dot(n, light1Dir), 0.0);
			highp vec3 light2Color = vec3(0.1, 0.1, 0.2);
			highp vec3 light2Dir = normalize(vec3(-0.85, -0.8, -0.75));
			highp float light2 = max(dot(n, light2Dir), 0.0);
			light = ambientLight + (light1Color * light1) + (light2Color * light2);
		}
	`

	fragmentShaderSource = `
		#version 410
		in vec3 color;
		in vec3 light;
		out vec4 frag_color;
		void main() {
			frag_color = vec4(light, 1.0);
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
		uniform sampler2D tex;
		out vec4 frag_color;

		void main(void) {
			vec4 texel = texture(tex, texcoord);
			if (texel.a < 0.5) {
				discard;
		  }
			frag_color = texel;
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
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

func drawFrame(h float32, player *person, text *screenText, over *overlay, planetRen *planetRenderer, peopleRen *peopleRenderer, window *glfw.Window) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	planetRen.draw(player, window)
	peopleRen.draw(player, window)
	over.draw(window)
	text.draw(player, window)

	glfw.PollEvents()
	window.SwapBuffers()
}

type peopleRenderer struct {
	api               *API
	program           uint32
	drawable          uint32
	projectionUniform int32
}

func newPeopleRenderer(api *API) *peopleRenderer {
	peopleRen := peopleRenderer{}
	peopleRen.api = api
	peopleRen.program = createProgram(vertexShaderSourcePeople, fragmentShaderSourcePeople)
	bindAttribute(peopleRen.program, 0, "vp")
	bindAttribute(peopleRen.program, 1, "n")
	peopleRen.projectionUniform = uniformLocation(peopleRen.program, "proj")
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

		peopleRen.drawable = makeVao(square, nms)

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

		gl.BindVertexArray(peopleRen.drawable)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
	}
}

type planetRenderer struct {
	planet            *geom.Planet
	program           uint32
	projectionUniform int32
}

func newPlanetRenderer(planet *geom.Planet) *planetRenderer {
	pr := planetRenderer{}
	pr.planet = planet
	pr.program = createProgram(vertexShaderSource, fragmentShaderSource)
	bindAttribute(pr.program, 0, "vp")
	bindAttribute(pr.program, 1, "n")
	pr.projectionUniform = uniformLocation(pr.program, "proj")
	return &pr
}

func (planetRen *planetRenderer) draw(player *person, w *glfw.Window) {
	gl.UseProgram(planetRen.program)
	lookDir := player.lookDir()
	view := mgl32.LookAtV(player.loc, player.loc.Add(lookDir), player.loc.Normalize())
	width, height := framebufferSize(w)
	perspective := mgl32.Perspective(45, float32(width)/float32(height), 0.01, 1000)
	proj := perspective.Mul4(view)
	gl.UniformMatrix4fv(planetRen.projectionUniform, 1, false, &proj[0])

	for key, chunk := range planetRen.planet.Chunks {
		if !chunk.GraphicsInitialized {
			initChunkGraphics(chunk, planetRen.planet, key.Lon, key.Lat, key.Alt)
		}
		drawChunk(chunk)
	}
}

func initChunkGraphics(c *geom.Chunk, planet *geom.Planet, lonIndex, latIndex, altIndex int) {
	cs := geom.ChunkSize
	points := []float32{}
	normals := []float32{}

	for cLon := 0; cLon < cs; cLon++ {
		for cLat := 0; cLat < cs; cLat++ {
			for cAlt := 0; cAlt < cs; cAlt++ {
				if c.Cells[cLon][cLat][cAlt].Material != geom.Air {
					pts := make([]float32, len(square))
					for i := 0; i < len(square); i += 3 {
						l := geom.CellLoc{
							Lon: float32(cs*lonIndex+cLon) + square[i+0],
							Lat: float32(cs*latIndex+cLat) + square[i+1],
							Alt: float32(cs*altIndex+cAlt) + square[i+2],
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
				}
			}
		}
	}
	c.Drawable = makeVao(points, normals)
	c.GraphicsInitialized = true
}

func drawChunk(chunk *geom.Chunk) {
	gl.BindVertexArray(chunk.Drawable)
	cs := geom.ChunkSize
	gl.DrawArrays(gl.TRIANGLES, 0, int32(cs*cs*cs*len(square)/3))
}

type overlay struct {
	program           uint32
	drawable          uint32
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
	over.drawable = makePointsVao(points, 3)

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

	gl.BindVertexArray(over.drawable)
	gl.DrawArrays(gl.LINES, 0, 4)
}

type screenText struct {
	charInfo       map[string]charInfo
	statusLine     textLine
	charCount      int
	textDrawable   uint32
	program        uint32
	texture        uint32
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

	text.textureUniform = uniformLocation(text.program, "texture")

	// Load up texture info
	var textMeta map[string]interface{}
	textMetaBytes, err := ioutil.ReadFile("font.json")
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

	gl.ActiveTexture(gl.TEXTURE0)
	gl.GenTextures(1, &text.texture)
	gl.BindTexture(gl.TEXTURE_2D, text.texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.SRGB_ALPHA,
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
	text.textDrawable = makePointsVao(points, 4)
}

func (text *screenText) draw(player *person, w *glfw.Window) {
	r, theta, phi := mgl32.CartesianToSpherical(player.loc)
	text.statusLine.str = fmt.Sprintf("LAT %v, LON %v, ALT %v", int(theta/math.Pi*180-90+0.5), int(phi/math.Pi*180+0.5), int(r+0.5))
	text.computeGeometry(framebufferSize(w))

	gl.UseProgram(text.program)
	gl.Uniform1i(text.textureUniform, int32(text.texture))
	gl.BindVertexArray(text.textDrawable)
	gl.DrawArrays(gl.TRIANGLES, 0, 6*int32(text.charCount))
}
