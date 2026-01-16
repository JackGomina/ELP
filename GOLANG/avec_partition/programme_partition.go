// programme_partition.go
package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"runtime"
	"sync"
)

type Job struct {
	yStart int // inclusive
	yEnd   int // exclusive
}

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
	outputPath := flag.String("out", "image_blur_partition.png", "output PNG file")
	k := flag.Int("k", 5, "kernel size (odd number >= 1), e.g. 3,5,9")
	workersFlag := flag.Int("workers", 0, "number of worker goroutines (0 = runtime.NumCPU())")
	chunk := flag.Int("chunk", 32, "number of rows per job chunk (e.g. 16, 32, 64)")
	flag.Parse()

	if *k < 1 || (*k)%2 == 0 {
		fmt.Fprintln(os.Stderr, "Error: -k must be an odd number >= 1 (e.g. 3, 5, 7, 9).")
		os.Exit(1)
	}
	if *chunk < 1 {
		fmt.Fprintln(os.Stderr, "Error: -chunk must be >= 1.")
		os.Exit(1)
	}

	workers := *workersFlag
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if workers < 1 {
		workers = 1
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
		os.Exit(1) //Ca va tuer la session : pas cool pour le serveur
	}

	// --- Convert to RGBA for fast pixel access ---
	bounds := srcImg.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	src := image.NewRGBA(bounds)
	draw.Draw(src, bounds, srcImg, bounds.Min, draw.Src)

	// --- Destination image (written by workers on disjoint rows) ---
	dst := image.NewRGBA(bounds)

	radius := (*k) / 2

	// --- Create jobs channel ---
	jobs := make(chan Job, workers*2)

	var wg sync.WaitGroup

	// --- Worker function ---
	workerFn := func() {
		defer wg.Done()

		for job := range jobs {
			// Process rows [job.yStart, job.yEnd]
			for y := job.yStart; y < job.yEnd; y++ {
				for x := 0; x < w; x++ {

					var sumR, sumG, sumB, sumA uint32
					var count uint32

					for dy := -radius; dy <= radius; dy++ {
						yy := clamp(y+dy, 0, h-1)
						for dx := -radius; dx <= radius; dx++ {
							xx := clamp(x+dx, 0, w-1)

							// Convert (xx,yy) to absolute coords in the RGBA buffer
							i := src.PixOffset(xx+bounds.Min.X, yy+bounds.Min.Y)
							sumR += uint32(src.Pix[i+0])
							sumG += uint32(src.Pix[i+1])
							sumB += uint32(src.Pix[i+2])
							sumA += uint32(src.Pix[i+3])
							count++
						}
					}

					di := dst.PixOffset(x+bounds.Min.X, y+bounds.Min.Y)
					dst.Pix[di+0] = uint8(sumR / count)
					dst.Pix[di+1] = uint8(sumG / count)
					dst.Pix[di+2] = uint8(sumB / count)
					dst.Pix[di+3] = uint8(sumA / count)
				}
			}
		}
	}

	// --- Start workers ---
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go workerFn()
	}

	// --- Send jobs (chunked by rows) ---
	for y := 0; y < h; y += *chunk {
		yEnd := y + *chunk
		if yEnd > h {
			yEnd = h
		}
		jobs <- Job{yStart: y, yEnd: yEnd}
	}

	// --- Close jobs channel so workers stop ---
	close(jobs)

	// --- Wait for all workers to finish ---
	wg.Wait()

	// --- Write output PNG ---
	outFile, err := os.Create(*outputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating output:", err)
		os.Exit(1) //Ca va tuer la session : pour le serveur cest pas dingue
	}
	defer outFile.Close()

	enc := png.Encoder{CompressionLevel: png.DefaultCompression}
	if err := enc.Encode(outFile, dst); err != nil {
		fmt.Fprintln(os.Stderr, "Error encoding PNG:", err)
		os.Exit(1)
	}

	fmt.Printf("Done: %s (workers=%d, chunk=%d, k=%d)\n", *outputPath, workers, *chunk, *k)
}
