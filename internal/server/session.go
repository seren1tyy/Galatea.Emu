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

	s.logger.Info("Client connected", "addr", conn.RemoteAddr())

	// 1. СРАЗУ отправляем machoNet.Hello (клиент ждёт именно этого)
	response := protocol.BuildServerHelloResponse()
	s.logger.Info("Sending Server Hello",
		"len", len(response),
		"hex", hex.EncodeToString(response),
	)

	n, err := conn.Write(response)
	if err != nil {
		s.logger.Error("Failed to send hello", "err", err)
		return
	}
	s.logger.Info("Server Hello sent", "bytes", n)

	// 2. Ждём ответ клиента (Client Hello)
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	pkt, err := protocol.ReadPacket(conn)
	if err != nil {
		s.logger.Warn("Client disconnected or timeout", "err", err)
		return
	}

	s.logger.Info("Received Client Hello",
		"len", len(pkt.Payload),
		"hex", hex.EncodeToString(pkt.Payload),
	)

	// ✅ Хендшейк пройден. Далее пойдёт svc.Login или auth challenge
	s.logger.Info("Handshake complete. Ready for authentication.")
}
