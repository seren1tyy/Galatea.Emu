package protocol

import (
	"encoding/binary"
)

// BuildServerHelloResponse собирает корректный пакет machoNet.Hello
func BuildServerHelloResponse() []byte {
	// 1. Marshal-поток (PyDict)
	m := NewMarshaler()
	entries := []DictEntry{
		// Попробуем более короткую версию, если EVE-TRANQUILITY@ccp не проходит
		{"version", EVEProjectVersion},
		{"build", EVEBuildVersion},
		{"region", EVEProjectRegion},
		{"codename", EVEProjectCodename},
		{"birthday", EVEBirthday},
	}
	m.WriteOrderedDict(entries)
	marshalPayload := m.Bytes()

	// 2. MachoNet Header (10 байт):
	// [2] MachoNet Version = 0x019E (414) (LittleEndian: 9E 01)
	// [2] Service ID       = 0x0000 (machoNet)
	// [2] Method ID        = 0x0001 (Hello)
	// [4] Call ID          = 0x00000000
	machoHeader := make([]byte, 10)
	binary.LittleEndian.PutUint16(machoHeader[0:2], MachoNetVersion) // MachoNet Version (414)
	binary.LittleEndian.PutUint16(machoHeader[2:4], 0x0000)           // Service ID
	binary.LittleEndian.PutUint16(machoHeader[4:6], 0x0001)           // Method ID
	binary.LittleEndian.PutUint32(machoHeader[6:10], 0x00000000)      // Call ID

	// 3. Собираем: [4 bytes Len LE] + [10 bytes MachoHeader] + [MarshalPayload]
	fullPayload := append(machoHeader, marshalPayload...)
	pkt := make([]byte, 4+len(fullPayload))
	binary.LittleEndian.PutUint32(pkt[:4], uint32(len(fullPayload)))
	copy(pkt[4:], fullPayload)

	return pkt
}
