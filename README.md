# smtp-tester

A tiny, dependency-free command-line tool for testing SMTP server connectivity. It performs a real SMTP handshake — connect, `EHLO`, `STARTTLS`, `AUTH`, `MAIL FROM`, `RCPT TO` — and reports exactly where it succeeds or fails, without sending an actual email unless you ask it to.

No installation is required: prebuilt binaries for Linux, macOS, and Windows are published automatically on every push, so you can download and run a single executable on any machine.

## Why

Diagnosing "email isn't sending" problems usually means reaching for a full mail client, writing a throwaway script, or digging through application logs. `smtp-tester` isolates the SMTP layer so you can answer a narrow question fast: can this host, on this port, with these credentials, actually talk to the mail server?

- Single static binary, no runtime or dependencies to install.
- Reports failures per SMTP step (connect, `STARTTLS`, `AUTH`, `MAIL FROM`, `RCPT TO`), not just a generic timeout.
- Safe by default — it does not send a message unless you explicitly pass `--send`.
- Scriptable: exits `0` on success and `1` on failure.

## Quick start

Download the binary for your platform from the [`latest` release](https://github.com/nooblk-98/smtp-tester/releases/tag/latest) and run it — no build step required.

**Linux (amd64):**

```bash
curl -fsSL -o smtp-tester https://github.com/nooblk-98/smtp-tester/releases/download/latest/smtp-tester-linux-amd64
chmod +x smtp-tester
```

**macOS (Apple Silicon):**

```bash
curl -fsSL -o smtp-tester https://github.com/nooblk-98/smtp-tester/releases/download/latest/smtp-tester-darwin-arm64
chmod +x smtp-tester
```

**Windows (PowerShell):**

```powershell
Invoke-WebRequest -Uri https://github.com/nooblk-98/smtp-tester/releases/download/latest/smtp-tester-windows-amd64.exe -OutFile smtp-tester.exe
```

Other available assets: `smtp-tester-linux-arm64` and `smtp-tester-darwin-amd64`. Tagged releases (`v1.2.3`, ...) get a pinned, versioned release with the same assets, if you'd rather not track `latest`.

> [!TIP]
> Binaries are also committed to [`bin/`](bin/) on `main` on every build, so `raw.githubusercontent.com/nooblk-98/smtp-tester/main/bin/<file>` works as an alternative download source.

## Usage

```bash
smtp-tester --smtphost=<host> --port=<port> --sender=<from> --receiver=<to> [flags]
```

| Flag | Default | Description |
| --- | --- | --- |
| `--smtphost` | — | SMTP server host (required). `--host` is an alias. |
| `--port` | `587` | SMTP server port. |
| `--sender` | — | `MAIL FROM` address (required). |
| `--receiver` | — | `RCPT TO` address (required). `--recipient` is an alias. |
| `--user` | — | Username for `AUTH`, if the server requires authentication. |
| `--password` | — | Password for `AUTH`. |
| `--tls` | `false` | Connect using implicit TLS (typically port `465`). |
| `--starttls` | `true` | Upgrade the connection with `STARTTLS` if the server offers it. |
| `--insecure` | `false` | Skip TLS certificate verification. Useful for self-signed or internal relays. |
| `--timeout` | `10s` | Connection and command timeout. |
| `--send` | `false` | Actually deliver a test email to `--receiver` instead of stopping at `RSET`. |
| `--verbose` | `false` | Print each SMTP step as it happens. |
| `--version` | — | Print the version and exit. |

### Examples

Check a handshake over STARTTLS without sending anything:

```bash
smtp-tester --smtphost=smtp.gmail.com --port=587 --sender=me@example.com --receiver=you@example.com
```

Implicit TLS on port 465, with authentication:

```bash
smtp-tester --smtphost=smtp.gmail.com --port=465 --tls \
  --user=me@example.com --password=secret \
  --sender=me@example.com --receiver=you@example.com
```

Self-signed or internal relay, verbose output:

```bash
smtp-tester --smtphost=mail.internal --port=25 --insecure --verbose \
  --sender=me@internal --receiver=you@internal
```

Actually deliver a test email:

```bash
smtp-tester --smtphost=smtp.gmail.com --port=587 \
  --user=me@example.com --password=secret --send \
  --sender=me@example.com --receiver=you@example.com
```

> [!NOTE]
> By default `smtp-tester` stops after `RCPT TO` and sends `RSET` — no message is transmitted. Pass `--send` to have it write a minimal test email via `DATA` and actually deliver it to `--receiver`.

Exit codes are `0` on success and `1` on failure, making the tool easy to drop into scripts or CI health checks.

## Building from source

Requires Go 1.22 or later.

```bash
go build -o smtp-tester .
```

## How releases are built

[`.github/workflows/build.yml`](.github/workflows/build.yml) cross-compiles the CLI for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, and `windows/amd64` on every push to `main`, then:

1. Publishes the binaries as assets on the rolling `latest` GitHub Release.
2. Commits the binaries into [`bin/`](bin/) on `main` as a secondary distribution path.
3. On version tags (`v*`), also cuts a pinned, versioned release with the same assets.
