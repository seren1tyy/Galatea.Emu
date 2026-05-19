package server

import (
	"encoding/hex"
	"net"
	"time"

	"Galatea.Emu/internal/protocol"
)

func (s *Server) handleSession(conn net.Conn) {
	defer conn.Close()

	// Мгновенная доставка (критично для Windows)
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
		s.logger.Printf("[TCP] SetNoDelay=true on %s", conn.RemoteAddr())
	}

	s.logger.Printf("[SESSION] Client connected from %s", conn.RemoteAddr())
	s.logger.Printf("[SESSION] Local addr: %s", conn.LocalAddr())

	// 1. Сначала ждём любой пакет от клиента (EVE client может отправлять init-пакет первым)
	s.logger.Printf("[SESSION] Waiting for initial packet from client (timeout=10s)...")
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	readStart := time.Now()
	clientPkt, err := protocol.ReadPacket(conn)
	readDuration := time.Since(readStart)

	if err != nil {
		s.logger.Printf("[SESSION] ❌ Failed to receive initial packet after %v: %v", readDuration, err)
		s.logger.Printf("[SESSION] Connection state: closed=%v", conn == nil)
		return
	}

	s.logger.Printf("[SESSION] ✅ Received initial packet after %v", readDuration)
	s.logger.Printf("[PACKET] Initial packet details:")
	s.logger.Printf("  - Total length: %d bytes", len(clientPkt.Payload))
	s.logger.Printf("  - Hex dump: %s", hex.EncodeToString(clientPkt.Payload))

	// Попробуем распарсить заголовок MachoNet если достаточно данных
	if len(clientPkt.Payload) >= 10 {
		s.logger.Printf("[MACHONET] Parsing header (first 10 bytes):")
		s.logger.Printf("  - Bytes: %s", hex.EncodeToString(clientPkt.Payload[:10]))
	}

	// 2. Теперь отправляем Server Hello в ответ
	response := protocol.BuildServerHelloResponse()
	s.logger.Printf("[SESSION] Building Server Hello response...")
	s.logger.Printf("[PACKET] Server Hello details:")
	s.logger.Printf("  - Total length: %d bytes", len(response))
	s.logger.Printf("  - Hex dump: %s", hex.EncodeToString(response))

	// Разбор Server Hello для отладки
	if len(response) >= 4 {
		var pktLen uint32
		for i := range response[:4] {
			pktLen |= uint32(response[i]) << (i * 8)
		}
		s.logger.Printf("  - Declared payload length: %d bytes", pktLen)

		if len(response) >= 14 {
			s.logger.Printf("  - MachoNet Version: 0x%04X", uint16(response[4])|uint16(response[5])<<8)
			s.logger.Printf("  - Service ID: 0x%04X", uint16(response[6])|uint16(response[7])<<8)
			s.logger.Printf("  - Method ID: 0x%04X", uint16(response[8])|uint16(response[9])<<8)
			s.logger.Printf("  - Call ID: 0x%08X", uint32(response[10])|uint32(response[11])<<8|uint32(response[12])<<16|uint32(response[13])<<24)
		}
	}

	writeStart := time.Now()
	n, err := conn.Write(response)
	writeDuration := time.Since(writeStart)

	if err != nil {
		s.logger.Printf("[SESSION] Failed to send Server Hello after %v: %v", writeDuration, err)
		return
	}
	s.logger.Printf("[SESSION] Server Hello sent: %d bytes in %v", n, writeDuration)

	// 3. Ждём ответный Client Hello
	s.logger.Printf("[SESSION] Waiting for Client Hello response (timeout=10s)...")
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	readStart = time.Now()
	clientHelloPkt, err := protocol.ReadPacket(conn)
	readDuration = time.Since(readStart)

	if err != nil {
		s.logger.Printf("[SESSION] Failed to receive Client Hello after %v: %v", readDuration, err)
		return
	}

	s.logger.Printf("[SESSION] Received Client Hello after %v", readDuration)
	s.logger.Printf("[PACKET] Client Hello details:")
	s.logger.Printf("  - Length: %d bytes", len(clientHelloPkt.Payload))
	s.logger.Printf("  - Hex dump: %s", hex.EncodeToString(clientHelloPkt.Payload))

	// ✅ Хендшейк пройден. Далее пойдёт svc.Login или auth challenge
	s.logger.Println("[SESSION] Handshake complete. Ready for authentication.")
	s.logger.Printf("[SESSION] Connection %s is now in authenticated state", conn.RemoteAddr())
}
