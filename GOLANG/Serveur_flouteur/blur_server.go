package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"net"
	"runtime"
	"sync"
)

// Protocol: client envoie uint64 BE length, puis PNG bytes. Le serveur renvoie uint64 BE length, puis PNG bytes.

type Job struct {
	yStart int
	yEnd   int
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

// blurImage applies a box blur with kernel size k, using a worker pool of 'workers' goroutines
// and job chunks of 'chunk' rows. It returns an *image.RGBA with the same bounds as src.
func blurImage(srcImg image.Image, k, workers, chunk int) *image.RGBA {
	if k < 1 || k%2 == 0 {
		k = 5
	}
	if chunk < 1 {
		chunk = 32
	}
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if workers < 1 {
		workers = 1
	}

	bounds := srcImg.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	src := image.NewRGBA(bounds)
	draw.Draw(src, bounds, srcImg, bounds.Min, draw.Src)

	dst := image.NewRGBA(bounds)

	radius := k / 2

	jobs := make(chan Job, workers*2)
	var wg sync.WaitGroup

	workerFn := func() {
		defer wg.Done()
		for job := range jobs {
			for y := job.yStart; y < job.yEnd; y++ {
				for x := 0; x < w; x++ {
					var sumR, sumG, sumB, sumA uint32
					var count uint32
					for dy := -radius; dy <= radius; dy++ {
						yy := clamp(y+dy, 0, h-1)
						for dx := -radius; dx <= radius; dx++ {
							xx := clamp(x+dx, 0, w-1)
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

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go workerFn()
	}

	for y := 0; y < h; y += chunk {
		yEnd := y + chunk
		if yEnd > h {
			yEnd = h
		}
		jobs <- Job{yStart: y, yEnd: yEnd}
	}
	close(jobs)
	wg.Wait()

	return dst
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Nouveau client: %s", conn.RemoteAddr())

	reader := bufio.NewReader(conn)

	for {
		// Read 8-byte length header
		var lenBuf [8]byte
		if _, err := io.ReadFull(reader, lenBuf[:]); err != nil { // Lit exactement 8 bytes : la taille de l'image
			if err == io.EOF { // Cette erreur signifie que le client a fermé la connexion
				log.Printf("Client déconnecté: %s", conn.RemoteAddr())
			} else { // Sinon, c'est une autre erreur
				log.Printf("Erreur lecture header: %v", err)
			}
			return // Ducoup on disconnecte le client (print(Client disconnected))
		}
		imgLen := binary.BigEndian.Uint64(lenBuf[:]) // Slicing du header converti en uint64 qui représente la taille atttendue de l'image
		if imgLen == 0 {
			// ignore empty
			continue
		}

		// Read image bytes
		imgBytes := make([]byte, imgLen)                         // Allocation d'un buffer de la taille de l'image
		if _, err := io.ReadFull(reader, imgBytes); err != nil { // Lecture complète de l'image dans le buffer grâce au buffer de taille Imglen en bytes
			log.Printf("Erreur lecture image (%d bytes): %v", imgLen, err) // Si erreur on log et on disconnecte le client
			return
		}

		// Debug: log length and first bytes to help diagnose invalid PNGs
		log.Printf("Received %d bytes from %s", len(imgBytes), conn.RemoteAddr())
		if len(imgBytes) >= 8 {
			log.Printf("Signature: % x", imgBytes[:8])
			pngSig := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
			if bytes.Equal(imgBytes[:8], pngSig) {
				log.Printf("Signature looks like PNG")
			} else {
				log.Printf("Signature does NOT match PNG")
			}
		} else {
			log.Printf("Received fewer than 8 bytes for signature: %d", len(imgBytes))
		}

		// Decode PNG
		srcImg, err := png.Decode(bytes.NewReader(imgBytes)) // Décodage de l'image PNG à partir des bytes reçus
		if err != nil {
			log.Printf("Erreur décodage PNG: %v", err) // Encore un traitement d'erreur yen a bcp
			return
		}

		// Use defaults; these could be negotiated per-client if desired
		k := 5
		chunk := 32
		workers := 0 // => runtime.NumCPU()

		// Blur using a worker pool dedicated to this client
		dst := blurImage(srcImg, k, workers, chunk)

		// Encode result to PNG
		var outBuf bytes.Buffer                                      // Buffer pour stocker l'image encodée
		enc := png.Encoder{CompressionLevel: png.DefaultCompression} // Création d'un encodeur PNG
		if err := enc.Encode(&outBuf, dst); err != nil {             // Encodage de l'image floutée dans le buffer
			log.Printf("Erreur encodage PNG: %v", err)
			return
		}

		// Send length-prefixed response
		respLen := uint64(outBuf.Len()) // Longueur de l'image encodée
		var respLenBuf [8]byte
		binary.BigEndian.PutUint64(respLenBuf[:], respLen)   // Conversion de la longueur en bytes (big-endian)
		if _, err := conn.Write(respLenBuf[:]); err != nil { // Envoi du header de longueur au client
			log.Printf("Erreur envoi header: %v", err) // Plus d'erreur handling
			return
		}
		if _, err := conn.Write(outBuf.Bytes()); err != nil { // Envoi des bytes de l'image encodée au client
			log.Printf("Erreur envoi image: %v", err)
			return
		}

		// Loop to allow client to send more images on same connection
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Serveur TCP en écoute sur le port 8080 (protocol: uint64 BE length + PNG bytes)")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Erreur d'acceptation: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}
