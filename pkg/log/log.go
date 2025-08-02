package log

import (
	"fmt"
	"os"
)

func Warningln(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
}

func Warningf(msg string, a ...any) {
	fmt.Fprintf(os.Stderr, "[WARN] "+msg, a)
}
