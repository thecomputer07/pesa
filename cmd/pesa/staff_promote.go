package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// runStaff dispatches `pesa staff promote|demote <email>` subcommands.
func runStaff(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: pesa staff {promote|demote|apply-tombstones} <arg>")
		os.Exit(2)
	}
	sub := args[0]
	rest := args[1:]
	switch sub {
	case "promote":
		runStaffSetFlag(rest, true)
	case "demote":
		runStaffSetFlag(rest, false)
	case "apply-tombstones":
		runApplyTombstones(rest)
	default:
		fmt.Fprintf(os.Stderr, "unknown staff subcommand: %s\n", sub)
		os.Exit(2)
	}
}

func runStaffSetFlag(args []string, promote bool) {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: pesa staff promote|demote <email>")
		os.Exit(2)
	}
	email := strings.ToLower(strings.TrimSpace(args[0]))
	if email == "" {
		fmt.Fprintln(os.Stderr, "email must not be empty")
		os.Exit(2)
	}

	cfg := mustConfig()
	db := openDB(cfg)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	res, err := db.ExecContext(ctx, `UPDATE users SET is_staff=? WHERE email=?`, promote, email)
	if err != nil {
		slog.Error("staff toggle", "err", err)
		os.Exit(1)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		fmt.Fprintf(os.Stderr, "no user with email %s\n", email)
		os.Exit(1)
	}
	flag := "demoted"
	if promote {
		flag = "promoted"
	}
	fmt.Printf("%s %s (is_staff=%v)\n", flag, email, promote)
}
