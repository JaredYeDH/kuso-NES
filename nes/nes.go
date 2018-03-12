package nes

import (
	"image"
	"log"
)

// Early support file for testing the CPU

type NES struct {
	FileName    string
	APU         *APU
	Cartridge   *Cartridge
	Controller1 *Controller
	Controller2 *Controller
	CPU         *CPU
	PPU         *PPU
	RAM         []byte
	Mapper      Mapper
	CPUMemory   Memory
	PPUMemory   Memory
}

func NewNES(path string) (*NES, error) {
	cartidge, err := LoadNES(path)

	if err != nil {
		return nil, err
	}

	ram := make([]byte, 2048)
	Controller1 := NewController()
	Controller2 := NewController()
	nes := NES{path, nil, cartidge, Controller1, Controller2, nil, nil, ram, nil, nil, nil}
	mapper, err := NewMapper(&nes)
	if err != nil {
		return nil, err
	}
	nes.Mapper = mapper
	nes.APU = NewAPU(&nes)
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
		nes.Mapper.Run()
	}
	for i := 0; i < cpuCycles; i++ {
		nes.APU.Run()
	}
	return cpuCycles
}

func (n *NES) RunSeconds(second float64) {
	cycles := int(CPUFrequency * second)
	for cycles > 0 {
		cycles -= n.Run()
	}
}

func (n *NES) Buffer() *image.RGBA {
	return n.PPU.front
}

func (n *NES) SetKeyPressed(controller, btn int, press bool) {
	switch controller {
	case 1:
		n.Controller1.SetPressed(btn, press)
	case 2:
		n.Controller2.SetPressed(btn, press)
	}
}

func (n *NES) SetAPUChannel(channel chan float32) {
	n.APU.channel = channel
}

func (n *NES) SetAPUSRate(sRate float64) {
	log.Print(sRate)
	if sRate != 0 {
		n.APU.sampleRate = CPUFrequency / sRate
		n.APU.fChain = FilterChain{
			HPassFilter(float32(sRate), 90),
			HPassFilter(float32(sRate), 440),
			LPassFilter(float32(sRate), 14000),
		}
	} else {
		n.APU.fChain = nil
	}
}
