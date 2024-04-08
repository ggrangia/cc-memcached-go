package parser

type Action string

const (
	ActionSet = "set"
	ActionGet = "get"
)

type Command struct {
	Action    Action
	Key       string
	Exptime   int
	Flags     int
	ByteCount int
	Noreply   bool
}

type GetCommand struct {
	Action Action
	Key    string
}
