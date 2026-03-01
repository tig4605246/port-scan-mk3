package main

import (
	"fmt"
	"os"

	"github.com/xuxiping/port-scan-mk3/pkg/config"
)

func main() {
	if _, err := config.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
