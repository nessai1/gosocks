package gosocks

import (
	"fmt"
	"go.uber.org/zap"
	"net"
)

func ListenAndServe(address string) error {

	l, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("cannot start listen address '%s': %w", address, err)
	}

	proxy := Proxy{
		listener: l,
		logger:   zap.Must(zap.NewProduction()),
	}

	err = proxy.Listen()
	if err != nil {
		return fmt.Errorf("cannot start proxy listen: %w", err)
	}

	return nil
}

type Proxy struct {
	listener net.Listener
	logger   *zap.Logger
}

func (p *Proxy) Listen() error {
	p.logger.Info("Application listening start")
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			p.logger.Error("Cannot start listen connection", zap.Error(err))
		}

		go p.handleConnection(conn)
	}
}

func (p *Proxy) handleConnection(conn net.Conn) {
	p.logger.Info("Got new connection", zap.String("remote_address", conn.RemoteAddr().String()))
	conn.Close()
}
