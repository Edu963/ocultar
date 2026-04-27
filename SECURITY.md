# Security Policy

## Supported Versions

We currently provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0.0 | :x:                |

## Reporting a Vulnerability

We take the security of Ocultar seriously. If you believe you have found a security vulnerability, please report it to us as described below.

**Please do not report security vulnerabilities through public GitHub issues.**

### Disclosure Process

1.  **Report**: Send an email to security@ocultar.dev (replace with your actual security contact if different) with a description of the vulnerability.
2.  **Acknowledge**: We will acknowledge receipt of your report within 48 hours.
3.  **Investigate**: We will investigate the issue and may contact you for further information.
4.  **Fix**: We will work on a fix and coordinate a release date with you.
5.  **Release**: We will release the fix and credit you for the discovery (unless you prefer to remain anonymous).

### Guidelines

- Provide detailed steps to reproduce the vulnerability.
- Do not exploit the vulnerability beyond what is necessary for a proof of concept.
- Do not disclose the vulnerability to the public or any third party until we have had a reasonable amount of time to fix it.

## Security Architecture

Ocultar is built with a **Zero-Egress Architecture**. Key security features include:
- **Fail-Closed Design**: If a security component fails, the system blocks the request rather than leaking PII.
- **Immutable Audit Logs**: All redaction events are signed with Ed25519 and chained to prevent tampering.
- **Deterministic Tokenization**: Enables analytical utility on encrypted data without ever exposing the plaintext.
- **SSRF Protection**: Hardened validation for all egress targets to prevent internal network scanning.

For more details, see our [Technical Documentation](docs/reference).
