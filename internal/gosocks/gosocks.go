package gosocks

import (
	"fmt"
	"github.com/nessai1/gosocks/internal/storage"
	"go.uber.org/zap"
	"net"
)

const SocksVersion int = 5

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

type UserStorage interface {
}

type Proxy struct {
	listener net.Listener
	logger   *zap.Logger
	storage  storage.Storage
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
	p.logger.Info("Got new client connection", zap.String("remote_address", conn.RemoteAddr().String()))

	p.handleSocks5(conn)

	defer func() {
		err := conn.Close()
		if err != nil {
			p.logger.Error("Cannot close client connection", zap.Error(err))
		}
	}()
}

func (p *Proxy) handleSocks5(conn net.Conn) error {

	client, err := p.handshake(conn)
	if err != nil {
		p.logger.Error("Cannot handshake with client", zap.String("remote_address", conn.RemoteAddr().String()), zap.Error(err))
	}

	if client.AuthMethod != Credentials {
		p.logger.Info("Successful authorized handshake", zap.String("remote_address", client.RemoteAddr.String()), zap.String("login", client.Credentials.Login))
	} else {
		p.logger.Info("Successful anon handshake", zap.String("remote_address", client.RemoteAddr.String()))
	}

	return nil
}
