package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Connecté au serveur")

	serverReader := bufio.NewReader(conn)
	inputReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Message > ")
		text, _ := inputReader.ReadString('\n')

		// Envoi au serveur
		conn.Write([]byte(text))

		// Réponse du serveur
		response, _ := serverReader.ReadString('\n')
		fmt.Print("Réponse serveur : ", response)
	}
}
