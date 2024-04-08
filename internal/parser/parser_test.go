package parser_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ggrangia/cc-memcached-go/internal/parser"
)

func TestParseSet(t *testing.T) {
	data := []struct {
		name     string
		strCmd   string
		expected parser.Command
		errMsg   string
	}{
		{"ok", "set foo 0 1 4", parser.Command{Action: "set", Key: "foo", Flags: 0, Exptime: 1, ByteCount: 4, Noreply: false}, ""},
		{"short", "set ", parser.Command{}, `incorrect number of elements for "set" action: 2`},
		{"long", "set foo 0 1 3 4 4 4", parser.Command{}, `incorrect number of elements for "set" action: 8`},
		{"noreply", "set foo 0 0 4 noreply", parser.Command{Action: "set", Key: "foo", Flags: 0, Exptime: 0, ByteCount: 4, Noreply: true}, ""},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			var errMsg string

			buffer := bytes.NewBufferString(d.strCmd)
			buffList := bytes.Split(buffer.Bytes(), []byte(" "))
			val, err := parser.ParseSet(buffList)
			if diff := cmp.Diff(d.expected, val); diff != "" {
				t.Errorf(diff)
			}
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != d.errMsg {
				t.Errorf("Expected error message `%s` got `%s`", d.errMsg, errMsg)
			}
		})
	}
}

func TestParseGet(t *testing.T) {
	data := []struct {
		name     string
		strCmd   string
		expected parser.GetCommand
		errMsg   string
	}{
		{"ok", "get mykey", parser.GetCommand{Action: "get", Key: "mykey"}, ""},
		{"short", "get", parser.GetCommand{}, `incorrect number of elements for "get" action: 1`},
		{"long", "get mykey another one", parser.GetCommand{}, `incorrect number of elements for "get" action: 4`},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {

			var errMsg string

			buffer := bytes.NewBufferString(d.strCmd)
			buffList := bytes.Split(buffer.Bytes(), []byte(" "))
			val, err := parser.ParseGet(buffList)
			if diff := cmp.Diff(d.expected, val); diff != "" {
				t.Errorf(diff)
			}
			if err != nil {
				errMsg = err.Error()
			}
			if d.errMsg != errMsg {
				t.Errorf("Expected error message `%s` got `%s`", d.errMsg, errMsg)
			}
		})

	}

}
