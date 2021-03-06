// Package radio implements the driver for Adafruit's packaging of
// the Si4713 FM transmitter. You can learn more about it here:
// https://www.adafruit.com/product/1958
//
// The main implementation is under the Si4713Driver and it requires
// some additional configuration via Si4713Config structure.
//
// The original drivers were written in C and Python and can be found
// at the following addresses:
//     - Python: https://github.com/adafruit/Adafruit_CircuitPython_SI4713 (MIT License)
//     - C: https://github.com/adafruit/Adafruit-Si4713-Library (BSD License)
//
// To read about the specifications of the transmitter, read the following documents:
// https://www.silabs.com/documents/public/data-sheets/Si4712-13-B30.pdf
// https://www.silabs.com/documents/public/application-notes/AN332.pdf
package radio

import (
	"fmt"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/drivers/i2c"
)

const (
	low  = 0x0
	high = 0x1
)

// Misc constants.
//
//goland:noinspection GoUnusedConst,GoUnnecessarilyExportedIdentifiers,GoSnakeCaseUsage
const (
	// Address is the device default address if SEN is high.
	Address = 0x63

	// AlternativeAddress if SEN is low.
	AlternativeAddress = 0x11

	// DEFAULT_RDS_PROGRAM_ID holds some random default for the RDS program ID
	DEFAULT_RDS_PROGRAM_ID = 0xADAF
)

// Different command identifiers that the transmitter supports.
//
//goland:noinspection GoUnusedConst,GoUnnecessarilyExportedIdentifiers,GoSnakeCaseUsage
const (
	// STATUS_CTS is the command to read the device status.
	STATUS_CTS = 0x80

	// CMD_POWER_UP commands the device power up and mode selection.
	// Modes include FM transmit and analog/digital audio interface configuration.
	CMD_POWER_UP = 0x01

	// CMD_GET_REV command returns revision information on the device.
	CMD_GET_REV = 0x10

	// CMD_POWER_DOWN commands the device to power down.
	CMD_POWER_DOWN = 0x11

	// CMD_SET_PROPERTY sets the value of a property.
	CMD_SET_PROPERTY = 0x12

	// CMD_GET_PROPERTY retrieves a property's value.
	CMD_GET_PROPERTY = 0x13

	// CMD_GET_INT_STATUS read interrupt status bits.
	CMD_GET_INT_STATUS = 0x14

	// CMD_TX_TUNE_FREQ tunes to the given transmit frequency.
	CMD_TX_TUNE_FREQ = 0x30

	// CMD_TX_TUNE_POWER sets the output power level and tunes the antenna capacitor.
	CMD_TX_TUNE_POWER = 0x31

	// CMD_TX_TUNE_MEASURE measures the received noise level at the specified frequency.
	CMD_TX_TUNE_MEASURE = 0x32

	// CMD_TX_TUNE_STATUS queries the status of a previously sent
	// TX Tune Freq, TX Tune Power, or TX Tune Measure command.
	CMD_TX_TUNE_STATUS = 0x33

	// CMD_TX_ASQ_STATUS queries the TX status and input audio signal metrics.
	CMD_TX_ASQ_STATUS = 0x34

	// CMD_TX_RDS_BUFF queries the status of the RDS Group Buffer
	// and loads new data into buffer.
	CMD_TX_RDS_BUFF = 0x35

	// CMD_TX_RDS_PS sets up default PS strings.
	CMD_TX_RDS_PS = 0x36

	// CMD_GPO_CTL configures GPO3 as output or Hi-Z.
	CMD_GPO_CTL = 0x80

	// CMD_GPO_SET sets GPO3 output level (low or high).
	CMD_GPO_SET = 0x81
)

