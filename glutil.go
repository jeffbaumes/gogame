package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const (
	width  = 500
	height = 500

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
	` + "\x00"

	fragmentShaderSource = `
		#version 410
		in vec3 color;
		in vec3 light;
		out vec4 frag_color;
		void main() {
			frag_color = vec4(light, 1.0);
		}
	` + "\x00"
)

// InitGlfw initializes GLFW and returns the window
func InitGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "World Blocks", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

// InitOpenGL initializes the OpenGL context and returns the program
func InitOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	gl.Enable(gl.DEPTH_TEST)

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	bindAttribute(prog, 0, "vp")
	bindAttribute(prog, 1, "n")
	gl.LinkProgram(prog)

	return prog
}

// MakeVao creates two buffers for the points and normals and returns a VAO for both
func MakeVao(points []float32, normals []float32) uint32 {
	var vbo = make([]uint32, 2)
	gl.GenBuffers(2, (*uint32)(gl.Ptr(vbo)))
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo[0])
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo[1])
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(normals), gl.Ptr(normals), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo[0])
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)
	gl.EnableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo[1])
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, nil)

	return vao
}

func bindAttribute(prog, location uint32, name string) {
	s, free := gl.Strs(name)
	gl.BindAttribLocation(prog, location, *s)
	free()
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

// UniformLocation retrieves a uniform location by name
func UniformLocation(program uint32, name string) int32 {
	glstr, free := gl.Strs(name)
	uniform := gl.GetUniformLocation(program, *glstr)
	free()
	return uniform
}
