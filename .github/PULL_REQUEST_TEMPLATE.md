## Description
What does this PR do and why?

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Refactor
- [ ] Documentation
- [ ] CI / tooling

## Privacy Pipeline Impact
Does this PR touch the 3-step masking pipeline (custom rules → allowlist → ocultar NER)?
- [ ] No — no pipeline code changed
- [ ] Yes — described below

If yes, explain what changed and confirm fail-closed behavior is preserved:

## Checklist
- [ ] Tests added or updated for the changed behavior
- [ ] `go test ./...` passes locally
- [ ] No raw PII appears in logs, error messages, or test fixtures
- [ ] STRIDE threat impact considered (document any new attack surface below)
- [ ] DCO sign-off included (`git commit -s`)

## STRIDE Notes
Any new threats introduced or mitigated by this change:
