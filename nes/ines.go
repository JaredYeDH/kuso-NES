package nes

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

const NESMagicMumber = 0x1a53454e //"NES\sub"

type NESFileHeader struct {
	MagicNumber		uint32	// NES Magic Number,must be 0x1a53454e
	PRGNum			byte	// PRG-ROM banks number
	CHRNum			byte	// CHR-ROM banks number
	Ctrl1			byte	// Control
	Ctrl2			byte	// Control too
	RAMNum			byte	// RAM number (8KB each)
	_				[7]byte // Empty bytes. Not used at this tume but MUST BE ALL ZEROS or games will not work.
}

/*
 * LoadNES function reads an iNES file from the given path and return a Cartidge
 * if success.
 */
func LoadNES(path string) (*Cartridge , error) {
	file,err := os.Open(path)

	if err != nil {
		return nil,err
	}

	defer file.Close()

	header := NESFileHeader{}

	// Read header

	if err := binary.Read(file,binary.LittleEndian,&header) ; err != nil {
		return nil, err
	}

	if header.MagicNumber != NESMagicMumber {
		return nil , errors.New("Magic Number is Wrong.Invilid iNES file.")
	}

	mapper1 := int(header.Ctrl1) >> 4
	mapper2 := int(header.Ctrl2) >> 4
	mapper := mapper1 | mapper2 << 4

	mirror1 := int(header.Ctrl1) & 1
	mirror2 := int(header.Ctrl1 >> 3) & 1
	mirror := mirror1 | mirror2 << 1

	battery := (header.Ctrl1 >> 1 & 1) == 1

	if header.Ctrl1 & 0x4 == 0x4 {
		trainer := make([]byte,512)

		if _, err := io.ReadFull(file, trainer); err != nil {
			return nil,err
		}
	}

	// PRG -- 16 KB each

	prg :=  make([]byte, int(header.PRGNum)*(16 * (1 << 10)))

	if _,err := io.ReadFull(file,prg); err != nil {
		return nil,err
	}

	// CHR -- 8 KB each

	chr := make([]byte,int(header.PRGNum)*(8 * (1 << 10)))

	if _,err := io.ReadFull(file,chr); err != nil {
		return nil,err
	}

	//Now every thing is OK, return thr cartridge

	cartridge := Cartridge{prg, chr, mapper, mirror, battery}
	return &cartridge, nil
}