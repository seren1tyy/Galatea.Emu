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

	// 1. СРАЗУ отправляем Server Hello (клиент EVE ждет инициативы от сервера!)
	s.logger.Printf("[SESSION] >>> Sending initial Server Hello to wake up client...")
	response := protocol.BuildServerHelloResponse()
	s.logger.Printf("[PACKET] Server Hello details:")
	s.logger.Printf("  - Total length: %d bytes", len(response))
	s.logger.Printf("  - Hex dump: %s", hex.EncodeToString(response))

	writeStart := time.Now()
	n, err := conn.Write(response)
	writeDuration := time.Since(writeStart)

	if err != nil {
		s.logger.Printf("[SESSION] Failed to send Server Hello after %v: %v", writeDuration, err)
		return
	}
	s.logger.Printf("[SESSION] Server Hello sent: %d bytes in %v", n, writeDuration)

	// 2. Теперь ждём ответный пакет от клиента (Client Hello или другой init-пакет)
	s.logger.Printf("[SESSION] Waiting for client response (timeout=10s)...")
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	readStart := time.Now()
	clientPkt, err := protocol.ReadPacket(conn)
	readDuration := time.Since(readStart)

	if err != nil {
		s.logger.Printf("[SESSION] Failed to receive client response after %v: %v", readDuration, err)
		s.logger.Printf("[SESSION] Connection state: closed=%v", conn == nil)
		return
	}

	s.logger.Printf("[SESSION] Received client response after %v", readDuration)
	s.logger.Printf("[PACKET] Client response details:")
	s.logger.Printf("  - Total length: %d bytes", len(clientPkt.Payload))
	s.logger.Printf("  - Hex dump: %s", hex.EncodeToString(clientPkt.Payload))

	// Попробуем распарсить заголовок MachoNet если достаточно данных
	if len(clientPkt.Payload) >= 10 {
		s.logger.Printf("[MACHONET] Parsing header (first 10 bytes):")
		s.logger.Printf("  - Bytes: %s", hex.EncodeToString(clientPkt.Payload[:10]))
		s.logger.Printf("  - MachoNet Version: 0x%04X", uint16(clientPkt.Payload[4])|uint16(clientPkt.Payload[5])<<8)
		s.logger.Printf("  - Service ID: 0x%04X", uint16(clientPkt.Payload[6])|uint16(clientPkt.Payload[7])<<8)
		s.logger.Printf("  - Method ID: 0x%04X", uint16(clientPkt.Payload[8])|uint16(clientPkt.Payload[9])<<8)
	}

	// ✅ Хендшейк пройден. Далее пойдёт svc.Login или auth challenge
	s.logger.Println("[SESSION] Handshake complete. Ready for authentication.")
	s.logger.Printf("[SESSION] Connection %s is now in authenticated state", conn.RemoteAddr())
}
