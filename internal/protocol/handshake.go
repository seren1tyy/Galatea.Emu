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

	// 2. MachoNet Header (9 байт!)
	// [1] MachoNet Version = 0x01 (КРИТИЧНО: был пропущен ранее)
	// [2] Service ID       = 0x0000 (machoNet)
	// [2] Method ID        = 0x0001 (Hello)
	// [4] Call ID          = 0x00000000
	machoHeader := []byte{
		0x01,       // MachoNet Version
		0x00, 0x00, // Service ID
		0x01, 0x00, // Method ID
		0x00, 0x00, 0x00, 0x00, // Call ID
	}

	// 3. Собираем: [4 bytes Len LE] + [9 bytes MachoHeader] + [MarshalPayload]
	fullPayload := append(machoHeader, marshalPayload...)
	pkt := make([]byte, 4+len(fullPayload))
	binary.LittleEndian.PutUint32(pkt[:4], uint32(len(fullPayload)))
	copy(pkt[4:], fullPayload)

	return pkt
}
