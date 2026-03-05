package editor

import (
	"fmt"
	"strconv"
)

var headingANSI256 = [6]int{228, 39, 39, 39, 39, 35}

const (
	linkLabelColor256 = 78
	linkDestColor256  = 44
)

func headingColor256(level int) (int, bool) {
	if level < 1 || level > len(headingANSI256) {
		return 0, false
	}
	return headingANSI256[level-1], true
}

func headingColorCode(level int) (string, bool) {
	v, ok := headingColor256(level)
	if !ok {
		return "", false
	}
	return strconv.Itoa(v), true
}

func headingAnsiFg(level int) (string, bool) {
	v, ok := headingColor256(level)
	if !ok {
		return "", false
	}
	return ansiFg256(v), true
}

func ansiFg256(v int) string {
	return fmt.Sprintf("\x1b[38;5;%dm", v)
}

func linkLabelAnsiFg() string {
	return ansiFg256(linkLabelColor256)
}

func linkDestAnsiFg() string {
	return ansiFg256(linkDestColor256)
}

func linkLabelColorCode() string {
	return strconv.Itoa(linkLabelColor256)
}

func linkDestColorCode() string {
	return strconv.Itoa(linkDestColor256)
}
