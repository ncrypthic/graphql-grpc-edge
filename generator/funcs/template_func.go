package funcs

import (
	"strings"

	"github.com/yoheimuta/go-protoparser/interpret/unordered"
)

func LCFirst(str string) string {
	head := str[0:1]
	tail := ""
	if len(str) > 1 {
		tail = str[1:len(str)]
	}
	return strings.ToLower(head) + tail
}

func UCFirst(str string) string {
	head := str[0:1]
	tail := ""
	if len(str) > 1 {
		tail = str[1:len(str)]
	}
	return strings.ToUpper(head) + tail
}

func Concat(str ...string) string {
	c := ""
	for _, s := range str {
		c = c + s
	}
	return c
}

func Title(str string) string {
	return strings.Title(str)
}

func LookUpMessage(name string, b *unordered.ProtoBody) *unordered.Message {
	for _, m := range b.Messages {
		if m.MessageName == name {
			return m
		}
	}
	return nil
}

func LookUpEnum(name string, b *unordered.ProtoBody) []*unordered.Enum {
	return nil
}
