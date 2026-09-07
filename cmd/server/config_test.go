package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
)

// configEnv is every variable parseConfig reads. Each test sets the ones it
// cares about and blanks the rest, so a developer's own shell cannot change
// the answer: cmp.Or treats empty as unset, which is the same as absent.
var configEnv = []string{
	"APP_DESCRIPTION", "APP_NAME", "APP_SHORTNAME", "APP_VERSION",
	"CREDENTIALS_DIRECTORY", "HOST", "MCP_CLIENT_ID", "OIDC_ISSUER", "PORT",
	"PAYMENT_DB_HOST", "PAYMENT_DB_NAME", "PAYMENT_DB_PASSWORD", "PAYMENT_DB_PORT", "PAYMENT_DB_SSLMODE", "PAYMENT_DB_USER",
	"RESERVATION_DB_HOST", "RESERVATION_DB_NAME", "RESERVATION_DB_PASSWORD", "RESERVATION_DB_PORT", "RESERVATION_DB_SSLMODE", "RESERVATION_DB_USER",
}

func emptyEnv(t *testing.T) {
	t.Helper()
	for _, name := range configEnv {
		t.Setenv(name, "")
	}
}

// ============================================================================
// Defaults
// ============================================================================

func Test_ParseConfig_With_Empty_Environment_Should_Return_Working_Defaults(t *testing.T) {
	// Arrange
	emptyEnv(t)

	// Act
	cfg, err := parseConfig(nil, io.Discard)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "port must default to 8080", cfg.Port, "8080")
	assert.That(t, "host must default to every interface", cfg.Host, "")
	assert.That(t, "app name must have a default", cfg.AppName, "Hotel Booking")
	assert.That(t, "reservation database must default", cfg.ReservationDB.Target(), "localhost:5432/reservation_db")
	assert.That(t, "payment database must default", cfg.PaymentDB.Target(), "localhost:5433/payment_db")
}

// ============================================================================
// Precedence: flags beat environment beats defaults
// ============================================================================

func Test_ParseConfig_Flag_Should_Beat_Environment_Variable(t *testing.T) {
	// This is the one that regresses silently: a wrong answer still starts.

	// Arrange
	emptyEnv(t)
	t.Setenv("PORT", "9999")

	// Act
	cfg, err := parseConfig([]string{"-port", "7777"}, io.Discard)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "the flag must win", cfg.Port, "7777")
}

func Test_ParseConfig_Environment_Should_Beat_Default(t *testing.T) {
	// Arrange
	emptyEnv(t)
	t.Setenv("PORT", "9999")
	t.Setenv("APP_NAME", "Env Named App")

	// Act
	cfg, err := parseConfig(nil, io.Discard)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "the environment must win over the default", cfg.Port, "9999")
	assert.That(t, "app name must come from the environment", cfg.AppName, "Env Named App")
}

// ============================================================================
// Validation
// ============================================================================

func Test_ParseConfig_With_Non_Numeric_Port_Should_Fail(t *testing.T) {
	// Arrange
	emptyEnv(t)

	// Act
	_, err := parseConfig([]string{"-port", "http"}, io.Discard)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "the message must name the flag", strings.Contains(err.Error(), "port"), true)
}

func Test_ParseConfig_With_Port_Above_The_Range_Should_Fail(t *testing.T) {
	// Arrange
	emptyEnv(t)

	// Act
	_, err := parseConfig([]string{"-port", "70000"}, io.Discard)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_ParseConfig_With_Non_Numeric_Database_Port_Should_Fail(t *testing.T) {
	// Arrange
	emptyEnv(t)
	t.Setenv("PAYMENT_DB_PORT", "postgres")

	// Act
	_, err := parseConfig(nil, io.Discard)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "the message must name the variable", strings.Contains(err.Error(), "PAYMENT_DB_PORT"), true)
}

func Test_ParseConfig_With_Empty_App_Name_Should_Fail(t *testing.T) {
	// Arrange
	emptyEnv(t)

	// Act
	_, err := parseConfig([]string{"-app-name", ""}, io.Discard)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_ParseConfig_With_Issuer_That_Is_Not_A_URL_Should_Fail(t *testing.T) {
	// A typo here would otherwise surface as a failed token verification,
	// which reads like an authentication bug instead of a configuration one.

	// Arrange
	emptyEnv(t)

	// Act
	_, err := parseConfig([]string{"-oidc-issuer", "keycloak:8180/realms/local"}, io.Discard)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "the message must name the flag", strings.Contains(err.Error(), "oidc-issuer"), true)
}

