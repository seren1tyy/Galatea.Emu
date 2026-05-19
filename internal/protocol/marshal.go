package protocol

import (
	"bytes"
	"encoding/binary"
)

// Opcodes (точные значения из opcodes.go)
const (
	MarshalHeader        = 0x7E
	OpPyNone             = 0x01
	OpPyToken            = 0x02
	OpPyLongLong         = 0x03
	OpPyLong             = 0x04
	OpPySignedShort      = 0x05
	OpPyByte             = 0x06
	OpPyMinusOne         = 0x07
	OpPyZeroInteger      = 0x08
	OpPyOneInteger       = 0x09
	OpPyReal             = 0x0A
	OpPyZeroReal         = 0x0B
	OpPyBuffer           = 0x0D
	OpPyEmptyString      = 0x0E
	OpPyCharString       = 0x0F
	OpPyShortString      = 0x10
	OpPyStringTableItem  = 0x11
	OpPyWStringUCS2      = 0x12
	OpPyLongString       = 0x13
	OpPyTuple            = 0x14
	OpPyList             = 0x15
	OpPyDict             = 0x16
	OpPyObject           = 0x17
	OpPySubStruct        = 0x19
	OpPySavedStreamElem  = 0x1B
	OpPyChecksumedStream = 0x1C
	OpPyTrue             = 0x1F
	OpPyFalse            = 0x20
	OpCPicked            = 0x21
	OpPyObjectEx1        = 0x22
	OpPyObjectEx2        = 0x23
	OpPyEmptyTuple       = 0x24
	OpPyOneTuple         = 0x25
	OpPyEmptyList        = 0x26
	OpPyOneList          = 0x27
	OpPyEmptyWString     = 0x28
	OpPyWStringUCS2Char  = 0x29
	OpPyPackedRow        = 0x2A
	OpPySubStream        = 0x2B
	OpPyTwoTuple         = 0x2C
	OpPackedTerminator   = 0x2D
	OpPyWStringUTF8      = 0x2E
	OpPyVarInteger       = 0x2F
)

type Marshaler struct {
	buf bytes.Buffer
}

func NewMarshaler() *Marshaler {
	m := &Marshaler{}
	m.buf.WriteByte(MarshalHeader)
	m.writeUint32(0) // mapcount = 0 (object sharing отключен)
	return m
}

func (m *Marshaler) writeUint32(v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	m.buf.Write(b[:])
}

func (m *Marshaler) writeUint16(v uint16) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], v)
	m.buf.Write(b[:])
}

// writeSizeEx: если размер < 255 → 1 байт, иначе 0xFF + uint32
func (m *Marshaler) writeSizeEx(size uint32) {
	if size < 0xFF {
		m.buf.WriteByte(byte(size))
	} else {
		m.buf.WriteByte(0xFF)
		_ = binary.Write(&m.buf, binary.LittleEndian, size)
	}
}

// WriteInt: строгая логика из marshal.go
func (m *Marshaler) WriteInt(v int32) {
	switch {
	case v == -1:
		m.buf.WriteByte(OpPyMinusOne)
	case v == 0:
		m.buf.WriteByte(OpPyZeroInteger)
	case v == 1:
		m.buf.WriteByte(OpPyOneInteger)
	case v >= -0x80 && v < 0x80:
		m.buf.WriteByte(OpPyByte)
		m.buf.WriteByte(byte(v))
	case v >= -0x8000 && v < 0x8000:
		m.buf.WriteByte(OpPySignedShort)
		m.writeUint16(uint16(v))
	default:
		m.buf.WriteByte(OpPyLong)
		m.writeUint32(uint32(v))
	}
}

// WriteString: без string-table lookup, но с точными опкодами
func (m *Marshaler) WriteString(s string) {
	if len(s) == 0 {
		m.buf.WriteByte(OpPyEmptyString)
	} else if len(s) == 1 {
		m.buf.WriteByte(OpPyCharString)
		m.buf.WriteByte(s[0])
	} else {
		m.buf.WriteByte(OpPyLongString)
		m.writeSizeEx(uint32(len(s)))
		m.buf.WriteString(s)
	}
}

// WriteDict: ЗНАЧЕНИЕ идёт ПЕРЕД КЛЮЧОМ (специфика EVE wire format)
func (m *Marshaler) WriteDict(items map[string]interface{}) {
	m.buf.WriteByte(OpPyDict)
	m.writeSizeEx(uint32(len(items)))

	for k, v := range items {
		// 1. Значение
		switch val := v.(type) {
		case string:
			m.WriteString(val)
		case int32:
			m.WriteInt(val)
		case int:
			m.WriteInt(int32(val))
		}
		// 2. Ключ (всегда строка)
		m.WriteString(k)
	}
}

func (m *Marshaler) Bytes() []byte {
	return m.buf.Bytes()
}

// DictEntry гарантирует порядок ключей в словаре (value → key)
type DictEntry struct {
	Key   string
	Value interface{}
}

// WriteOrderedDict записывает словарь в строго заданном порядке
func (m *Marshaler) WriteOrderedDict(entries []DictEntry) {
	m.buf.WriteByte(OpPyDict)
	m.writeSizeEx(uint32(len(entries)))

	for _, e := range entries {
		// 1. Значение
		switch v := e.Value.(type) {
		case string:
			m.WriteString(v)
		case int32:
			m.WriteInt(v)
		case int:
			m.WriteInt(int32(v))
		}
		// 2. Ключ
		m.WriteString(e.Key)
	}
}
