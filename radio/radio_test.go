package radio

import (
	"math/rand"
)


func NewI2cTestAdaptor() *I2CTestAdaptor {
	val := &I2CTestAdaptor{
		i2cConnectErr:  false,
	}

	val.i2cReadImpl = func() func(t *I2CTestAdaptor, buff []byte) (int, error) {
		getRevPassCount := 0
		return func(t *I2CTestAdaptor, buff []byte) (int, error) {
			switch t.lastWritten[0] {
			case CMD_POWER_UP:
				buff[0] = STATUS_CTS
				return 1, nil

			case CMD_SET_PROPERTY:
				buff[0] = STATUS_CTS
				return 1, nil

			case CMD_TX_TUNE_POWER:
				buff[0] = STATUS_CTS
				return 1, nil

			case CMD_TX_TUNE_FREQ:
				buff[0] = STATUS_CTS
				return 1, nil

			case CMD_GET_INT_STATUS:
				buff[0] = 0x81
				return 1, nil

			case CMD_TX_TUNE_STATUS:
				buff[0] = STATUS_CTS
				return 1, nil

			case CMD_TX_TUNE_MEASURE:
				buff[0] = STATUS_CTS
				return 1, nil

			case CMD_TX_RDS_PS:
				buff[0] = STATUS_CTS
				return 1, nil

			case CMD_TX_RDS_BUFF:
				buff[0] = STATUS_CTS
				return 1, nil

			case CMD_GPO_CTL:
				buff[0] = STATUS_CTS
				return 1, nil

			case CMD_TX_ASQ_STATUS:
				buff[0] = STATUS_CTS
				return 1, nil

			case CMD_GPO_SET:
				buff[0] = STATUS_CTS
				return 1, nil

			case CMD_GET_REV:
				getRevPassCount++
				buff[0] = STATUS_CTS
				if getRevPassCount == 3 {
					buff[0] = 13
				}
				return 1, nil
			default:
				for i := range buff {
					buff[i] = byte(rand.Intn(2))
				}
				return len(buff), nil
			}
		}
	}()

	val.i2cWriteImpl = func(t *I2CTestAdaptor, buff []byte) (int, error) {
		t.lastWritten = make([]byte, len(buff))
		copy(t.lastWritten, buff)
		return len(buff), nil
	}

	return val
}

