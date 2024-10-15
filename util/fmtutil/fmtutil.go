package fmtutil

import (
	"fmt"
	"os"
)

func Eprintf(format string, a ...any) {
	_, err := fmt.Fprintf(os.Stderr, format, a...)
	if err != nil {
		panic(err)
	}
}
