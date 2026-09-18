# smtp-tester

A tiny, dependency-free CLI to test SMTP connectivity: connects, does
EHLO/STARTTLS/AUTH/MAIL FROM/RCPT TO, then RSET+QUIT by default — no
message is actually sent unless you pass `--send`, in which case a real
test email is delivered to `--receiver`.

Prebuilt binaries are published as GitHub Release assets on every push to
`main`, so you can run it without installing Go.

## Quick start (no install)

Binaries are attached to the rolling [`latest` release](https://github.com/nooblk-98/smtp-tester/releases/tag/latest),
which is updated on every push to `main`.

**Linux (amd64):**
```bash
curl -fsSL -o smtp-tester https://github.com/nooblk-98/smtp-tester/releases/download/latest/smtp-tester-linux-amd64
chmod +x smtp-tester
./smtp-tester --smtphost=smtp.example.com --port=587 --sender=a@x.com --receiver=b@y.com
```

**macOS (Apple Silicon):**
```bash
curl -fsSL -o smtp-tester https://github.com/nooblk-98/smtp-tester/releases/download/latest/smtp-tester-darwin-arm64
chmod +x smtp-tester
./smtp-tester --smtphost=smtp.example.com --port=587 --sender=a@x.com --receiver=b@y.com
```

**Windows (PowerShell):**
```powershell
Invoke-WebRequest -Uri https://github.com/nooblk-98/smtp-tester/releases/download/latest/smtp-tester-windows-amd64.exe -OutFile smtp-tester.exe
.\smtp-tester.exe --smtphost=smtp.example.com --port=587 --sender=a@x.com --receiver=b@y.com
```

Available release assets:

| OS      | Arch  | File                              |
|---------|-------|------------------------------------|
| Linux   | amd64 | `smtp-tester-linux-amd64`         |
| Linux   | arm64 | `smtp-tester-linux-arm64`         |
| macOS   | amd64 | `smtp-tester-darwin-amd64`        |
| macOS   | arm64 | `smtp-tester-darwin-arm64`        |
| Windows | amd64 | `smtp-tester-windows-amd64.exe`   |

Tagged releases (`v1.2.3`, etc.) get a pinned, versioned release with the
same assets, if you'd rather not track `latest`.

## Flags

| Flag          | Default | Description                                      |
|---------------|---------|---------------------------------------------------|
| `--smtphost`  | -       | SMTP server host (required, `--host` also works)  |
| `--port`      | `587`   | SMTP server port                                   |
| `--sender`    | -       | MAIL FROM address (required)                       |
| `--receiver`  | -       | RCPT TO address (required, `--recipient` also works) |
| `--user`      | -       | Username for AUTH (optional)                       |
| `--password`  | -       | Password for AUTH (optional)                       |
| `--tls`       | `false` | Use implicit TLS (e.g. port 465)                   |
| `--starttls`  | `true`  | Upgrade with STARTTLS if offered                   |
| `--insecure`  | `false` | Skip TLS certificate verification                  |
| `--timeout`   | `10s`   | Connection/command timeout                         |
| `--verbose`   | `false` | Print each SMTP step                               |
| `--send`      | `false` | Actually deliver a test email to `--receiver` instead of RSET |
| `--version`   | -       | Print version and exit                             |

## Examples

Basic check on port 587 with STARTTLS:
```bash
./smtp-tester --smtphost=smtp.gmail.com --port=587 --sender=me@example.com --receiver=you@example.com
```

Implicit TLS on port 465 with authentication:
```bash
./smtp-tester --smtphost=smtp.gmail.com --port=465 --tls --user=me@example.com --password=secret \
  --sender=me@example.com --receiver=you@example.com
```

Verbose, self-signed/internal relay:
```bash
./smtp-tester --smtphost=mail.internal --port=25 --insecure --verbose \
  --sender=me@internal --receiver=you@internal
```

Actually send a test email:
```bash
./smtp-tester --smtphost=smtp.gmail.com --port=587 --tls=false --starttls \
  --user=me@example.com --password=secret --send \
  --sender=me@example.com --receiver=you@example.com
```

Exit code is `0` on success and `1` on failure, so it's easy to use in scripts/CI.

## Building from source

```bash
go build -o smtp-tester .
```

## How the binaries get published

`.github/workflows/build.yml` cross-compiles the CLI for Linux/macOS/Windows
on every push to `main`, then publishes the binaries as assets on the
rolling `latest` GitHub Release. Tagged pushes (`v*`) additionally create a
pinned, versioned Release with the same assets attached. The binaries are
also committed into `bin/` on `main` as a secondary way to fetch them via
`raw.githubusercontent.com`.
