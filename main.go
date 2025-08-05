package main

import (
	"TinyGW/boot"
	"log"
)

func main() {
	if err := boot.Run(); err != nil {
		log.Fatal(err)
	}
}
