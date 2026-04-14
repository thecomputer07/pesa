-- Substitute @@APP_PASSWORD@@ / @@MIG_PASSWORD@@ / @@BAK_PASSWORD@@ before applying.
-- Run as root after the mariadb container has initialised the pesa database.

CREATE USER IF NOT EXISTS 'pesa_app'@'%'       IDENTIFIED BY '@@APP_PASSWORD@@' REQUIRE SSL;
CREATE USER IF NOT EXISTS 'pesa_migration'@'%' IDENTIFIED BY '@@MIG_PASSWORD@@' REQUIRE SSL;
CREATE USER IF NOT EXISTS 'pesa_backup'@'%'    IDENTIFIED BY '@@BAK_PASSWORD@@' REQUIRE SSL;

GRANT SELECT, INSERT, UPDATE, DELETE, EXECUTE ON pesa.* TO 'pesa_app'@'%';
GRANT CREATE, DROP, ALTER, REFERENCES, INDEX, TRIGGER ON pesa.* TO 'pesa_migration'@'%';
GRANT SELECT, LOCK TABLES, PROCESS, RELOAD, REPLICATION CLIENT ON *.* TO 'pesa_backup'@'%';
FLUSH PRIVILEGES;
