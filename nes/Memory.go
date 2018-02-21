package nes

// Early implemention of 6502 CPU Memory to support the development of CPU

type Memory interface {
	Read(address uint16) byte
	Write(address uint16, value byte)
	Read16(address uint16) uint16
}

type CPUMemory struct {
	NES *NES
	RAM []byte
}

func NewCPUMemory(nes *NES) Memory {
	ram := make([]byte,2048)
	return &CPUMemory{nes,ram}
}

func (mem *CPUMemory)Read(address uint16) byte {
	if address < 0x2000 {
		return mem.RAM[address % 0x0800]
	} else if address >= 0x8000 {
		return mem.NES.Cartridge.Read(address)
	} else {
		return 0 // Failed
	}
}

func (mem *CPUMemory) Write(address uint16, value byte) {
	switch {
	case address < 0x2000:
		mem.RAM[address%0x0800] = value
	}
}

func (mem *CPUMemory) Read16(address uint16) uint16 {
	l := uint16(mem.Read(address))
	h := uint16(mem.Read(address))

	return (h << 8 | l)
}