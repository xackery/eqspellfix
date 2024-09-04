package main

import (
	"fmt"
	"os"

	"github.com/xackery/eqspellfix/fix"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	err = fix.Spell(path)
	if err != nil {
		return err
	}
	return nil
}
