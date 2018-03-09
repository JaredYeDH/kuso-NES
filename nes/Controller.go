package nes

type Controller struct {
	button [8]byte
	index   byte
	strobe  byte
}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) Press(button int) {
	c.button[button] = 1
}

func (c *Controller) Release(button int) {
	c.button[button] = 0
}

func (c *Controller) SetPressed(button int, pressed bool) {
	if pressed {
		c.button[button] = 1
	} else {
		c.button[button] = 0
	}
}

func (c *Controller) Read() byte {
	var val byte
	if c.index < 8 {
		val = c.button[c.index]
	} else {
		val = 0
	}
	c.index++
	if c.strobe&1 == 1 {
		c.index = 0
	}
	return val
}

func (c *Controller) Write(val byte) {
	c.strobe = val
	if c.strobe&1 == 1 {
		c.index = 0
	}
}