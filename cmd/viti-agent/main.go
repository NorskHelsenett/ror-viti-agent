package main

import "github.com/NorskHelsenett/ror-viti-agent/internal/config"

func main() {

	_, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

}
