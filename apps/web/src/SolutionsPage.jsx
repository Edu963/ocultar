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
        tagColor: "emerald",
        source: "pkg/handler/slack_app.go"
    },
    {
        name: "SharePoint via MS Graph",
        icon: <FileText className="w-5 h-5" />,
        desc: "Polls SharePoint document libraries via Microsoft Graph OAuth2. Enterprise-licensed capability — all document content is processed through the Refinery before any downstream action.",
        tag: "Enterprise",
        tagColor: "slate",
        source: "services/refinery/pkg/connector/sharepoint.go"
    },
    {
        name: "Generic REST API",
        icon: <Globe className="w-5 h-5" />,
        desc: "Zero-config adapter for any authenticated REST endpoint. Supports Bearer tokens and API keys, with per-connector data policy enforcement and model allowlists.",
        tag: "Native",
        tagColor: "emerald",
        source: "pkg/connector/api_connector.go"
    },
    {
        name: "File Connector",
        icon: <Box className="w-5 h-5" />,
        desc: "Processes uploaded documents through the Refinery before storage or forwarding. Supports JSON, CSV, and raw text.",
        tag: "Native",
        tagColor: "emerald",
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
    { tier: "Tier 0",   name: "Custom Dictionary Engine", desc: "Admin-configurable blocklists for VIP names, internal codenames, and org-specific sensitive terms. Backed by a live CRM/LDAP sync that updates the dictionary without restart." },
    { tier: "Tier 1",   name: "Deterministic Regex Pipeline", desc: "30+ pre-compiled, Luhn-validated patterns: SSNs, IBANs, credit cards, phone numbers, addresses, SIRET/SIREN, health references, and regional IDs across 10+ countries. Credit card false positives are eliminated with Luhn checksum verification." },
    { tier: "Tier 1.5", name: "Contextual Heuristics", desc: "Phone libphonenumber validation, heuristic address parsing, greeting/signature detection (name in 'Dear John' or 'Best, Sarah'). Catches PII that regex misses without any model inference overhead." },
    { tier: "Tier 2",   name: "OpenAI Privacy Filter (Deep Scan)", desc: "A 1.5B-parameter local bidirectional token classifier (Apache 2.0) performs Named-Entity Recognition for contextual PII — conversational names, implicit references, and finance-specific entities. 97% F1. Runs entirely in your infrastructure. Swappable with any compatible model via TIER2_ENGINE env var." },
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
        <div className="min-h-screen bg-[#050505]">

            {/* ── Hero ── */}
            <section className="py-32 bg-[#050505] border-b border-white/5 relative overflow-hidden">
                <div className="absolute inset-0 bg-[radial-gradient(ellipse_80%_50%_at_50%_-10%,rgba(16,185,129,0.05),transparent)]" />
                <div className="max-container relative z-10 flex flex-col items-center text-center gap-6">
                    <p className="text-xs font-mono font-semibold uppercase tracking-widest text-emerald-500">
                        Product Suite
                    </p>
                    <h1 className="text-5xl md:text-6xl tracking-tight text-balance max-w-4xl text-white">
                        Product Suite
                    </h1>
                    <p className="text-xl text-slate-400 leading-relaxed max-w-2xl">
                        Every capability listed here runs in your infrastructure. No external dependencies.
                    </p>
                </div>
            </section>

            {/* ── PILLAR 1: REFINERY ── */}
            <section className="py-24 bg-[#0A0A0C] border-b border-white/5">
                <div className="max-container flex flex-col gap-16">
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-20 items-start">
                        <div className="flex flex-col gap-6">
                            <div className="flex items-center gap-3 text-emerald-500">
                                <Cpu className="w-5 h-5" />
                                <p className="text-xs font-mono font-semibold uppercase tracking-widest">Product 01</p>
                            </div>
                            <h2 className="text-4xl md:text-5xl font-bold text-white tracking-tight text-balance">
                                Ocultar Refinery
                            </h2>
                            <h3 className="text-xl text-slate-400 text-balance">The High-Throughput Privacy Engine</h3>
                            <p className="text-lg text-slate-400 leading-relaxed">
                                A 4-tier detection pipeline that sequentially escalates from deterministic
                                pattern matching to local neural inference — achieving near-zero false negative
                                rates with sub-millisecond latency for most payload sizes.
                            </p>
                            <div className="flex flex-col gap-3 pt-2">
                                {[
                                    "AES-256-GCM encryption for every vault entry",
                                    "Concurrent batch processing — bounded 100-worker pool",
                                    "Session cache eliminates redundant lookups within a run",
                                    "Base64, URL-encoded, and nested JSON evasion detection",
                                    "Fail-Closed: any batch error blocks the entire payload",
                                ].map(f => (
                                    <div key={f} className="flex items-start gap-3 text-base text-slate-400">
                                        <ShieldCheck className="w-4 h-4 text-emerald-500 shrink-0 mt-0.5" />
                                        {f}
                                    </div>
                                ))}
                            </div>

                            {/* Refinery terminal — left column, below bullet list */}
                            <div className="bg-[#050505] ring-1 ring-white/10 rounded-xl p-8 w-full mt-12 shadow-2xl">
                                <div className="flex items-center justify-between mb-6 pb-4 border-b border-white/5">
                                    <p className="text-xs font-mono font-semibold uppercase tracking-widest text-slate-500">
                                        Refinery Flow — Fail-Closed Batch
                                    </p>
                                    <Terminal className="w-4 h-4 text-emerald-500" />
                                </div>
                                <pre className="text-xs font-mono leading-relaxed overflow-x-auto">
                                    <span className="text-slate-500">{'// RefineBatch — bounded 100-worker goroutine pool\n'}</span>
                                    <span className="text-slate-300">{'for _, item := range items {\n'}</span>
                                    <span className="text-slate-300">{'    go func(idx int, val interface{}) {\n'}</span>
                                    <span className="text-slate-300">{'        res, err := e.'}</span><span className="text-emerald-400">{'ProcessInterface'}</span><span className="text-slate-300">{'(val, actor)\n'}</span>
                                    <span className="text-slate-500">{'        // ...\n'}</span>
                                    <span className="text-slate-300">{'    }(i, item)\n'}</span>
                                    <span className="text-slate-300">{'}\n\n'}</span>
                                    <span className="text-slate-500">{'// Fail-Closed: any single error blocks the entire batch\n'}</span>
                                    <span className="text-slate-300">{'for _, err := range errs {\n'}</span>
                                    <span className="text-slate-300">{'    if err != nil {\n'}</span>
                                    <span className="text-slate-300">{'        return nil, fmt.Errorf('}</span><span className="text-emerald-400">{'"secure block: %w"'}</span><span className="text-slate-300">{', err)\n'}</span>
                                    <span className="text-slate-300">{'    }\n'}</span>
                                    <span className="text-slate-300">{'}'}</span>
                                </pre>
                            </div>
                        </div>
                        <div className="flex flex-col gap-4">
                            {REFINERY_TIERS.map(t => (
                                <div
                                    key={t.tier}
                                    className="bg-[#111114] ring-1 ring-white/5 rounded-xl p-6 hover:ring-emerald-500/30 transition-all duration-300 shadow-xl flex flex-col gap-3"
                                >
                                    <div className="flex items-center gap-3">
                                        <span className="text-xs font-mono font-semibold text-emerald-400 bg-emerald-500/10 ring-1 ring-emerald-500/20 px-2.5 py-0.5 rounded-full">
                                            {t.tier}
                                        </span>
                                        <span className="text-sm font-bold text-white">{t.name}</span>
                                    </div>
                                    <p className="text-sm text-slate-400 leading-relaxed">{t.desc}</p>
                                </div>
                            ))}

                        </div>
                    </div>
                </div>
            </section>

            {/* ── PILLAR 2: SOMBRA ── */}
            <section className="py-24 bg-[#050505]">
                <div className="max-container flex flex-col gap-24">

                    {/* Header */}
                    <div className="max-w-3xl flex flex-col gap-4">
                        <div className="flex items-center gap-3 text-emerald-500">
                            <Network className="w-5 h-5" />
                            <p className="text-xs font-mono font-semibold uppercase tracking-widest">Product 02</p>
                        </div>
                        <h2 className="text-4xl md:text-5xl font-bold text-white tracking-tight text-balance">Ocultar Sombra</h2>
                        <h3 className="text-xl text-slate-400 text-balance">The Agentic Privacy Gateway</h3>
                        <p className="text-lg text-slate-400 leading-relaxed max-w-2xl">
                            Sombra sits between your infrastructure and every AI provider. It intercepts,
                            sanitizes, routes, and re-hydrates — ensuring only redacted prompts are ever
                            transmitted outbound, and only restored responses are delivered inbound. It is
                            a wire-compatible drop-in for the OpenAI API: change one URL, protect every model call.
                        </p>
                    </div>

                    {/* Multi-Model Router */}
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
                        <div className="flex flex-col gap-5">
                            <p className="text-xs font-mono font-semibold uppercase tracking-widest text-emerald-500">Multi-Model Router</p>
                            <p className="text-slate-400 leading-relaxed">
                                A pluggable{' '}
                                <code className="text-emerald-400 bg-white/5 px-1.5 py-0.5 rounded text-xs">ModelAdapter</code>{' '}
                                interface dispatches sanitized prompts to any registered backend. Zero-Egress domain
                                validation is enforced per-adapter — any unregistered domain is blocked before transmission.
                            </p>
                            <p className="text-sm text-slate-500 font-mono">
                                <span className="text-emerald-500/70">Also acts as a drop-in OpenAI-compatible proxy</span>{' '}—
                                existing SDKs need zero code changes.
                            </p>
                        </div>
                        <div className="bg-[#0A0A0C] ring-1 ring-white/5 rounded-xl p-8 shadow-xl">
                            <p className="text-xs font-mono font-semibold uppercase tracking-widest text-slate-500 mb-6">
                                Registered Adapters
                            </p>
                            <div className="flex flex-col">
                                {MODELS.map(m => (
                                    <div key={m.name} className="flex items-center justify-between py-4 border-b border-white/5 last:border-0">
                                        <div className="flex flex-col gap-1">
                                            <div className="text-sm font-bold text-white">{m.name}</div>
                                            <div className="text-xs text-slate-600 font-mono">{m.file}</div>
                                        </div>
                                        <span className="text-xs font-mono text-slate-400 bg-white/5 px-3 py-1 rounded-full ring-1 ring-white/10">
                                            {m.provider}
                                        </span>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>

                    {/* Data Source Connectors */}
                    <div className="flex flex-col gap-8">
                        <div className="flex flex-col gap-3">
                            <p className="text-xs font-mono font-semibold uppercase tracking-widest text-emerald-500">
                                Data Source Connectors
                            </p>
                            <p className="text-slate-400 leading-relaxed max-w-2xl">
                                Each connector declares a{' '}
                                <code className="text-emerald-400 bg-white/5 px-1.5 py-0.5 rounded text-xs">DataPolicy</code>{' '}
                                that enforces which PII categories are stripped, which models may receive its data, and enforces size limits.
                            </p>
                        </div>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            {CONNECTORS.map(c => (
                                <div
                                    key={c.name}
                                    className="bg-[#0A0A0C] ring-1 ring-white/5 rounded-xl p-8 hover:ring-emerald-500/30 transition-all duration-300 shadow-xl flex flex-col gap-5"
                                >
                                    <div className="flex items-start justify-between">
                                        <div className="w-10 h-10 flex items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-500">
                                            {c.icon}
                                        </div>
                                        <span className={`text-xs font-mono font-semibold uppercase tracking-widest px-3 py-1 rounded-full ring-1 ${
                                            c.tagColor === 'slate'
                                                ? 'ring-white/10 text-slate-400 bg-white/5'
                                                : 'ring-emerald-500/20 text-emerald-400 bg-emerald-500/5'
                                        }`}>
                                            {c.tag}
                                        </span>
                                    </div>
                                    <div className="flex flex-col gap-2">
                                        <h3 className="text-base font-bold text-white text-balance">{c.name}</h3>
                                        <p className="text-sm text-slate-400 leading-relaxed">{c.desc}</p>
                                    </div>
                                    <div className="text-xs font-mono text-slate-600 pt-2 border-t border-white/5">{c.source}</div>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Policy Enforcement Matrix */}
                    <div className="flex flex-col gap-8">
                        <div className="flex flex-col gap-3">
                            <div className="flex items-center gap-3 text-emerald-500">
                                <ClipboardCheck className="w-4 h-4" />
                                <p className="text-xs font-mono font-semibold uppercase tracking-widest">
                                    Per-Connector Policy Enforcement
                                </p>
                            </div>
                            <p className="text-slate-400 leading-relaxed max-w-2xl">
                                Every connector enforces a{' '}
                                <code className="text-emerald-400 bg-white/5 px-1.5 py-0.5 rounded text-xs">DataPolicy</code>{' '}
                                at the gateway layer — before a single byte reaches the Refinery. Sensitive data sources are hard-restricted to local inference; no configuration drift can bypass this.
                            </p>
                        </div>
                        <div className="overflow-x-auto rounded-xl ring-1 ring-white/5">
                            <table className="w-full text-sm">
                                <thead>
                                    <tr className="bg-white/[0.02]">
                                        <th className="text-left px-6 py-4 text-xs font-mono font-semibold uppercase tracking-widest text-slate-500">Connector</th>
                                        <th className="text-left px-6 py-4 text-xs font-mono font-semibold uppercase tracking-widest text-slate-500">Strip Categories</th>
                                        <th className="text-left px-6 py-4 text-xs font-mono font-semibold uppercase tracking-widest text-slate-500">Allowed Models</th>
                                        <th className="text-left px-6 py-4 text-xs font-mono font-semibold uppercase tracking-widest text-slate-500">Size Limit</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {POLICY_MATRIX.map(row => (
                                        <tr key={row.connector} className="border-b border-white/5 last:border-0 hover:bg-white/[0.02] transition-colors">
                                            <td className="px-6 py-4 font-mono text-xs text-emerald-400">{row.connector}</td>
                                            <td className="px-6 py-4">
                                                <div className="flex flex-wrap gap-1.5">
                                                    {row.strip.map(cat => (
                                                        <span key={cat} className="text-xs font-mono font-semibold uppercase tracking-wide px-2 py-0.5 rounded bg-red-500/10 text-red-400 ring-1 ring-red-500/20">
                                                            {cat}
                                                        </span>
                                                    ))}
                                                </div>
                                            </td>
                                            <td className="px-6 py-4">
                                                <div className="flex flex-wrap gap-1.5">
                                                    {row.models.map(m => (
                                                        <span key={m} className="text-xs font-mono px-2 py-0.5 rounded bg-white/5 text-slate-400 ring-1 ring-white/10">
                                                            {m}
                                                        </span>
                                                    ))}
                                                </div>
                                            </td>
                                            <td className="px-6 py-4 font-mono text-xs text-slate-500">{row.limit}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                        <p className="text-xs font-mono text-slate-600">
                            source: apps/sombra/configs/sombra.yaml · pkg/connector/connector.go
                        </p>
                    </div>

                    {/* Drop-in OpenAI Compatibility */}
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
                        <div className="flex flex-col gap-5">
                            <div className="flex items-center gap-3 text-emerald-500">
                                <GitMerge className="w-4 h-4" />
                                <p className="text-xs font-mono font-semibold uppercase tracking-widest">
                                    Drop-in OpenAI Compatibility
                                </p>
                            </div>
                            <p className="text-slate-400 leading-relaxed">
                                Sombra exposes a{' '}
                                <code className="text-emerald-400 bg-white/5 px-1.5 py-0.5 rounded text-xs">/v1/chat/completions</code>{' '}
                                endpoint that is wire-compatible with the OpenAI API. Change one URL — every existing SDK, agent, or tool works without modification. No wrapper libraries. No forked SDKs.
                            </p>
                            <p className="text-slate-400 leading-relaxed">
                                All traffic is scrubbed by the Refinery before dispatch. Responses are optionally rehydrated before returning to the caller. Your application never sees tokenized output unless you opt in.
                            </p>
                            <p className="text-xs font-mono text-slate-600">
                                source: apps/sombra/pkg/handler/handler.go · HandleV1ChatCompletions
                            </p>
                        </div>
                        <div className="bg-[#050505] ring-1 ring-white/10 rounded-xl p-6 font-mono text-xs flex flex-col">
                            <p className="text-xs font-mono font-semibold uppercase tracking-widest text-slate-500 mb-5 pb-4 border-b border-white/5">
                                Migration — one line changes
                            </p>
                            {OPENAI_COMPAT_STEPS.map((step, i) => (
                                <div key={i} className="py-4 border-b border-white/5 last:border-0 flex flex-col gap-2">
                                    <div className="text-xs uppercase tracking-widest text-slate-600 mb-1">{step.label}</div>
                                    <div className="flex items-center gap-2">
                                        <span className="text-red-500/50">−</span>
                                        <span className={`text-slate-400 ${step.before !== step.after ? 'line-through decoration-red-500/40' : ''}`}>
                                            {step.before}
                                        </span>
                                    </div>
                                    {step.before !== step.after && (
                                        <div className="flex items-center gap-2">
                                            <span className="text-emerald-500/60">+</span>
                                            <span className="text-emerald-400">{step.after}</span>
                                        </div>
                                    )}
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Immutable Audit Trail */}
                    <div className="relative overflow-hidden rounded-2xl ring-1 ring-emerald-500/20 bg-[#0A0A0C] p-10 flex flex-col gap-8">
                        <div className="absolute inset-0 bg-[radial-gradient(ellipse_60%_80%_at_80%_20%,rgba(16,185,129,0.04),transparent)]" />
                        <div className="relative z-10 flex flex-col gap-4 max-w-2xl">
                            <div className="flex items-center gap-3 text-emerald-500">
                                <Fingerprint className="w-4 h-4" />
                                <p className="text-xs font-mono font-semibold uppercase tracking-widest">Immutable Audit Trail</p>
                            </div>
                            <h3 className="text-2xl font-bold text-white text-balance tracking-tight">
                                Every vault event is cryptographically signed and hash-chained.
                            </h3>
                            <p className="text-slate-400 leading-relaxed">
                                Sombra initializes an{' '}
                                <code className="text-emerald-400 bg-white/5 px-1.5 py-0.5 rounded text-xs ring-1 ring-white/10">ImmutableLogger</code>{' '}
                                backed by an ephemeral Ed25519 keypair on startup. Every vault, match, and rehydration event carries the SHA-256 hash of the previous entry — a tamper-evident chain. The public key is printed to stdout at boot; auditors verify the entire log offline without any access to your infrastructure.
                            </p>
                        </div>
                        <div className="relative z-10 grid grid-cols-1 md:grid-cols-3 gap-4">
                            {[
                                { label: 'Signature',    value: 'Ed25519',              sub: 'FIPS 186-5 compliant' },
                                { label: 'Chain',        value: 'SHA-256 hash chain',   sub: 'Append-only JSON log' },
                                { label: 'Verification', value: 'Offline / Air-gapped', sub: 'No infra access required' },
                            ].map(item => (
                                <div key={item.label} className="bg-[#111114] ring-1 ring-white/5 rounded-xl p-5 flex flex-col gap-1">
                                    <div className="text-xs font-mono font-semibold uppercase tracking-widest text-slate-500">{item.label}</div>
                                    <div className="text-sm font-bold text-white">{item.value}</div>
                                    <div className="text-xs font-mono text-slate-600">{item.sub}</div>
                                </div>
                            ))}
                        </div>
                        <p className="relative z-10 text-xs font-mono text-slate-600">
                            source: pkg/audit/audit.go · audit.NewImmutableLogger
                        </p>
                    </div>

                </div>
            </section>

            {/* ── PILLAR 3: ENTERPRISE ── */}
            <section className="py-24 bg-[#050505] relative overflow-hidden">
                <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-full h-[600px] bg-emerald-500/[0.03] blur-[140px] rounded-full pointer-events-none" />
                <div className="max-container relative z-10 flex flex-col gap-16">
                    <div className="max-w-3xl flex flex-col gap-4">
                        <div className="flex items-center gap-3 text-emerald-500">
                            <Layout className="w-5 h-5" />
                            <p className="text-xs font-mono font-semibold uppercase tracking-widest">Product 03</p>
                        </div>
                        <h2 className="text-4xl md:text-5xl font-bold text-white tracking-tight text-balance">
                            Ocultar Enterprise
                        </h2>
                        <h3 className="text-xl text-slate-400 text-balance">Governance, Audit, and Sovereign Operations</h3>
                        <p className="text-lg text-slate-400 leading-relaxed">
                            Extended capabilities, license-gated at the binary level. Enterprise unlocks
                            deep-scan NER, the SharePoint connector, structured audit logging, and
                            SIEM-compatible syslog forwarding.
                        </p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        {ENTERPRISE_FEATURES.map(f => (
                            <div
                                key={f.title}
                                className="bg-[#0A0A0C] ring-1 ring-white/5 rounded-xl p-8 hover:ring-emerald-500/30 transition-all duration-300 shadow-xl flex flex-col gap-5"
                            >
                                <div className="w-12 h-12 bg-emerald-500/10 ring-1 ring-emerald-500/20 rounded-lg flex items-center justify-center text-emerald-500">
                                    {f.icon}
                                </div>
                                <h3 className="text-lg font-bold text-white text-balance">{f.title}</h3>
                                <p className="text-base text-slate-400 leading-relaxed">{f.desc}</p>
                            </div>
                        ))}
                    </div>

                    {/* Syslog SIEM terminal */}
                    <div className="bg-[#050505] ring-1 ring-white/10 rounded-xl p-8 max-w-3xl shadow-2xl">
                        <p className="text-xs font-mono font-semibold uppercase tracking-widest text-slate-500 mb-6 pb-4 border-b border-white/5">
                            syslog.go — SIEM Forward Architecture
                        </p>
                        <pre className="text-xs font-mono leading-relaxed overflow-x-auto">
                            <span className="text-slate-500">{'// SyslogServer: PII-scrub every log line before SIEM delivery\n'}</span>
                            <span className="text-slate-300">{'refined, err := s.eng.'}</span><span className="text-emerald-400">{'RefineString'}</span><span className="text-slate-300">{'(msg, '}</span><span className="text-emerald-400">{'"syslog_proxy"'}</span><span className="text-slate-300">{', nil)\n'}</span>
                            <span className="text-slate-300">{'if err != nil {\n'}</span>
                            <span className="text-slate-500">{'    // Fail-Closed: drop on refinery error — never forward raw\n'}</span>
                            <span className="text-slate-300">{'    continue\n'}</span>
                            <span className="text-slate-300">{'}\n'}</span>
                            <span className="text-slate-300">{'if upstream != nil {\n'}</span>
                            <span className="text-slate-500">{'    // Forward clean message to any UDP SIEM (Splunk HEC, Elastic, etc.)\n'}</span>
                            <span className="text-slate-300">{'    upstreamConn.'}</span><span className="text-emerald-400">{'Write'}</span><span className="text-slate-300">{'([]byte(refined))\n'}</span>
                            <span className="text-slate-300">{'}'}</span>
                        </pre>
                        <p className="text-xs text-slate-500 font-mono mt-6 pt-4 border-t border-white/5">
                            Note: Syslog forwarding is SIEM-agnostic via standard UDP. Tested with any compliant upstream endpoint.
                        </p>
                    </div>

                    <div className="pt-4 text-center">
                        <Link
                            to="/risk-assessment"
                            className="inline-flex items-center gap-3 bg-emerald-500 hover:bg-emerald-400 text-black font-bold px-12 py-5 text-base rounded-lg transition-all duration-200 hover:-translate-y-0.5 hover:shadow-xl hover:shadow-emerald-500/20 group"
                        >
                            Request Enterprise Access
                            <ChevronRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
                        </Link>
                    </div>
                </div>
            </section>

            {/* ── Footer status bar ── */}
            <section className="py-8 border-t border-white/5 bg-[#050505]">
                <div className="max-container flex flex-col md:flex-row justify-between items-center gap-4 text-xs font-mono text-slate-600 uppercase tracking-widest">
                    <div className="flex items-center gap-6">
                        <span className="flex items-center gap-2 text-emerald-500/50">
                            <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                            Sombra_Gateway_Active
                        </span>
                        <span>Build: 4.12.26</span>
                    </div>
                    <div className="flex gap-10">
                        <span>Features sourced from /pkg</span>
                        <span className="text-emerald-500/30">Fail_Closed_By_Default</span>
                    </div>
                </div>
            </section>

        </div>
    );
}
