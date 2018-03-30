package util

import (
	"errors"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

func IsExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well
	return isExported(t.Name()) || t.PkgPath() == ""
}

func IsContextType(t reflect.Type) bool {
	return true
}

func ExtractServiceAndMethod(ServiceMethod string) (serviceName string, methodName string, err error) {
	var sep = "."
	dot := strings.LastIndex(ServiceMethod, sep)
	if dot < 0 {
		err = errors.New("rpc: service/Method request ill-formed: " + ServiceMethod)
		return
	}
	serviceName = ServiceMethod[:dot]
	methodName = ServiceMethod[dot+1:]

	return
}
