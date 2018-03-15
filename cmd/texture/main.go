package main

import (
	"image/png"
	"log"
	"os"

	"github.com/jeffbaumes/gogame/pkg/client"
)

func main() {
	v := client.LoadTextures()
	f, err := os.Create("textures.png")
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, v); err != nil {
		f.Close()
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
