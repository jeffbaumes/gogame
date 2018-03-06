package client

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
)

func createProgram(vertexSource, fragmentSource string) uint32 {
	vertexShader, err := compileShader(vertexSource+"\x00", gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	fragmentShader, err := compileShader(fragmentSource+"\x00", gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)
	return program
}

func makePointsVao(points []float32, size int32) uint32 {
	var vbo = make([]uint32, 2)
	gl.GenBuffers(1, (*uint32)(gl.Ptr(vbo)))
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo[0])
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo[0])
	gl.VertexAttribPointer(0, size, gl.FLOAT, false, 0, nil)

	return vao
}

func makeVao(points []float32, normals []float32) uint32 {
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
func uniformLocation(program uint32, name string) int32 {
	glstr, free := gl.Strs(name)
	uniform := gl.GetUniformLocation(program, *glstr)
	free()
	return uniform
}
