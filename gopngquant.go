package gopngquant

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	_ "image/png"
	"log"
	"os"

	"github.com/fish1/golibimagequant"
)

type gopngImage struct {
	data   []byte
	width  int
	height int
}

func CompressFile(input string, output string) error {
	image := imageFromFile(input)
	bytes, err := compress(image)
	if err != nil {
		return err
	}
	err = bytesToFile(output, bytes)
	if err != nil {
		return err
	}
	return nil
}

func CompressBytes(data []byte) ([]byte, error) {
	image := imageFromBytes(data)
	bytes, err := compress(image)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func compress(gopngimage gopngImage) ([]byte, error) {
	cattr := golibimagequant.CreateAttr()
	defer golibimagequant.DestroyAttr(cattr)
	cerr := golibimagequant.SetQuality(cattr, 0, 25)
	if cerr != 0 {
		return nil, errors.New(fmt.Sprintf("error setting quality. liq error: %d", cerr))
	}
	cerr = golibimagequant.SetSpeed(cattr, 1)
	if cerr != 0 {
		return nil, errors.New(fmt.Sprintf("error setting speed. liq error: %d", cerr))
	}
	cimage := golibimagequant.CreateImageRGBA(cattr, &gopngimage.data[0], gopngimage.width, gopngimage.height, 0)
	defer golibimagequant.DestroyImage(cimage)

	var cresult *golibimagequant.LiqResult
	cerr = golibimagequant.QuantizeImage(cattr, cimage, &cresult)
	defer golibimagequant.DestroyResult(cresult)
	if cerr != 0 {
		return nil, errors.New(fmt.Sprintf("error quantizing image. liq error: %d", cerr))
	}

	pixels := make([]byte, gopngimage.width*gopngimage.height)
	cerr = golibimagequant.WriteRemappedImage(cresult, cimage, &pixels[0], uint64(len(pixels)))
	if cerr != 0 {
		return nil, errors.New(fmt.Sprintf("error writing remapped image. liq error: %d", cerr))
	}

	cpalette := golibimagequant.GetPalette(cresult)

	rectangle := image.Rect(0, 0, gopngimage.width, gopngimage.height)
	palette := make(color.Palette, 0)
	for index := 0; index < int(cpalette.Count); index += 1 {
		entry := cpalette.Entries[index]
		palette = append(palette, golibimagequant.NewRGBA(entry))
	}

	image := image.NewPaletted(rectangle, palette)
	image.Pix = pixels

	buffer := new(bytes.Buffer)
	encoder := png.Encoder{
		CompressionLevel: png.BestCompression,
	}
	err := encoder.Encode(buffer, image)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func imageFromBytes(data []byte) gopngImage {
	image, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Fatalln(err)
	}
	size := image.Bounds().Size()
	width := size.X
	height := size.Y
	raw := make([]byte, width*height*4)

	for y := 0; y < height; y += 1 {
		for x := 0; x < width; x += 1 {
			index := (y + x) * 4
			r, g, b, a := image.At(x, y).RGBA()
			raw[index], raw[index+1], raw[index+2], raw[index+3] = uint8(r), uint8(g), uint8(b), uint8(a)
		}
	}

	return gopngImage{
		data:   raw,
		width:  width,
		height: height,
	}
}

func bytesToFile(filename string, data []byte) error {
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func imageFromFile(filename string) gopngImage {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	image, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}

	size := image.Bounds().Size()
	width := size.X
	height := size.Y
	raw := make([]byte, width*height*4)

	for y := 0; y < height; y += 1 {
		for x := 0; x < width; x += 1 {
			index := (y*width + x) * 4
			r, g, b, a := image.At(x, y).RGBA()
			raw[index], raw[index+1], raw[index+2], raw[index+3] = byte(r), byte(g), byte(b), byte(a)
		}
	}

	return gopngImage{
		data:   raw,
		width:  width,
		height: height,
	}
}