// This section holds all the constants that mark the various properties that
// our transmitter has.
//
//goland:noinspection GoUnusedConst,GoUnnecessarilyExportedIdentifiers,GoSnakeCaseUsage
const (
	// PROP_GPO_IEN enables interrupt sources.
	PROP_GPO_IEN = 0x0001

	// PROP_DIGITAL_INPUT_FORMAT configures the digital input format.
	PROP_DIGITAL_INPUT_FORMAT = 0x0101

	// PROP_DIGITAL_INPUT_SAMPLE_RATE configures the digital input
	// sample rate in 10 Hz steps.
	// Default is 0 Hz.
	PROP_DIGITAL_INPUT_SAMPLE_RATE = 0x0103

	// PROP_REFCLK_FREQ sets frequency of the reference clock in Hz.
	// The range is 31130 to 34406 Hz, or 0 to disable the AFC.
	// Default is 32768 Hz.
	PROP_REFCLK_FREQ = 0x0201

	// PROP_REFCLK_PRESCALE sets the prescaler value for the reference clock.
	PROP_REFCLK_PRESCALE = 0x0202

	// PROP_TX_COMPONENT_ENABLE enables transmit multiplex signal components.
	// Default has pilot and L-R enabled.
	PROP_TX_COMPONENT_ENABLE = 0x2100

	// PROP_TX_AUDIO_DEVIATION configures the audio frequency deviation level.
	// Units are in 10 Hz increments.
	// Default is 6285 (68.25 kHz).
	PROP_TX_AUDIO_DEVIATION = 0x2101

	// PROP_TX_PILOT_DEVIATION configures the pilot tone frequency deviation level.
	// Units are in 10 Hz increments.
	// Default is 675 (6.75 kHz)
	PROP_TX_PILOT_DEVIATION = 0x2102

	// PROP_TX_RDS_DEVIATION configures the RDS/RBDS frequency deviation level.
	// Units are in 10 Hz increments.
	// Default is 2 kHz.
	PROP_TX_RDS_DEVIATION = 0x2103

	// PROP_TX_LINE_LEVEL_INPUT_LEVEL configures the maximum analog line input
	// level to the LIN/RIN pins to reach the maximum deviation level
	// programmed into the audio deviation property TX Audio Deviation.
	// Default is 636 mVPK.
	PROP_TX_LINE_LEVEL_INPUT_LEVEL = 0x2104

	// PROP_TX_LINE_INPUT_MUTE sets the line input mute.
	// L and R inputs may be independently muted.
	// Default is not muted.
	PROP_TX_LINE_INPUT_MUTE = 0x2105

	// PROP_TX_PREEMPHASIS configures the pre-emphasis time constant.
	// Default is 0 (75 μS).
	PROP_TX_PREEMPHASIS = 0x2106

	// PROP_TX_PILOT_FREQUENCY configures the frequency of the stereo pilot.
	// Default is 19000 Hz.
	PROP_TX_PILOT_FREQUENCY = 0x2107

	// PROP_TX_ACOMP_ENABLE enables the audio dynamic range control.
	// Default is 0 (disabled).
	PROP_TX_ACOMP_ENABLE = 0x2200

	// PROP_TX_ACOMP_THRESHOLD sets the threshold level for audio dynamic range control.
	// Default is –40 dB.
	PROP_TX_ACOMP_THRESHOLD = 0x2201

	// PROP_TX_ATTACK_TIME sets the attack time for audio dynamic range control.
	// Default is 0 (0.5 ms).
	PROP_TX_ATTACK_TIME = 0x2202

	// PROP_TX_RELEASE_TIME sets the release time for audio dynamic range control.
	// Default is 4 (1000 ms).
	PROP_TX_RELEASE_TIME = 0x2203

	// PROP_TX_ACOMP_GAIN sets the gain for audio dynamic range control.
	// Default is 15 dB.
	PROP_TX_ACOMP_GAIN = 0x2204

	// PROP_TX_LIMITER_RELEASE_TIME sets the limiter release time.
	// Default is 102 (5.01 ms)
	PROP_TX_LIMITER_RELEASE_TIME = 0x2205

	// PROP_TX_ASQ_INTERRUPT_SOURCE configures measurements related to signal quality metrics.
	// Default is none selected.
	PROP_TX_ASQ_INTERRUPT_SOURCE = 0x2300

	// PROP_TX_ASQ_LEVEL_LOW configures low audio input level detection threshold.
	// This threshold can be used to detect silence on the incoming audio.
	PROP_TX_ASQ_LEVEL_LOW = 0x2301

	// PROP_TX_ASQ_DURATION_LOW configures the duration which the input audio level must be below
	// the low threshold in order to detect a low audio condition.
	PROP_TX_ASQ_DURATION_LOW = 0x2302

	// PROP_TX_AQS_LEVEL_HIGH configures the high audio input level detection threshold.
	// This threshold can be used to detect activity on the incoming audio.
	PROP_TX_AQS_LEVEL_HIGH = 0x2303

	// PROP_TX_AQS_DURATION_HIGH configures the duration which the input audio level must be above
	// the high threshold to detect a high audio condition.
	PROP_TX_AQS_DURATION_HIGH = 0x2304

	// PROP_TX_RDS_INTERRUPT_SOURCE configures the RDS interrupt sources.
	// Default is none selected.
	PROP_TX_RDS_INTERRUPT_SOURCE = 0x2C00

	// PROP_TX_RDS_PI sets the transmit RDS program identifier.
	PROP_TX_RDS_PI = 0x2C01

	// PROP_TX_RDS_PS_MIX configures the mix of RDS PS Group with RDS Group Buffer.
	PROP_TX_RDS_PS_MIX = 0x2C02

	// PROP_TX_RDS_PS_MISC sets miscellaneous bits to transmit along with RDS_PS Groups.
	PROP_TX_RDS_PS_MISC = 0x2C03

	// PROP_TX_RDS_PS_REPEAT_COUNT sets number of times to repeat transmission
	// of a PS message before transmitting the next PS message.
	PROP_TX_RDS_PS_REPEAT_COUNT = 0x2C04

	// PROP_TX_RDS_MESSAGE_COUNT gets the number of PS messages in use.
	PROP_TX_RDS_MESSAGE_COUNT = 0x2C05

	// PROP_TX_RDS_PS_AF sets the RDS Program Service Alternate Frequency.
	// This provides the ability to inform the receiver of a single
	// alternate frequency using AF Method A coding and is transmitted
	// along with the RDS_PS Groups.
	PROP_TX_RDS_PS_AF = 0x2C06

	// PROP_TX_RDS_FIFO_SIZE sets the number of blocks reserved for the FIFO.
	// Note that the value written must be one larger than the desired FIFO size.
	PROP_TX_RDS_FIFO_SIZE = 0x2C07
)

