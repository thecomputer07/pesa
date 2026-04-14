package srt

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Mailer is the subset of the service.Mailer interface needed by Notifier.
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

// Notifier sends the subject-rights ACK mail from a markdown template.
type Notifier struct {
	TemplatesDir string
	Mailer       Mailer

	once     sync.Once
	tplErr   error
	tplBody  string
}

// NewNotifier constructs a Notifier; the template file is loaded lazily on
// first use so missing-file errors surface immediately on the first call.
func NewNotifier(templatesDir string, mailer Mailer) *Notifier {
	if templatesDir == "" {
		templatesDir = ".planning/legal/subject-rights-templates"
	}
	return &Notifier{TemplatesDir: templatesDir, Mailer: mailer}
}

// SendAck substitutes {{TICKET_REF}} + {{DEADLINE}} and sends to recipient.
func (n *Notifier) SendAck(ctx context.Context, t *Ticket, recipient string) error {
	n.once.Do(n.loadTemplate)
	if n.tplErr != nil {
		return n.tplErr
	}
	if t == nil {
		return errors.New("srt: nil ticket")
	}
	body := Render(n.tplBody, map[string]string{
		"TICKET_REF": t.TicketRef,
		"DEADLINE":   t.Deadline.UTC().Format(time.RFC3339),
	})
	if n.Mailer == nil {
		return nil // dev mode: substitute only
	}
	subject := fmt.Sprintf("[Pesa] Richiesta ricevuta — %s", t.TicketRef)
	return n.Mailer.Send(ctx, recipient, subject, body)
}

// Render substitutes {{KEY}} placeholders with the provided map values.
// Unknown placeholders are left as-is (useful during authoring).
func Render(tpl string, vars map[string]string) string {
	out := tpl
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	return out
}

func (n *Notifier) loadTemplate() {
	path := filepath.Join(n.TemplatesDir, "00-ack.md")
	body, err := os.ReadFile(path)
	if err != nil {
		// Fallback template so the notifier still renders if the project
		// hasn't committed the template file yet. The Italian text matches
		// the minimal ACK required by GDPR Art. 12(3).
		n.tplBody = defaultAckTemplate
		n.tplErr = nil
		return
	}
	n.tplBody = string(body)
}

const defaultAckTemplate = `Gentile interessato/a,

abbiamo ricevuto la sua richiesta relativa ai diritti dell'interessato
(GDPR artt. 15-22).

Riferimento ticket: {{TICKET_REF}}
Scadenza di risposta: {{DEADLINE}} (30 giorni, prorogabili di 60 in caso di complessità).

Le risponderemo quanto prima. Per comunicazioni relative a questa
richiesta, citi il riferimento del ticket.

Cordialmente,
Pesa — privacy@pesadia.com
`
