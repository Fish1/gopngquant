package gopngquant

import (
	Bytes "bytes"
	"errors"
	"fmt"
	Image "image"
	"image/color"
	"image/png"
	_ "image/png"
	"os"

	"github.com/fish1/golibimagequant"
)

type Options struct {
	Speed         int
	MinQuality    int
	TargetQuality int
}

func CompressImage(input Image.Image, options Options) (Image.Image, error) {
	if err := validateOptions(options); err != nil {
		return nil, err
	}
	bytes, width, height, err := bytesFromImage(input)
	if err != nil {
		return nil, err
	}
	bytes, err = compress(bytes, width, height, options)
	if err != nil {
		return nil, err
	}
	image, err := png.Decode(Bytes.NewReader(bytes))
	if err != nil {
		return nil, err
	}
	return image, nil
}

func CompressFile(input string, output string, options Options) error {
	if err := validateOptions(options); err != nil {
		return err
	}
	bytes, width, height, err := bytesFromFile(input)
	if err != nil {
		return err
	}
	bytes, err = compress(bytes, width, height, options)
	if err != nil {
		return err
	}
	err = bytesToFile(output, bytes)
	if err != nil {
		return err
	}
	return nil
}

func validateOptions(options Options) error {
	if options.Speed < 1 || options.Speed > 11 {
		return errors.New("speed must be between 1 and 11")
	}
	if options.TargetQuality < options.MinQuality {
		return errors.New("target quality must be greater than or equal to min quality")
	}
	if options.MinQuality < 0 || options.MinQuality > 100 {
		return errors.New("min quality must be greater than or equal to min quality")
	}
	if options.TargetQuality < 0 || options.TargetQuality > 100 {
		return errors.New("target quality must be greater than or equal to min quality")
	}
	return nil
}

func compress(input []byte, width int, height int, options Options) ([]byte, error) {
	cattr := golibimagequant.CreateAttr()
	defer golibimagequant.DestroyAttr(cattr)
	cerr := golibimagequant.SetQuality(cattr, options.MinQuality, options.TargetQuality)
	if cerr != 0 {
		return nil, errors.New(fmt.Sprintf("error setting quality. liq error: %d", cerr))
	}
	cerr = golibimagequant.SetSpeed(cattr, options.Speed)
	if cerr != 0 {
		return nil, errors.New(fmt.Sprintf("error setting speed. liq error: %d", cerr))
	}
	cimage := golibimagequant.CreateImageRGBA(cattr, &input[0], width, height, 0)
	defer golibimagequant.DestroyImage(cimage)

	var cresult *golibimagequant.LiqResult
	cerr = golibimagequant.QuantizeImage(cattr, cimage, &cresult)
	defer golibimagequant.DestroyResult(cresult)
	if cerr != 0 {
		return nil, errors.New(fmt.Sprintf("error quantizing image. liq error: %d", cerr))
	}

	newPixels := make([]byte, width*height)
	cerr = golibimagequant.WriteRemappedImage(cresult, cimage, &newPixels[0], uint64(len(newPixels)))
	if cerr != 0 {
		return nil, errors.New(fmt.Sprintf("error writing remapped image. liq error: %d", cerr))
	}

	cpalette := golibimagequant.GetPalette(cresult)

	rectangle := Image.Rect(0, 0, width, height)
	palette := make(color.Palette, 0)
	for index := 0; index < int(cpalette.Count); index += 1 {
		entry := cpalette.Entries[index]
		palette = append(palette, golibimagequant.NewRGBA(entry))
	}

	paletted := Image.NewPaletted(rectangle, palette)
	paletted.Pix = newPixels

	buffer := new(Bytes.Buffer)
	encoder := png.Encoder{
		CompressionLevel: png.BestCompression,
	}
	err := encoder.Encode(buffer, paletted)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
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

func bytesFromFile(filename string) ([]byte, int, int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, 0, 0, err
	}
	defer file.Close()
	image, err := png.Decode(file)
	if err != nil {
		return nil, 0, 0, err
	}
	return bytesFromImage(image)
}

func bytesFromImage(image Image.Image) ([]byte, int, int, error) {
	switch image := image.(type) {
	case *Image.NRGBA:
		return image.Pix, image.Rect.Dx(), image.Rect.Dy(), nil
	case *Image.RGBA:
		return image.Pix, image.Rect.Dx(), image.Rect.Dy(), nil
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
	return raw, width, height, nil
}
