package nes

import (
	"image"
)

// Early support file for testing the CPU

type NES struct {
	FileName  string
	CPU       *CPU
	Cartridge *Cartridge
	PPU       *PPU
	RAM       []byte
	CPUMemory Memory
	PPUMemory Memory
}

func NewNES(path string) (*NES, error) {
	cartidge, err := LoadNES(path)

	if err != nil {
		return nil, err
	}

	ram := make([]byte, 2048)
	nes := NES{path, nil, cartidge, nil, ram, nil, nil}

	nes.CPUMemory = NewCPUMemory(&nes)
	nes.PPUMemory = NewPPUMemory(&nes)
	nes.CPU = NewCPU(nes.CPUMemory)
	nes.PPU = NewPPU(&nes)
	return &nes, nil
}

func (n *NES) Reset() {
	n.CPU.Reset()
}

func (nes *NES) Run() int {
	cpuCycles := nes.CPU.Run()
	for i := 0; i < cpuCycles*3; i++ {
		nes.PPU.Run()
	}
	return cpuCycles
}

func (n *NES) FrameRun() {
	frame := n.PPU.Frame
	for frame == n.PPU.Frame {
		n.Run()
	}
}

func (n *NES) Buffer() *image.RGBA {
	return n.PPU.back
}
