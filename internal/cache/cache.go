package cache

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/ggrangia/cc-memcached-go/internal/parser"
)

type Cache struct {
	port  int
	Store map[string]Data
}

type Data struct {
	Value     []byte
	Flags     int
	ExpTime   int
	ByteCount int
}

func (c *Cache) Start() {
	listenSoc := &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: c.port,
	}
	tcpListener, err := net.ListenTCP("tcp", listenSoc)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("Listening on port", c.port)
	defer tcpListener.Close()

	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			fmt.Println("Error accepting connections: ", err.Error())
			os.Exit(1)
		}
		go c.handleRequest(conn)
	}
}

func NewCache(port int) *Cache {
	return &Cache{
		port:  port,
		Store: make(map[string]Data, 50),
	}
}

func (c Cache) handleRequest(conn net.Conn) {
	chunkSize := 4096

	// listen for multiple messages loop
	for {
		buffer := bytes.NewBuffer(nil)
		dataSize := 0
		// Read data in chucks
		for {
			chunk := make([]byte, chunkSize)
			read, err := conn.Read(chunk)
			if err != nil {
				// Check for EOF
				if err == io.EOF {
					fmt.Println("Client closed the connection")
				} else {
					fmt.Println("Error reading: ", err.Error())
				}
				break
			}
			buffer.Write(chunk[:read])
			dataSize += read
			if read == 0 || read < chunkSize {
				break
			}
		}

		//strCmd := buffer.String()
		fmt.Println("got: ", buffer.Bytes())
		cmd, err := parser.Parse(buffer)
		if err != nil {
			fmt.Println(err.Error())
			continue // keep listening
		}
		//fmt.Println("%v", cmd)
		c.ProcessCommand(cmd, conn)
	}
}

func (c Cache) ProcessCommand(cmd parser.Command, conn net.Conn) {
	var message []byte

	if cmd.Action == "get" {
		d, exist := c.Store[cmd.Key]
		if exist {
			fmt.Println(string(d.Value))
			s := fmt.Sprintf("VALUE %s %d %d\r\n", d.Value, d.Flags, d.ByteCount)
			message = []byte(s)
		} else {
			message = []byte("END\r\n")
		}
		_, err := conn.Write(message)
		if err != nil {
			fmt.Println("Error writing: ", err.Error())
		}
	}
}