// Define the format for the command to send to the transmitter
type command []uint8

// The list of the different commands.
func cmdPowerUp() command {
	return command{
		CMD_POWER_UP,
		0x12,
		// CTS interrupt disabled
		// GPO2 output disabled
		// Boot normally
		// Cristal oscillator Enabled
		// FM transmit
		0x50, // analog input mode
	}
}

func cmdPowerDown() command {
	return command{
		CMD_POWER_DOWN,
		0,
	}
}

func cmdGetRev() command {
	return command{
		CMD_GET_REV,
		0,
	}
}

func cmdTuneFM(h, l uint8) command {
	return command{
		CMD_TX_TUNE_FREQ,
		0,
		h,
		l,
	}
}

func cmdReadTuneStatus() command {
	return command{
		CMD_TX_TUNE_STATUS,
		0x1, // INTACK
	}
}

func cmdTuneMeasure(h, l uint8) command {
	return command{
		CMD_TX_TUNE_MEASURE,
		0,
		h,
		l,
		0,
	}
}

func cmdSetTxPower(pwr, antCap uint8) command {
	return command{
		CMD_TX_TUNE_POWER,
		0,
		0,
		pwr,
		antCap,
	}
}

func cmdSetProperty() command {
	return command{
		CMD_SET_PROPERTY,
		0,
		0,
		0,
		0,
		0,
	}
}

func cmdSetRDSStationName(slotName uint8, n1, n2, n3, n4 byte) command {
	return command{
		CMD_TX_RDS_PS,
		slotName,
		n1,
		n2,
		n3,
		n4,
	}
}

func cmdSetRDSMessage(messageType, msgType, msgP, slot uint8, n1 byte, n2 byte, n3 byte, n4 byte) command {
	return command{
		messageType,
		msgType,
		msgP,
		slot,
		n1,
		n2,
		n3,
		n4,
	}
}

func cmdSetGPIOCtrl(pin uint8) command {
	return command{
		CMD_GPO_CTL,
		pin,
	}
}

func cmdSetGPIO(pin uint8) command {
	return command{
		CMD_GPO_SET,
		pin,
	}
}

func cmdASQStatus() command {
	return command{
		CMD_TX_ASQ_STATUS,
		0x1,
	}
}

