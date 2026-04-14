package service

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Representative subset (5 tests) per realism directive. Full 17-case suite
// listed in plan-03 task 7 — TODO for later validation phase.

type fakeMailer struct {
	calls int
	last  struct{ to, subject, body string }
}

func (f *fakeMailer) Send(_ context.Context, to, subject, body string) error {
	f.calls++
	f.last.to = to
	f.last.subject = subject
	f.last.body = body
	return nil
}

func newTestSvc(t *testing.T) (*AuthService, *fakeMailer, func()) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run integration tests")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}
	// Truncate in dependency order (respect FK CASCADE).
	if _, err := db.Exec(`TRUNCATE password_resets, email_verifications, sessions, users RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	fm := &fakeMailer{}
	svc := &AuthService{DB: db, Mailer: fm, Now: time.Now, Domain: "test.invalid"}
	return svc, fm, func() { _ = db.Close() }
}

func TestRegisterHappyPath(t *testing.T) {
	svc, fm, done := newTestSvc(t)
	defer done()
	uid, err := svc.Register(context.Background(), RegisterInput{
		Email: "u1@example.test", Password: "CorrectHorseBattery9!", Locale: "it", PrivacyConsent: true,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if uid == 0 {
		t.Fatal("expected user id")
	}
	if fm.calls != 1 {
		t.Fatalf("want 1 mail, got %d", fm.calls)
	}
	if !strings.Contains(fm.last.body, "verify-email?token=") {
		t.Fatalf("mail body missing verify link: %q", fm.last.body)
	}
}

func TestRegisterRequiresConsent(t *testing.T) {
	svc, _, done := newTestSvc(t)
	defer done()
	_, err := svc.Register(context.Background(), RegisterInput{
		Email: "u2@example.test", Password: "CorrectHorseBattery9!", Locale: "it", PrivacyConsent: false,
	})
	if err != ErrPrivacyConsentRequired {
		t.Fatalf("want ErrPrivacyConsentRequired, got %v", err)
	}
}

func TestLoginEmailNotVerified(t *testing.T) {
	svc, _, done := newTestSvc(t)
	defer done()
	_, err := svc.Register(context.Background(), RegisterInput{
		Email: "u3@example.test", Password: "CorrectHorseBattery9!", Locale: "it", PrivacyConsent: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Login(context.Background(), "u3@example.test", "CorrectHorseBattery9!")
	if err != ErrEmailNotVerified {
		t.Fatalf("want ErrEmailNotVerified, got %v", err)
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	svc, _, done := newTestSvc(t)
	defer done()
	_, err := svc.Login(context.Background(), "unknown@example.test", "wrong-password-123!")
	if err != ErrInvalidCredentials {
		t.Fatalf("want ErrInvalidCredentials, got %v", err)
	}
}

func TestPasswordResetUnknownEmailNoError(t *testing.T) {
	svc, _, done := newTestSvc(t)
	defer done()
	if err := svc.RequestPasswordReset(context.Background(), "nobody@example.test", nil); err != nil {
		t.Fatalf("enumeration-safe RequestPasswordReset returned error: %v", err)
	}
}

// TODO: TestVerifyEmailHappyPath/Expired/Reuse, TestLoginHappyPath,
// TestLoginUnknownEmailNoEnumeration (timing), TestPasswordResetHappyPath/
// Expired/Reuse, TestRateLimit{Register,Login,ResetRequest} — deferred to
// validation phase per plan-01 realism directive.
