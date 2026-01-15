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
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run client.go <server_url> <pin> [username]")
		fmt.Println("Example: go run client.go ws://127.0.0.1:8080/ws REDTEAM01 operator1")
		return
	}

	serverURL := os.Args[1]
	pin := os.Args[2]
	username := "anonymous"
	if len(os.Args) > 3 {
		username = os.Args[3]
	}

	// Add pin and username to URL
	u, err := url.Parse(serverURL)
	if err != nil {
		log.Fatal("Invalid server URL:", err)
	}
	q := u.Query()
	q.Set("pin", pin)
	q.Set("username", username)
	u.RawQuery = q.Encode()

	fmt.Printf("🔗 Connecting to room %s as %s...\n", pin, username)

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
					fmt.Println("  /kick <username> - Kick a user (Admin only)")
					fmt.Println("  /help - Show this help")
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
