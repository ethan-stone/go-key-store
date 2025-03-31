package main

import (
	"log"

	client "github.com/ethan-stone/go-key-store/internal/cli"
)

func main() {
	log.Default().SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)

	client.Execute()
}
