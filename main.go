package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) >= 3 && os.Args[1] == "store" {
		if err := RunStore(os.Args[2]); err != nil {
			log.Printf("store error: %v", err)
		}
		return
	}
	RunMCPServer()
}
