package main

import (
	"fmt"
	"net"
	"bufio"
	"sync"
	"strings"
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

	// Handled the case where the client disconnects at the exact moment broadcast is looping over clients list
    for _, client := range s.clients {
		if client.conn != sender.conn {
			_, err := client.conn.Write([]byte(fmt.Sprintf("[%s]: %s\n", sender.username, message)))
			if err != nil {
				fmt.Printf("Failed to send message to %s: %s\n", client.username, err)
			}
		}
	}
}

func (s *Server) handleConnection(conn net.Conn){
	defer conn.Close();

	scanner := bufio.NewScanner(conn);

	conn.Write([]byte("Enter your username: "));
	var username string;
	for {
		if !scanner.Scan() {
			// client disconnected before entering username
			fmt.Println("Client disconnected before entering username")
			return;
		}
		username = strings.TrimSpace(scanner.Text())
		if username != "" {
			break;
		}
		conn.Write([]byte("Username cannot be empty. Enter your username: "))
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

	// Distinguish clean disconnect vs error
	if err := scanner.Err(); err != nil {
		fmt.Printf("%s lost connection due to error: %s\n", client.username, err)
	} else {
		fmt.Printf("%s disconnected cleanly\n", client.username)
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