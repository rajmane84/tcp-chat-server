package main

import (
	"fmt"
	"net"
	"bufio"
)

type Client struct {
	conn net.Conn
	username string
}

func hanldeConnection(conn net.Conn){
	defer conn.Close();

	scanner := bufio.NewScanner(conn);

	err := scanner.Err();
	if err != nil {
		fmt.Printf("Error while scanning", err);
	}

	conn.Write([]byte("Enter your username: "));
	var username string;
	if scanner.Scan() {
		username = scanner.Text();
	}

	client := Client{conn: conn, username: username};
	fmt.Printf("%s joined the server ( from %s )\n", client.username, client.conn.RemoteAddr());


	for scanner.Scan() {
		message := scanner.Text();
		fmt.Printf("[%s]: %s\n", client.username, message)
	}

	fmt.Printf("Client disconnected: %s\n [%s]", client.username, client.conn.RemoteAddr());
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