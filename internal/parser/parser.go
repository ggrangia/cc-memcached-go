package parser

import (
	"bytes"
	"fmt"
	"strconv"
)

func Parse(buffer *bytes.Buffer) [][]byte {

	splitted := bytes.Split(buffer.Bytes(), []byte(" "))
	// Remove empty arrays (multiple spaces in the command)
	cmdParts := make([][]byte, 0)
	for _, v := range splitted {
		if len(v) > 0 {
			cmdParts = append(cmdParts, v)
		}
	}
	fmt.Println(cmdParts)
	return cmdParts
}

/*
	// The first block is the action (set, get)
	action := string(cmdParts[0])
	fmt.Println(action)

	switch action {
	case "set":
		ParseSet(cmdParts)
	case "get":
		ParseGet(cmdParts)
	default:
		return Command{}, fmt.Errorf("invalid action: %s", action)
	}

	return Command{}, errors.New("TODO: complete the function")
*/

func ParseSet(actionParams [][]byte) (Command, error) {

	actionsLength := len(actionParams)
	fmt.Println(actionsLength)
	// <command name> <key> <flags> <exptime> <byte count> [noreply]\r\n
	if actionsLength > 6 || actionsLength < 5 {
		return Command{}, fmt.Errorf("incorrect number of elements for \"set\" action: %d", actionsLength)
	}

	key := string(actionParams[1])
	flags, err := strconv.Atoi(string(actionParams[2]))
	if err != nil {
		return Command{}, fmt.Errorf("flags conversion error: %s", err.Error())
	}
	exptime, err := strconv.Atoi(string(actionParams[3]))
	if err != nil {
		return Command{}, fmt.Errorf("exptime conversion error: %s", err.Error())
	}
	byteCount, err := strconv.Atoi(string(actionParams[4]))
	if err != nil {
		return Command{}, fmt.Errorf("bytecount conversion error: %s", err.Error())
	}

	noreply := false
	if actionsLength == 6 && string(actionParams[5]) == "noreply" {
		noreply = true
	}

	fmt.Println(key, flags, exptime, byteCount, noreply)

	return Command{
		Action:    "set",
		Key:       key,
		Flags:     flags,
		Exptime:   exptime,
		ByteCount: byteCount,
		Noreply:   noreply,
	}, nil
}

func ParseGet(actionParams [][]byte) (GetCommand, error) {
	// <command name> <key>\r\n
	actionsLength := len(actionParams)
	if actionsLength != 2 {
		return GetCommand{}, fmt.Errorf(`incorrect number of elements for "get" action: %d`, actionsLength)
	}

	key := string(actionParams[1])

	fmt.Println(key)

	return GetCommand{Action: "get", Key: key}, nil
}
