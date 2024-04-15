package parser

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func Parse(buffer *bytes.Buffer) (Command, error) {
	var err error = nil
	var command Command = Command{}

	splitted := bytes.Split(buffer.Bytes(), []byte(" "))
	// Remove empty arrays (multiple spaces in the command)
	cmdParts := make([][]byte, 0)
	for _, v := range splitted {
		if len(v) > 0 {
			cmdParts = append(cmdParts, v)
		}
	}
	if len(cmdParts) == 0 {
		return Command{}, fmt.Errorf("Empty command")
	}

	action := strings.TrimSpace(string(cmdParts[0]))

	switch action {
	case "set", "add", "replace", "append", "prepend":
		command, err = ParseCommandAction(action, cmdParts)
	case "get":
		command, err = ParseGet(cmdParts)
	default:
		err = fmt.Errorf("invalid action: %s", action)
	}
	return command, err
}

func ParseCommandAction(action string, actionParams [][]byte) (Command, error) {

	actionsLength := len(actionParams)
	// <command name> <key> <flags> <exptime> <byte count> [noreply]\r\n
	// <data block>\r\n
	if actionsLength > 6 || actionsLength < 5 {
		return Command{}, fmt.Errorf("incorrect number of elements for \"%s\" action: %d", action, actionsLength)
	}

	key := strings.TrimSpace(string(actionParams[1]))
	flags, err := strconv.Atoi(strings.TrimSpace(string(actionParams[2])))
	if err != nil {
		return Command{}, fmt.Errorf("flags conversion error: %s", err.Error())
	}
	exptime, err := strconv.Atoi(strings.TrimSpace(string(actionParams[3])))
	if err != nil {
		return Command{}, fmt.Errorf("exptime conversion error: %s", err.Error())
	}
	byteCount, err := strconv.Atoi(strings.TrimSpace(string(actionParams[4])))
	if err != nil {
		return Command{}, fmt.Errorf("bytecount conversion error: %s", err.Error())
	}

	noreply := false
	if actionsLength == 6 && strings.TrimSpace(string(actionParams[5])) == "noreply" {
		noreply = true
	}

	return Command{
		Action:    action,
		Key:       key,
		Flags:     flags,
		Exptime:   exptime,
		ByteCount: byteCount,
		Noreply:   noreply,
	}, nil
}

func ParseGet(actionParams [][]byte) (Command, error) {
	// <command name> <key>\r\n
	actionsLength := len(actionParams)
	if actionsLength != 2 {
		return Command{}, fmt.Errorf(`incorrect number of elements for "get" action: %d`, actionsLength)
	}

	key := strings.TrimSpace(string(actionParams[1]))

	return Command{Action: "get", Key: key, Complete: true}, nil
}
