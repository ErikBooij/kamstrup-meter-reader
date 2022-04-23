package kamstrup

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/tarm/serial"
)

type KamstrupClient interface {
	ClosePort()
	ReadRegister(register int16) RegisterValue
	ReadRegisterWithRetry(register int16, retries int, backoff time.Duration) RegisterValue
}

type kamstrupClient struct {
	serialPort *serial.Port
}

func CreateKamstrupClient(serialPort string, readTimeout time.Duration) (KamstrupClient, error) {
	serialConfig := &serial.Config{
		Name:        serialPort,
		Baud:        1200,
		ReadTimeout: readTimeout,
		Parity:      serial.ParityNone,
		StopBits:    serial.Stop2,
	}

	port, err := serial.OpenPort(serialConfig)

	if err != nil {
		return nil, err
	}

	return &kamstrupClient{
		serialPort: port,
	}, nil
}

func (c *kamstrupClient) ClosePort() {
	c.serialPort.Close()
}

func (c *kamstrupClient) ReadRegister(register int16) RegisterValue {
	if err := c.send(register); err != nil {
		return errorValue(err)
	}

	return c.receive(register)
}

func (c *kamstrupClient) ReadRegisterWithRetry(register int16, retries int, backoff time.Duration) RegisterValue {
	regValue := c.ReadRegister(register)

	for regValue.Error() != nil && retries > 0 {
		time.Sleep(backoff)

		regValue = c.ReadRegister(register)

		retries--
	}

	return regValue
}

func (c *kamstrupClient) crc1021(msg []byte) int32 {
	var poly int32 = 0x1021
	var reg int32 = 0x0000

	for _, b := range msg {
		var mask byte = 0x80

		for mask > 0 {
			reg <<= 1

			if b&mask > 0 {
				reg |= 1
			}

			mask >>= 1

			if reg&0x10000 > 0 {
				reg &= 0xffff
				reg ^= poly
			}
		}
	}

	return reg
}

func (c *kamstrupClient) decodeBase(parsed []byte) uint64 {
	var x uint64 = 0

	i := 0

	for i = 0; i < int(parsed[5]); i++ {
		x <<= 8
		x |= uint64(parsed[i+7])
	}

	return x
}

func (c *kamstrupClient) decodeExp(parsed []byte) float64 {
	j := int(parsed[6] & 0x3f)

	if parsed[6]&0x40 > 0 {
		j *= -1
	}

	exp := math.Pow(10, float64(j))

	if parsed[6]&0x80 > 0 {
		exp *= -1
	}

	return exp
}

func (c *kamstrupClient) decodeUnit(parsed []byte) string {
	if u, ok := units[parsed[4]]; ok {
		return u
	}

	return ""
}

func (c *kamstrupClient) parse(raw []byte) ([]byte, error) {
	i := 1

	parsed := make([]byte, 0)

	for i < len(raw)-1 {
		if raw[i] == 0x1b {
			v := raw[i+1] ^ 0xff

			parsed = append(parsed, v)

			i += 2
		} else {
			parsed = append(parsed, raw[i])
			i += 1
		}
	}

	if c.crc1021(parsed) != 0 {
		return parsed, errors.New("CRC error in returned data")
	}

	return parsed, nil
}

func (c *kamstrupClient) receive(register int16) RegisterValue {
	r := bufio.NewReader(c.serialPort)

	buf, err := r.ReadBytes(0x0d)

	if err != nil {
		return errorValue(err)
	}

	parsed, err := c.parse(buf)

	if err != nil {
		return errorValue(err)
	}

	if !c.validate(parsed, register) {
		return errorValue(fmt.Errorf("parsed message does not appear to be a valid response: %x", parsed))
	}

	unit := c.decodeUnit(parsed)
	base := c.decodeBase(parsed)
	exp := c.decodeExp(parsed)

	return registerValue(float64(int(base))*exp, unit)
}

func (c *kamstrupClient) send(register int16) error {
	var prefix byte = 0x80
	msg := []byte{0x3f, 0x10, 0x01, (byte)(register >> 8), (byte)(register & 0xff)}

	msg = append(msg, 0)
	msg = append(msg, 0)

	crc := c.crc1021(msg)

	msg[5] = (byte)(crc >> 8)
	msg[6] = (byte)(crc & 0xff)

	s := []byte{prefix}

	for _, b := range msg {
		if _, ok := escapes[b]; ok {
			s = append(s, 0x1b)
			s = append(s, b^0xff)
		} else {
			s = append(s, b)
		}
	}

	s = append(s, 0x0d)

	_, err := c.serialPort.Write(s)

	return err
}

func (c *kamstrupClient) validate(parsed []byte, register int16) bool {
	if len(parsed) < 6 {
		return false
	}

	if len(parsed) < int(parsed[5])+7 {
		return false
	}

	if parsed[0] != 0x3F || parsed[1] != 0x10 {
		return false
	}

	if parsed[2] != (byte)(register>>8) || parsed[3] != (byte)(register&0xFF) {
		return false
	}

	return true
}
