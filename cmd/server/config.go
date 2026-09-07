package main

import (
	"cmp"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config is every knob this binary reads. Once parseConfig returns, no code in
// this repository touches the environment again — the struct is the contract.
//
// One gap is not ours to close: cloud-native-utils reads its own variables
// directly, and they are listed under "Read by cloud-native-utils" in the usage
// text so that -h stays the complete contract anyway.
type Config struct {
	AppDescription string
	AppName        string
	AppShortname   string
	AppVersion     string
	Host           string
	MCPClientID    string
	OIDCIssuer     string
	PaymentDB      DatabaseConfig
	Port           string
	ReservationDB  DatabaseConfig
}

// DatabaseConfig is one bounded context's PostgreSQL connection. Reservation
// and Payment each own a database and never query the other's.
type DatabaseConfig struct {
	Host     string
	Name     string
	Password string // secret: a file in $CREDENTIALS_DIRECTORY wins over the environment
	Port     string
	SSLMode  string
	User     string
}

// DSN renders the connection string the pgx driver takes.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

// Target is host:port/name — what makes two DatabaseConfigs the same database.
// It holds no secret, so it is what the logs and the pair check below use.
func (d DatabaseConfig) Target() string {
	return net.JoinHostPort(d.Host, d.Port) + "/" + d.Name
}

// LogValue is what slog prints for a Config: an allowlist, so a secret added to
// the struct next year is not logged by default. A blocklist forgets.
func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("app_name", c.AppName),
		slog.String("app_version", c.AppVersion),
		slog.String("host", c.Host),
		slog.String("port", c.Port),
		slog.String("oidc_issuer", c.OIDCIssuer),
		slog.String("mcp_client_id", c.MCPClientID),
		slog.String("reservation_db", c.ReservationDB.Target()),
		slog.String("payment_db", c.PaymentDB.Target()),
	)
}

// errUsage means the FlagSet already printed what was wrong, so main only has
// to pick the exit code.
var errUsage = errors.New("usage error")

