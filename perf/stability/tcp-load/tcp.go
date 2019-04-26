package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

var addr = os.Getenv("ADDRESS")

func main() {
	load()
}

func load() {
	conns := 0
	for {
		_, err := net.Dial("tcp", addr)
		fmt.Printf("Connected [%v]: %v\n", conns, err)
		if err == nil {
			conns++
		} else {
			time.Sleep(time.Second)
		}
	}
}
