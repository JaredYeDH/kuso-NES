package nes

import (
	"log"
	"strings"
	"testing"
)

func (c *CPU) ReadString(address uint16) string {
	var s []byte
	for {
		char := c.Read(address)

		if char == 0 {
			break
		}

		s = append(s, char)
		address++
	}

	return string(s)
}

// File from : http://blargg.8bitalley.com/nes-tests/instr_test-v5.zip

func TestOfficialInstructions(t *testing.T) {

	nes, err := NewNES("official_only.nes")

	if err != nil {
		t.Fatal(err)
	}

	cpu := nes.CPU
	cpu.Write(0x6000, 0xFF)
	for {
		for i := 0; i < 0x1DEAD; i++ {
			cpu.Run()
		}
		if cpu.Read(0x6000) < 0x80 {
			break
		}
		msg := cpu.ReadString(0x6004)
		if len(msg) > 0 {
			log.Print(strings.TrimSpace(msg))
		}
	}
}
