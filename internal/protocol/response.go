package protocol

import (
	"bytes"
	"encoding/binary"
	"io"
)

// SendServerHello отправляет самый первый пакет клиенту
func SendServerHello(w io.Writer) error {
	// Формируем payload.
	// В MachoNet это обычно сериализованный объект, но для простого хендшейка
	// часто достаточно отправить "голые" данные версии или пустой пакет с правильным Opcode.
	// Попробуем отправить структуру, похожую на EVEmu Handshake.

	buf := &bytes.Buffer{}

	// Примерная структура Handshake ответа:
	// 1. MachoNet Version (uint16)
	if err := binary.Write(buf, binary.LittleEndian, uint16(MachoNetVersion)); err != nil {
		return err
	}

	// 2. Version String (сначала длина uint16, потом строка)
	// Клиент ожидает строку версии
	if err := writeStringWithLen(buf, EVEProjectVersion); err != nil {
		return err
	}

	// 3. Build Number (int32)
	if err := binary.Write(buf, binary.LittleEndian, EVEBuildVersion); err != nil {
		return err
	}

	// 4. Region (string)
	if err := writeStringWithLen(buf, EVEProjectRegion); err != nil {
		return err
	}

	// 5. CodeName (string)
	if err := writeStringWithLen(buf, EVEProjectCodename); err != nil {
		return err
	}

	// 6. Birthday (int32)
	if err := binary.Write(buf, binary.LittleEndian, EVEBirthday); err != nil {
		return err
	}

	// Теперь оборачиваем в пакет с длиной
	packetLen := uint32(buf.Len())

	// Записываем длину (4 байта)
	if err := binary.Write(w, binary.LittleEndian, packetLen); err != nil {
		return err
	}

	// Записываем тело
	_, err := w.Write(buf.Bytes())
	return err
}

func writeStringWithLen(w *bytes.Buffer, s string) error {
	if err := binary.Write(w, binary.LittleEndian, uint16(len(s))); err != nil {
		return err
	}
	_, err := w.Write([]byte(s))
	return err
}
