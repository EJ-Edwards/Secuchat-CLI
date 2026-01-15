package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/term"
)

const (
	UserDBFile = "users.json"
	SaltSize   = 16
	KeySize    = 32
)

type User struct {
	PasswordHash string    `json:"password_hash"`
	Salt         string    `json:"salt"`
	IsAdmin      bool      `json:"is_admin"`
	DisplayName  string    `json:"display_name"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    string    `json:"created_by"`
}

type UserDatabase struct {
	Users map[string]User `json:"users"`
}

func hashPassword(password string, salt []byte) string {
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, KeySize)
	return base64.StdEncoding.EncodeToString(hash)
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	_, err := rand.Read(salt)
	return salt, err
}

func loadUsers() (UserDatabase, error) {
	var db UserDatabase
	db.Users = make(map[string]User)

	if _, err := os.Stat(UserDBFile); os.IsNotExist(err) {
		return db, nil
	}

	data, err := os.ReadFile(UserDBFile)
	if err != nil {
		return db, err
	}

	err = json.Unmarshal(data, &db)
	if err != nil {
		db.Users = make(map[string]User)
	}

	return db, nil
}

func saveUsers(db UserDatabase) error {
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(UserDBFile, data, 0600)
}

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(bytePassword), nil
}

func InitialSetup() error {
	db, err := loadUsers()
	if err != nil {
		return err
	}

	if len(db.Users) > 0 {
		fmt.Println("❌ Users already exist. Use admin account to create more users.")
		return fmt.Errorf("users already exist")
	}

	fmt.Println("🚀 Secuchat-CLI Initial Setup")
	fmt.Println("==============================")
	fmt.Println("Creating first admin account...")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Admin username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	username = strings.TrimSpace(username)

	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	password, err := readPassword("Admin password: ")
	if err != nil {
		return err
	}

	confirmPassword, err := readPassword("Confirm password: ")
	if err != nil {
		return err
	}

	if password != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	fmt.Print("Display name: ")
	displayName, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	displayName = strings.TrimSpace(displayName)

	if displayName == "" {
		displayName = username
	}

	salt, err := generateSalt()
	if err != nil {
		return err
	}

	db.Users[username] = User{
		PasswordHash: hashPassword(password, salt),
		Salt:         base64.StdEncoding.EncodeToString(salt),
		IsAdmin:      true,
		DisplayName:  displayName,
		CreatedAt:    time.Now(),
		CreatedBy:    "SYSTEM",
	}

	err = saveUsers(db)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Admin account '%s' created successfully!\n", username)
	return nil
}

func CreateUser(creatorUsername string, isCreatorAdmin bool) error {
	if !isCreatorAdmin {
		return fmt.Errorf("only admins can create new users")
	}

	db, err := loadUsers()
	if err != nil {
		return err
	}

	fmt.Println("\n➕ Create New User")
	fmt.Println("===================")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("New username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	username = strings.TrimSpace(username)

	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if _, exists := db.Users[username]; exists {
		return fmt.Errorf("username '%s' already exists", username)
	}

	password, err := readPassword("Password: ")
	if err != nil {
		return err
	}

	confirmPassword, err := readPassword("Confirm password: ")
	if err != nil {
		return err
	}

	if password != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	fmt.Print("Display name: ")
	displayName, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	displayName = strings.TrimSpace(displayName)

	if displayName == "" {
		displayName = username
	}

	fmt.Print("Grant admin privileges? (y/N): ")
	adminChoice, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	isAdmin := strings.ToLower(strings.TrimSpace(adminChoice)) == "y"

	salt, err := generateSalt()
	if err != nil {
		return err
	}

	db.Users[username] = User{
		PasswordHash: hashPassword(password, salt),
		Salt:         base64.StdEncoding.EncodeToString(salt),
		IsAdmin:      isAdmin,
		DisplayName:  displayName,
		CreatedAt:    time.Now(),
		CreatedBy:    creatorUsername,
	}

	err = saveUsers(db)
	if err != nil {
		return err
	}

	role := "USER"
	if isAdmin {
		role = "ADMIN"
	}

	fmt.Printf("✅ User '%s' (%s) [%s] created successfully!\n\n", username, displayName, role)
	return nil
}

func Login() (string, bool, error) {
	db, err := loadUsers()
	if err != nil {
		return "", false, err
	}

	if len(db.Users) == 0 {
		fmt.Println("⚠️  No users found. Please run initial setup first:")
		fmt.Println("   go run *.go --setup")
		return "", false, fmt.Errorf("no users configured")
	}

	fmt.Println("🔒 Secuchat-CLI Authentication")
	fmt.Println("===============================")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", false, err
	}
	username = strings.TrimSpace(username)

	password, err := readPassword("Password: ")
	if err != nil {
		return "", false, err
	}

	user, exists := db.Users[username]
	if !exists {
		fmt.Println("❌ Invalid username or password")
		return "", false, fmt.Errorf("authentication failed")
	}

	salt, err := base64.StdEncoding.DecodeString(user.Salt)
	if err != nil {
		return "", false, err
	}

	hash := hashPassword(password, salt)
	if hash != user.PasswordHash {
		fmt.Println("❌ Invalid username or password")
		return "", false, fmt.Errorf("authentication failed")
	}

	role := "USER"
	if user.IsAdmin {
		role = "ADMIN"
	}

	fmt.Printf("✅ Authentication successful! Welcome, %s [%s]\n", user.DisplayName, role)
	return username, user.IsAdmin, nil
}

func ListUsers() error {
	db, err := loadUsers()
	if err != nil {
		return err
	}

	fmt.Println("\n👥 Registered Users")
	fmt.Println("====================")

	if len(db.Users) == 0 {
		fmt.Println("No users found.")
		return nil
	}

	fmt.Println("\nADMIN ACCOUNTS:")
	hasAdmins := false
	for username, user := range db.Users {
		if user.IsAdmin {
			hasAdmins = true
			fmt.Printf("  • %s (%s) - Created: %s by %s\n",
				username, user.DisplayName,
				user.CreatedAt.Format("2006-01-02"), user.CreatedBy)
		}
	}
	if !hasAdmins {
		fmt.Println("  (none)")
	}

	fmt.Println("\nUSER ACCOUNTS:")
	hasUsers := false
	for username, user := range db.Users {
		if !user.IsAdmin {
			hasUsers = true
			fmt.Printf("  • %s (%s) - Created: %s by %s\n",
				username, user.DisplayName,
				user.CreatedAt.Format("2006-01-02"), user.CreatedBy)
		}
	}
	if !hasUsers {
		fmt.Println("  (none)")
	}
	fmt.Println()

	return nil
}
