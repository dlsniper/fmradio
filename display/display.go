package display

import (
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
)

const (
	// command signals that we want to send a command to the screen
	command = 0x04

	// data signals that we want to send a command to the screen
	data = 0x05

	// address is our default address
	address = 0x27
)

// SunFounderLCD1602Driver controls the LCD 1602 from SunFounder
//
//goland:noinspection GoUnnecessarilyExportedIdentifiers
type SunFounderLCD1602Driver struct {
	name         string
	i2cConnector i2c.Connector
	i2c.Config
	gobot.Commander

	i2cAddr int
	conn    i2c.Connection

	backlightEnabled bool
}

// Name of our device
func (lcd *SunFounderLCD1602Driver) Name() string {
	return lcd.name
}

// SetName set the name of our device
func (lcd *SunFounderLCD1602Driver) SetName(name string) {
	lcd.name = name
}

// Start the device work
func (lcd *SunFounderLCD1602Driver) Start() error {
	bus := lcd.GetBusOrDefault(lcd.i2cConnector.GetDefaultBus())

	var err error
	lcd.conn, err = lcd.i2cConnector.GetConnection(lcd.i2cAddr, bus)
	if err != nil {
		return err
	}

	commands := []byte{0x33, 0x32, 0x28, 0x0C}
	for _, cmd := range commands {
		if err = lcd.sendCommand(cmd); err != nil {
			return err
		}
		time.Sleep(5 * time.Millisecond)
	}

	return lcd.ClearScreen()
}

// Halt stops the device in a graceful way
func (lcd *SunFounderLCD1602Driver) Halt() error {
	lcd.backlightEnabled = false
	return lcd.ClearScreen()
}

// Connection retrieves the i2c connection to the device
func (lcd *SunFounderLCD1602Driver) Connection() gobot.Connection {
	return lcd.i2cConnector.(gobot.Connection)
}

// Send a command to the LCD
func (lcd *SunFounderLCD1602Driver) sendCommand(cmd byte) (err error) {
	return lcd.communicate(command, cmd)
}

// Send data to the LCD
func (lcd *SunFounderLCD1602Driver) sendData(cmd byte) (err error) {
	return lcd.communicate(data, cmd)
}

// write handles the actual data writing to the LCD i2c connection
func (lcd *SunFounderLCD1602Driver) write(data byte) error {
	temp := data
	if lcd.backlightEnabled {
		temp |= 0x08
	} else {
		temp |= 0x07
	}

	return lcd.conn.WriteByte(temp)
}

// Communicate with the LCD by sending either a command or data
func (lcd *SunFounderLCD1602Driver) communicate(cmdType byte, cmd byte) error {
	// Send bit7-4 firstly
	buf := cmd & 0xF0
	buf |= cmdType // RS = 0, RW = 0, EN = 1
	if err := lcd.write(buf); err != nil {
		return err
	}

	time.Sleep(2 * time.Millisecond)

	buf &= 0xFB // Make EN = 0
	if err := lcd.write(buf); err != nil {
		return err
	}

	// Send bit3-0 secondly
	buf = (cmd & 0x0F) << 4
	buf |= cmdType // RS = 0, RW = 0, EN = 1
	if err := lcd.write(buf); err != nil {
		return err
	}

	time.Sleep(2 * time.Millisecond)
	buf &= 0xFB // Make EN = 0
	return lcd.write(buf)
}

// EnableBacklight turns on the screen backlight
func (lcd *SunFounderLCD1602Driver) EnableBacklight() error {
	err := lcd.write(0x08)
	time.Sleep(2 * time.Millisecond)
	return err
}

// DisableBacklight turns off the screen backlight
func (lcd *SunFounderLCD1602Driver) DisableBacklight() error {
	err := lcd.write(0x07)
	time.Sleep(2 * time.Millisecond)
	return err
}

// ClearScreen removes any message from the LCD screen
func (lcd *SunFounderLCD1602Driver) ClearScreen() error {
	// The screen clearing commands needs to be
	// sent with the backlight turned on
	tmp := lcd.backlightEnabled
	lcd.backlightEnabled = true
	if err := lcd.sendCommand(0x01); err != nil {
		return err
	}

	time.Sleep(2 * time.Millisecond)

	lcd.backlightEnabled = tmp

	if lcd.backlightEnabled {
		return lcd.EnableBacklight()
	}
	return lcd.DisableBacklight()
}

// DisplayMessageWithCoordinates renders our message on the display
func (lcd *SunFounderLCD1602Driver) DisplayMessageWithCoordinates(x, y int, msg string) error {
	if x < 0 {
		x = 0
	}

	if x > 15 {
		x = 15
	}

	if y < 0 {
		y = 0
	}

	if y > 1 {
		y = 1
	}

	// Move cursor
	addr := byte(0x80 + 0x40*y + x)
	if err := lcd.sendCommand(addr); err != nil {
		return err
	}

	for _, ch := range msg {
		if err := lcd.sendData(byte(ch)); err != nil {
			return err
		}
	}
	return nil
}

// DisplayMessage renders our message on the display
func (lcd *SunFounderLCD1602Driver) DisplayMessage(msg string) error {
	// Pad the message
	if len(msg) < 32 {
		iLen := 32 - len(msg)
		for i := 0; i < iLen; i++ {
			msg += " "
		}
	}

	addr := byte(0x80)
	if err := lcd.sendCommand(addr); err != nil {
		return err
	}

	for _, ch := range msg[:16] {
		if err := lcd.sendData(byte(ch)); err != nil {
			return err
		}
	}

	addr = byte(0x80 + 0x40)
	if err := lcd.sendCommand(addr); err != nil {
		return err
	}

	for _, ch := range msg[16:32] {
		if err := lcd.sendData(byte(ch)); err != nil {
			return err
		}
	}

	return nil
}

// NewLCD1602Driver creates a new GoBot driver for our FM transmitter
func NewLCD1602Driver(connector i2c.Connector, options ...func(i2c.Config)) (*SunFounderLCD1602Driver, error) {
	lcd := &SunFounderLCD1602Driver{
		name:             gobot.DefaultName("SunFounderLCD1602Driver"),
		i2cConnector:     connector,
		Config:           i2c.NewConfig(),
		i2cAddr:          address,
		backlightEnabled: true,
	}

	for _, option := range options {
		option(lcd)
	}

	return lcd, nil
}
