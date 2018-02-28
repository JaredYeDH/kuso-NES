package nes

// Early support file for testing the CPU

type NES struct {
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
	nes := NES{nil, cartidge, nil, ram, nil, nil}

	nes.CPUMemory = NewCPUMemory(&nes)
	nes.PPUMemory = NewPPUMemory(&nes)
	nes.CPU = NewCPU(nes.CPUMemory)
	nes.PPU = NewPPU(&nes)
	return &nes, nil
}
