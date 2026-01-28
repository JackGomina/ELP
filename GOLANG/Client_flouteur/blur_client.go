package main

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: %s input.png output.png", os.Args[0])
	}
	inPath := os.Args[1]
	outPath := os.Args[2]

	data, err := os.ReadFile(inPath)
	if err != nil {
		log.Fatalf("read input: %v", err)
	}

	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close()

	var hdr [8]byte
	binary.BigEndian.PutUint64(hdr[:], uint64(len(data)))
	if _, err := conn.Write(hdr[:]); err != nil {
		log.Fatalf("send header: %v", err)
	}
	if _, err := conn.Write(data); err != nil {
		log.Fatalf("send image: %v", err)
	}

	// Read response length
	if _, err := io.ReadFull(conn, hdr[:]); err != nil {
		log.Fatalf("read resp header: %v", err)
	}
	respLen := binary.BigEndian.Uint64(hdr[:])
	resp := make([]byte, respLen)
	if _, err := io.ReadFull(conn, resp); err != nil {
		log.Fatalf("read resp body: %v", err)
	}

	if err := os.WriteFile(outPath, resp, 0644); err != nil {
		log.Fatalf("write output: %v", err)
	}
}
