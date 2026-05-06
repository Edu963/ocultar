import { Link } from "react-router-dom";
import {
  Cpu, Network, Activity, Server, ShieldCheck, Database, Terminal,
  ChevronRight, Layout, MessageSquare, FileText, Globe, Radio, Key,
  Box, GitMerge, ClipboardCheck, Fingerprint,
} from "lucide-react";
import { SiteNav } from "@/components/site/SiteNav";
import { SiteFooter } from "@/components/site/SiteFooter";
import { SectionHeader } from "@/components/site/ProblemSection";

const MODELS = [
  { name: "GPT-4o", provider: "OpenAI", file: "router/openai.go" },
  { name: "Mistral Large", provider: "Mistral AI", file: "router/openai.go" },
  { name: "Claude Sonnet", provider: "Anthropic", file: "router/claude.go" },
  { name: "Gemini Flash", provider: "Google", file: "router/gemini.go" },
  { name: "Local SLM", provider: "privacy-filter", file: "apps/slm-engine/main.go" },
];

const CONNECTORS = [
  {
    name: "Slack Events API",
    icon: MessageSquare,
    desc: "End-to-end agentic Slack bot with fail-closed PII redaction on inbound messages and secure re-hydration on AI responses before delivery.",
    tag: "Native",
    accent: true,
    source: "pkg/handler/slack_app.go",
  },
  {
    name: "SharePoint via MS Graph",
    icon: FileText,
    desc: "Polls SharePoint document libraries via Microsoft Graph OAuth2. Enterprise-licensed — all document content is processed through the Refinery before any downstream action.",
    tag: "Enterprise",
    accent: false,
    source: "services/refinery/pkg/connector/sharepoint.go",
  },
  {
    name: "Generic REST API",
    icon: Globe,
    desc: "Zero-config adapter for any authenticated REST endpoint. Bearer tokens and API keys, with per-connector data policy enforcement and model allowlists.",
    tag: "Native",
    accent: true,
    source: "pkg/connector/api_connector.go",
  },
  {
    name: "File Connector",
    icon: Box,
    desc: "Processes uploaded documents through the Refinery before storage or forwarding. Supports JSON, CSV, and raw text.",
    tag: "Native",
    accent: true,
    source: "pkg/connector/file_connector.go",
  },
];

const POLICY_MATRIX = [
  { connector: "file", strip: ["SSN", "ACCOUNT_NUMBER"], models: ["local-slm", "gpt-4o", "gemini-flash-latest"], limit: "10 MB" },
  { connector: "banking", strip: ["SSN", "ROUTING_NUMBER"], models: ["local-slm"], limit: "10 MB" },
  { connector: "slack-prod", strip: ["SSN", "CREDENTIAL", "SECRET", "PHONE_NUMBER"], models: ["gemini-flash-latest"], limit: "—" },
];

const OPENAI_COMPAT_STEPS = [
  { before: "https://api.openai.com/v1", after: "http://localhost:8086/v1", label: "Base URL" },
  { before: "Authorization: Bearer sk-...", after: "Authorization: Bearer sk-...", label: "Auth header (unchanged)" },
  { before: "model: gpt-4o", after: "model: gpt-4o", label: "Model name (unchanged)" },
];

const REFINERY_TIERS = [
  { tier: "Tier 0.1", name: "Evasion Shield", desc: "Recursively decodes Base64, JWT, and URL-encoded payloads and rescans the decoded content. Catches attackers who wrap PII in encoding layers to bypass downstream filters." },
  { tier: "Tier 0", name: "Custom Dictionary Engine", desc: "Admin-configurable blocklists for VIP names, internal codenames, and org-specific sensitive terms. Backed by a live CRM/LDAP sync that updates the dictionary without restart." },
  { tier: "Tier 1", name: "Deterministic Regex Pipeline", desc: "30+ pre-compiled, Luhn-validated patterns: SSNs, IBANs, credit cards, phone numbers, addresses, SIRET/SIREN, health references, and regional IDs across 10+ countries." },
  { tier: "Tier 1.5", name: "Contextual Heuristics", desc: "libphonenumber validation, heuristic address parsing, greeting/signature detection. Catches PII that regex misses without any model inference overhead." },
  { tier: "Tier 2", name: "EU-Sovereign Deep Scan (SLM)", desc: "A local bidirectional token classifier (openai/privacy-filter) performs NER for contextual PII. Domain-specific fine-tunes available (fr-finance). ~65% F1 on financial corpora; roadmap targets >90% via fine-tuning." },
];

