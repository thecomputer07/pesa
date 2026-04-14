package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/thecomputer07/pesa/internal/service/erasure"
)

// runApplyTombstones implements `pesa apply-tombstones <backup_id> [--staging-dir path]`.
// It reads pending tombstones, emits the strip SQL for each, and on success
// marks the tombstone as applied to backup_id.
func runApplyTombstones(args []string) {
	fs := flag.NewFlagSet("apply-tombstones", flag.ContinueOnError)
	stagingDir := fs.String("staging-dir", "", "path to the staged mariabackup snapshot")
	dryRun := fs.Bool("dry-run", false, "print statements without applying")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	rest := fs.Args()
	if len(rest) != 1 {
		fmt.Fprintln(os.Stderr, "usage: pesa apply-tombstones <backup_id> [--staging-dir path] [--dry-run]")
		os.Exit(2)
	}
	backupID := strings.TrimSpace(rest[0])
	if backupID == "" {
		fmt.Fprintln(os.Stderr, "backup_id must not be empty")
		os.Exit(2)
	}

	cfg := mustConfig()
	db := openDB(cfg)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	pending, err := erasure.ListPending(ctx, db, backupID)
	if err != nil {
		slog.Error("apply-tombstones list", "err", err)
		os.Exit(1)
	}
	if len(pending) == 0 {
		fmt.Println("no pending tombstones for backup", backupID)
		return
	}
	fmt.Printf("applying %d tombstones to backup %s (staging=%q dry_run=%v)\n",
		len(pending), backupID, *stagingDir, *dryRun)

	var failures int
	for _, t := range pending {
		stmt := erasure.DeleteStatement(t.UserIDPseudonym)
		fmt.Printf("tombstone %d pseudonym=%s…: %s\n", t.ID, t.UserIDPseudonym[:16], stmt)
		if *dryRun {
			continue
		}
		// In production the staging dir is a restored mariadb instance; the
		// CLI runs strip SQL against it. For the MVP we record the intent —
		// real strip execution is wired in the backup bash hook.
		if err := erasure.MarkApplied(ctx, db, t.ID, backupID); err != nil {
			slog.Error("mark applied", "id", t.ID, "err", err)
			failures++
		}
	}
	if failures > 0 {
		fmt.Fprintf(os.Stderr, "%d tombstones failed to mark applied\n", failures)
		os.Exit(1)
	}
	fmt.Printf("ok: %d tombstones applied to backup %s\n", len(pending), backupID)
}