// Si4713Config holds the additional configuration needed for Si4713Driver.
type Si4713Config struct {
	// DebugMode allows for greater details to be available during debugging
	DebugMode bool

	// DebugLog allows for debugging message handling
	DebugLog func(format string, v ...interface{})

	// Log provides access to any log data produced by the device
	Log func(format string, v ...interface{})

	// AlternateFrequency specifies transmission frequency.
	// Must be between 8750 and 10800.
	// Value * 10 = value in MHz
	AlternateFrequency uint16

	// HasRDS enables the RDS support
	HasRDS bool

	// RDSProgramID specifies the ID of our station for RDS transmission
	RDSProgramID uint16

	// RDSMessage is the message sent out via RDS
	RDSMessage string

	// ResetPin marks the pin used for resetting the device. Default is 29
	ResetPin string

	// RDSStationName is the name of the station that shows up in RDS information
	RDSStationName string

	// StopAfterFrequencyScan enables us exit after a quick frequency scan.
	// Must be combined with WithFrequencyScan flag.
	StopAfterFrequencyScan bool

	// TransmitFrequency is our main transmission frequency.
	// Must be between 8750 and 10800.
	// Value * 10 = value in MHz
	TransmitFrequency uint16

	// TransmitPower is our transmission power.
	// Must be between 88-115, value is in dBuV
	TransmitPower uint8

	// WithFrequencyScan enables scanning of frequencies before transmission.
	// Can be used with StopAfterFrequencyScan.
	WithFrequencyScan bool
}

// Si4713Driver holds the implementation to talk to the
// Adafruit Si 4713 FM Radio Transmitter breakout.
//
//goland:noinspection GoUnnecessarilyExportedIdentifiers
type Si4713Driver struct {
	name         string
	i2cAddr      int
	conn         i2c.Connection
	i2cConnector i2c.Connector
	i2c.Config

	Si4713Config
}

// Name of our device.
func (s *Si4713Driver) Name() string {
	return s.name
}

// SetName set the name of our device.
func (s *Si4713Driver) SetName(name string) {
	s.name = name
}

// Start the device work.
func (s *Si4713Driver) Start() error {
	// Run validation again, just in case the driver was not created
	// via the New function
	if err := s.Validate(); err != nil {
		return err
	}

	bus := s.GetBusOrDefault(s.i2cConnector.GetDefaultBus())

	if conn, err := s.i2cConnector.GetConnection(s.i2cAddr, bus); err != nil {
		return err
	} else {
		s.conn = conn
	}

	if begun, err := s.begin(); err != nil {
		return err
	} else if !begun { // begin with address 0x63 (CS high default)
		return fmt.Errorf("couldn't find radio")
	}

	if s.WithFrequencyScan {
		if err := s.scanFrequencies(); err != nil {
			return err
		}
	}

	if s.StopAfterFrequencyScan {
		return fmt.Errorf("forced stop due to configuration option")
	}

	if s.WithFrequencyScan {
		if err := s.scanTransmitFrequency(); err != nil {
			return err
		}
	}

	if s.DebugMode {
		s.DebugLog("Set TX power %d\n", s.TransmitPower)
	}
	if err := s.setTxPower(s.TransmitPower, 0); err != nil {
		return err
	}

	if s.DebugMode {
		s.DebugLog("Tuning into %.2f\n", float32(s.TransmitFrequency)/100)
	}
	if err := s.tuneFM(s.TransmitFrequency); err != nil {
		return err
	}

	// This will tell you the status in case you want to read it from the chip
	if currFreq, currdBuV, currAntCap, currNoiseLevel, err := s.readTuneStatus(); err != nil {
		return err
	} else if s.DebugMode {
		s.DebugLog("Curr freq: %.2f\n", float32(currFreq)/100)
		s.DebugLog("Curr freq dBuV: %d\n", currdBuV)
		s.DebugLog("Curr ANT cap: %d\n", currAntCap)
		s.DebugLog("Curr noise level: %d\n", currNoiseLevel)
	}

	if s.HasRDS {
		if err := s.EnableRDS(); err != nil {
			return err
		}
	}

	// set GP1 and GP2 to output
	return s.setGPIOCtrl(1<<1 | 1<<2)
}

// Halt stops the device in a graceful way.
func (s *Si4713Driver) Halt() error {
	return s.powerDown()
}

// Connection retrieves the i2c connection to the device.
func (s *Si4713Driver) Connection() gobot.Connection {
	return s.i2cConnector.(gobot.Connection)
}

// EnableRDS will configure then turn on the RDS/RDBS transmission.
func (s *Si4713Driver) EnableRDS() error {
	if err := s.beginRDS(s.RDSProgramID); err != nil {
		return err
	}
	if err := s.SetRDSStation(s.RDSStationName); err != nil {
		return err
	}
	if err := s.SetRDSMessage(s.RDSMessage); err != nil {
		return err
	}

	if s.DebugMode {
		s.DebugLog("RDS on!\n")
	}

	return nil
}

