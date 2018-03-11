package ui

import (
	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/kuso-kodo/kuso-NES/nes"
	"image"
	"log"
	"runtime"
	"time"
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

// Window
const (
	Width   = 256
	Height  = 240
	Scale   = 4
	Padding = 0
)

func init() {
	runtime.LockOSThread()
}

// TODO: Change Keys Dynamically

func getKeys(window *glfw.Window, nes *nes.NES) {
	nes.SetKeyPressed(1, BA, window.GetKey(glfw.KeyJ) == glfw.Press)
	nes.SetKeyPressed(1, BB, window.GetKey(glfw.KeyK) == glfw.Press)
	nes.SetKeyPressed(1, BSelect, window.GetKey(glfw.KeyF) == glfw.Press)
	nes.SetKeyPressed(1, BStart, window.GetKey(glfw.KeyH) == glfw.Press)
	nes.SetKeyPressed(1, BUp, window.GetKey(glfw.KeyW) == glfw.Press)
	nes.SetKeyPressed(1, BDown, window.GetKey(glfw.KeyS) == glfw.Press)
	nes.SetKeyPressed(1, BLeft, window.GetKey(glfw.KeyA) == glfw.Press)
	nes.SetKeyPressed(1, BRight, window.GetKey(glfw.KeyD) == glfw.Press)
}

func Run(nes *nes.NES) {
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	window, err := glfw.CreateWindow(Width*Scale, Height*Scale, "KUSO-NES - "+nes.FileName, nil, nil)
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


	var Frame uint64
	for window.ShouldClose() == false {
		getKeys(window, nes)
		nes.FrameRun()
		setTexture(texture, nes.Buffer())
		// render frame
		time.Sleep(time.Millisecond*8)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		draw(window)
		Frame ++
		if Frame % 60 == 0{
			log.Print("60 Frame generated.")
		}
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
	s1 := float32(w) / float32(Width)
	s2 := float32(h) / float32(Height)
	f := float32(1 - Padding)
	var x, y float32
	if s1 < s2 {
		x = f
		y = f * s1 / s2
	} else {
		x = f * s2 / s1
		y = f
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
