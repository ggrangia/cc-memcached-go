package cache_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/ggrangia/cc-memcached-go/internal/cache"
	"github.com/ggrangia/cc-memcached-go/internal/parser"
)

type StoreMock struct {
}

func (s *StoreMock) Delete(key string) error {
	return nil
}

func (s *StoreMock) Save(key string, data cache.Data) error {
	return nil
}

func (s *StoreMock) Get(key string) (cache.Data, bool, error) {
	return cache.Data{}, false, nil
}

type FullNetConnMock struct {
	read func(b []byte) (n int, err error)
	// write            func(b []byte) (n int, err error)
	close            func() error
	localAddr        func() net.Addr
	remoteAddr       func() net.Addr
	setDeadline      func(t time.Time) error
	setReadDeadline  func(t time.Time) error
	setWriteDeadline func(t time.Time) error
}

func (f FullNetConnMock) Write(b []byte) (n int, err error) {
	fmt.Println("MOCKING WRITE")
	l := len(b)

	fmt.Println("Writing ", b)
	fmt.Println("Str: ", string(b))
	return l, nil
}

func (f FullNetConnMock) Read(b []byte) (n int, err error) {
	return f.read(b)
}

func (f FullNetConnMock) Close() error {
	return f.close()
}

func (f FullNetConnMock) LocalAddr() net.Addr {
	return f.localAddr()
}

func (f FullNetConnMock) RemoteAddr() net.Addr {
	return f.remoteAddr()
}

func (f FullNetConnMock) SetDeadline(t time.Time) error {
	return f.setDeadline(t)
}

func (f FullNetConnMock) SetReadDeadline(t time.Time) error {
	return f.setReadDeadline(t)
}

func (f FullNetConnMock) SetWriteDeadline(t time.Time) error {
	return f.setWriteDeadline(t)
}

func TestProcessCommands(t *testing.T) {
	s := &StoreMock{}
	c := cache.NewCache(400, s)
	myData := cache.Data{
		Value:     []byte("100"),
		Flags:     12,
		ExpTime:   0,
		ByteCount: 3,
	}
	c.Store.Save("foo", myData)
	cmd := parser.Command{
		Action: "get",
		Key:    "foo",
	}
	connMock := FullNetConnMock{}
	c.ProcessCommand(cmd, connMock)
}
