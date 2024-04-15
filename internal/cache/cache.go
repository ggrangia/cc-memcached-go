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
	Store Store
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

func NewCache(port int, store Store) *Cache {
	return &Cache{
		port:  port,
		Store: store,
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
	const chunkSize = 4096
	var activeCmd parser.Command

	waitForData := false

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

		if waitForData {
			activeCmd, waitForData = c.handleMoreData(buffer, activeCmd, waitForData, conn)
		} else {
			activeCmd, waitForData = c.handleNewCommand(buffer, activeCmd, waitForData, conn)
		}
	}
}

func (c Cache) handleMoreData(buffer *bytes.Buffer, activeCmd parser.Command, waitForData bool, conn net.Conn) (parser.Command, bool) {
	// remove \r\n
	buffData := buffer.Bytes()[:buffer.Len()-2]
	fmt.Println(buffData, string(buffData), activeCmd.ByteCount)
	activeCmd.Data = append(activeCmd.Data, buffData...)
	if len(activeCmd.Data) >= activeCmd.ByteCount {
		waitForData = false
		activeCmd = c.ProcessCommand(activeCmd, conn)
	}
	return activeCmd, waitForData
}

func (c Cache) handleNewCommand(buffer *bytes.Buffer, activeCmd parser.Command, waitForData bool, conn net.Conn) (parser.Command, bool) {
	var err error

	activeCmd, err = parser.Parse(buffer)

	if err != nil {
		fmt.Println("Parse error:", err)
		conn.Write([]byte("ERROR\r\n"))
		// "reset" cmd and waitForData
		return parser.Command{}, false
	}

	if activeCmd.Complete {
		waitForData = false
		activeCmd = c.ProcessCommand(activeCmd, conn)
	} else {
		waitForData = true
	}
	return activeCmd, waitForData
}

func (c *Cache) sendMessage(conn net.Conn, message string) {
	if _, err := conn.Write([]byte(message)); err != nil {
		fmt.Println("Error writing to connection:", err)
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
	case "replace":
		c.ProcessReplace(conn, cmd)
	case "append":
		c.ProcessExtendData(conn, cmd, true)
	case "prepend":
		c.ProcessExtendData(conn, cmd, false)
	}
	return parser.Command{}
}

func (c *Cache) ProcessExtendData(conn net.Conn, cmd parser.Command, appendData bool) {
	if len(cmd.Data) > cmd.ByteCount {
		c.sendMessage(conn, "CLIENT_ERROR bad data chunk\r\n")
		return
	}

	data, exists, err := c.Store.Get(cmd.Key)

	if err != nil {
		fmt.Println("Error accessing data store:", err)
		c.sendMessage(conn, "ERROR\r\n")
		return
	}

	if !exists {
		c.sendMessage(conn, "NOT_STORED\r\n")
		return
	}

	if appendData {
		data.Value = append(data.Value, cmd.Data...)
	} else {
		data.Value = append(cmd.Data, data.Value...)
	}
	data.ByteCount += len(cmd.Data)
	data.ExpTime = calculateExpiryTime(cmd.Exptime)

	if err := c.Store.Save(cmd.Key, data); err != nil {
		fmt.Println("Error saving data:", err)
		c.sendMessage(conn, "ERROR\r\n")
		return
	}

	if !cmd.Noreply {
		c.sendMessage(conn, "STORED\r\n")
	}

}

func (c *Cache) ProcessReplace(conn net.Conn, cmd parser.Command) {
	if len(cmd.Data) > cmd.ByteCount {
		c.sendMessage(conn, "CLIENT_ERROR bad data chunk\r\n")
		return
	}
	_, exists, getErr := c.Store.Get(cmd.Key)

	if getErr != nil {
		fmt.Println("ProcessAdd: ", getErr.Error())
	}

	if !exists {
		c.sendMessage(conn, "NOT_STORED\r\n")
		return
	}

	data := Data{
		Value:     cmd.Data,
		ExpTime:   calculateExpiryTime(cmd.Exptime),
		Flags:     cmd.Flags,
		ByteCount: len(cmd.Data),
	}

	if err := c.Store.Save(cmd.Key, data); err != nil {
		fmt.Println("Error saving data:", err)
		c.sendMessage(conn, "ERROR\r\n")
		return
	}

	if !cmd.Noreply {
		c.sendMessage(conn, "STORED\r\n")
	}
}

func (c *Cache) ProcessAdd(conn net.Conn, cmd parser.Command) {
	if len(cmd.Data) > cmd.ByteCount {
		c.sendMessage(conn, "CLIENT_ERROR bad data chunk\r\n")
		return
	}

	_, exists, getErr := c.Store.Get(cmd.Key)

	if getErr != nil {
		fmt.Println("ProcessAdd: ", getErr.Error())
	}
	if exists {
		c.sendMessage(conn, "NOT_STORED\r\n")
		return
	}

	data := Data{
		Value:     cmd.Data,
		ExpTime:   calculateExpiryTime(cmd.Exptime),
		Flags:     cmd.Flags,
		ByteCount: len(cmd.Data),
	}

	if err := c.Store.Save(cmd.Key, data); err != nil {
		fmt.Println("Error saving data:", err)
		c.sendMessage(conn, "ERROR\r\n")
		return
	}

	if !cmd.Noreply {
		c.sendMessage(conn, "STORED\r\n")
	}
}

func (c *Cache) ProcessSet(conn net.Conn, cmd parser.Command) {
	if len(cmd.Data) > cmd.ByteCount {
		c.sendMessage(conn, "CLIENT_ERROR bad data chunk\r\n")
		return
	}

	data := Data{
		Value:     cmd.Data,
		ExpTime:   calculateExpiryTime(cmd.Exptime),
		Flags:     cmd.Flags,
		ByteCount: len(cmd.Data),
	}

	if err := c.Store.Save(cmd.Key, data); err != nil {
		fmt.Println("Error saving data:", err)
		c.sendMessage(conn, "ERROR\r\n")
		return
	}

	if !cmd.Noreply {
		c.sendMessage(conn, "END\r\n")
	}
}

func (c *Cache) ProcessGet(conn net.Conn, cmd parser.Command) {
	message := bytes.NewBuffer([]byte{})
	d, exist, getErr := c.Store.Get(cmd.Key)

	if getErr != nil {
		fmt.Println("Error reading data :", getErr.Error())
		return
	}

	if exist && d.IsExpired() {
		delErr := c.Store.Delete(cmd.Key)
		if delErr != nil {
			fmt.Printf("Error deleting key %s: %s\n", cmd.Key, delErr.Error())
		}
		exist = false
	}

	if exist {
		fmt.Println(string(d.Value))
		s := fmt.Sprintf("VALUE %s %d %d\n%s\n", cmd.Key, d.Flags, d.ByteCount, d.Value)
		message.Write([]byte(s))
	}
	message.Write([]byte("END\r\n"))

	_, err := conn.Write(message.Bytes())
	if err != nil {
		fmt.Println("Error writing: ", err.Error())
	}
}

func calculateExpiryTime(exptime int) int {
	if exptime == 0 {
		return 0
	}
	return int(time.Now().Unix()) + exptime
}
