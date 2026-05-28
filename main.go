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

func (s *Server) sendPrivate(sender Client, targetUsername string, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sender.username == targetUsername {
		sender.conn.Write([]byte("You cannot send a private message to yourself\n"))
		return
	}

	// find the target client
	for _, client := range s.clients {
		if client.username == targetUsername {
			client.conn.Write([]byte(fmt.Sprintf("[PM from %s]: %s\n", sender.username, message)))
			sender.conn.Write([]byte(fmt.Sprintf("[PM to %s]: %s\n", targetUsername, message)))
			return
		}
	}

	// target not found
	sender.conn.Write([]byte(fmt.Sprintf("User '%s' not found\n", targetUsername)))
}

func (s *Server) listUsers(requester Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	requester.conn.Write([]byte("--- Online Users ---\n"))
	for _, client := range s.clients {
		if client.conn == requester.conn {
			requester.conn.Write([]byte(fmt.Sprintf("  %s (you)\n", client.username)))
		} else {
			requester.conn.Write([]byte(fmt.Sprintf("  %s\n", client.username)))
		}
	}
	requester.conn.Write([]byte("--------------------\n"))
}

func (s *Server) handleCommand(client Client, message string) bool {
	// not a command
	if !strings.HasPrefix(message, "/") {
		return false
	}

	parts := strings.SplitN(message, " ", 3)
	command := parts[0]

	switch command {
	case "/msg":
		if len(parts) < 3 {
			client.conn.Write([]byte("Usage: /msg <username> <message>\n"))
			return true
		}
		targetUsername := parts[1]
		message := parts[2]
		s.sendPrivate(client, targetUsername, message)

	case "/users":
		s.listUsers(client)

	case "/help":
		client.conn.Write([]byte("--- Commands ---\n"))
		client.conn.Write([]byte("/msg <username> <message> - Send a private message\n"))
		client.conn.Write([]byte("/users                    - List online users\n"))
		client.conn.Write([]byte("/help                     - Show this help message\n"))
		client.conn.Write([]byte("----------------\n"))

	default:
		client.conn.Write([]byte(fmt.Sprintf("Unknown command '%s'. Type /help for available commands\n", command)))
	}

	return true
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	// Ask for username
	conn.Write([]byte("Enter your username: "))
	var username string
	for {
		if !scanner.Scan() {
			fmt.Println("Client disconnected before entering username")
			return
		}
		username = strings.TrimSpace(scanner.Text())
		if username != "" {
			break
		}
		conn.Write([]byte("Username cannot be empty. Enter your username: "))
	}

	client := Client{conn: conn, username: username}
	s.addClient(client)
	fmt.Printf("%s joined the server\n", client.username)
	s.broadcast(Client{username: "Server"}, fmt.Sprintf("%s joined the chat", username))
	conn.Write([]byte("Type /help for available commands\n"))

	// Message loop
	for scanner.Scan() {
		message := scanner.Text()

		// check if it's a command, if not broadcast
		if !s.handleCommand(client, message) {
			fmt.Printf("[%s]: %s\n", client.username, message)
			s.broadcast(client, message)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("%s lost connection due to error: %s\n", client.username, err)
	} else {
		fmt.Printf("%s disconnected cleanly\n", client.username)
	}

	s.removeClient(client)
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