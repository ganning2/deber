package log

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
)

const (
	blue   = "\033[0;34m"
	red    = "\033[0;31m"
	normal = "\033[0m"
)

var (
	NoColor         = false
	ExtraInfoIndent = "  "
	newLine         = true
)

func Drop() {
	newLine = true
	fmt.Println()
}

func None() error {
	return nil
}

func Failed(err error) error {
	if !newLine {
		fmt.Printf("failed\n")
	}

	return err
}

func Done() error {
	if !newLine {
		fmt.Printf("done\n")
	}

	return nil
}

func Skipped() error {
	if !newLine {
		fmt.Printf("skipped\n")
	}

	return nil
}

func ExtraInfo(v interface{}) {
	fmt.Printf("%s%s ...", ExtraInfoIndent, v)

	newLine = false
}

func Info(v interface{}) {
	if !NoColor {
		fmt.Printf("%s%s:info:%s %s ...", blue, app.Name, normal, v)
	} else {
		fmt.Printf("%s:info: %s ...", app.Name, v)
	}

	newLine = false
}

func Error(v interface{}) {
	if !NoColor {
		fmt.Printf("%s%s:error:%s %s\n", red, app.Name, normal, v)
	} else {
		fmt.Printf("%s:error: %s\n", app.Name, v)
	}
}
