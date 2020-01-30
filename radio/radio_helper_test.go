package radio

import (
	"errors"
	"fmt"
	"sync"

	"gobot.io/x/gobot/drivers/i2c"
)

// I2CTestAdaptor is useful to implement tests for
// passing i2c messages back and forth.
type I2CTestAdaptor struct {
	name          string
	written       []byte
	lastWritten   []byte
	mtx           sync.Mutex
	i2cConnectErr bool
	i2cReadImpl   func(*I2CTestAdaptor, []byte) (int, error)
	i2cWriteImpl  func(*I2CTestAdaptor, []byte) (int, error)
}

func (t *I2CTestAdaptor) DigitalWrite(/* s */ string, /* b */ byte) (err error) {
	return nil
}

func (t *I2CTestAdaptor) Read(b []byte) (count int, err error) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	return t.i2cReadImpl(t, b)
}

func (t *I2CTestAdaptor) Write(b []byte) (count int, err error) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	t.written = append(t.written, b...)
	return t.i2cWriteImpl(t, b)
}

func (t *I2CTestAdaptor) Close() error {
	return nil
}

func (t *I2CTestAdaptor) ReadByte() (val byte, err error) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	bytes := []byte{0}
	bytesRead, err := t.i2cReadImpl(t, bytes)
	if err != nil {
		return 0, err
	}
	if bytesRead != 1 {
		return 0, fmt.Errorf("buffer underrun")
	}
	val = bytes[0]
	return
}

func (t *I2CTestAdaptor) ReadByteData(/* reg */ uint8) (val uint8, err error) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	bytes := []byte{0}
	bytesRead, err := t.i2cReadImpl(t, bytes)
	if err != nil {
		return 0, err
	}
	if bytesRead != 1 {
		return 0, fmt.Errorf("buffer underrun")
	}
	val = bytes[0]
	return
}

func (t *I2CTestAdaptor) ReadWordData(/* reg */ uint8) (val uint16, err error) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	bytes := []byte{0, 0}
	bytesRead, err := t.i2cReadImpl(t, bytes)
	if err != nil {
		return 0, err
	}
	if bytesRead != 2 {
		return 0, fmt.Errorf("buffer underrun")
	}
	l, h := bytes[0], bytes[1]
	return (uint16(h) << 8) | uint16(l), err
}

func (t *I2CTestAdaptor) WriteByte(val byte) (err error) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	t.written = append(t.written, val)
	bytes := []byte{val}
	_, err = t.i2cWriteImpl(t, bytes)
	return
}

func (t *I2CTestAdaptor) WriteByteData(reg uint8, val uint8) (err error) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	t.written = append(t.written, reg)
	t.written = append(t.written, val)
	bytes := []byte{val}
	_, err = t.i2cWriteImpl(t, bytes)
	return
}

func (t *I2CTestAdaptor) WriteWordData(reg uint8, val uint16) (err error) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	t.written = append(t.written, reg)
	l := uint8(val & 0xff)
	h := uint8((val >> 8) & 0xff)
	t.written = append(t.written, l)
	t.written = append(t.written, h)
	bytes := []byte{l, h}
	_, err = t.i2cWriteImpl(t, bytes)
	return
}

func (t *I2CTestAdaptor) WriteBlockData(reg uint8, b []byte) (err error) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	t.written = append(t.written, reg)
	t.written = append(t.written, b...)
	_, err = t.i2cWriteImpl(t, b)
	return
}

func (t *I2CTestAdaptor) GetConnection( /* address */ int, /* bus */ int) (connection i2c.Connection, err error) {
	if t.i2cConnectErr {
		return nil, errors.New("invalid i2c connection")
	}
	return t, nil
}

func (t *I2CTestAdaptor) GetDefaultBus() int {
	return 0
}

func (t *I2CTestAdaptor) Name() string          { return t.name }
func (t *I2CTestAdaptor) SetName(n string)      { t.name = n }
func (t *I2CTestAdaptor) Connect() (err error)  { return }
func (t *I2CTestAdaptor) Finalize() (err error) { return }
