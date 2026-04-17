# Changelog

All notable changes to gosre-cli are documented here.
Format: [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [0.1.0] - 2026-04-16

### Added
- HTTP checker with status code and latency capture
- TCP checker with dial timeout
- DNS checker with A, AAAA, CNAME, MX record type support
- TLS checker with certificate expiry validation (configurable threshold, default 14 days)
- Table and JSON output formatters
- Config file support (`~/.gosre.yaml`) via Viper
- `--target-name` flag for named target resolution from config file
- `gosre targets list` command
- `--output`, `--quiet`, `--timeout` global flags
