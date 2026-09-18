// Command smtp-tester checks connectivity and basic deliverability handshake
// against an SMTP server without sending a real message body.
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

const version = "1.0.0"

type config struct {
	host        string
	port        int
	sender      string
	receiver    string
	username    string
	password    string
	timeout     time.Duration
	useTLS      bool
	useStartTLS bool
	insecure    bool
	verbose     bool
}

func main() {
	cfg := parseFlags()

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ FAILED: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ SUCCESS: SMTP server accepted the full handshake.")
}

func parseFlags() config {
	var cfg config
	var showVersion bool

	flag.StringVar(&cfg.host, "smtphost", "", "SMTP server host (required)")
	flag.StringVar(&cfg.host, "host", "", "Alias for -smtphost")
	flag.IntVar(&cfg.port, "port", 587, "SMTP server port")
	flag.StringVar(&cfg.sender, "sender", "", "Sender (MAIL FROM) address (required)")
	flag.StringVar(&cfg.receiver, "receiver", "", "Receiver (RCPT TO) address (required)")
	flag.StringVar(&cfg.receiver, "recipient", "", "Alias for -receiver")
	flag.StringVar(&cfg.username, "user", "", "Username for AUTH (optional)")
	flag.StringVar(&cfg.password, "password", "", "Password for AUTH (optional)")
	flag.DurationVar(&cfg.timeout, "timeout", 10*time.Second, "Connection/command timeout")
	flag.BoolVar(&cfg.useTLS, "tls", false, "Connect using implicit TLS (e.g. port 465)")
	flag.BoolVar(&cfg.useStartTLS, "starttls", true, "Upgrade with STARTTLS if the server offers it")
	flag.BoolVar(&cfg.insecure, "insecure", false, "Skip TLS certificate verification")
	flag.BoolVar(&cfg.verbose, "verbose", false, "Print each SMTP step")
	flag.BoolVar(&showVersion, "version", false, "Print version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "smtp-tester %s\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage:\n  %s --smtphost=smtp.example.com --port=587 --sender=a@x.com --receiver=b@y.com\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if showVersion {
		fmt.Println("smtp-tester " + version)
		os.Exit(0)
	}

	var missing []string
	if cfg.host == "" {
		missing = append(missing, "-smtphost")
	}
	if cfg.sender == "" {
		missing = append(missing, "-sender")
	}
	if cfg.receiver == "" {
		missing = append(missing, "-receiver")
	}
	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "Missing required flag(s): %s\n\n", strings.Join(missing, ", "))
		flag.Usage()
		os.Exit(2)
	}

	return cfg
}

func logStep(cfg config, format string, args ...any) {
	if cfg.verbose {
		fmt.Printf(format+"\n", args...)
	}
}

func run(cfg config) error {
	addr := net.JoinHostPort(cfg.host, fmt.Sprintf("%d", cfg.port))
	dialer := net.Dialer{Timeout: cfg.timeout}

	fmt.Printf("Connecting to %s ...\n", addr)

	var conn net.Conn
	var err error

	if cfg.useTLS {
		tlsConf := &tls.Config{ServerName: cfg.host, InsecureSkipVerify: cfg.insecure}
		conn, err = tls.DialWithDialer(&dialer, "tcp", addr, tlsConf)
	} else {
		conn, err = dialer.Dial("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("connecting to %s: %w", addr, err)
	}
	defer conn.Close()

	logStep(cfg, "→ TCP connection established")

	_ = conn.SetDeadline(time.Now().Add(cfg.timeout))

	client, err := smtp.NewClient(conn, cfg.host)
	if err != nil {
		return fmt.Errorf("creating SMTP client: %w", err)
	}
	defer client.Close()

	logStep(cfg, "→ SMTP session opened, sending EHLO")

	if err := client.Hello("smtp-tester"); err != nil {
		return fmt.Errorf("EHLO/HELO failed: %w", err)
	}

	if cfg.useStartTLS && !cfg.useTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			logStep(cfg, "→ Server supports STARTTLS, upgrading connection")
			tlsConf := &tls.Config{ServerName: cfg.host, InsecureSkipVerify: cfg.insecure}
			if err := client.StartTLS(tlsConf); err != nil {
				return fmt.Errorf("STARTTLS failed: %w", err)
			}
			logStep(cfg, "→ TLS handshake complete")
		} else {
			logStep(cfg, "→ Server does not advertise STARTTLS, continuing in plaintext")
		}
	}

	if cfg.username != "" {
		logStep(cfg, "→ Authenticating as %s", cfg.username)
		auth := smtp.PlainAuth("", cfg.username, cfg.password, cfg.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("AUTH failed: %w", err)
		}
		logStep(cfg, "→ Authentication succeeded")
	}

	logStep(cfg, "→ Sending MAIL FROM:<%s>", cfg.sender)
	if err := client.Mail(cfg.sender); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	logStep(cfg, "→ Sending RCPT TO:<%s>", cfg.receiver)
	if err := client.Rcpt(cfg.receiver); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}

	logStep(cfg, "→ Resetting session (no message body sent)")
	if err := client.Reset(); err != nil {
		return fmt.Errorf("RSET failed: %w", err)
	}

	logStep(cfg, "→ Sending QUIT")
	if err := client.Quit(); err != nil {
		return fmt.Errorf("QUIT failed: %w", err)
	}

	return nil
}
