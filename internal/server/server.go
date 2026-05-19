package server

import (
"context"
"fmt"
"log"
"net"
)

type Server struct {
host   string
port   int
logger *log.Logger
}

func New(host string, port int, logger *log.Logger) *Server {
return &Server{host: host, port: port, logger: logger}
}

func (s *Server) Start(ctx context.Context) error {
addr := fmt.Sprintf("%s:%d", s.host, s.port)
listener, err := net.Listen("tcp", addr)
if err != nil {
return fmt.Errorf("listen tcp: %w", err)
}
s.logger.Printf("Server started on %s", addr)

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
s.logger.Printf("accept error: %v", err)
continue
}
}

// Каждое соединение в отдельной горутине!
s.logger.Printf("New connection from %s", conn.RemoteAddr())
go s.handleSession(conn)
}
}
