package nes

// Early support file for testing the CPU

type NES struct {
	CPU       *CPU
	Cartridge *Cartridge
}

func NewNES(path string) (*NES, error) {
	cartidge, err := LoadNES(path)

	if err != nil {
		return nil, err
	}

	nes := NES{nil, cartidge}

	nes.CPU = NewCPU(NewCPUMemory(&nes))

	return &nes, nil
}
