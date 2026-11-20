package mqtt

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"time"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func loadImage(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}

	img = addTimestamp(img)

	resized := resize(img, 400, 300)

	buffer := make([]byte, 400*300/8)

	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			r, g, b, _ := resized.At(x, y).RGBA()

			// Gleiche Formel wie PIL convert("1")
			gray := (r>>8)*299/1000 + (g>>8)*587/1000 + (b>>8)*114/1000

			idx := (y*400 + x) / 8
			bit := uint(7 - ((y*400 + x) % 8)) // MSB first = np.packbits

			if gray > 127 { // hell = weiß = 1
				buffer[idx] |= 1 << bit
			}
		}
	}

	return buffer, nil
}

func resize(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

func addTimestamp(img image.Image) image.Image {
	// In bearbeitbares Bild umwandeln
	bounds := img.Bounds()
	dst := image.NewRGBA(bounds)

	// Original kopieren
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dst.Set(x, y, img.At(x, y))
		}
	}

	// Timestamp Text
	timestamp := time.Now().Format("02.01.2006 15:04:05")

	// Text schreiben
	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(color.Black), // Textfarbe
		Face: basicfont.Face7x13,
		Dot:  fixed.P(5, bounds.Min.Y-5), // Position (x, y) – unten links
	}
	d.DrawString(timestamp)

	return dst
}
