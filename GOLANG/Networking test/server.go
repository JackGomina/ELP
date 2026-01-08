package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Nouveau client :", conn.RemoteAddr())

	reader := bufio.NewReader(conn)

	for {
		// Lecture jusqu'au saut de ligne
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client déconnecté :", conn.RemoteAddr())
			return
		}

		// Traitement
		message = strings.TrimSpace(message)
		response := strings.ToUpper(message)

		// Envoi au client
		conn.Write([]byte(response + "\n"))
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Serveur TCP en écoute sur le port 8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Erreur d'acceptation :", err)
			continue
		}

		// Une goroutine par client
		go handleConnection(conn)
	}
}
