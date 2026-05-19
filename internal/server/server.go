package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
)

type Server struct {
	host   string
	port   int
	logger *slog.Logger
}

func New(host string, port int, logger *slog.Logger) *Server {
	return &Server{host: host, port: port, logger: logger}
}

func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen tcp: %w", err)
	}
	s.logger.Info("Server started", "addr", addr)

	// Отложенная отмена
	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil // Нормальный выход
			default:
				s.logger.Error("accept error", "err", err)
				continue
			}
		}

		// Каждое соединение в отдельной горутине!
		s.logger.Info("New connection", "remote", conn.RemoteAddr())
		go s.handleSession(conn)
	}
}
