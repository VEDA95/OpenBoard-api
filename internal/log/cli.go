package log

import (
	"fmt"
	"os"
)

func PrintErrorString(err string) {
	fmt.Println(err)
	os.Exit(1)
}

func PrintError(err error) {
	PrintErrorString(err.Error())
}
