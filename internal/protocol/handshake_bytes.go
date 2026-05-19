package protocol

import (
	"bytes"
	"encoding/binary"
)

// BuildServerHelloBytes собирает точные байты для отправки клиенту
// Временно для теста
func BuildServerHelloBytes() []byte {
	testPayload := []byte{0x01, 0x02, 0x03, 0x04}
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, uint32(len(testPayload)))
	buf.Write(testPayload)
	return buf.Bytes()
}

func writeString(buf *bytes.Buffer, s string) {
	binary.Write(buf, binary.LittleEndian, uint16(len(s)))
	buf.WriteString(s)
}
