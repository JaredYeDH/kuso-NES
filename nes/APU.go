package nes

// Audio Process Unit
// Ref: http://nesdev.com/apu_ref.txt

var (
	lengthTable = []byte{
		10, 254, 20, 2, 40, 4, 80, 6, 160, 8, 60, 10, 14, 12, 26, 14,
		12, 16, 24, 18, 48, 20, 96, 22, 192, 24, 72, 26, 16, 28, 32, 30,
	}

	dutyTable = [][]byte{
		{0, 1, 0, 0, 0, 0, 0, 0},
		{0, 1, 1, 0, 0, 0, 0, 0},
		{0, 1, 1, 1, 1, 0, 0, 0},
		{1, 0, 0, 1, 1, 1, 1, 1},
	}

	triangleTable = []byte{
		15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	}

	noiseTable = []uint16{
		4, 8, 16, 32, 64, 96, 128, 160, 202, 254, 380, 508, 762, 1016, 2034, 4068,
	}

	dmcTable = []byte{
		214, 190, 170, 160, 143, 127, 113, 107, 95, 80, 71, 64, 53, 42, 36, 27,
	}
)

type APU struct {
	nes        *NES
	channel    chan float32
	sampleRate float64
	square1    Square
	square2    Square
	triangle   Triangle
	noise      Noise
	dmc        DMC
	cycle      uint64
	fPeriod    byte
	fValue     byte
	fIRQ       bool
	fChain     FilterChain
}

type Square struct {
	enabled  bool
	channel  byte
	lEnabled bool
	lValue   byte
	tPeriod  uint16
	tValue   uint16
	dMode    byte
	dValue   byte
	sReload  bool
	sEnabled bool
	sNegate  bool
	sShift   byte
	sPeriod  byte
	sValue   byte
	eEnabled bool
	eLoop    bool
	eStart   bool
	ePeriod  byte
	eValue   byte
	eVolume  byte
	cVolume  byte
}

type Triangle struct {
	enabled  bool
	lEnabled bool
	lValue   byte
	tPeriod  uint16
	tValue   uint16
	dValue   byte
	cPeriod  byte
	cValue   byte
	cReload  bool
}

type Noise struct {
	enabled       bool
	mode          bool
	shiftRegister uint16
	lEnabled      bool
	lValue        byte
	tPeriod       uint16
	tValue        uint16
	eEnabled      bool
	eLoop         bool
	eStart        bool
	ePeriod       byte
	eValue        byte
	eVolume       byte
	cVolume       byte
}

type DMC struct {
	cpu           *CPU
	enabled       bool
	value         byte
	sAddress      uint16
	sLength       uint16
	cAddress      uint16
	cLength       uint16
	shiftRegister byte
	bCount        byte
	tPeriod       byte
	tValue        byte
	loop          bool
	irq           bool
}

type Filter interface {
	Run(x float32) float32
}

type FilterChain []Filter

func NewAPU(nes *NES) *APU {
	apu := APU{}
	apu.nes = nes
	apu.noise.shiftRegister = 1
	apu.square1.channel = 1
	apu.square2.channel = 2
	apu.dmc.cpu = nes.CPU
	return &apu
}
