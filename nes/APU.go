package nes

// Audio Process Unit
// Ref: http://nesdev.com/apu_ref.txt

const FrameCounter = CPUFrequency / 240

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

var (
	squareTable [32]float32
	tndTable    [203]float32
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

func init() {
	for i := 0; i < 31; i++ {
		squareTable[i] = 95.52 / (8128.0/float32(i) + 100.0)
	}
	for i := 0; i < 203; i++ {
		tndTable[i] = 163.67 / (24329.0/float32(i) + 100)
	}
}

// Square

func (s *Square) wCtrl(val byte) {
	s.dMode = (val >> 6) & 3
	s.lEnabled = (val>>5)&1 == 0
	s.eLoop = !s.lEnabled
	s.eEnabled = (val>>4)&1 == 0
	s.ePeriod = val & 15
	s.eVolume = val & 15
	s.eStart = true
}

func (s *Square) wSweep(val byte) {
	s.sEnabled = (val>>7)&1 == 0
	s.sPeriod = (val&4)&7 + 1
	s.sNegate = (val>>3)&1 == 1
	s.sShift = val & 7
	s.sReload = true
}

func (s *Square) wTimerLow(val byte) {
	s.tPeriod = (s.tPeriod & 0xFF00) | uint16(val)
}

func (s *Square) wTimerHigh(val byte) {
	s.lValue = lengthTable[val>>3]
	s.tPeriod = (s.tPeriod & 0xFF00) | uint16((val&7)<<8)
	s.eStart = true
	s.dValue = 0
}

func (s *Square) rTimer() {
	if s.tValue == 0 {
		s.tValue = s.tValue
		s.dValue = (s.dValue + 1) % 8
	} else {
		s.tValue--
	}
}

func (s *Square) rEnvelope() {
	if s.eStart {
		s.eVolume = 0
		s.eValue = s.ePeriod
		s.eStart = false
	} else if s.eValue > 0 {
		s.eValue--
	} else {
		if s.eVolume > 0 {
			s.eVolume--
		} else if s.eLoop {
			s.eVolume = 1<<4 - 1
		}
		s.eValue = s.ePeriod
	}
}

func (s *Square) rLength() {
	if s.lEnabled && s.lValue > 0 {
		s.lValue--
	}
}

func (s *Square) sweep() {
	a := s.tPeriod >> s.sShift
	if s.sNegate {
		s.tPeriod = s.tPeriod - a
		if s.channel == 1 {
			s.tPeriod--
		}
	} else {
		s.tPeriod = s.tPeriod + a
	}
}

func (s *Square) out() byte {
	if s.enabled == false || s.lValue == 0 || dutyTable[s.dMode][s.dValue] == 0 || s.tPeriod < 8 || s.tPeriod > 0x7FF {
		return 0
	}
	if s.eEnabled {
		return s.eVolume
	} else {
		return s.cVolume
	}
}

// Triangle

func (t *Triangle) wCtrl(val byte) {
	t.lEnabled = val>>7&1 == 0
}

func (t *Triangle) wTimerLow(val byte) {
	t.tPeriod = (t.tPeriod & 0xFF00) | uint16(val)
}

func (t *Triangle) wTimerHigh(val byte) {
	t.lValue = lengthTable[val>>3]
	t.tPeriod = (t.tPeriod & 0x00FF) | (uint16(val&7) << 8)
	t.tValue = t.tPeriod
	t.cReload = true
}

func (t *Triangle) rTimer() {
	if t.tValue == 0 {
		t.tValue = t.tPeriod
		if t.lValue > 0 && t.cValue > 0 {
			t.dValue = (t.dValue + 1) % 32
		}
	} else {
		t.tValue--
	}
}

func (t *Triangle) rLength() {
	if t.lEnabled && t.lValue > 0 {
		t.lValue--
	}
}

func (t *Triangle) rCounter() {
	if t.cReload {
		t.cValue = t.cPeriod
	} else if t.cValue > 0 {
		t.cValue--
	}
	if t.lEnabled {
		t.cReload = false
	}
}

func (t *Triangle) out() byte {
	if !t.enabled || t.lValue == 0 || t.cValue == 0 {
		return 0
	}
	return triangleTable[t.dValue]
}

// Noise

func (n *Noise) wCtrl(value byte) {
	n.lEnabled = (value>>5)&1 == 0
	n.eLoop = (value>>5)&1 == 1
	n.eEnabled = (value>>4)&1 == 0
	n.ePeriod = value & 15
	n.cVolume = value & 15
	n.eStart = true
}

func (n *Noise) wPeriod(value byte) {
	n.mode = value&0x80 == 0x80
	n.tPeriod = noiseTable[value&0x0F]
}

func (n *Noise) wLength(value byte) {
	n.lValue = lengthTable[value>>3]
	n.eStart = true
}

func (n *Noise) rTimer() {
	if n.tValue == 0 {
		n.tValue = n.tPeriod
		var shift byte
		if n.mode {
			shift = 6
		} else {
			shift = 1
		}
		b1 := n.shiftRegister & 1
		b2 := (n.shiftRegister >> shift) & 1
		n.shiftRegister = n.shiftRegister >> 1
		n.shiftRegister |= (b1 ^ b2) << 14
	} else {
		n.tValue--
	}
}

func (n *Noise) rEnvelope() {
	if n.eStart {
		n.eVolume = 15
		n.eValue = n.ePeriod
		n.eStart = false
	} else if n.eValue > 0 {
		n.eValue--
	} else {
		if n.eVolume > 0 {
			n.eVolume--
		} else if n.eLoop {
			n.eVolume = 15
		}
		n.eValue = n.ePeriod
	}
}

func (n *Noise) rLength() {
	if n.lEnabled && n.lValue > 0 {
		n.lValue--
	}
}

func (n *Noise) out() byte {
	if !n.enabled || n.lValue == 0 || n.shiftRegister&1 == 1 {
		return 0
	}
	if n.eEnabled {
		return n.eVolume
	} else {
		return n.cVolume
	}
}

// DMC

func (d *DMC) wCtrl(val byte) {
	d.irq = val&0x80 == 0x80
	d.loop = val&0x40 == 0x40
	d.tPeriod = dmcTable[val&0x0F]
}

func (d *DMC) wValue(val byte) {
	d.value = val & 0x7F
}

func (d *DMC) wAddress(value byte) {
	d.sAddress = 0xC000 | (uint16(value) << 6)
}

func (d *DMC) wLength(value byte) {
	d.sLength = (uint16(value) << 4) | 1
}


func (d *DMC) stepTimer() {
	if !d.enabled {
		return
	}
	d.rReader()
	if d.tValue == 0 {
		d.tValue = d.tPeriod
		d.rShifter()
	} else {
		d.tValue--
	}
}

func (d *DMC) rReader() {
	if d.cLength > 0 && d.bCount == 0 {
		d.cpu.stall += 4
		d.shiftRegister = d.cpu.Read(d.cAddress)
		d.bCount = 8
		d.cAddress++
		if d.cAddress == 0 {
			d.cAddress = 0x8000
		}
		d.cLength--
		if d.cLength == 0 && d.loop {
			d.cAddress = d.sAddress
			d.cLength = d.sLength
		}
	}
}

func (d *DMC) rShifter() {
	if d.bCount == 0 {
		return
	}
	if d.shiftRegister&1 == 1 {
		if d.value <= 125 {
			d.value += 2
		}
	} else {
		if d.value >= 2 {
			d.value -= 2
		}
	}
	d.shiftRegister = d.shiftRegister >> 1
	d.bCount--
}

func (d *DMC) out() byte {
	return d.value
}