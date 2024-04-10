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
	var activeCmd parser.Command
	var waitForData bool
	var err error
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

		fmt.Println("got: ", buffer.Bytes())
		// get command is complete after 1 line
		// set cmd parse first line and keep listening for incoming data
		// TODO: handle multiple messages as long as they are < bytecount
		if waitForData {
			// parse data
			// remove \r\n
			buffData := buffer.Bytes()[:buffer.Len()-2]
			fmt.Println(buffData, string(buffData), activeCmd.ByteCount)
			if len(buffData) > activeCmd.ByteCount {
				conn.Write([]byte("CLIENT_ERROR bad data chunk\r\n"))
			} else {
				activeCmd.Data = buffData
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
				//fmt.Println("%v", cmd)
			} else {
				waitForData = true
			}
		}
	}
}

func (c Cache) ProcessCommand(cmd parser.Command, conn net.Conn) parser.Command {
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
	} else if cmd.Action == "set" {
		c.Store[cmd.Key] = Data{
			Value:     cmd.Data,
			ExpTime:   cmd.Exptime,
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
	return parser.Command{}
}
