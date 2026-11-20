package main

import (
	"log"

	"github.com/waygu-music-plugin/sdk/desktop"
	"github.com/waygu-music-plugin/subsonic"
)

func main() {
	if err := desktop.RunPlatformPackage("subsonic", subsonic.Platforms()); err != nil {
		log.Fatal(err)
	}
}
