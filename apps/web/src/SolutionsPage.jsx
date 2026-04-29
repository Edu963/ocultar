import { Link } from 'react-router-dom';
import {
    Shield, Lock, Zap, Network, Cpu, Activity, Server,
    ShieldCheck, Database, Terminal, ChevronRight, Layout,
    MessageSquare, FileText, Globe, Radio, Key, Box,
    GitMerge, ClipboardCheck, Fingerprint
} from 'lucide-react';

// ---- Data (all verified against codebase) ----

const MODELS = [
    { name: "GPT-4o", provider: "OpenAI", file: "router/openai.go" },
    { name: "Claude 3.x", provider: "Anthropic", file: "router/claude.go" },
    { name: "Gemini 1.5 Pro", provider: "Google", file: "router/gemini.go" },
    { name: "Local SLM", provider: "llama.cpp", file: "router/local.go" },
];

const CONNECTORS = [
    {
        name: "Slack Events API",
        icon: <MessageSquare className="w-5 h-5" />,
        desc: "End-to-end agentic Slack bot with fail-closed PII redaction on inbound messages and secure re-hydration on AI responses before delivery.",
        tag: "Native",
        tagColor: "cyan",
        source: "pkg/handler/slack_app.go"
    },
    {
        name: "SharePoint via MS Graph",
        icon: <FileText className="w-5 h-5" />,
        desc: "Polls SharePoint document libraries via Microsoft Graph OAuth2. Enterprise-licensed capability — all document content is processed through the Refinery before any downstream action.",
        tag: "Enterprise",
        tagColor: "sky",
        source: "services/refinery/pkg/connector/sharepoint.go"
    },
    {
        name: "Generic REST API",
        icon: <Globe className="w-5 h-5" />,
        desc: "Zero-config adapter for any authenticated REST endpoint. Supports Bearer tokens and API keys, with per-connector data policy enforcement and model allowlists.",
        tag: "Native",
        tagColor: "cyan",
        source: "pkg/connector/api_connector.go"
    },
    {
        name: "File Connector",
        icon: <Box className="w-5 h-5" />,
        desc: "Processes uploaded documents through the Refinery before storage or forwarding. Supports JSON, CSV, and raw text.",
        tag: "Native",
        tagColor: "cyan",
        source: "pkg/connector/file_connector.go"
    },
];

const POLICY_MATRIX = [
    {
        connector: "file",
        strip: ["SSN", "ACCOUNT_NUMBER"],
        models: ["local-slm", "gpt-4o", "gemini-flash-latest"],
        limit: "10 MB",
    },
    {
        connector: "banking",
        strip: ["SSN", "ROUTING_NUMBER"],
        models: ["local-slm"],
        limit: "10 MB",
    },
    {
        connector: "slack-prod",
        strip: ["SSN", "CREDENTIAL", "SECRET", "PHONE_NUMBER"],
        models: ["gemini-flash-latest"],
        limit: "—",
    },
];

const OPENAI_COMPAT_STEPS = [
    { before: "https://api.openai.com/v1", after: "http://localhost:8086/v1", label: "Base URL" },
    { before: "Authorization: Bearer sk-...", after: "Authorization: Bearer sk-...", label: "Auth header (unchanged)" },
    { before: "model: gpt-4o", after: "model: gpt-4o", label: "Model name (unchanged)" },
];

const REFINERY_TIERS = [
    { tier: "Tier 0.1", name: "Evasion Shield", desc: "Recursively decodes Base64, JWT, and URL-encoded payloads and rescans the decoded content. Catches attackers who wrap PII in encoding layers to bypass downstream filters." },
    { tier: "Tier 0", name: "Custom Dictionary Engine", desc: "Admin-configurable blocklists for VIP names, internal codenames, and org-specific sensitive terms. Backed by a live CRM/LDAP sync that updates the dictionary without restart." },
    { tier: "Tier 1", name: "Deterministic Regex Pipeline", desc: "30+ pre-compiled, Luhn-validated patterns: SSNs, IBANs, credit cards, phone numbers, addresses, SIRET/SIREN, health references, and regional IDs across 10+ countries. Credit card false positives are eliminated with Luhn checksum verification." },
    { tier: "Tier 1.5", name: "Contextual Heuristics", desc: "Phone libphonenumber validation, heuristic address parsing, greeting/signature detection (name in 'Dear John' or 'Best, Sarah'). Catches PII that regex misses without any model inference overhead." },
    { tier: "Tier 2", name: "OpenAI Privacy Filter (Deep Scan)", desc: "A 1.5B-parameter local bidirectional token classifier (Apache 2.0) performs Named-Entity Recognition for contextual PII — conversational names, implicit references, and finance-specific entities. 97% F1. Runs entirely in your infrastructure. Swappable with any compatible model via TIER2_ENGINE env var." },
];