// Scan transmission power of entire range from 87.5 to 108.0 MHz.
func (s *Si4713Driver) scanFrequencies() error {
	for f := uint16(7600); f < 10800; f += 10 {
		if err := s.readTuneMeasure(f); err != nil {
			return err
		}

		_, _, _, currNoiseLevel, err := s.readTuneStatus()
		if err != nil {
			return err
		}
		if s.DebugMode {
			s.DebugLog("Noise level on %.2f MHz is %d\n", float32(f)/100, currNoiseLevel)
		}
	}
	return nil
}

// Scan the power of existing transmissions over our transmission frequency.
func (s *Si4713Driver) scanTransmitFrequency() error {
	if err := s.readTuneMeasure(s.TransmitFrequency); err != nil {
		return err
	}

	_, _, _, currNoiseLevel, err := s.readTuneStatus()
	if err != nil {
		return err
	}
	if s.DebugMode {
		s.DebugLog("Noise level on %.2f MHz is %d\n", float32(s.TransmitFrequency)/100, currNoiseLevel)
	}
	return nil
}

// SetGPIO controls the GPIO pins on the device
// You can toggle both off by sending 1<<0, or both.
//
//goland:noinspection GoUnnecessarilyExportedIdentifiers
func (s *Si4713Driver) SetGPIO(pin uint8) error {
	return s.sendCommand(cmdSetGPIO(pin))
}

// readASQ performs a status read for the TxAsqStatus.
func (s *Si4713Driver) readASQ() (status, currASQ, currInLevel byte, err error) {
	if err = s.sendCommand(cmdASQStatus()); err != nil {
		return 0, 0, 0, err
	}

	values, err := s.buffRead(5)
	if err != nil {
		return 0, 0, 0, err
	}

	status = values[0]
	currASQ = values[1]

	// discard
	_, _ = values[2], values[3]

	currInLevel = values[4]

	return status, currASQ, currInLevel, nil
}

// Queries the status of a previously sent TX Tune Freq, TX Tune
// Power, or TX Tune Measure using CMD_TX_TUNE_STATUS command.
func (s *Si4713Driver) readTuneStatus() (currFreq uint16, currdBuV, currAntCap, currNoiseLevel uint8, err error) {
	if err = s.sendCommand(cmdReadTuneStatus()); err != nil {
		return 0, 0, 0, 0, err
	}

	// status and resp1
	if _, err = s.conn.ReadByte(); err != nil {
		return 0, 0, 0, 0, err
	}
	if _, err = s.conn.ReadByte(); err != nil {
		return 0, 0, 0, 0, err
	}

	val, err := s.conn.ReadByte()
	if err != nil {
		return 0, 0, 0, 0, err
	}
	currFreq = uint16(val) << 8
	val, err = s.conn.ReadByte()
	if err != nil {
		return 0, 0, 0, 0, err
	}
	currFreq |= uint16(val) // resp3

	// resp4
	if _, err = s.conn.ReadByte(); err != nil {
		return 0, 0, 0, 0, err
	}

	currdBuV, err = s.conn.ReadByte()
	if err != nil {
		return 0, 0, 0, 0, err
	}

	currAntCap, err = s.conn.ReadByte()
	if err != nil {
		return 0, 0, 0, 0, err
	}

	currNoiseLevel, err = s.conn.ReadByte()
	return currFreq, currdBuV, currAntCap, currNoiseLevel, err
}

// SetRDSStation sets up the RDS station string.
//
//goland:noinspection GoUnnecessarilyExportedIdentifiers
func (s *Si4713Driver) SetRDSStation(stationName string) error {
	j := len(stationName) / 4
	name := []byte(stationName)
	// pad the name so that we can add nulls at the end of the command, if needed
	for i := len(stationName) - j*4; i > 0 && i < 4; i++ {
		name = append(name, ' ')
	}

	slots := uint8((len(stationName) + 3) / 4)
	j = 0
	for i := uint8(0); i < slots; i++ {
		// set slot number, then the message
		c := cmdSetRDSStationName(i, name[j], name[j+1], name[j+2], name[j+3])
		if err := s.sendCommand(c); err != nil {
			return err
		}
		j += 4
	}

	return nil
}

