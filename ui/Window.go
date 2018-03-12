package ui

import (
	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/gordonklaus/portaudio"
	"github.com/kuso-kodo/kuso-NES/nes"
	"image"
	"log"
	"runtime"
)

// Window
const (
	Width   = 256
	Height  = 240
	Scale   = 4
	Padding = 0
)

func init() {
	runtime.GOMAXPROCS(2)
	runtime.LockOSThread()
}

// TODO: Change Keys Dynamically

func readKey(window *glfw.Window, key glfw.Key) bool {
	return window.GetKey(key) == glfw.Press
}

func getKeys(window *glfw.Window, n *nes.NES) {
	n.SetKeyPressed(1, nes.BA, readKey(window, glfw.KeyK))
	n.SetKeyPressed(1, nes.BB, readKey(window, glfw.KeyJ))
	n.SetKeyPressed(1, nes.BSelect, readKey(window, glfw.KeyF))
	n.SetKeyPressed(1, nes.BStart, readKey(window, glfw.KeyH))
	n.SetKeyPressed(1, nes.BUp, readKey(window, glfw.KeyW))
	n.SetKeyPressed(1, nes.BDown, readKey(window, glfw.KeyS))
	n.SetKeyPressed(1, nes.BLeft, readKey(window, glfw.KeyA))
	n.SetKeyPressed(1, nes.BRight, readKey(window, glfw.KeyD))
}

func Run(nes *nes.NES) {
	portaudio.Initialize()
	defer portaudio.Terminate()
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

	t1 := glfw.GetTime()

	audio := NewAudio()
	nes.SetAPUChannel(audio.channel)
	if audio.Start() != nil {
		log.Panic(err)
	}
	nes.SetAPUSRate(audio.sampleRate)
	defer audio.Stop()

	for window.ShouldClose() == false {
		now := glfw.GetTime()
		d := now - t1
		t1 = now
		getKeys(window, nes)
		nes.RunSeconds(d)
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
	gl.BindTexture(gl.TEXTURE_2D, texture)
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
