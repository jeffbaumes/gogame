package scene

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/jeffbaumes/buildorb/pkg/common"
)

// Health draws the player's hotbar
type Health struct {
	drawableVAO    uint32
	pointsVBO      uint32
	numPoints      int32
	program        uint32
	texture        uint32
	textureUnit    int32
	textureUniform int32
}

// NewHealth creates a new health bar
func NewHealth() *Health {
	const vertexShader = `
		#version 410

		in vec4 coord;
		out vec2 tcoord;

		void main(void) {
			gl_Position = vec4(coord.xy, 0, 1);
			tcoord = coord.zw;
		}
	`

	const fragmentShader = `
		#version 410

		in vec2 tcoord;
		uniform sampler2D tex;
		out vec4 frag_color;

		void main(void) {
			vec4 texel = texture(tex, tcoord);
			frag_color = texel;
		}
	`

	h := Health{}

	h.program = createProgram(vertexShader, fragmentShader)
	bindAttribute(h.program, 0, "coord")

	h.textureUniform = uniformLocation(h.program, "tex")
	h.pointsVBO = newVBO()
	h.drawableVAO = newPointsVAO(h.pointsVBO, 4)

	rgba, _ := LoadImages([]string{"textures/health.png", "textures/health-empty.png"}, 16)
	h.textureUnit = 4
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

func (h *Health) computeGeometry(player *common.Player, width, height int) {
	aspect := float32(width) / float32(height)
	sq := []float32{-1, -1, -1, 1, 1, 1, 1, 1, 1, -1, -1, -1}
	points := []float32{}
	sz := float32(0.02)
	for m := 0; m < common.MaxHealth; m++ {
		px := 1.25 * 2 * sz * (float32(m+1) - float32(common.MaxHealth)/2)
		py := -1 + 0.1*aspect
		scale := sz
		pts := make([]float32, 2*len(sq))
		tx := float32(0)
		if m >= player.Health {
			tx = 0.5
		}
		for i := 0; i < len(sq); i += 2 {
			pts = append(pts, []float32{
				px + sq[i+0]*scale,
				py + sq[i+1]*scale*aspect,
				tx + 0.5*(sq[i+0]+1)/2,
				0.5 - 0.5*(sq[i+1]+1)/2,
			}...)
		}
		points = append(points, pts...)
	}
	h.numPoints = int32(len(points) / 4)
	fillVBO(h.pointsVBO, points)
}

// Draw draws the hotbar
func (h *Health) Draw(player *common.Player, w *glfw.Window) {
	if player.HotbarOn {
		width, height := FramebufferSize(w)
		h.textureUniform = 0
		h.computeGeometry(player, width, height)
		gl.UseProgram(h.program)
		gl.Uniform1i(h.textureUniform, h.textureUnit)
		gl.BindVertexArray(h.drawableVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, h.numPoints)
	}
}
