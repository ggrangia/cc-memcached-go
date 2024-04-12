package parser_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ggrangia/cc-memcached-go/internal/parser"
)

func TestParser(t *testing.T) {
	data := []struct {
		name     string
		strCmd   string
		expected parser.Command
		errMsg   string
	}{
		{"empty str", "", parser.Command{}, ""},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			var errMsg string

			buff := bytes.NewBuffer([]byte(d.strCmd))
			got, err := parser.Parse(buff)
			if diff := cmp.Diff(d.expected, got); diff != "" {
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

func TestParseSet(t *testing.T) {
	data := []struct {
		name     string
		strCmd   string
		expected parser.Command
		errMsg   string
	}{
		{"ok", "set foo 0 1 4\r\n", parser.Command{Action: "set", Key: "foo", Flags: 0, Exptime: 1, ByteCount: 4, Noreply: false}, ""},
		{"short", "set \r\n", parser.Command{}, `incorrect number of elements for "set" action: 2`},
		{"long", "set foo 0 1 3 4 4 4\r\n", parser.Command{}, `incorrect number of elements for "set" action: 8`},
		{"noreply", "set foo 0 0 4 noreply\r\n", parser.Command{Action: "set", Key: "foo", Flags: 0, Exptime: 0, ByteCount: 4, Noreply: true}, ""},
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
		expected parser.Command
		errMsg   string
	}{
		{"ok", "get mykey\r\n", parser.Command{Action: "get", Key: "mykey"}, ""},
		{"short", "get\r\n", parser.Command{}, `incorrect number of elements for "get" action: 1`},
		{"long", "get mykey another one\r\n", parser.Command{}, `incorrect number of elements for "get" action: 4`},
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
