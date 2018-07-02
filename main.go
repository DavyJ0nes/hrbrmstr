package main

import (
	"fmt"
	"os"

	"github.com/davyj0nes/docker-image-search/imageinfo"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please Supply an Image Name")
		os.Exit(1)
	}
	imageRepo := os.Args[1]

	wantImage, err := imageinfo.NewImageInfo(imageRepo)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	fmt.Println(wantImage)
}
