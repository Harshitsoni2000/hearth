package main

import (
	"fmt"

	"hearth/internal/config"
)

func main() {
	cfg := config.Parse()
	fmt.Printf("%+v\n", cfg)
}