// SetRDSMessage queries the status of the RDS Group Buffer and loads new data into buffer.
func (s *Si4713Driver) SetRDSMessage(message string) error {
	j := len(message) / 4
	msg := []byte(message)
	// pad the name so that we can add nulls at the end of the command, if needed
	for i := len(message) - j*4; i > 0 && i < 4; i++ {
		msg = append(msg, ' ')
	}

	slots := uint8((len(message) + 3) / 4)
	j = 0
	for i := uint8(0); i < slots; i++ {
		msgType := uint8(0x04)
		if i == 0 {
			msgType = 0x06
		}

		c := cmdSetRDSMessage(CMD_TX_RDS_BUFF, msgType, 0x20, i, msg[j], msg[j+1], msg[j+2], msg[j+3])
		j += 4

		if err := s.sendCommand(c); err != nil {
			return err
		}
	}

	if err := s.setRDSTime(); err != nil {
		return err
	}

	if s.DebugMode {
		s.DebugLog("Enabling the RDS subsystem...\n")
	}

	// stereo, pilot+rds
	return s.setProperty(PROP_TX_COMPONENT_ENABLE, 0x0007)
}

// Configures GP1 / GP2 as output or Hi-Z.
func (s *Si4713Driver) setGPIOCtrl(pin uint8) error {
	return s.sendCommand(cmdSetGPIOCtrl(pin))
}

// Resets the registers to default settings and puts chip in.
func (s *Si4713Driver) reset() (err error) {
	dw, ok := s.i2cConnector.(gpio.DigitalWriter)
	if !ok {
		return fmt.Errorf("i2c connector does not have a digital writter capability")
	}

	if err = dw.DigitalWrite(s.ResetPin, high); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)

	if err = dw.DigitalWrite(s.ResetPin, low); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)

	return dw.DigitalWrite(s.ResetPin, high)
}

// Sends power up command to the breakout, then CTS and GPO2 output
// is disabled and then enable cristal oscillator.
// Also, it sets properties:
//            PROP_REFCLK_FREQ: 32.768
//            PROP_TX_PREEMPHASIS: 74uS pre-emphasis (USA standard)
//            PROP_TX_ACOMP_GAIN: max gain
//            PROP_TX_ACOMP_ENABLE: turned on limiter and AGC
//
func (s *Si4713Driver) powerUp() error {
	if err := s.sendCommand(cmdPowerUp()); err != nil {
		return err
	}

	// Crystal is 32.768
	if err := s.setProperty(PROP_REFCLK_FREQ, 32768); err != nil {
		return err
	}

	// 74uS pre-emphasis (USA std)
	if err := s.setProperty(PROP_TX_PREEMPHASIS, 0); err != nil {
		return err
	}

	// max gain?
	if err := s.setProperty(PROP_TX_ACOMP_ENABLE, 0x02); err != nil {
		return err
	}

	// turn on the limiter, but no dynamic ranging
	if err := s.setProperty(PROP_TX_ACOMP_GAIN, 10); err != nil {
		return err
	}

	// turn on the limiter and AGC
	return s.setProperty(PROP_TX_ACOMP_ENABLE, 0x02)
}

// Turn off the device.
func (s *Si4713Driver) powerDown() error {
	return s.sendCommand(cmdPowerDown())
}

// Setups the i2cConnector and calls powerUp function.
// Returns true if initialization was successful, otherwise false.
func (s *Si4713Driver) begin() (bool, error) {
	if err := s.reset(); err != nil {
		return false, err
	}
	if err := s.powerUp(); err != nil {
		return false, err
	}

	// check for Si4713Driver
	status, err := s.getRev()
	return status == 13, err
}

// Get the hardware revision code from the device using CMD_GET_REV.
func (s *Si4713Driver) getRev() (uint8, error) {
	if err := s.sendCommand(cmdGetRev()); err != nil {
		return 0, err
	}

	values, err := s.buffRead(9)
	if err != nil {
		return 0, err
	}

	partNumber := values[1]

	fw := uint16(values[2])
	fw <<= 8
	fw |= uint16(values[3])

	patch := uint16(values[4])
	patch <<= 8
	patch |= uint16(values[5])

	cmp := uint16(values[6])
	cmp <<= 8
	cmp |= uint16(values[7])

	chipRev := values[8]

	if s.DebugMode {
		s.DebugLog("Part # Si47%d-%x", partNumber, fw)
		s.DebugLog("Firmware %x\n", fw)
		s.DebugLog("Patch %x\n", patch)
		s.DebugLog("Chip rev %d\n", chipRev)
	}

	return partNumber, nil
}