func Test_ParseConfig_With_Https_Issuer_Should_Succeed(t *testing.T) {
	// Arrange
	emptyEnv(t)

	// Act
	cfg, err := parseConfig([]string{"-oidc-issuer", "https://auth.example.com/realms/prod"}, io.Discard)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "issuer must be kept", cfg.OIDCIssuer, "https://auth.example.com/realms/prod")
}

// ============================================================================
// The paired setting: two contexts, two databases
// ============================================================================

func Test_ParseConfig_With_Both_Contexts_On_One_Database_Should_Fail(t *testing.T) {
	// Neither field is wrong on its own, so only a check that sees both can
	// catch it. Left alone, a payment row would answer a reservation lookup.

	// Arrange
	emptyEnv(t)
	t.Setenv("PAYMENT_DB_HOST", "localhost")
	t.Setenv("PAYMENT_DB_PORT", "5432")
	t.Setenv("PAYMENT_DB_NAME", "reservation_db")

	// Act
	_, err := parseConfig(nil, io.Discard)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "the message must name the shared database", strings.Contains(err.Error(), "localhost:5432/reservation_db"), true)
}

func Test_ParseConfig_With_Same_Host_But_Different_Database_Should_Succeed(t *testing.T) {
	// Arrange
	emptyEnv(t)
	t.Setenv("PAYMENT_DB_PORT", "5432")

	// Act
	_, err := parseConfig(nil, io.Discard)

	// Assert
	assert.That(t, "error must be nil", err, nil)
}

// ============================================================================
// Secrets
// ============================================================================

func Test_ParseConfig_Should_Prefer_The_Credentials_Directory_Over_The_Environment(t *testing.T) {
	// Arrange
	emptyEnv(t)
	dir := t.TempDir()
	writeCredential(t, dir, "reservation-db-password", "from-the-file\n")
	writeCredential(t, dir, "payment-db-password", "payment-from-the-file")
	t.Setenv("CREDENTIALS_DIRECTORY", dir)
	t.Setenv("RESERVATION_DB_PASSWORD", "from-the-environment")

	// Act
	cfg, err := parseConfig(nil, io.Discard)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "the file must win, and its newline must go", cfg.ReservationDB.Password, "from-the-file")
	assert.That(t, "the payment secret must come from its own file", cfg.PaymentDB.Password, "payment-from-the-file")
}

func Test_ParseConfig_With_Missing_Credential_File_Should_Fail(t *testing.T) {
	// A deployment that names a directory and forgets a file must not start
	// with an empty password and fail at the first query instead.

	// Arrange
	emptyEnv(t)
	t.Setenv("CREDENTIALS_DIRECTORY", t.TempDir())

	// Act
	_, err := parseConfig(nil, io.Discard)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "the message must name the credential", strings.Contains(err.Error(), "reservation-db-password"), true)
}

func Test_ParseConfig_Without_A_Credentials_Directory_Should_Fall_Back_To_The_Environment(t *testing.T) {
	// Arrange
	emptyEnv(t)
	t.Setenv("RESERVATION_DB_PASSWORD", "dev-secret")

	// Act
	cfg, err := parseConfig(nil, io.Discard)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "the environment must supply the development secret", cfg.ReservationDB.Password, "dev-secret")
}

func Test_Config_LogValue_Should_Not_Leak_A_Database_Password(t *testing.T) {
	// The allowlist in LogValue is what stops a field added next year from
	// being logged. This test fails the day somebody replaces it with a
	// blocklist and forgets an entry.

	// Arrange
	cfg := Config{
		AppName:       "Hotel Booking",
		Port:          "8080",
		ReservationDB: DatabaseConfig{Host: "db", Port: "5432", Name: "reservation_db", Password: "hunter2-reservation"},
		PaymentDB:     DatabaseConfig{Host: "db", Port: "5433", Name: "payment_db", Password: "hunter2-payment"},
	}
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Act
	logger.Info("configuration loaded", slog.Any("config", cfg))

	// Assert
	logged := buf.String()
	assert.That(t, "the reservation password must not be logged", strings.Contains(logged, "hunter2-reservation"), false)
	assert.That(t, "the payment password must not be logged", strings.Contains(logged, "hunter2-payment"), false)
	assert.That(t, "the safe fields must still be logged", strings.Contains(logged, "Hotel Booking"), true)
}

