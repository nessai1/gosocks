package gosocks

import (
	"context"
	"fmt"
	"github.com/nessai1/gosocks/internal/storage"
	"net"
	"time"
)

type AuthMethod int

// Supported auth methods for this proxy
// TODO: add GSS-API support for compliant implementation
const (
	NoAuth      AuthMethod = 0x00
	Credentials AuthMethod = 0x02
)

type Client struct {
	AuthMethod  AuthMethod
	Conn        net.Conn
	Credentials *storage.Client // not nil of AuthMethod == Credentials
}

func (p *Proxy) handshake(conn net.Conn) (*Client, error) {
	head := make([]byte, 2)
	n, err := conn.Read(head)

	if err != nil {
		return nil, fmt.Errorf("cannot read 2 header bytes: %w", err)
	}

	if n != 2 {
		return nil, fmt.Errorf("incorrect header format: read less that 2 bytes (%d)", n)
	}

	if int(head[0]) != SocksVersion {
		return nil, fmt.Errorf("incorrect socks version: require %d, got %d", SocksVersion, int(head[0]))
	}

	methodsCount := int(head[1])
	if methodsCount <= 0 {
		return nil, fmt.Errorf("client has no supported auth methods")
	}

	var supportCredentials, supportNoAuth bool
	methods := make([]byte, methodsCount)
	n, err = conn.Read(methods)
	if err != nil {
		return nil, fmt.Errorf("cannot read auth methods of client: %w", err)
	}
	if n != methodsCount {
		return nil, fmt.Errorf("incorrect count of read methods: require %d, got %d", methodsCount, n)
	}

	for _, method := range methods {
		if AuthMethod(method) == NoAuth {
			supportNoAuth = true
		}

		if AuthMethod(method) == Credentials {
			supportCredentials = true
		}
	}

	if p.storage != nil && !supportCredentials {
		return nil, fmt.Errorf("proxy require auth method, but client doesn't support it")
	}

	if supportCredentials {
		_, err = conn.Write([]byte{byte(SocksVersion), byte(Credentials)})
		if err != nil {
			return nil, fmt.Errorf("cannot send selected auth method")
		}

		client, err := p.fetchCredentials(conn)
		if err != nil {
			return nil, fmt.Errorf("cannot get credentials from client: %w", err)
		}

		return &Client{
			AuthMethod:  Credentials,
			Conn:        conn,
			Credentials: client,
		}, nil
	} else if supportNoAuth {
		_, err = conn.Write([]byte{byte(SocksVersion), byte(NoAuth)})
		if err != nil {
			return nil, fmt.Errorf("cannot send selected auth method")
		}

		return &Client{
			AuthMethod:  NoAuth,
			Conn:        conn,
			Credentials: nil,
		}, nil
	}

	return nil, fmt.Errorf("no supported methods by proxy")
}

func (p *Proxy) fetchCredentials(conn net.Conn) (*storage.Client, error) {
	vAndULen := make([]byte, 2)
	n, err := conn.Read(vAndULen)
	if err != nil {
		return nil, fmt.Errorf("cannot read username len: %w", err)
	}
	if n != 2 {
		return nil, fmt.Errorf("count of read username len message must be equal 2, got %d", n)
	}

	usernameLen := int(vAndULen[1])
	username := make([]byte, usernameLen)
	n, err = conn.Read(username)
	if err != nil {
		return nil, fmt.Errorf("cannot read username: %w", err)
	}
	if n != usernameLen {
		return nil, fmt.Errorf("count of read username message must be equal %d, got %d", usernameLen, n)
	}

	passwordLenBuffer := make([]byte, 1)
	n, err = conn.Read(passwordLenBuffer)
	if err != nil {
		return nil, fmt.Errorf("cannot read password len: %w", err)
	}
	if n != 1 {
		return nil, fmt.Errorf("count of read password len message must be equal 1, got %d", n)
	}

	passwordLen := int(passwordLenBuffer[0])
	password := make([]byte, passwordLen)
	n, err = conn.Read(password)
	if err != nil {
		return nil, fmt.Errorf("cannot read password: %w", err)
	}
	if n != passwordLen {
		return nil, fmt.Errorf("count of read password message must be equal %d, got %d", passwordLen, n)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	client, err := p.storage.GetClient(ctx, string(username), string(password))
	if err != nil {
		conn.Write([]byte{0x01, 0x01})
		return nil, fmt.Errorf("cannot get client from storage: %w", err)
	}

	conn.Write([]byte{0x01, 0x00})

	return client, nil
}
