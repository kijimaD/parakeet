// Package main is the entry point of the application.
package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if _, err := fmt.Fprintln(os.Stdout, "Hello world"); err != nil {
		log.Fatal(err)
	}
}
