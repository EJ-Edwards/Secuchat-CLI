package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type      string `json:"type"`
	Message   string `json:"msg,omitempty"`
	Username  string `json:"user,omitempty"`
	Timestamp string `json:"ts,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Secuchat-CLI v1.2.0 - User Management System")
		fmt.Println("==============================================")
		fmt.Println("Usage:")
		fmt.Println("  go run *.go <server_url> [pin]          - Join chat")
		fmt.Println("  go run *.go --setup                     - Initial admin setup")
		fmt.Println("  go run *.go --create-user               - Create new user (admin only)")
		fmt.Println("  go run *.go --list-users                - List all users")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  go run *.go ws://127.0.0.1:8080/ws REDTEAM01")
		fmt.Println("  go run *.go --setup")
		return
	}

	switch os.Args[1] {
	case "--setup":
		err := InitialSetup()
		if err != nil {
			fmt.Printf("❌ Setup failed: %v\n", err)
		}
		return
	case "--list-users":
		err := ListUsers()
		if err != nil {
			fmt.Printf("❌ Failed to list users: %v\n", err)
		}
		return
	case "--create-user":
		// Need to authenticate first
		username, isAdmin, err := Login()
		if err != nil {
			fmt.Printf("❌ Authentication failed: %v\n", err)
			return
		}
		err = CreateUser(username, isAdmin)
		if err != nil {
			fmt.Printf("❌ Failed to create user: %v\n", err)
		}
		return
	}

	serverURL := os.Args[1]
	pin := "GENERAL"
	if len(os.Args) > 2 {
		pin = os.Args[2]
	}

	// Authenticate user
	username, isAdmin, err := Login()
	if err != nil {
		fmt.Printf("❌ Login failed: %v\n", err)
		return
	}

	// Add pin, username and admin status to URL
	u, err := url.Parse(serverURL)
	if err != nil {
		log.Fatal("Invalid server URL:", err)
	}
	q := u.Query()
	q.Set("pin", pin)
	q.Set("username", username)
	if isAdmin {
		q.Set("admin", "true")
	}
	u.RawQuery = q.Encode()

	role := "USER"
	if isAdmin {
		role = "ADMIN"
	}
	fmt.Printf("🔗 Connecting to room %s as %s [%s]...\n", pin, username, role)

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Connection failed:", err)
	}
	defer conn.Close()

	fmt.Printf("✅ Connected to Secuchat-CLI room: %s\n", pin)
	fmt.Println("📝 Type messages and press Enter. Type '/quit' to exit.")
	fmt.Println("---")

	// Handle incoming messages
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var msg Message
				err := conn.ReadJSON(&msg)
				if err != nil {
					if !websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
						fmt.Printf("❌ Read error: %v\n", err)
					}
					cancel()
					return
				}

				switch msg.Type {
				case "system":
					fmt.Printf("🔔 %s\n", msg.Message)
				case "message":
					timestamp := ""
					if msg.Timestamp != "" {
						if t, err := time.Parse(time.RFC3339, msg.Timestamp); err == nil {
							timestamp = t.Format("15:04")
						}
					}
					if timestamp != "" {
						fmt.Printf("[%s] %s: %s\n", timestamp, msg.Username, msg.Message)
					} else {
						fmt.Printf("%s: %s\n", msg.Username, msg.Message)
					}
				case "pong":
					// Handle pong silently
				default:
					fmt.Printf("📨 %s\n", msg.Message)
				}
			}
		}
	}()

	// Handle user input
	scanner := bufio.NewScanner(os.Stdin)

	// Handle Ctrl+C gracefully
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("\n👋 Disconnecting...")
		cancel()
		conn.Close()
		os.Exit(0)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if scanner.Scan() {
				input := strings.TrimSpace(scanner.Text())
				if input == "" {
					continue
				}
				if input == "/quit" {
					fmt.Println("👋 Goodbye!")
					return
				}

				if input == "/help" {
					fmt.Println("📋 Available commands:")
					fmt.Println("  /quit - Exit the chat")
					if isAdmin {
						fmt.Println("  /kick <username> - Kick a user (Admin only)")
						fmt.Println("  /create-user - Create new user account (Admin only)")
						fmt.Println("  /list-users - List all registered users (Admin only)")
					}
					fmt.Println("  /help - Show this help")
					continue
				}

				if input == "/create-user" && isAdmin {
					fmt.Println("🔄 Creating user... (This will interrupt chat temporarily)")
					err := CreateUser(username, isAdmin)
					if err != nil {
						fmt.Printf("❌ Failed to create user: %v\n", err)
					}
					continue
				}

				if input == "/list-users" && isAdmin {
					err := ListUsers()
					if err != nil {
						fmt.Printf("❌ Failed to list users: %v\n", err)
					}
					continue
				}

				if strings.HasPrefix(input, "/kick ") {
					// Let the server handle kick validation
					msg := Message{
						Type:      "message",
						Message:   input,
						Username:  username,
						Timestamp: time.Now().UTC().Format(time.RFC3339),
					}
					err := conn.WriteJSON(msg)
					if err != nil {
						fmt.Printf("❌ Send error: %v\n", err)
						return
					}
					continue
				}

				msg := Message{
					Type:      "message",
					Message:   input,
					Username:  username,
					Timestamp: time.Now().UTC().Format(time.RFC3339),
				}

				err := conn.WriteJSON(msg)
				if err != nil {
					fmt.Printf("❌ Send error: %v\n", err)
					return
				}
			} else {
				if err := scanner.Err(); err != nil {
					fmt.Printf("❌ Input error: %v\n", err)
				}
				return
			}
		}
	}
}
