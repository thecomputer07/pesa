-- +goose Up
-- scs postgresstore schema — see https://github.com/alexedwards/scs/tree/master/postgresstore
CREATE TABLE sessions (
  token  CHAR(43) NOT NULL PRIMARY KEY,
  data   BYTEA NOT NULL,
  expiry TIMESTAMPTZ(6) NOT NULL
);
CREATE INDEX sessions_expiry_idx ON sessions (expiry);

-- +goose Down
DROP TABLE sessions;
