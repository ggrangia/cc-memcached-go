package parser

type Command struct {
	Action    string
	Data      []byte
	Key       string
	Exptime   int
	Flags     int
	ByteCount int
	Noreply   bool
	Complete  bool
}
