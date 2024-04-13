package cache

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/ggrangia/cc-memcached-go/internal/parser"
)

type Cache struct {
	port  int
	Store map[string]Data
}

type Store interface {
	Get() (Data, error)
	Set() (Data, error)
	Delete(error)
}

type Data struct {
	Value     []byte
	Flags     int
	ExpTime   int
	ByteCount int
}

func (d Data) IsExpired() bool {
	return d.ExpTime > 0 && d.ExpTime < int(time.Now().Unix())
}

func NewCache(port int) *Cache {
	return &Cache{
		port:  port,
		Store: make(map[string]Data, 50),
	}
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

func (c *Cache) ReadChunks(conn net.Conn, buffer *bytes.Buffer, chunkSize int, dataSize int) (int, error) {
	for {
		chunk := make([]byte, chunkSize)
		read, err := conn.Read(chunk)
		if err != nil {
			var errMsg string
			// Check for EOF
			if err == io.EOF {
				errMsg = "Client closed the connection"
			} else {
				errMsg = fmt.Sprintf("Error reading: %s", err.Error())
			}
			return -1, fmt.Errorf("%s", errMsg)
		}
		buffer.Write(chunk[:read])
		dataSize += read
		if read == 0 || read < chunkSize {
			return dataSize, nil
		}
	}
}

func (c Cache) handleRequest(conn net.Conn) {
	chunkSize := 4096
	var activeCmd parser.Command
	var waitForData bool
	var err error

	defer conn.Close()
	// listen for multiple messages loop
	for {
		buffer := bytes.NewBuffer(nil)
		dataSize := 0
		// Read data in chucks
		_, readErr := c.ReadChunks(conn, buffer, chunkSize, dataSize)
		if readErr != nil {
			fmt.Println(readErr.Error())
			break
		}

		fmt.Println("got: ", buffer.Bytes())
		if waitForData {
			// parse data
			// remove \r\n
			buffData := buffer.Bytes()[:buffer.Len()-2]
			fmt.Println(buffData, string(buffData), activeCmd.ByteCount)
			activeCmd.Data = append(activeCmd.Data, buffData...)
			if len(activeCmd.Data) >= activeCmd.ByteCount {
				waitForData = false
				activeCmd = c.ProcessCommand(activeCmd, conn)
			}
		} else {
			activeCmd, err = parser.Parse(buffer)
			if err != nil {
				fmt.Println(err.Error())
				conn.Write([]byte("ERROR\r\n"))
			} else if activeCmd.Complete {
				waitForData = false
				activeCmd = c.ProcessCommand(activeCmd, conn)
			} else {
				waitForData = true
			}
		}
	}
}

func (c Cache) ProcessCommand(cmd parser.Command, conn net.Conn) parser.Command {
	switch cmd.Action {
	case "get":
		c.ProcessGet(conn, cmd)
	case "set":
		c.ProcessSet(conn, cmd)
	case "add":
		c.ProcessAdd(conn, cmd)
	}
	return parser.Command{}
}

func (c *Cache) ProcessAdd(conn net.Conn, cmd parser.Command) {
	if len(cmd.Data) > cmd.ByteCount {
		conn.Write([]byte("CLIENT_ERROR bad data chunk\r\n"))
		return
	}
	if _, exists := c.Store[cmd.Key]; exists {
		conn.Write([]byte("NOT_STORED\r\n"))
		return
	}
	c.Store[cmd.Key] = Data{
		Value:     cmd.Data,
		ExpTime:   int(time.Now().Unix()) + cmd.Exptime,
		Flags:     cmd.Flags,
		ByteCount: len(cmd.Data),
	}
	if !cmd.Noreply {
		_, err := conn.Write([]byte("STORED\r\n"))
		if err != nil {
			fmt.Println("Error writing: ", err.Error())
		}
	}
}

func (c *Cache) ProcessSet(conn net.Conn, cmd parser.Command) {
	var exptime int

	if len(cmd.Data) > cmd.ByteCount {
		conn.Write([]byte("CLIENT_ERROR bad data chunk\r\n"))
		return
	}

	if cmd.Exptime == 0 {
		exptime = 0
	} else {
		exptime = int(time.Now().Unix()) + cmd.Exptime
	}

	c.Store[cmd.Key] = Data{
		Value:     cmd.Data,
		ExpTime:   exptime,
		Flags:     cmd.Flags,
		ByteCount: len(cmd.Data),
	}
	if !cmd.Noreply {
		_, err := conn.Write([]byte("END\r\n"))
		if err != nil {
			fmt.Println("Error writing: ", err.Error())
		}
	}
}

func (c *Cache) ProcessGet(conn net.Conn, cmd parser.Command) {
	var message []byte
	d, exist := c.Store[cmd.Key]
	if exist && d.IsExpired() {
		delete(c.Store, cmd.Key)
		exist = false
	}
	if exist {
		fmt.Println(string(d.Value))
		s := fmt.Sprintf("VALUE %s %d %d\n%s\n", cmd.Key, d.Flags, d.ByteCount, d.Value)
		message = []byte(s)
	} else {
		message = []byte("END\r\n")
	}
	_, err := conn.Write(message)
	if err != nil {
		fmt.Println("Error writing: ", err.Error())
	}
}