// parseConfig reads the flags, then the environment, then the built-in
// defaults — in that order, which is what cmp.Or expresses. It validates
// everything local before main opens a database or binds a listener.
func parseConfig(args []string, stderr io.Writer) (Config, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var c Config
	fs.StringVar(&c.AppDescription, "app-description", cmp.Or(os.Getenv("APP_DESCRIPTION"), "Hotel reservation and payment management system"), "description shown in the page title (env APP_DESCRIPTION)")
	fs.StringVar(&c.AppName, "app-name", cmp.Or(os.Getenv("APP_NAME"), "Hotel Booking"), "name shown in the UI (env APP_NAME)")
	fs.StringVar(&c.AppShortname, "app-shortname", cmp.Or(os.Getenv("APP_SHORTNAME"), "hotel-booking"), "short name for the MCP server and image tags (env APP_SHORTNAME)")
	fs.StringVar(&c.AppVersion, "app-version", cmp.Or(os.Getenv("APP_VERSION"), buildVersion()), "version the MCP server reports (env APP_VERSION)")
	fs.StringVar(&c.Host, "host", os.Getenv("HOST"), "bind address; empty means every interface, which is what the container needs (env HOST)")
	fs.StringVar(&c.MCPClientID, "mcp-client-id", cmp.Or(os.Getenv("MCP_CLIENT_ID"), "hotel-booking-mcp"), "OAuth client id for /mcp bearer tokens (env MCP_CLIENT_ID)")
	fs.StringVar(&c.OIDCIssuer, "oidc-issuer", cmp.Or(os.Getenv("OIDC_ISSUER"), "http://localhost:8180/realms/local"), "OIDC issuer URL (env OIDC_ISSUER)")
	fs.StringVar(&c.Port, "port", cmp.Or(os.Getenv("PORT"), "8080"), "listener port (env PORT)")

	// The variables that are not flags have nowhere else to be documented, so
	// -h names them too. Without this, -h is a partial contract.
	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage of server:\n")
		fs.PrintDefaults()
		fmt.Fprint(stderr, "\nRead from the environment only:\n"+
			"  RESERVATION_DB_HOST, RESERVATION_DB_PORT, RESERVATION_DB_USER, RESERVATION_DB_NAME, RESERVATION_DB_SSLMODE\n"+
			"  PAYMENT_DB_HOST, PAYMENT_DB_PORT, PAYMENT_DB_USER, PAYMENT_DB_NAME, PAYMENT_DB_SSLMODE\n"+
			"\tthe two bounded contexts' databases; they must not be the same one\n"+
			"  CREDENTIALS_DIRECTORY\n"+
			"\tdirectory holding one file per secret, set by the deployment. When it is\n"+
			"\tset, reservation-db-password and payment-db-password are read from it and\n"+
			"\tthe two _PASSWORD variables below are ignored\n"+
			"  RESERVATION_DB_PASSWORD, PAYMENT_DB_PASSWORD\n"+
			"\tthe development fallback for those two secrets\n"+
			"\nRead by cloud-native-utils, not by this struct:\n"+
			"  OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, OIDC_REDIRECT_URL, REDIRECT_URL\n"+
			"  KAFKA_BROKERS, KAFKA_CONSUMER_GROUP_ID\n"+
			"  SERVICE_BREAKER_THRESHOLD, SERVICE_DEBOUNCE_PER_SEC, SERVICE_RETRY_DELAY, SERVICE_RETRY_MAX, SERVICE_TIMEOUT\n")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Config{}, err // -h: usage printed, exit 0
		}
		return Config{}, errUsage // the FlagSet already said what was wrong
	}

	// Not flags: a deployment sets these, and twelve more flags would bury the
	// eight above.
	c.ReservationDB = DatabaseConfig{
		Host:    cmp.Or(os.Getenv("RESERVATION_DB_HOST"), "localhost"),
		Name:    cmp.Or(os.Getenv("RESERVATION_DB_NAME"), "reservation_db"),
		Port:    cmp.Or(os.Getenv("RESERVATION_DB_PORT"), "5432"),
		SSLMode: cmp.Or(os.Getenv("RESERVATION_DB_SSLMODE"), "disable"),
		User:    cmp.Or(os.Getenv("RESERVATION_DB_USER"), "reservation"),
	}
	c.PaymentDB = DatabaseConfig{
		Host:    cmp.Or(os.Getenv("PAYMENT_DB_HOST"), "localhost"),
		Name:    cmp.Or(os.Getenv("PAYMENT_DB_NAME"), "payment_db"),
		Port:    cmp.Or(os.Getenv("PAYMENT_DB_PORT"), "5433"),
		SSLMode: cmp.Or(os.Getenv("PAYMENT_DB_SSLMODE"), "disable"),
		User:    cmp.Or(os.Getenv("PAYMENT_DB_USER"), "payment"),
	}

	// Cheap checks first: a typo in a flag should not wait on a file read.
	if c.AppName == "" {
		return Config{}, errors.New("app-name is empty: set -app-name or APP_NAME")
	}
	if err := validatePort("port", c.Port); err != nil {
		return Config{}, err
	}
	if err := validatePort("RESERVATION_DB_PORT", c.ReservationDB.Port); err != nil {
		return Config{}, err
	}
	if err := validatePort("PAYMENT_DB_PORT", c.PaymentDB.Port); err != nil {
		return Config{}, err
	}
	if err := validateIssuer(c.OIDCIssuer); err != nil {
		return Config{}, err
	}

	// The two databases are one setting, not two: pointing both contexts at the
	// same database silently merges them, and the first sign is a payment row
	// answering a reservation lookup. Neither field is wrong on its own, so
	// nothing above can catch it.
	if c.ReservationDB.Target() == c.PaymentDB.Target() {
		return Config{}, fmt.Errorf("both bounded contexts point at %s: give the payment context its own database, or change PAYMENT_DB_NAME", c.PaymentDB.Target())
	}

	reservationPassword, err := readCredential("reservation-db-password", "RESERVATION_DB_PASSWORD")
	if err != nil {
		return Config{}, err
	}
	c.ReservationDB.Password = reservationPassword

	paymentPassword, err := readCredential("payment-db-password", "PAYMENT_DB_PASSWORD")
	if err != nil {
		return Config{}, err
	}
	c.PaymentDB.Password = paymentPassword

	return c, nil
}

// validatePort keeps the parsed value where it belongs — at the edge — so no
// later code re-parses it and fails somewhere less useful.
func validatePort(name, value string) error {
	if _, err := strconv.ParseUint(value, 10, 16); err != nil {
		return fmt.Errorf("%s %q: want a number from 0 to 65535", name, value)
	}
	return nil
}

// validateIssuer checks the OIDC issuer here rather than at the first request.
// A typo would otherwise surface as a failed token verification, which reads
// like an authentication problem instead of a configuration one.
func validateIssuer(issuer string) error {
	u, err := url.Parse(issuer)
	if err != nil {
		return fmt.Errorf("oidc-issuer %q: %w", issuer, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("oidc-issuer %q: want an http or https URL, for example http://localhost:8180/realms/local", issuer)
	}
	if u.Host == "" {
		return fmt.Errorf("oidc-issuer %q: no host", issuer)
	}
	return nil
}

// readCredential returns a secret, preferring a file over the environment.
//
// A secret belongs in a file: a file is not inherited by every child process
// and does not appear in a process listing, which is true of neither a flag
// value nor an environment variable. The deployment puts one file per secret in
// $CREDENTIALS_DIRECTORY: docker-compose.yml mounts them at /run/secrets, and
// .env points `make run` at ./secrets, so both paths read files. The
// environment fallback below is what is left for a machine that sets neither.
//
// An empty result is not an error: the binary must start with an empty
// environment, and the database says clearly enough what a missing password
// costs at the first query.
func readCredential(file, envVar string) (string, error) {
	dir := os.Getenv("CREDENTIALS_DIRECTORY")
	if dir == "" {
		return os.Getenv(envVar), nil
	}
	b, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return "", fmt.Errorf("reading credential %q from $CREDENTIALS_DIRECTORY: %w", file, err)
	}
	// Credential files usually end in a newline.
	return strings.TrimSpace(string(b)), nil
}
