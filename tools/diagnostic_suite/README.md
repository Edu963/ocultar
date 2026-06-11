# Diagnostic Suite — Adversarial Evasion Tests

This tool is a **defensive red-team harness** for the Ocultar refinery engine. It verifies that the engine correctly catches PII that has been deliberately obfuscated — it is not a bypass tool.

## What it tests

| Vector | Description |
|---|---|
| `ENCODING_BASE64` | PII encoded in Base64 before submission |
| `OBFUSCATION_UNICODE` | Email addresses using Unicode homoglyphs (fullwidth chars) |
| `PROMPT_INJECTION` | Attacker instructs the model to ignore filtering and output PII |
| `SPLITTING` | SSN split across multiple tokens to defeat regex |

Each test passes if the PII is redacted (`[PASS]`) and raises an alarm if it leaks (`[ALARM]`).

## How to run

```bash
# From the repo root
go run ./tools/diagnostic_suite/evasion.go
```

Expected output on a healthy engine:

```
--- OCULTAR ADVERSARIAL DIAGNOSTIC REPORT ---
[PASS] ENCODING_BASE64: Base64 Email Encoding -> REDACTED
[PASS] OBFUSCATION_UNICODE: Unicode Homoglyph Email -> REDACTED
[PASS] PROMPT_INJECTION: Prompt Injection Bypass -> REDACTED
[PASS] SPLITTING: PII Splitting -> REDACTED

Summary: 4 / 4 tests passed.
```

An `[ALARM]` result means the engine missed that vector and a fix is needed in the refinery before release.
