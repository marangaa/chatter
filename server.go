package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "sync"
)

// Client represents a connected client
type Client struct {
    conn net.Conn
    name string
}

var (
    clients   = make(map[*Client]bool) // Map to track connected clients
    broadcast = make(chan string)      // Channel for broadcasting messages
    mutex     = sync.Mutex{}           // Mutex to protect the clients map
)

func main() {
    // Start the server on a given port
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        fmt.Println("Error starting server:", err)
        os.Exit(1)
    }
    defer listener.Close()
    fmt.Println("Server started on :8080")

    // Goroutine to handle broadcasting messages
    go handleBroadcast()

    // Accept incoming connections
    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting connection:", err)
            continue
        }
        client := &Client{conn: conn}
        go handleConnection(client)
    }
}

// handleBroadcast sends messages to all connected clients
func handleBroadcast() {
    for {
        message := <-broadcast
        mutex.Lock()
        for client := range clients {
            client.conn.Write([]byte(message))
        }
        mutex.Unlock()
    }
}

// handleConnection handles communication with a single client
func handleConnection(client *Client) {
    defer client.conn.Close()

    // Add client to the clients map
    mutex.Lock()
    clients[client] = true
    mutex.Unlock()

    client.conn.Write([]byte("Enter your name: "))
    name, _ := bufio.NewReader(client.conn).ReadString('\n')
    client.name = name

    welcomeMessage := fmt.Sprintf("%s has joined the chat\n", client.name)
    broadcast <- welcomeMessage

    for {
        message, err := bufio.NewReader(client.conn).ReadString('\n')
        if err != nil {
            fmt.Printf("%s left the chat\n", client.name)
            mutex.Lock()
            delete(clients, client)
            mutex.Unlock()
            broadcast <- fmt.Sprintf("%s has left the chat\n", client.name)
            return
        }

		// check for commands
		if message == "/exit\n" {
			fmt.Printf("%s left the chat\n", client.name)
			mutex.Lock()
			delete(clients, client)
			mutex.Unlock()
			return
		}

        broadcast <- fmt.Sprintf("%s: %s", client.name, message)
    }
}