const ENTERPRISE_FEATURES = [
  { icon: Radio, title: "Syslog → Any Upstream SIEM", desc: "A fail-closed UDP Syslog proxy strips PII from every log line before forwarding to any compliant SIEM (Splunk, Elastic, Wazuh). If the Refinery errors, the line is dropped." },
  { icon: Activity, title: "Ed25519-Signed Immutable Audit Log", desc: "Every vault transaction is written to a SHA-256 hash-chained, Ed25519-signed append-only log. Auditors verify the full chain offline. GDPR Article 5(2) compliant." },
  { icon: Database, title: "PostgreSQL HA Vault", desc: "Production deployments replace the local DuckDB vault with a shared PostgreSQL backend, enabling multi-node Refinery pools, read replicas, and encrypted WAL." },
  { icon: Globe, title: "CRM / LDAP Identity Sync", desc: "Background polling ingests protected identities directly from your CRM or LDAP into Tier 0's dictionary — without any manual list management." },
  { icon: Key, title: "AES-256-GCM Vault Encryption", desc: "Every PII token is encrypted at rest with AES-256-GCM, key derived via HKDF-SHA256. Master key is operator-controlled and never leaves process memory." },
  { icon: Server, title: "License-Gated at the Binary Level", desc: "Enterprise connectors and Tier 2 NER require a signed Ed25519 license payload. Unlicensed access is blocked before any code path executes." },
];

const ProductEyebrow = ({ icon: Icon, label }: { icon: any; label: string }) => (
  <div className="flex items-center gap-2 text-primary">
    <Icon className="h-4 w-4" />
    <p className="font-mono text-[11px] font-semibold uppercase tracking-[0.2em]">{label}</p>
  </div>
);

