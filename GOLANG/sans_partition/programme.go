package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"time"
)

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func main() {
	inputPath := flag.String("in", "image.png", "input PNG file")
	outputPath := flag.String("out", "image_blur.png", "output PNG file")
	k := flag.Int("k", 5, "kernel size (odd number >= 1), e.g. 3,5,9")
	flag.Parse()

	if *k < 1 || (*k)%2 == 0 {
		fmt.Fprintln(os.Stderr, "Error: -k must be an odd number >= 1 (e.g. 3, 5, 7, 9).")
		os.Exit(1)
	}

	// --- Open input file ---
	inFile, err := os.Open(*inputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error opening input:", err)
		os.Exit(1)
	}
	defer inFile.Close()

	// --- Decode PNG ---
	srcImg, err := png.Decode(inFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error decoding PNG:", err)
		os.Exit(1)
	}

	// --- Convert to RGBA for easy pixel access ---
	bounds := srcImg.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	src := image.NewRGBA(bounds)
	draw.Draw(src, bounds, srcImg, bounds.Min, draw.Src)

	// --- Prepare destination image ---
	dst := image.NewRGBA(bounds)

	radius := (*k) / 2

	// --- Box blur (mean filter) ---
	// For each pixel (x,y), average all pixels in the square neighborhood:
	// x in [x-radius, x+radius], y in [y-radius, y+radius]
	// Border handling: clamp coordinates into the image.
	startTime := time.Now()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {

			var sumR, sumG, sumB, sumA uint32
			var count uint32

			for dy := -radius; dy <= radius; dy++ {
				yy := clamp(y+dy, 0, h-1)
				for dx := -radius; dx <= radius; dx++ {
					xx := clamp(x+dx, 0, w-1)

					i := src.PixOffset(xx+bounds.Min.X, yy+bounds.Min.Y)
					// src.Pix stores bytes in RGBA order
					sumR += uint32(src.Pix[i+0])
					sumG += uint32(src.Pix[i+1])
					sumB += uint32(src.Pix[i+2])
					sumA += uint32(src.Pix[i+3])
					count++
				}
			}

			// Average and write to dst
			di := dst.PixOffset(x+bounds.Min.X, y+bounds.Min.Y)
			dst.Pix[di+0] = uint8(sumR / count)
			dst.Pix[di+1] = uint8(sumG / count)
			dst.Pix[di+2] = uint8(sumB / count)
			dst.Pix[di+3] = uint8(sumA / count)
		}
	}
	elapsedTime := time.Since(startTime)
	fmt.Printf("Temps de traitement de l'image: %v\n", elapsedTime)

	// --- Write output PNG ---
	outFile, err := os.Create(*outputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating output:", err)
		os.Exit(1)
	}
	defer outFile.Close()

	enc := png.Encoder{CompressionLevel: png.DefaultCompression}
	if err := enc.Encode(outFile, dst); err != nil {
		fmt.Fprintln(os.Stderr, "Error encoding PNG:", err)
		os.Exit(1)
	}

	fmt.Println("Done:", *outputPath)
}
