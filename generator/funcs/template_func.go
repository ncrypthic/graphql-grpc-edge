package funcs

import (
	"path/filepath"
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

func NormalizedFileName(s string) string {
	replaceStr := []string{".", "/", "_"}
	for _, c := range replaceStr {
		s = strings.ReplaceAll(s, c, "")
	}
	return strings.Title(s)
}

func GenerateImportAlias(s string) (string, string) {
	dir := filepath.Dir(s)
	filename := filepath.Base(s)
	ext := filepath.Ext(filename)
	suffix := strings.Title(filename[0 : len(filename)-len(ext)])
	return strings.ReplaceAll(dir, "/", "."), "pb" + suffix
}