const Solutions = () => {
  return (
    <main className="min-h-screen bg-background text-foreground">
      <SiteNav />

      {/* Hero */}
      <section className="relative overflow-hidden pt-32 pb-20 md:pt-40 md:pb-24 border-b border-border">
        <div className="pointer-events-none absolute inset-0 bg-grid bg-grid-fade opacity-50" />
        <div className="pointer-events-none absolute inset-x-0 top-0 h-[520px] bg-gradient-hero" />
        <div className="container-page relative flex flex-col items-center text-center">
          <span className="pill pill-accent">
            <span className="h-1.5 w-1.5 rounded-full bg-primary animate-pulse-dot" />
            Product Suite
          </span>
          <h1 className="mt-7 max-w-3xl text-balance text-[44px] sm:text-[56px] md:text-[64px] font-semibold leading-[1.04] tracking-tightest text-gradient">
            Three products. One sovereign boundary.
          </h1>
          <p className="mt-6 max-w-2xl text-[17px] leading-relaxed text-muted-foreground">
            Every capability listed here runs in your infrastructure. No external dependencies, no telemetry, no exfiltration paths.
          </p>
        </div>
      </section>

      {/* Pillar 1 — Refinery */}
      <section className="section border-b border-border bg-surface/30">
        <div className="container-page grid grid-cols-1 lg:grid-cols-2 gap-16 items-start">
          <div className="flex flex-col gap-6">
            <ProductEyebrow icon={Cpu} label="Product 01" />
            <h2 className="text-[36px] sm:text-[44px] font-semibold tracking-tight text-foreground text-balance">
              OCULTAR Refinery
            </h2>
            <h3 className="text-[18px] text-muted-foreground">The high-throughput privacy engine</h3>
            <p className="text-[15px] leading-relaxed text-muted-foreground">
              A 4-tier detection pipeline that escalates from deterministic pattern matching to local neural inference — near-zero false negatives, sub-millisecond latency for most payloads.
            </p>
            <div className="flex flex-col gap-2.5 pt-2">
              {[
                "AES-256-GCM encryption for every vault entry",
                "Concurrent batch processing — bounded 100-worker pool",
                "Session cache eliminates redundant lookups within a run",
                "Base64, URL-encoded, nested JSON evasion detection",
                "Fail-Closed: any batch error blocks the entire payload",
              ].map((f) => (
                <div key={f} className="flex items-start gap-3 text-[14px] text-muted-foreground">
                  <ShieldCheck className="h-4 w-4 text-primary shrink-0 mt-0.5" />
                  {f}
                </div>
              ))}
            </div>

            <div className="card-surface mt-8 p-6">
              <div className="flex items-center justify-between pb-4 mb-4 border-b border-border">
                <span className="mono-label">Refinery Flow — Fail-Closed Batch</span>
                <Terminal className="h-4 w-4 text-primary" />
              </div>
              <pre className="font-mono text-[12px] leading-relaxed overflow-x-auto text-muted-foreground">
{`// RefineBatch — bounded 100-worker goroutine pool
for _, item := range items {
    go func(idx int, val interface{}) {
        res, err := e.`}<span className="text-primary">ProcessInterface</span>{`(val, actor)
        // ...
    }(i, item)
}

// Fail-Closed: any single error blocks the entire batch
for _, err := range errs {
    if err != nil {
        return nil, fmt.Errorf(`}<span className="text-primary">"secure block: %w"</span>{`, err)
    }
}`}
              </pre>
            </div>
          </div>

          <div className="flex flex-col gap-3">
            {REFINERY_TIERS.map((t) => (
              <div key={t.tier} className="card-surface card-surface-hover p-5 flex flex-col gap-2.5">
                <div className="flex items-center gap-3">
                  <span className="font-mono text-[10px] font-semibold uppercase tracking-[0.18em] text-primary bg-primary/10 border border-primary/20 px-2.5 py-1 rounded-full">
                    {t.tier}
                  </span>
                  <span className="text-[14px] font-semibold text-foreground">{t.name}</span>
                </div>
                <p className="text-[13.5px] leading-relaxed text-muted-foreground">{t.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Pillar 2 — Sombra */}
      <section className="section border-b border-border">
        <div className="container-page flex flex-col gap-20">
          <div className="flex flex-col gap-4 max-w-3xl">
            <ProductEyebrow icon={Network} label="Product 02" />
            <h2 className="text-[36px] sm:text-[44px] font-semibold tracking-tight text-foreground text-balance">
              OCULTAR Sombra
            </h2>
            <h3 className="text-[18px] text-muted-foreground">The agentic privacy gateway</h3>
            <p className="text-[15px] leading-relaxed text-muted-foreground">
              Sombra sits between your infrastructure and every AI provider. It intercepts, sanitizes, routes, and re-hydrates — only redacted prompts are ever transmitted outbound. Fail-closed by construction.
            </p>
          </div>

          {/* Multi-Model Router */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
            <div className="flex flex-col gap-4">
              <span className="mono-label text-primary">Multi-Model Router</span>
              <p className="text-[15px] leading-relaxed text-muted-foreground">
                A pluggable <code className="font-mono text-[12px] text-primary bg-primary/10 px-1.5 py-0.5 rounded">ModelAdapter</code> interface dispatches sanitized prompts to any registered backend. Zero-Egress domain validation per-adapter — unregistered domains blocked before transmission.
              </p>
              <p className="text-[13px] font-mono text-muted-foreground">
                <span className="text-primary/80">Drop-in OpenAI-compatible proxy</span> — existing SDKs need zero code changes.
              </p>
            </div>
            <div className="card-surface p-6">
              <div className="mono-label pb-4 mb-2 border-b border-border">Registered Adapters</div>
              <div className="flex flex-col">
                {MODELS.map((m) => (
                  <div key={m.name} className="flex items-center justify-between py-3.5 border-b border-border last:border-0">
                    <div className="flex flex-col gap-0.5">
                      <div className="text-[14px] font-semibold text-foreground">{m.name}</div>
                      <div className="text-[11px] font-mono text-muted-foreground">{m.file}</div>
                    </div>
                    <span className="text-[11px] font-mono text-muted-foreground bg-surface-elevated border border-border px-2.5 py-1 rounded-full">
                      {m.provider}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Connectors */}
          <div className="flex flex-col gap-8">
            <SectionHeader
              eyebrow="Data Source Connectors"
              title="Per-connector data policies, enforced at the gateway."
              description="Each connector declares a DataPolicy that governs which PII categories are stripped, which models may receive its data, and size limits."
            />
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {CONNECTORS.map((c) => (
                <div key={c.name} className="card-surface card-surface-hover p-6 flex flex-col gap-4">
                  <div className="flex items-start justify-between">
                    <div className="flex h-10 w-10 items-center justify-center rounded-md bg-primary/10 border border-primary/20 text-primary">
                      <c.icon className="h-5 w-5" />
                    </div>
                    <span
                      className={`text-[10px] font-mono font-semibold uppercase tracking-[0.18em] px-2.5 py-1 rounded-full border ${
                        c.accent
                          ? "border-primary/20 bg-primary/10 text-primary"
                          : "border-border bg-surface-elevated text-muted-foreground"
                      }`}
                    >
                      {c.tag}
                    </span>
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <h3 className="text-[15px] font-semibold text-foreground">{c.name}</h3>
                    <p className="text-[13.5px] leading-relaxed text-muted-foreground">{c.desc}</p>
                  </div>
                  <div className="text-[11px] font-mono text-muted-foreground pt-3 border-t border-border">{c.source}</div>
                </div>
              ))}
            </div>
          </div>

          {/* Policy Matrix */}
          <div className="flex flex-col gap-6">
            <div className="flex flex-col gap-3">
              <ProductEyebrow icon={ClipboardCheck} label="Per-Connector Policy Enforcement" />
              <p className="text-[15px] leading-relaxed text-muted-foreground max-w-2xl">
                Policies are enforced at the gateway layer — before a single byte reaches the Refinery. Sensitive sources are hard-restricted to local inference; no configuration drift can bypass this.
              </p>
            </div>
            <div className="card-surface overflow-x-auto">
              <table className="w-full text-[13px]">
                <thead>
                  <tr className="bg-surface-elevated border-b border-border">
                    <th className="text-left px-5 py-3.5 mono-label">Connector</th>
                    <th className="text-left px-5 py-3.5 mono-label">Strip Categories</th>
                    <th className="text-left px-5 py-3.5 mono-label">Allowed Models</th>
                    <th className="text-left px-5 py-3.5 mono-label">Size Limit</th>
                  </tr>
                </thead>
                <tbody>
                  {POLICY_MATRIX.map((row) => (
                    <tr key={row.connector} className="border-b border-border last:border-0 hover:bg-surface/50 transition-colors">
                      <td className="px-5 py-4 font-mono text-[12px] text-primary">{row.connector}</td>
                      <td className="px-5 py-4">
                        <div className="flex flex-wrap gap-1.5">
                          {row.strip.map((cat) => (
                            <span key={cat} className="text-[10px] font-mono font-semibold uppercase tracking-wide px-2 py-0.5 rounded bg-destructive/10 text-destructive border border-destructive/20">
                              {cat}
                            </span>
                          ))}
                        </div>
                      </td>
                      <td className="px-5 py-4">
                        <div className="flex flex-wrap gap-1.5">
                          {row.models.map((m) => (
                            <span key={m} className="text-[11px] font-mono px-2 py-0.5 rounded bg-surface-elevated text-muted-foreground border border-border">
                              {m}
                            </span>
                          ))}
                        </div>
                      </td>
                      <td className="px-5 py-4 font-mono text-[12px] text-muted-foreground">{row.limit}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <p className="text-[11px] font-mono text-muted-foreground">
              source: apps/sombra/configs/sombra.yaml · pkg/connector/connector.go
            </p>
          </div>

          {/* OpenAI Compat */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
            <div className="flex flex-col gap-4">
              <ProductEyebrow icon={GitMerge} label="Drop-in OpenAI Compatibility" />
              <p className="text-[15px] leading-relaxed text-muted-foreground">
                Sombra exposes a <code className="font-mono text-[12px] text-primary bg-primary/10 px-1.5 py-0.5 rounded">/v1/chat/completions</code> endpoint that is wire-compatible with the OpenAI API. Change one URL — every existing SDK, agent, or tool works without modification.
              </p>
              <p className="text-[15px] leading-relaxed text-muted-foreground">
                All traffic is scrubbed by the Refinery before dispatch. Responses are optionally rehydrated before returning to the caller.
              </p>
              <p className="text-[11px] font-mono text-muted-foreground">
                source: apps/sombra/pkg/handler/handler.go · HandleV1ChatCompletions
              </p>
            </div>
            <div className="card-surface p-6 font-mono text-[12px]">
              <div className="mono-label pb-4 mb-2 border-b border-border">Migration — one line changes</div>
              {OPENAI_COMPAT_STEPS.map((step, i) => (
                <div key={i} className="py-3 border-b border-border last:border-0 flex flex-col gap-1.5">
                  <div className="text-[10px] uppercase tracking-[0.2em] text-muted-foreground">{step.label}</div>
                  <div className="flex items-center gap-2">
                    <span className="text-destructive/60">−</span>
                    <span className={`text-muted-foreground ${step.before !== step.after ? "line-through decoration-destructive/40" : ""}`}>
                      {step.before}
                    </span>
                  </div>
                  {step.before !== step.after && (
                    <div className="flex items-center gap-2">
                      <span className="text-primary/70">+</span>
                      <span className="text-primary">{step.after}</span>
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>

          {/* Audit Trail */}
          <div className="relative overflow-hidden rounded-xl border border-primary/20 bg-surface/50 p-8 md:p-10 flex flex-col gap-6">
            <div className="absolute inset-0 bg-[radial-gradient(ellipse_60%_80%_at_80%_20%,hsl(var(--primary)/0.06),transparent)]" />
            <div className="relative z-10 flex flex-col gap-4 max-w-2xl">
              <ProductEyebrow icon={Fingerprint} label="Immutable Audit Trail" />
              <h3 className="text-[24px] sm:text-[28px] font-semibold tracking-tight text-foreground text-balance">
                Every vault event is cryptographically signed and hash-chained.
              </h3>
              <p className="text-[15px] leading-relaxed text-muted-foreground">
                Sombra initializes an <code className="font-mono text-[12px] text-primary bg-primary/10 px-1.5 py-0.5 rounded">ImmutableLogger</code> backed by an ephemeral Ed25519 keypair on startup. Every event carries the SHA-256 hash of the previous entry — a tamper-evident chain. The public key is printed at boot; auditors verify the entire log offline.
              </p>
            </div>
            <div className="relative z-10 grid grid-cols-1 md:grid-cols-3 gap-3">
              {[
                { label: "Signature", value: "Ed25519", sub: "FIPS 186-5 compliant" },
                { label: "Chain", value: "SHA-256 hash chain", sub: "Append-only JSON log" },
                { label: "Verification", value: "Offline / Air-gapped", sub: "No infra access required" },
              ].map((item) => (
                <div key={item.label} className="card-surface p-4 flex flex-col gap-1">
                  <div className="mono-label">{item.label}</div>
                  <div className="text-[14px] font-semibold text-foreground">{item.value}</div>
                  <div className="text-[11px] font-mono text-muted-foreground">{item.sub}</div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* Pillar 3 — Enterprise */}
      <section className="section border-b border-border bg-surface/30 relative overflow-hidden">
        <div className="pointer-events-none absolute bottom-0 left-1/2 -translate-x-1/2 w-full h-[500px] bg-primary/5 blur-[140px] rounded-full" />
        <div className="container-page relative z-10 flex flex-col gap-12">
          <div className="max-w-3xl flex flex-col gap-4">
            <ProductEyebrow icon={Layout} label="Product 03" />
            <h2 className="text-[36px] sm:text-[44px] font-semibold tracking-tight text-foreground text-balance">
              OCULTAR Enterprise
            </h2>
            <h3 className="text-[18px] text-muted-foreground">Governance, audit, and sovereign operations</h3>
            <p className="text-[15px] leading-relaxed text-muted-foreground">
              Extended capabilities, license-gated at the binary level. Enterprise unlocks deep-scan NER, the SharePoint connector, structured audit logging, and SIEM-compatible syslog forwarding.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {ENTERPRISE_FEATURES.map((f) => (
              <div key={f.title} className="card-surface card-surface-hover p-6 flex flex-col gap-4">
                <div className="h-11 w-11 bg-primary/10 border border-primary/20 rounded-md flex items-center justify-center text-primary">
                  <f.icon className="h-5 w-5" />
                </div>
                <h3 className="text-[16px] font-semibold text-foreground">{f.title}</h3>
                <p className="text-[14px] leading-relaxed text-muted-foreground">{f.desc}</p>
              </div>
            ))}
          </div>

          <div className="card-surface p-6 max-w-3xl">
            <div className="mono-label pb-4 mb-4 border-b border-border">syslog.go — SIEM Forward Architecture</div>
            <pre className="font-mono text-[12px] leading-relaxed overflow-x-auto text-muted-foreground">
{`// SyslogServer: PII-scrub every log line before SIEM delivery
refined, err := s.eng.`}<span className="text-primary">RefineString</span>{`(msg, `}<span className="text-primary">"syslog_proxy"</span>{`, nil)
if err != nil {
    // Fail-Closed: drop on refinery error — never forward raw
    continue
}
if upstream != nil {
    // Forward clean message to any UDP SIEM (Splunk HEC, Elastic, etc.)
    upstreamConn.`}<span className="text-primary">Write</span>{`([]byte(refined))
}`}
            </pre>
            <p className="text-[11px] font-mono text-muted-foreground mt-4 pt-4 border-t border-border">
              SIEM-agnostic via standard UDP. Tested with any compliant upstream endpoint.
            </p>
          </div>

          <div className="pt-4 flex justify-center">
            <Link
              to="/"
              className="group inline-flex h-11 items-center gap-2 rounded-md bg-foreground px-6 text-[13px] font-semibold text-background hover:bg-foreground/90 transition-colors"
            >
              Request enterprise access
              <ChevronRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5" />
            </Link>
          </div>
        </div>
      </section>

      {/* Status bar */}
      <section className="py-6 border-t border-border bg-background">
        <div className="container-page flex flex-col md:flex-row justify-between items-center gap-3 mono-label">
          <div className="flex items-center gap-6">
            <span className="flex items-center gap-2 text-primary">
              <span className="h-1.5 w-1.5 rounded-full bg-primary animate-pulse-dot" />
              Sombra_Gateway_Active
            </span>
            <span>Build: 4.30.26</span>
          </div>
          <div className="flex gap-8">
            <span>Features sourced from /pkg</span>
            <span className="text-primary/70">Fail_Closed_By_Default</span>
          </div>
        </div>
      </section>

      <SiteFooter />
    </main>
  );
};

export default Solutions;
