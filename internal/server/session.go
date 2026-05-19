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
	}

	s.logger.Printf("Client connected from %s", conn.RemoteAddr())

	// 1. Сначала ждём любой пакет от клиента (EVE client может отправлять init-пакет первым)
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	clientPkt, err := protocol.ReadPacket(conn)
	if err != nil {
		s.logger.Printf("Failed to receive initial packet: %v", err)
		return
	}

	s.logger.Printf("Received initial packet from client: len=%d hex=%s", len(clientPkt.Payload), hex.EncodeToString(clientPkt.Payload))

	// 2. Теперь отправляем Server Hello в ответ
	response := protocol.BuildServerHelloResponse()
	s.logger.Printf("Sending Server Hello: len=%d hex=%s", len(response), hex.EncodeToString(response))

	n, err := conn.Write(response)
	if err != nil {
		s.logger.Printf("Failed to send hello: %v", err)
		return
	}
	s.logger.Printf("Server Hello sent: bytes=%d", n)

	// 3. Ждём ответный Client Hello
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	clientHelloPkt, err := protocol.ReadPacket(conn)
	if err != nil {
		s.logger.Printf("Failed to receive Client Hello: %v", err)
		return
	}

	s.logger.Printf("Received Client Hello: len=%d hex=%s", len(clientHelloPkt.Payload), hex.EncodeToString(clientHelloPkt.Payload))

	// ✅ Хендшейк пройден. Далее пойдёт svc.Login или auth challenge
	s.logger.Println("Handshake complete. Ready for authentication.")
}
