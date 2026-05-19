package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Packet представляет сетевой пакет EVE (StacklessIO)
type Packet struct {
	Payload []byte
}

// ReadPacket читает пакет из соединения.
// В протоколе EVE длина кодируется 4 байтами (uint32, LittleEndian).
func ReadPacket(r io.Reader) (*Packet, error) {
	// 1. Читаем заголовок длины (4 байта)
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return nil, fmt.Errorf("read length header: %w", err)
	}

	// Пустые пакеты бывают (keepalive/heartbeat), пропускаем молча
	if length == 0 {
		return &Packet{Payload: nil}, nil
	}

	// Защита от аномально больших пакетов
	if length > 1024*1024 { // 1MB
		return nil, fmt.Errorf("packet too large: %d bytes", length)
	}

	// 2. Читаем тело пакета
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, fmt.Errorf("read payload: %w", err)
	}

	return &Packet{Payload: payload}, nil
}
