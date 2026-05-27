package main

import (
	"fmt"
	"net"
	"bufio"
	"log"
)

func hanldeConnection(conn net.Conn){
	defer conn.Close();

	fmt.Printf("Client connected: %s\n", conn.RemoteAddr());

	scanner := bufio.NewScanner(conn);

	err := scanner.Err();

	if err != nil {
		log.Fatal("Error while scanning", err);
	}


	for scanner.Scan() {
		message := scanner.Text();
		fmt.Printf("[%s]: %s\n", conn.RemoteAddr(), message)
	}

	fmt.Printf("Client disconnected: %s\n", conn.RemoteAddr());
}

func main() {
	listener, err := net.Listen("tcp", ":8080");
	if err != nil {
		fmt.Println("Error starting TCP Server", err);
	}

	defer listener.Close();

	fmt.Println("TCP Server started on PORT 8080");

	for {
		conn, err := listener.Accept();
		if err != nil {
			fmt.Println("Error accepting connection:", err);
			continue;
		}

		go hanldeConnection(conn);
	}
}