const ENTERPRISE_FEATURES = [
    {
        icon: <Radio className="w-6 h-6" />,
        title: "Syslog → Any Upstream SIEM",
        desc: "A fail-closed UDP Syslog proxy strips PII from every log line before forwarding to any compliant SIEM (Splunk, Elastic, Wazuh). If the Refinery errors, the log line is dropped — raw PII is never forwarded.",
    },
    {
        icon: <Activity className="w-6 h-6" />,
        title: "Ed25519-Signed Immutable Audit Log",
        desc: "Every vault transaction is written to a SHA-256 hash-chained, Ed25519-signed append-only log. Auditors verify the full chain offline — no infrastructure access required. GDPR Article 5(2) compliant out of the box.",
    },
    {
        icon: <Database className="w-6 h-6" />,
        title: "PostgreSQL HA Vault",
        desc: "Production deployments replace the local DuckDB vault with a shared PostgreSQL backend, enabling multi-node Refinery pools, read replicas, and encrypted WAL for point-in-time recovery.",
    },
    {
        icon: <Globe className="w-6 h-6" />,
        title: "CRM / LDAP Identity Sync",
        desc: "Background polling ingests protected identities directly from your CRM or LDAP directory into Tier 0's dictionary — ensuring executive names, codenames, and customer IDs are blocked without any manual list management.",
    },
    {
        icon: <Key className="w-6 h-6" />,
        title: "AES-256-GCM Vault Encryption",
        desc: "Every PII token is encrypted at rest with AES-256-GCM and a key derived via HKDF-SHA256. The master key is operator-controlled and never leaves your process memory — not stored in config, not logged.",
    },
    {
        icon: <Server className="w-6 h-6" />,
        title: "License-Gated at the Binary Level",
        desc: "Enterprise connectors and Tier 2 NER require a signed Ed25519 license payload. Unlicensed access is blocked before any code path executes — the binary itself enforces the boundary.",
    },
];

