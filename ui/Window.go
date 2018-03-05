package ui

import (
	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/kuso-kodo/kuso-NES/nes"
	"log"
	"runtime"
)

func Run(nes *nes.NES) {
	runtime.LockOSThread()
	err := glfw.Init()
	if err != nil {
		log.Panic("GLFW Init error: ", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	window, err := glfw.CreateWindow(256*4, 240*4, "KUSO-NES - "+nes.FileName, nil, nil)
	if err != nil {
		log.Panic("GLFW CreateWindow error: ", err)
	}

	window.MakeContextCurrent()
	err = gl.Init()
	if err != nil {
		log.Panic("OPenGL Init error: ", err)
	}

	gl.Enable(gl.TEXTURE_2D)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	for window.ShouldClose() == false {
		nes.FrameRun()
		gl.Clear(gl.COLOR_BUFFER_BIT)
		buf := nes.Buffer()
		size := buf.Rect.Size()
		gl.TexImage2D(
			gl.TEXTURE_2D, 0, gl.RGBA,
			int32(size.X), int32(size.Y),
			0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(buf.Pix))
	}
	gl.Begin(gl.QUADS)
	gl.TexCoord2f(0, 1)
	gl.Vertex3f(-1, -1, 1)
	gl.TexCoord2f(1, 1)
	gl.Vertex3f(1, -1, 1)
	gl.TexCoord2f(1, 0)
	gl.Vertex3f(1, 1, 1)
	gl.TexCoord2f(0, 0)
	gl.Vertex3f(-1, 1, 1)
	gl.End()
	window.SwapBuffers()
	glfw.PollEvents()
}
