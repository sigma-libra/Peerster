package helper

import (
	"strconv"
	. "strings"
)

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func SeparateIPAndPort(address string) (string, string) {
	elems := Split(address, ":")
	if(len(elems) != 2) {
		println("Helper function Error: expected 2 parts in address, got " + strconv.Itoa(len(elems)) + " instead")
		return address, ""
	}

	return elems[0], elems[1]
}

func Equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}