// repoFile reads a committed file at the repository root. `make ci` runs the
// gates against `git archive HEAD` in an empty directory, so a test that reads
// one of these sees the commit and not the working tree.
func repoFile(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", name))
	assert.That(t, "must be able to read "+name, err, nil)
	return string(b)
}

func Test_Deployment_Should_Not_Pass_A_Database_Password_As_An_Environment_Variable(t *testing.T) {
	// readCredential preferring a file is worth nothing while the deployment
	// still hands the binary a variable. A compose file is exactly the place
	// where one gets put back "just for now", so this test watches both files
	// that could do it.

	for _, name := range []string{"docker-compose.yml", ".env.example"} {
		for i, line := range strings.Split(repoFile(t, name), "\n") {
			// A comment may name the variable; only an assignment counts.
			stmt := strings.TrimPrefix(strings.TrimSpace(line), "- ")
			if strings.HasPrefix(stmt, "#") {
				continue
			}
			for _, secret := range []string{"RESERVATION_DB_PASSWORD", "PAYMENT_DB_PASSWORD"} {
				assigned := strings.HasPrefix(stmt, secret+"=") || strings.HasPrefix(stmt, secret+":")
				assert.That(t,
					fmt.Sprintf("%s:%d must not assign %s: the password is a file in $CREDENTIALS_DIRECTORY", name, i+1, secret),
					assigned, false)
			}
		}
	}
}

func Test_Compose_Should_Mount_One_File_Per_Secret(t *testing.T) {
	// Arrange
	compose := repoFile(t, "docker-compose.yml")

	// Assert
	assert.That(t, "the app must be told where compose mounted the secrets",
		strings.Contains(compose, "CREDENTIALS_DIRECTORY: /run/secrets"), true)
	for _, secret := range []string{"reservation-db-password", "payment-db-password"} {
		assert.That(t, "compose must source "+secret+" from a file",
			strings.Contains(compose, "file: ./secrets/"+secret), true)
		assert.That(t, "the database owning "+secret+" must read the same file",
			strings.Contains(compose, "POSTGRES_PASSWORD_FILE: /run/secrets/"+secret), true)
	}
}

func Test_DatabaseConfig_DSN_Should_Carry_Every_Field(t *testing.T) {
	// Arrange
	db := DatabaseConfig{Host: "db", Port: "5432", User: "reservation", Password: "secret", Name: "reservation_db", SSLMode: "require"}

	// Act
	dsn := db.DSN()

	// Assert
	assert.That(t, "dsn must match", dsn, "host=db port=5432 user=reservation password=secret dbname=reservation_db sslmode=require")
}

// ============================================================================
// Flag parsing outcomes
// ============================================================================

func Test_ParseConfig_With_Help_Should_Return_ErrHelp(t *testing.T) {
	// Arrange
	emptyEnv(t)
	var stderr bytes.Buffer

	// Act
	_, err := parseConfig([]string{"-h"}, &stderr)

	// Assert
	assert.That(t, "error must be flag.ErrHelp", errors.Is(err, flag.ErrHelp), true)
	assert.That(t, "usage must name the variables that are not flags", strings.Contains(stderr.String(), "CREDENTIALS_DIRECTORY"), true)
	assert.That(t, "usage must name the library's own variables", strings.Contains(stderr.String(), "KAFKA_BROKERS"), true)
}

func Test_ParseConfig_With_Unknown_Flag_Should_Return_Usage_Error(t *testing.T) {
	// Arrange
	emptyEnv(t)

	// Act
	_, err := parseConfig([]string{"-nope"}, io.Discard)

	// Assert
	assert.That(t, "error must be errUsage", errors.Is(err, errUsage), true)
}

func writeCredential(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write credential %q: %v", name, err)
	}
}
