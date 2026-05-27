package main

import (
	"fmt"
	"net"
	"bufio"
	"sync"
)

type Client struct {
	conn net.Conn
	username string
}

type Server struct {
	clients []Client
	mu      sync.Mutex
}

func (s *Server) addClient(client Client) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.clients = append(s.clients, client)
}

func (s *Server) removeClient(client Client){
	s.mu.Lock();
	defer s.mu.Unlock();

	for i, c := range s.clients {
        if c.conn == client.conn {
            s.clients = append(s.clients[:i], s.clients[i+1:]...)
            break
        }
    }
}

func (s *Server) broadcast(sender Client, message string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    for _, client := range s.clients {
        if client.conn != sender.conn {
            client.conn.Write([]byte(fmt.Sprintf("[%s]: %s\n", sender.username, message)))
        }
    }
}

func (s *Server) handleConnection(conn net.Conn){
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
	s.addClient(client)
	fmt.Printf("%s joined the server ( from %s )\n", client.username, client.conn.RemoteAddr());
	s.broadcast(Client{username: "Server"}, fmt.Sprintf("%s joined the chat", username))


	for scanner.Scan() {
		message := scanner.Text();
		// fmt.Printf("[%s]: %s\n", client.username, message)
		s.broadcast(client, message);
	}

	s.removeClient(client);
	fmt.Printf("Client disconnected: %s\n [%s]", client.username, client.conn.RemoteAddr());
	s.broadcast(Client{username: "Server"}, fmt.Sprintf("%s left the chat", username))
}

func main() {
	listener, err := net.Listen("tcp", ":8080");
	if err != nil {
		fmt.Println("Error starting TCP Server", err);
	}

	defer listener.Close();

	fmt.Println("TCP Server started on PORT 8080");

	server := &Server{}

	for {
		conn, err := listener.Accept();
		if err != nil {
			fmt.Println("Error accepting connection:", err);
			continue;
		}

		go server.handleConnection(conn);
	}
}