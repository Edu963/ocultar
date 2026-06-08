# 🚨 Incident Response: Potential Data Leak

If you suspect that PII has bypassed the OCULTAR refinery and reached an upstream AI provider, follow this playbook immediately.

---

## Phase 1: Identification & Triage
1.  **Isolate the Payload**: Identify the specific request ID from the Sombra logs.
2.  **Verify the Leak**: Check the `ai_saw` metadata (if debug mode is on) or the outgoing proxy logs to confirm if plaintext PII was sent.
3.  **Identify the Root Cause**:
    - Was it a regex failure (Tier 1)?
    - Did the SLM (Tier 2) miss the context?
    - Was there an evasion technique used (e.g., non-standard encoding)?

---

## Phase 2: Containment
1.  **Revoke Upstream Access**: If the leak is massive, rotate your OpenAI/Anthropic API keys immediately to stop further traffic.
2.  **Purge Upstream Logs**: Contact the AI provider support (OpenAI, etc.) to request a purge of the specific conversation history/logs containing the PII.
3.  **Block the Vector**: Add a temporary high-priority rule to the OCULTAR Dictionary Shield (Tier 0) to block the specific leaked entities.

---

## Phase 3: Eradication & Recovery
1.  **Update the Refinery**: Create a regression test case for the leaked PII and update the detection rules or SLM fine-tuning.
2.  **Audit the Vault**: Ensure no other similar patterns are currently bypassing the filter.
3.  **Verify the Fix**: Run the updated refinery against the offending payload in a test environment.

---

## Phase 4: Post-Mortem
1.  **Compliance Notification**: Consult your legal/DPO team to determine if a formal data breach notification is required under GDPR/CCPA.
2.  **Update the Academy**: Add a new "Lessons Learned" module to the Academy to prevent future occurrences.
3.  **Harden the Fail-Closed**: Evaluate if the system should have blocked the request entirely (Fail-Closed) and adjust settings if necessary.

---

## 📞 Emergency Contacts
- **DPO**: Your organisation's data protection officer
- **Security Team**: Your organisation's security operations team
- **OCULTAR Engineering**: edu@ocultar.dev