export default function SolutionsPage() {
    return (
        <div className="min-h-screen">

            {/* Hero */}
            <section className="py-16 pb-20 border-b border-zinc-800 relative overflow-hidden">
                <div className="max-container relative z-10 space-y-6 text-center flex flex-col items-center">
                    <h1 className="text-5xl md:text-6xl tracking-tighter max-w-4xl mx-auto">
                        Product Suite
                    </h1>
                    <p className="max-w-2xl text-xl text-zinc-400 mx-auto">
                        Every capability listed here runs in your infrastructure. No external dependencies.
                    </p>
                </div>
            </section>

            {/* ── PILLAR 1: SOMBRA ── */}
            <section className="section-padding bg-zinc-950">
                <div className="max-container space-y-20">
                    <div className="max-w-3xl">
                        <div className="flex items-center gap-3 text-orange-500 mb-6">
                            <Network className="w-6 h-6" />
                            <span className="text-[10px] font-bold uppercase tracking-widest">Product 01</span>
                        </div>
                        <h2 className="text-4xl md:text-5xl mb-4">Ocultar Sombra</h2>
                        <h3 className="text-xl text-zinc-400 mb-6">The Agentic Privacy Gateway</h3>
                        <p className="text-lg text-zinc-400 leading-relaxed max-w-2xl">
                            Sombra sits between your infrastructure and every AI provider. It intercepts,
                            sanitizes, routes, and re-hydrates — ensuring only redacted prompts are ever
                            transmitted outbound, and only restored responses are delivered inbound. It is
                            a wire-compatible drop-in for the OpenAI API: change one URL, protect every model call.
                        </p>
                    </div>

                    {/* Multi-Model Router */}
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
                        <div className="space-y-6">
                            <h4 className="text-[10px] font-bold uppercase tracking-widest text-orange-500">Multi-Model Router</h4>
                            <p className="text-zinc-400 leading-relaxed">
                                A pluggable <code className="text-orange-400 bg-zinc-800 px-1.5 py-0.5 rounded text-xs">ModelAdapter</code> interface
                                dispatches sanitized prompts to any registered backend. Zero-Egress domain
                                validation is enforced per-adapter — any unregistered domain is blocked before transmission.
                            </p>
                            <p className="text-xs text-zinc-500 font-mono">
                                <span className="text-orange-500/60">Also acts as a drop-in OpenAI-compatible proxy</span> —
                                existing SDKs need zero code changes.
                            </p>
                        </div>
                        <div className="card bg-zinc-950 border-white/5 p-8">
                            <div className="text-[10px] font-mono text-zinc-600 uppercase tracking-widest mb-6">Registered Adapters</div>
                            <div className="space-y-3">
                                {MODELS.map(m => (
                                    <div key={m.name} className="flex items-center justify-between py-3 border-b border-white/5">
                                        <div>
                                            <div className="text-sm font-bold">{m.name}</div>
                                            <div className="text-[10px] text-zinc-600 font-mono mt-0.5">{m.file}</div>
                                        </div>
                                        <div className="text-[10px] font-mono text-zinc-400 bg-zinc-800 px-3 py-1 rounded-full border border-zinc-800">
                                            {m.provider}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>

                    {/* Connectors */}
                    <div className="space-y-8">
                        <div>
                            <h4 className="text-[10px] font-bold uppercase tracking-widest text-orange-500 mb-2">Data Source Connectors</h4>
                            <p className="text-zinc-400 text-sm">Each connector declares a <code className="text-orange-400 bg-zinc-800 px-1 rounded text-xs">DataPolicy</code> that enforces which PII categories are stripped, which models may receive its data, and enforces size limits.</p>
                        </div>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            {CONNECTORS.map(c => (
                                <div key={c.name} className="card p-8 space-y-5 hover:border-orange-500/30 transition-all">
                                    <div className="flex items-start justify-between">
                                        <div className="w-10 h-10 flex items-center justify-center rounded bg-orange-500/10 text-orange-500">
                                            {c.icon}
                                        </div>
                                        <span className={`text-[9px] font-bold uppercase tracking-widest px-3 py-1 rounded-full border ${
                                            c.tag === 'Enterprise'
                                                ? 'border-zinc-700 text-zinc-400 bg-zinc-800/50'
                                                : 'border-orange-500/20 text-orange-400 bg-orange-500/5'
                                        }`}>
                                            {c.tag}
                                        </span>
                                    </div>
                                    <div>
                                        <h3 className="text-base font-bold text-white mb-2">{c.name}</h3>
                                        <p className="text-sm text-zinc-400 leading-relaxed">{c.desc}</p>
                                    </div>
                                    <div className="text-[9px] font-mono text-zinc-600 pt-2 border-t border-zinc-800">{c.source}</div>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Policy Enforcement Matrix */}
                    <div className="space-y-8">
                        <div>
                            <div className="flex items-center gap-3 text-orange-500 mb-2">
                                <ClipboardCheck className="w-5 h-5" />
                                <h4 className="text-[10px] font-bold uppercase tracking-widest">Per-Connector Policy Enforcement</h4>
                            </div>
                            <p className="text-zinc-400 text-sm max-w-2xl">
                                Every connector enforces a <code className="text-orange-400 bg-zinc-800 px-1 rounded text-xs">DataPolicy</code> at the gateway layer — before a single byte reaches the Refinery. Sensitive data sources are hard-restricted to local inference; no configuration drift can bypass this.
                            </p>
                        </div>
                        <div className="overflow-x-auto rounded-xl border border-zinc-800">
                            <table className="w-full text-sm">
                                <thead>
                                    <tr className="border-b border-zinc-800 bg-zinc-900">
                                        <th className="text-left px-6 py-4 text-[10px] font-bold uppercase tracking-widest text-zinc-500">Connector</th>
                                        <th className="text-left px-6 py-4 text-[10px] font-bold uppercase tracking-widest text-zinc-500">Strip Categories</th>
                                        <th className="text-left px-6 py-4 text-[10px] font-bold uppercase tracking-widest text-zinc-500">Allowed Models</th>
                                        <th className="text-left px-6 py-4 text-[10px] font-bold uppercase tracking-widest text-zinc-500">Size Limit</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-zinc-800 bg-zinc-950">
                                    {POLICY_MATRIX.map(row => (
                                        <tr key={row.connector} className="hover:bg-zinc-900 transition-colors">
                                            <td className="px-6 py-4 font-mono text-xs text-orange-400">{row.connector}</td>
                                            <td className="px-6 py-4">
                                                <div className="flex flex-wrap gap-1.5">
                                                    {row.strip.map(cat => (
                                                        <span key={cat} className="text-[9px] font-bold uppercase tracking-widest px-2 py-0.5 rounded bg-red-500/10 text-red-400 border border-red-500/20">{cat}</span>
                                                    ))}
                                                </div>
                                            </td>
                                            <td className="px-6 py-4">
                                                <div className="flex flex-wrap gap-1.5">
                                                    {row.models.map(m => (
                                                        <span key={m} className="text-[9px] font-mono px-2 py-0.5 rounded bg-zinc-800 text-zinc-400 border border-zinc-700">{m}</span>
                                                    ))}
                                                </div>
                                            </td>
                                            <td className="px-6 py-4 font-mono text-xs text-zinc-500">{row.limit}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                        <p className="text-[10px] font-mono text-zinc-500">source: apps/sombra/configs/sombra.yaml · pkg/connector/connector.go</p>
                    </div>

                    {/* OpenAI Drop-in Compatibility */}
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
                        <div className="space-y-6">
                            <div className="flex items-center gap-3 text-orange-500">
                                <GitMerge className="w-5 h-5" />
                                <h4 className="text-[10px] font-bold uppercase tracking-widest">Drop-in OpenAI Compatibility</h4>
                            </div>
                            <p className="text-zinc-400 leading-relaxed">
                                Sombra exposes a <code className="text-orange-400 bg-zinc-800 px-1.5 py-0.5 rounded text-xs">/v1/chat/completions</code> endpoint that is wire-compatible with the OpenAI API. Change one URL — every existing SDK, agent, or tool works without modification. No wrapper libraries. No forked SDKs.
                            </p>
                            <p className="text-zinc-400 text-sm leading-relaxed">
                                All traffic is scrubbed by the Refinery before dispatch. Responses are optionally rehydrated before returning to the caller. Your application never sees tokenized output unless you opt in.
                            </p>
                            <p className="text-[10px] font-mono text-zinc-500">source: apps/sombra/pkg/handler/handler.go · HandleV1ChatCompletions</p>
                        </div>
                        <div className="card p-6 space-y-0 font-mono text-xs">
                            <div className="text-[10px] uppercase tracking-widest text-zinc-500 mb-5">Migration — one line changes</div>
                            {OPENAI_COMPAT_STEPS.map((step, i) => (
                                <div key={i} className="py-4 border-b border-zinc-800 last:border-0">
                                    <div className="text-[9px] uppercase tracking-widest text-zinc-500 mb-2">{step.label}</div>
                                    <div className="flex flex-col gap-1.5">
                                        <div className="flex items-center gap-2">
                                            <span className="text-red-500/60 text-[10px]">−</span>
                                            <span className={`text-zinc-400 ${step.before !== step.after ? 'line-through decoration-red-500/40' : ''}`}>{step.before}</span>
                                        </div>
                                        {step.before !== step.after && (
                                            <div className="flex items-center gap-2">
                                                <span className="text-orange-500/60 text-[10px]">+</span>
                                                <span className="text-orange-500">{step.after}</span>
                                            </div>
                                        )}
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Immutable Audit Trail */}
                    <div className="relative overflow-hidden rounded-2xl border border-orange-500/20 bg-zinc-900 p-10 space-y-8">
                        <div className="relative z-10 space-y-4 max-w-2xl">
                            <div className="flex items-center gap-3 text-orange-500">
                                <Fingerprint className="w-5 h-5" />
                                <h4 className="text-[10px] font-bold uppercase tracking-widest">Immutable Audit Trail</h4>
                            </div>
                            <h3 className="text-2xl font-bold text-white">Every vault event is cryptographically signed and hash-chained.</h3>
                            <p className="text-zinc-400 leading-relaxed">
                                Sombra initializes an <code className="text-orange-400 bg-zinc-800 px-1.5 py-0.5 rounded text-xs border border-zinc-700">ImmutableLogger</code> backed by an ephemeral Ed25519 keypair on startup. Every vault, match, and rehydration event carries the SHA-256 hash of the previous entry — a tamper-evident chain. The public key is printed to stdout at boot; auditors verify the entire log offline without any access to your infrastructure.
                            </p>
                        </div>
                        <div className="relative z-10 grid grid-cols-1 md:grid-cols-3 gap-4">
                            {[
                                { label: 'Signature',     value: 'Ed25519',               sub: 'FIPS 186-5 compliant' },
                                { label: 'Chain',         value: 'SHA-256 hash chain',    sub: 'Append-only JSON log' },
                                { label: 'Verification',  value: 'Offline / Air-gapped',  sub: 'No infra access required' },
                            ].map(item => (
                                <div key={item.label} className="bg-zinc-800 border border-zinc-700 rounded-xl p-5 space-y-1">
                                    <div className="text-[9px] uppercase tracking-widest text-zinc-500">{item.label}</div>
                                    <div className="text-sm font-bold text-white">{item.value}</div>
                                    <div className="text-[10px] font-mono text-zinc-500">{item.sub}</div>
                                </div>
                            ))}
                        </div>
                        <div className="relative z-10 text-[10px] font-mono text-zinc-600">source: pkg/audit/audit.go · audit.NewImmutableLogger</div>
                    </div>

                </div>
            </section>

            {/* ── PILLAR 2: REFINERY ── */}
            <section className="section-padding border-y border-zinc-800 bg-zinc-950">
                <div className="max-container space-y-16">
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-20 items-start">
                        <div className="space-y-6">
                            <div className="flex items-center gap-3 text-orange-500">
                                <Cpu className="w-6 h-6" />
                                <span className="text-[10px] font-bold uppercase tracking-widest">Product 02</span>
                            </div>
                            <h2 className="text-4xl md:text-5xl">Ocultar Refinery</h2>
                            <h3 className="text-xl text-zinc-400">The High-Throughput Privacy Engine</h3>
                            <p className="text-lg text-zinc-400 leading-relaxed">
                                A 4-tier detection pipeline that sequentially escalates from deterministic
                                pattern matching to local neural inference — achieving near-zero false negative
                                rates with sub-millisecond latency for most payload sizes.
                            </p>
                            <div className="space-y-3 pt-4">
                                {[
                                    "AES-256-GCM encryption for every vault entry",
                                    "Concurrent batch processing — bounded 100-worker pool",
                                    "Session cache eliminates redundant lookups within a run",
                                    "Base64, URL-encoded, and nested JSON evasion detection",
                                    "Fail-Closed: any batch error blocks the entire payload",
                                ].map(f => (
                                    <div key={f} className="flex items-center gap-3 text-sm text-zinc-400">
                                        <ShieldCheck className="w-4 h-4 text-orange-500 flex-shrink-0" />
                                        {f}
                                    </div>
                                ))}
                            </div>
                        </div>
                        <div className="space-y-4">
                            {REFINERY_TIERS.map(t => (
                                <div key={t.tier} className="card p-6 space-y-2">
                                    <div className="flex items-center gap-3 mb-3">
                                        <span className="text-[9px] font-mono font-bold text-orange-400 bg-orange-500/10 border border-orange-500/20 px-2 py-0.5 rounded-full">{t.tier}</span>
                                        <span className="text-sm font-bold">{t.name}</span>
                                    </div>
                                    <p className="text-xs text-zinc-400 leading-relaxed">{t.desc}</p>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Code Visual — intentionally dark terminal */}
                    <div className="card bg-gray-900 border-gray-700 p-8 max-w-3xl mx-auto">
                        <div className="flex items-center justify-between mb-6 border-b border-white/10 pb-4">
                            <span className="text-[10px] font-mono text-zinc-400 uppercase tracking-widest">Refinery Flow — Fail-Closed Batch</span>
                            <Terminal className="w-4 h-4 text-orange-500" />
                        </div>
                        <pre className="text-xs font-mono text-orange-500/70 leading-relaxed overflow-x-auto">
                            <code>{`// RefineBatch — bounded 100-worker goroutine pool
for _, item := range items {
    go func(idx int, val interface{}) {
        res, err := e.ProcessInterface(val, actor)
        // ...
    }(i, item)
}

// Fail-Closed: any single error blocks the entire batch
for _, err := range errs {
    if err != nil {
        return nil, fmt.Errorf("secure block: %w", err)
    }
}`}</code>
                        </pre>
                    </div>
                </div>
            </section>

            {/* ── PILLAR 3: ENTERPRISE ── */}
            <section className="section-padding bg-zinc-950 relative overflow-hidden">
                <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-full h-[600px] bg-orange-500/100/[0.04] blur-[140px] rounded-full pointer-events-none"></div>
                <div className="max-container relative z-10 space-y-16">
                    <div className="max-w-3xl">
                        <div className="flex items-center gap-3 text-orange-500 mb-6">
                            <Layout className="w-6 h-6" />
                            <span className="text-[10px] font-bold uppercase tracking-widest">Product 03</span>
                        </div>
                        <h2 className="text-4xl md:text-5xl mb-4">Ocultar Enterprise</h2>
                        <h3 className="text-xl text-zinc-400 mb-6">Governance, Audit, and Sovereign Operations</h3>
                        <p className="text-lg text-zinc-400 leading-relaxed">
                            Extended capabilities, license-gated at the binary level. Enterprise unlocks
                            deep-scan NER, the SharePoint connector, structured audit logging, and
                            SIEM-compatible syslog forwarding.
                        </p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        {ENTERPRISE_FEATURES.map(f => (
                            <div key={f.title} className="card hover:border-orange-500/30 p-8 space-y-5 transition-all">
                                <div className="w-12 h-12 bg-orange-500/10 border border-orange-500/20 rounded flex items-center justify-center text-orange-500">
                                    {f.icon}
                                </div>
                                <h3 className="text-lg font-bold text-white">{f.title}</h3>
                                <p className="text-sm text-zinc-400 leading-relaxed">{f.desc}</p>
                            </div>
                        ))}
                    </div>

                    {/* Syslog SIEM callout — intentionally dark terminal */}
                    <div className="card bg-gray-900 border-gray-700 p-8 max-w-3xl">
                        <div className="text-[10px] font-mono text-zinc-400 uppercase tracking-widest mb-6">syslog.go — SIEM Forward Architecture</div>
                        <pre className="text-xs font-mono text-orange-400/80 leading-relaxed overflow-x-auto">
                            <code>{`// SyslogServer: PII-scrub every log line before SIEM delivery
refined, err := s.eng.RefineString(msg, "syslog_proxy", nil)
if err != nil {
    // Fail-Closed: drop on refinery error — never forward raw
    continue
}
if upstream != nil {
    // Forward clean message to any UDP SIEM (Splunk HEC, Elastic, etc.)
    upstreamConn.Write([]byte(refined))
}`}</code>
                        </pre>
                        <p className="text-xs text-zinc-400 mt-4 font-mono">
                            Note: Syslog forwarding is SIEM-agnostic via standard UDP. Tested with any compliant upstream endpoint.
                        </p>
                    </div>

                    <div className="pt-8 text-center">
                        <Link to="/risk-assessment" className="btn btn-primary px-14 py-5 text-lg">
                            Request Enterprise Access
                        </Link>
                    </div>
                </div>
            </section>

            {/* Footer status bar */}
            <section className="py-16 border-t border-zinc-800 bg-zinc-950">
                <div className="max-container flex flex-col md:flex-row justify-between items-center gap-6 text-[9px] font-mono text-zinc-600 uppercase tracking-[0.2em]">
                    <div className="flex items-center gap-6">
                        <span className="flex items-center gap-2 text-orange-500/50">
                            <div className="w-1.5 h-1.5 rounded-full bg-orange-500 animate-pulse"></div>
                            Sombra_Gateway_Active
                        </span>
                        <span>Build: 4.12.26</span>
                    </div>
                    <div className="flex gap-10">
                        <span>Features sourced from /pkg</span>
                        <span className="text-orange-500/40">Fail_Closed_By_Default</span>
                    </div>
                </div>
            </section>
        </div>
    );
}
