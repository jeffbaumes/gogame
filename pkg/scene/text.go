package scene

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"math"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jeffbaumes/gogame/pkg/common"
)

const (
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
)

// Text draws overlay text
type Text struct {
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

// NewText creates a new Text object
func NewText() *Text {
	text := Text{}

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
	textMetaBytes, e := ioutil.ReadFile("textures/font/font.json")
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
	existingImageFile, err := os.Open("textures/font/font.png")
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

func (text *Text) computeGeometry(width, height int, where float32) {
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
		y1 := where - sy*float32(line.y) + -sy*float32(aInfo.originY)/12
		y2 := where - sy*float32(line.y) + sy*float32(aInfo.height)/12 - sy*float32(aInfo.originY)/12
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

// Draw draws the overlay text
func (text *Text) Draw(player *common.Player, w *glfw.Window) {

	wi, h := FramebufferSize(w)
	if player.Intext == true {
		text.statusLine.str = player.Text
		text.computeGeometry(wi, h, 0.8)
		gl.UseProgram(text.program)
		gl.Uniform1i(text.textureUniform, text.textureUnit)
		gl.BindVertexArray(text.drawableVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 6*int32(text.charCount))
	}

	text.statusLine.str = player.DrawText
	text.computeGeometry(wi, h, 0.7)
	gl.UseProgram(text.program)
	gl.Uniform1i(text.textureUniform, text.textureUnit)
	gl.BindVertexArray(text.drawableVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6*int32(text.charCount))

	r, theta, phi := mgl32.CartesianToSpherical(player.Loc)
	text.statusLine.str = fmt.Sprintf("LAT %v, LON %v, ALT %v", int(theta/math.Pi*180-90+0.5), int(phi/math.Pi*180+0.5), int(r+0.5))
	text.computeGeometry(wi, h, 1)
	gl.UseProgram(text.program)
	gl.Uniform1i(text.textureUniform, text.textureUnit)
	gl.BindVertexArray(text.drawableVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6*int32(text.charCount))
}