// Tunes to given transmit frequency.
func (s *Si4713Driver) tuneFM(freqKHz uint16) error {
	h := uint8(freqKHz >> 8)
	l := uint8(freqKHz & 0xFF)
	if err := s.sendCommand(cmdTuneFM(h, l)); err != nil {
		return err
	}

	for {
		status, err := s.getStatus()
		if err != nil {
			return nil
		}
		if status&0x81 == 0x81 {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
}

//  Read interrupt status bits.
func (s *Si4713Driver) getStatus() (uint8, error) {
	if err := s.conn.WriteByte(CMD_GET_INT_STATUS); err != nil {
		return 0, err
	}

	return s.conn.ReadByte()
}

// Get the device status.
func (s *Si4713Driver) deviceStatus() (err error) {
	values, err := s.buffRead(6)
	if err != nil {
		return err
	}

	// values[0] discarded
	s.DebugLog("Circular avail: %d used: %d\n", values[2], values[3])
	s.DebugLog("FIFO avail: %d used: %d overflow: %d\n", values[4], values[5], values[1])
	return nil
}

// Measure the received noise level at the specified frequency.
func (s *Si4713Driver) readTuneMeasure(freq uint16) error {
	// check freq is multiple of 50khz
	if freq%5 != 0 {
		freq -= freq % 5
	}
	if s.DebugMode {
		s.DebugLog("Measuring frequency: %.2f MHz\n", float32(freq)/100)
	}

	h := uint8(freq >> 8)
	l := uint8(freq & 0xFF)
	if err := s.sendCommand(cmdTuneMeasure(h, l)); err != nil {
		return err
	}

	for {
		status, err := s.getStatus()
		if err != nil {
			return err
		}
		if status == 0x81 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

// Sets the output power level and tunes the antenna capacitor.
func (s *Si4713Driver) setTxPower(pwr, antCap uint8) error {
	return s.sendCommand(cmdSetTxPower(pwr, antCap))
}

// Set chip property over I2C.
func (s *Si4713Driver) setProperty(property uint16, value uint16) error {
	if s.DebugMode {
		s.DebugLog("Set Prop 0x%x = 0x%x (%d)\n", property, value, value)
	}

	p := cmdSetProperty()
	p[2] = uint8(property >> 8)
	p[3] = uint8(property & 0xFF)
	p[4] = uint8(value >> 8)
	p[5] = uint8(value & 0xFF)

	return s.sendCommand(p)
}

//  Begin RDS
//
//  Sets properties as follows:
//  	PROP_TX_AUDIO_DEVIATION: 66.25KHz,
//  	PROP_TX_RDS_DEVIATION: 2KHz,
//  	PROP_TX_RDS_INTERRUPT_SOURCE: 1,
//  	PROP_TX_RDS_PS_MIX: 50% mix (default value),
//  	PROP_TX_RDS_PS_MISC: 6152,
//  	PROP_TX_RDS_PS_REPEAT_COUNT: 3,
//  	PROP_TX_RDS_MESSAGE_COUNT: 1,
//  	PROP_TX_RDS_PS_AF: 57568,
//  	PROP_TX_RDS_FIFO_SIZE: 0,
//  	PROP_TX_COMPONENT_ENABLE: 7
func (s *Si4713Driver) beginRDS(programID uint16) error {
	// 66.25KHz (default is 68.25)
	if err := s.setProperty(PROP_TX_AUDIO_DEVIATION, 6625); err != nil {
		return err
	}

	// 2KHz (default)
	if err := s.setProperty(PROP_TX_RDS_DEVIATION, 200); err != nil {
		return err
	}

	// RDS IRQ
	if err := s.setProperty(PROP_TX_RDS_INTERRUPT_SOURCE, 0x0001); err != nil {
		return err
	}
	// program identifier
	if err := s.setProperty(PROP_TX_RDS_PI, programID); err != nil {
		return err
	}
	// 50% mix (default)
	if err := s.setProperty(PROP_TX_RDS_PS_MIX, 0x03); err != nil {
		return err
	}
	// RDSD0 & RDSMS (default)
	if err := s.setProperty(PROP_TX_RDS_PS_MISC, 0x1808); err != nil {
		return err
	}
	// 3 repeats (default)
	if err := s.setProperty(PROP_TX_RDS_PS_REPEAT_COUNT, 3); err != nil {
		return err
	}

	if err := s.setProperty(PROP_TX_RDS_MESSAGE_COUNT, 1); err != nil {
		return err
	}

	if err := s.setProperty(PROP_TX_RDS_PS_AF, s.AlternateFrequency); err != nil {
		return err
	}
	if err := s.setProperty(PROP_TX_RDS_FIFO_SIZE, 0); err != nil {
		return err
	}

	return s.setProperty(PROP_TX_COMPONENT_ENABLE, 0x0007)
}

// Send command to the radio chip.
func (s *Si4713Driver) sendCommand(cmd command) (err error) {
	if s.DebugMode {
		s.DebugLog("*** Command: %s\n", s.sliceToString(cmd))
	}
	if _, err = s.conn.Write(cmd); err != nil {
		return err
	}

	if cmd[0] == CMD_POWER_DOWN {
		return nil
	}

	// Wait for status CTS bit
	status := byte(0)
	for status&STATUS_CTS == 0 {
		status, err = s.conn.ReadByte()
		if err != nil {
			return err
		}
		if s.DebugMode {
			s.DebugLog("status: %x (%d)\n", status, status)
		}
	}

	return nil
}

func (s *Si4713Driver) setRDSTime() error {
	return s.sendCommand(cmdSetRDSMessage(CMD_TX_RDS_BUFF, 0x84, 0x40, 01, 0xA7, 0x0B, 0x2D, 0x6C))
}

// Loop performs the main application loop to transmit data and check the device status.
func (s *Si4713Driver) Loop() error {
	if !s.DebugMode {
		return nil
	}

	status, currASQ, currInLevel, err := s.readASQ()
	if err != nil {
		return err
	}

	s.DebugLog("Curr Status: 0x%x ASQ: 0x%x InLevel: %d dBfs\n", status, currASQ, int8(currInLevel))

	// toggle GPO1 and GPO2
	if err = s.SetGPIO(1 << 1); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)

	if err = s.SetGPIO(1 << 2); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)

	return s.deviceStatus()
}

func (s *Si4713Driver) buffRead(size int) ([]byte, error) {
	values := make([]byte, size)
	nValues, err := s.conn.Read(values)
	if err != nil {
		return nil, err
	}

	if nValues != size {
		return nil, fmt.Errorf("failed to read %d bytes from the line, read %d -> %s", size, len(values), s.sliceToString(values))
	}

	if s.DebugMode {
		s.DebugLog("read %d bytes: %s", size, s.sliceToString(values))
	}
	return values, nil
}

func (s *Si4713Driver) sliceToString(val []byte) string {
	res := ""
	for idx := range val {
		res += fmt.Sprintf("[%d]=0x%x(%d) ", idx, val[idx], val[idx])
	}
	return res
}

// Validate ensures that our Si4713Driver configuration is valid.
//noinspection GoUnnecessarilyExportedIdentifiers
func (c *Si4713Config) Validate() error {
	if c.Log == nil {
		panic("logging function cannot be nil. Use something like log.Printf or an empty function instead")
	}
	if c.DebugMode && c.DebugLog == nil {
		panic("cannot use debugging mode without configuring a DebugLog function, e.g. log.Printf")
	}

	if c.ResetPin == "" {
		c.ResetPin = "29"
	}

	if c.TransmitFrequency == 0 {
		return fmt.Errorf("FM transmission frequency not set")
	}

	if c.TransmitFrequency < 8750 || c.TransmitFrequency > 10800 {
		return fmt.Errorf("FM transmission frequency not in 87.50 MHz ... 108 MHz bounds")
	}

	if c.AlternateFrequency < 8750 || c.AlternateFrequency > 10800 {
		c.Log("FM alternate transmission frequency not in 87.50 MHz ... 108 MHz bounds, defaulting to %d\n", 8750)
		c.AlternateFrequency = 8750
	}

	// dBuV, 88-115 max
	if c.TransmitPower < 88 {
		c.Log("Transmit power %d < 88. Adjusting to minimum of 88.\n", c.TransmitPower)
		c.TransmitPower = 88
	} else if c.TransmitPower > 115 {
		c.Log("Transmit power %d > 115. Adjusting to maximum of 115.\n", c.TransmitPower)
		c.TransmitPower = 115
	}

	// If we don't have a valid program ID, then we can set a default one
	if c.RDSProgramID < 1 {
		c.RDSProgramID = 0x3104
	}

	return nil
}

// NewSi4713Driver creates a new Gobot driver for our FM transmitter
func NewSi4713Driver(connector i2c.Connector, cfg Si4713Config, options ...func(i2c.Config)) (*Si4713Driver, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	res := &Si4713Driver{
		name:         gobot.DefaultName("Si4713Driver"),
		i2cConnector: connector,
		Config:       i2c.NewConfig(),
		i2cAddr:      Address,

		Si4713Config: cfg,
	}

	for _, option := range options {
		option(res)
	}

	return res, nil
}
