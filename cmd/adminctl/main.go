package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/JustARandomBadDev/captive-portal-admin/internal/adminauth"
	"github.com/JustARandomBadDev/captive-portal-admin/internal/config"
	"github.com/JustARandomBadDev/captive-portal-admin/internal/database"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/term"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] != "create-admin" {
		fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/adminctl create-admin")
		os.Exit(2)
	}

	if err := createAdmin(); err != nil {
		fmt.Fprintf(os.Stderr, "create admin: %v\n", err)
		os.Exit(1)
	}
}

func createAdmin() error {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, database.Config{URL: cfg.DatabaseURL})
	if err != nil {
		return err
	}
	defer db.Close()

	reader := bufio.NewReader(os.Stdin)
	username, err := readLine(reader, "Username: ")
	if err != nil {
		return err
	}
	username = strings.TrimSpace(username)
	if username == "" {
		return errors.New("username is required")
	}

	password, err := readPassword(reader, "Password: ")
	if err != nil {
		return err
	}
	if strings.TrimSpace(password) == "" {
		return errors.New("password is required")
	}

	confirm, err := readPassword(reader, "Confirm password: ")
	if err != nil {
		return err
	}
	if password != confirm {
		return errors.New("password confirmation does not match")
	}

	hash, err := adminauth.HashPassword(password)
	if err != nil {
		return err
	}

	id, err := database.NewUUID()
	if err != nil {
		return err
	}

	_, err = db.Pool().Exec(ctx, `
INSERT INTO admin_users (id, username, password_hash, display_name, is_active)
VALUES ($1, $2, $3, $4, true)
`, id, username, hash, username)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("admin username %q already exists", username)
		}
		return err
	}

	fmt.Printf("Admin created: %s\n", username)
	return nil
}

func readLine(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(value, "\r\n"), nil
}

func readPassword(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	if term.IsTerminal(int(syscall.Stdin)) {
		bytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}

	return readLine(reader, "")
}
