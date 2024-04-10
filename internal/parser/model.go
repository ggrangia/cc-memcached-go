package parser

type Action string

const (
	ActionSet = "set"
	ActionGet = "get"
)

type Command struct {
	Action    Action
	Data      []byte
	Key       string
	Exptime   int
	Flags     int
	ByteCount int
	Noreply   bool
	Complete  bool
}

type GetCommand struct {
	Action Action
	Key    string
}
