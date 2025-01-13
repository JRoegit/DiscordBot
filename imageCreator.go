package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
)

func CreateCollageFromImages(ImagesURLs []string) string {
	var images []image.Image
	var minSize = image.Point{X: 230, Y: 100000}
	for _, URL := range ImagesURLs {
		response, err := http.Get(URL)
		if err != nil {
			fmt.Println(err)
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			fmt.Println(err)
		}

		img, format, err := image.Decode(response.Body)
		fmt.Println(format)
		if err != nil {
			fmt.Println(err)
		}
		images = append(images, img)

		fmt.Println(img.Bounds().String())
		if imgSize := img.Bounds().Size(); imgSize.Y < minSize.Y {
			minSize.Y = imgSize.Y
		}
	}
	newImageRect := image.Rect(0, 0, minSize.X*3, minSize.Y*3)
	newImage := image.NewRGBA64(newImageRect)
	fmt.Println(minSize.String())
	fmt.Println(len(images))
	row := 0
	column := 0
	for idx := 0; idx < 9; idx++ {
		img := images[idx]
		for i := 0; i < minSize.X; i++ {
			// for i := minSize.X * row; i < minSize.X * (row + 1); i++{
			x := i + (minSize.X * row)

			for j := 0; j < minSize.Y; j++ {
				// for j := minSize.Y * col; j < minSize.Y * (col + 1);j++{
				c := img.At(i, j)
				red, green, blue, alpha := c.RGBA()
				rgb := color.RGBA64{
					R: uint16(red),
					G: uint16(green),
					B: uint16(blue),
					A: uint16(alpha),
				}
				y := j + (minSize.Y * column)
				newImage.SetRGBA64(x, y, rgb)
			}
		}
		if row == 2 {
			row = 0
			column++
		} else {
			row++
		}
	}
	f, _ := os.Create("c3b3.jpg")
	jpeg.Encode(f, newImage, &jpeg.Options{Quality: 100})
	return ""
}

func DownloadImageFromURL() {
	URL := "https://s4.anilist.co/file/anilistcdn/user/avatar/large/b6302825-EYDJhL4yNDS2.jpg"

	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		fmt.Println(err)
	}

	img, _, err := image.Decode(response.Body)
	if err != nil {
		fmt.Println(err)
	}
	f, _ := os.Create("image2.png")
	png.Encode(f, img)
}

func CreateImage() {
	width := 200
	height := 100

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Colors are defined by Red, Green, Blue, Alpha uint8 values.
	cyan := color.RGBA{100, 200, 200, 0xff}

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			switch {
			case x < width/2 && y < height/2: // upper left quadrant
				img.Set(x, y, cyan)
			case x >= width/2 && y >= height/2: // lower right quadrant
				img.Set(x, y, color.White)
			default:
				// Use zero value.
			}
		}
	}

	// Encode as PNG.
	f, _ := os.Create("image.png")
	png.Encode(f, img)
}
