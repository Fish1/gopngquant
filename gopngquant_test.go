package gopngquant

import (
	"image/png"
	"os"
	"testing"
)

func Test_FileCompression(t *testing.T) {
	options := Options{
		Speed:         1,
		MinQuality:    0,
		TargetQuality: 25,
	}
	CompressFile("./images/example.png", "./images/output_gopngquant_file.png", options)
}

func Test_ImageCompression(t *testing.T) {
	file, err := os.Open("./images/example.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		panic(err)
	}

	img, err = CompressImage(img, Options{
		Speed:         1,
		MinQuality:    0,
		TargetQuality: 25,
	})
	if err != nil {
		panic(err)
	}

	file, err = os.Create("./images/output_gopngquant_byte.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	encoder := png.Encoder{
		CompressionLevel: png.BestCompression,
	}
	err = encoder.Encode(file, img)
	if err != nil {
		panic(err)
	}
}
