package nes

import "log"

type Memory interface {
	Read(address uint16) byte
	Write(address uint16, value byte)
	Read16(address uint16) uint16
}

// CPU
type CPUMemory struct {
	nes *NES
}

func NewCPUMemory(nes *NES) Memory {
	return &CPUMemory{nes}
}

func (mem *CPUMemory) Read(address uint16) byte {
	switch {
	case address < 0x2000:
		return mem.nes.RAM[address%0x0800]
	case address < 0x4000:
		return mem.nes.PPU.ReadRegister(0x2000 + address%8)
	case address == 0x4014:
		return mem.nes.PPU.ReadRegister(address)
	case address == 0x4015:
		return mem.nes.APU.ReadRegister(address)
	case address == 0x4016:
		mem.nes.Controller1.Read()
	case address == 0x4017:
		mem.nes.Controller2.Read()
	case address >= 0x6000:
		return mem.nes.Mapper.Read(address)
	default:
		log.Fatalf("Illegal CPU memory read at address: $%04X", address)
	}
	return 0
}

func (mem *CPUMemory) Write(address uint16, val byte) {
	switch {
	case address < 0x2000:
		mem.nes.RAM[address%0x0800] = val
	case address < 0x4000:
		mem.nes.PPU.WriteRegister(0x2000+address%8, val)
	case address < 0x4014:
		mem.nes.APU.WriteRegister(address, val)
	case address == 0x4014:
		mem.nes.PPU.WriteRegister(address, val)
	case address == 0x4015:
		mem.nes.APU.WriteRegister(address, val)
	case address == 0x4016:
		mem.nes.Controller1.Write(val)
	case address == 0x4017:
		mem.nes.Controller2.Write(val)
	case address < 0x4020:
		return
	case address >= 0x6000:
		mem.nes.Mapper.Write(address, val)
		return
	default:
		log.Fatalf("Illegal CPU memory write at address: $%04X", address)
	}
}

func (mem *CPUMemory) Read16(address uint16) uint16 {
	l := uint16(mem.Read(address))
	h := uint16(mem.Read(address + 1))

	return h<<8 | l
}

// PPU

type PPUMemory struct {
	nes *NES
}

func NewPPUMemory(nes *NES) Memory {
	return &PPUMemory{nes}
}

func (mem *PPUMemory) Read(address uint16) byte {
	switch {
	case address < 0x2000:
		return mem.nes.Cartridge.Read(address)
	case address < 0x3F00:
		mode := mem.nes.Cartridge.Mirror
		return mem.nes.PPU.nameTableData[MirrorAddress(mode, address)%2048]
	case address < 0x4000:
		return mem.nes.PPU.palette[address%32]
	default:
		log.Fatalf("PPUMemory: Unknown read at address: 0x%04X", address)
	}
	return 0
}

func (mem *PPUMemory) Write(address uint16, val byte) {
	address %= 0x4000
	switch {
	case address < 0x2000:
		mem.nes.Cartridge.Write(address, val)
		return
	case address < 0x3F00:
		mode := mem.nes.Cartridge.Mirror
		mem.nes.PPU.nameTableData[MirrorAddress(mode, address)%2048] = val
		return
	case address < 0x4000:
		mem.nes.PPU.wPalette(address%32, val)
		return
	default:
		log.Fatalf("PPUMemory: Unknown write at address: 0x%04X", address)
	}
}

func (mem *PPUMemory) Read16(address uint16) uint16 {
	l := uint16(mem.Read(address))
	h := uint16(mem.Read(address + 1))

	return h<<8 | l
}
