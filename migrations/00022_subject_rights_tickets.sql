-- +goose Up
-- LEGAL-09/10: Subject Rights Tickets (privacy@ inbox) with 30d SLA.
CREATE TABLE subject_rights_tickets (
  id                   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  ticket_ref           VARCHAR(64) NOT NULL UNIQUE,
  requester_email_hash CHAR(64) NOT NULL,
  kind                 TEXT NOT NULL
                       CHECK (kind IN ('access','rectification','erasure','portability','restriction','objection','complaint')),
  status               TEXT NOT NULL DEFAULT 'new'
                       CHECK (status IN ('new','in_progress','closed','rejected')),
  received_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deadline             TIMESTAMPTZ NOT NULL,
  closed_at            TIMESTAMPTZ NULL,
  notes_md             TEXT NULL,
  evidence_path        VARCHAR(512) NULL
);
CREATE INDEX idx_srt_status ON subject_rights_tickets (status);
CREATE INDEX idx_srt_deadline ON subject_rights_tickets (deadline);
CREATE INDEX idx_srt_requester ON subject_rights_tickets (requester_email_hash);

-- +goose Down
DROP TABLE subject_rights_tickets;
