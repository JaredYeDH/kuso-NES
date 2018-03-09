package ui

import (
	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/kuso-kodo/kuso-NES/nes"
	"image"
	"log"
	"runtime"
)

// Controllers
const (
	BA      = iota // J
	BB             // K
	BSelect        // F
	BStart         // H
	BUp            // W
	BDown          // S
	BLeft          // A
	BRight         // D
)

// TODO: Change Keys Dynamically

func getKeys(window *glfw.Window, nes *nes.NES) {
	nes.SetKeyPressed(1, BA, window.GetKey(glfw.KeyZ) == glfw.Press)
	nes.SetKeyPressed(1, BB, window.GetKey(glfw.KeyX) == glfw.Press)
	nes.SetKeyPressed(1, BSelect, window.GetKey(glfw.KeyRightShift) == glfw.Press)
	nes.SetKeyPressed(1, BStart, window.GetKey(glfw.KeyEnter) == glfw.Press)
	nes.SetKeyPressed(1, BUp, window.GetKey(glfw.KeyUp) == glfw.Press)
	nes.SetKeyPressed(1, BDown, window.GetKey(glfw.KeyDown) == glfw.Press)
	nes.SetKeyPressed(1, BLeft, window.GetKey(glfw.KeyLeft) == glfw.Press)
	nes.SetKeyPressed(1, BRight, window.GetKey(glfw.KeyRight) == glfw.Press)
}

func Run(nes *nes.NES) {
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

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
		getKeys(window, nes)
		nes.FrameRun()
		setTexture(texture, nes.Buffer())
		// render frame
		gl.Clear(gl.COLOR_BUFFER_BIT)
		draw(window)
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

// Textures

func setTexture(texture uint32, im *image.RGBA) {
	size := im.Rect.Size()
	gl.TexImage2D(
		gl.TEXTURE_2D, 0, gl.RGBA,
		int32(size.X), int32(size.Y),
		0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(im.Pix))
}

// Draw

func draw(win *glfw.Window) {
	w, h := win.GetFramebufferSize()
	aspect := float32(w) / float32(h)
	var x, y, size float32
	size = 0.95
	if aspect >= 1 {
		x = size / aspect
		y = size
	} else {
		x = size
		y = size * aspect
	}
	gl.Begin(gl.QUADS)
	gl.TexCoord2f(0, 1)
	gl.Vertex3f(-x, -y, 1)
	gl.TexCoord2f(1, 1)
	gl.Vertex3f(x, -y, 1)
	gl.TexCoord2f(1, 0)
	gl.Vertex3f(x, y, 1)
	gl.TexCoord2f(0, 0)
	gl.Vertex3f(-x, y, 1)
	gl.End()
}
