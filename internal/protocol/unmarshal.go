package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Unmarshaler парсит marshal-поток обратно
type Unmarshaler struct {
	data []byte
	pos  int
}

func NewUnmarshaler(data []byte) *Unmarshaler {
	return &Unmarshaler{data: data, pos: 0}
}

func (u *Unmarshaler) readByte() (byte, error) {
	if u.pos >= len(u.data) {
		return 0, errors.New("unexpected end of stream")
	}
	b := u.data[u.pos]
	u.pos++
	return b, nil
}

func (u *Unmarshaler) readUint32() (uint32, error) {
	if u.pos+4 > len(u.data) {
		return 0, errors.New("short read uint32")
	}
	v := binary.LittleEndian.Uint32(u.data[u.pos:])
	u.pos += 4
	return v, nil
}

func (u *Unmarshaler) readSizeEx() (uint32, error) {
	b, err := u.readByte()
	if err != nil {
		return 0, err
	}
	if b == 0xFF {
		return u.readUint32()
	}
	return uint32(b), nil
}

func (u *Unmarshaler) ReadString() (string, error) {
	op, err := u.readByte()
	if err != nil {
		return "", err
	}

	switch op {
	case OpPyEmptyString:
		return "", nil
	case OpPyCharString:
		b, _ := u.readByte()
		return string([]byte{b}), nil
	case OpPyLongString:
		ln, err := u.readSizeEx()
		if err != nil {
			return "", err
		}
		if u.pos+int(ln) > len(u.data) {
			return "", errors.New("short read string")
		}
		s := string(u.data[u.pos : u.pos+int(ln)])
		u.pos += int(ln)
		return s, nil
	default:
		return "", fmt.Errorf("unexpected string opcode 0x%02X", op)
	}
}

func (u *Unmarshaler) ReadInt() (int32, error) {
	op, err := u.readByte()
	if err != nil {
		return 0, err
	}

	switch op {
	case OpPyMinusOne:
		return -1, nil
	case OpPyZeroInteger:
		return 0, nil
	case OpPyOneInteger:
		return 1, nil
	case OpPyByte:
		b, _ := u.readByte()
		return int32(int8(b)), nil
	case OpPySignedShort:
		if u.pos+2 > len(u.data) {
			return 0, errors.New("short read short")
		}
		v := int32(int16(binary.LittleEndian.Uint16(u.data[u.pos:])))
		u.pos += 2
		return v, nil
	case OpPyLong:
		v, err := u.readUint32()
		return int32(v), err
	default:
		return 0, fmt.Errorf("unexpected int opcode 0x%02X", op)
	}
}

// ReadDict читает словарь и возвращает map[string]interface{}
func (u *Unmarshaler) ReadDict() (map[string]interface{}, error) {
	op, err := u.readByte()
	if err != nil {
		return nil, err
	}
	if op != OpPyDict {
		return nil, fmt.Errorf("expected dict opcode 0x%02X, got 0x%02X", OpPyDict, op)
	}

	count, err := u.readSizeEx()
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for i := uint32(0); i < count; i++ {
		// Читаем значение (любого типа)
		valOp, _ := u.readByte()
		u.pos-- // Возвращаем байт назад

		var val interface{}
		switch {
		case valOp == OpPyLongString || valOp == OpPyEmptyString || valOp == OpPyCharString:
			val, err = u.ReadString()
		case valOp == OpPyLong || valOp == OpPySignedShort || valOp == OpPyByte ||
			valOp == OpPyMinusOne || valOp == OpPyZeroInteger || valOp == OpPyOneInteger:
			val, err = u.ReadInt()
		default:
			return nil, fmt.Errorf("unsupported dict value type 0x%02X", valOp)
		}
		if err != nil {
			return nil, err
		}

		// Читаем ключ (всегда строка)
		key, err := u.ReadString()
		if err != nil {
			return nil, err
		}

		result[key] = val
	}

	return result, nil
}

// ParseClientHandshake парсит ответ клиента
func ParseClientHandshake(data []byte) (map[string]interface{}, error) {
	// Проверяем заголовок
	if len(data) < 5 || data[0] != MarshalHeader {
		return nil, fmt.Errorf("invalid marshal header: 0x%02X", data[0])
	}

	u := NewUnmarshaler(data)
	// Пропускаем header и mapcount
	u.pos = 5

	return u.ReadDict()
}

// ReadPacketWithMarshal читает пакет и проверяет marshal-заголовок
func ReadPacketWithMarshal(r io.Reader) ([]byte, error) {
	pkt, err := ReadPacket(r)
	if err != nil {
		return nil, err
	}

	// Проверяем, что это marshal-поток
	if len(pkt.Payload) == 0 || pkt.Payload[0] != MarshalHeader {
		return nil, fmt.Errorf("not a marshal packet: 0x%02X", pkt.Payload[0])
	}

	return pkt.Payload, nil
}
