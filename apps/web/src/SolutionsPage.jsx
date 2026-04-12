import { Link } from 'react-router-dom';
import {
    Shield, Lock, Zap, Network, Cpu, Activity, Server,
    ShieldCheck, Database, Terminal, ChevronRight, Layout,
    MessageSquare, FileText, Globe, Radio, Key, Box
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
        tagColor: "sky",
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

const REFINERY_TIERS = [
    { tier: "Tier 0", name: "Custom Dictionary Engine", desc: "Admin-configurable blocklists for VIP names, internal codenames, and org-specific sensitive terms." },
    { tier: "Tier 1", name: "Deterministic Regex Pipeline", desc: "30+ pre-compiled, validated patterns for SSNs, IBANs, credit cards, phone numbers, addresses, and regional IDs. Luhn validation prevents false positives on credit cards." },
    { tier: "Tier 1.5", name: "Structural Heuristics", desc: "Conjunction expansion, professional title detection, possessive patterns — catches PII that regex misses without model inference overhead." },
    { tier: "Tier 2", name: "SLM NER (Deep Scan)", desc: "A local Small Language Model performs Named-Entity Recognition for contextual PII (conversational names, implicit references). Processes entire records in a single inference pass." },
];

const ENTERPRISE_FEATURES = [
    {
        icon: <Radio className="w-6 h-6" />,
        title: "Syslog → Any Upstream SIEM",
        desc: "Ocultar runs a fail-closed UDP Syslog proxy that strips PII from every log line before forwarding to any compliant SIEM via standard protocol.",
    },
    {
        icon: <Activity className="w-6 h-6" />,
        title: "JSON Audit Logger",
        desc: "Every vault transaction (match, vault, re-hydrate) is written to a structured JSON audit stream. Queryable by any log aggregation platform.",
    },
    {
        icon: <Key className="w-6 h-6" />,
        title: "AES-256-GCM Vault",
        desc: "All PII is encrypted at rest with FIPS-aware AES-256-GCM before touching storage. The Master Key never leaves the operator's infrastructure.",
    },
    {
        icon: <Server className="w-6 h-6" />,
        title: "License-Gated Capabilities",
        desc: "Enterprise connectors (SharePoint, Deep Scan NER) require a signed license payload. Unlicensed access is blocked at the binary level — Fail-Closed by design.",
    },
];

// ---- Page ----

export default function SolutionsPage() {
    return (
        <div className="min-h-screen bg-black text-white selection:bg-emerald-500 selection:text-black">

            {/* Hero */}
            <section className="pt-40 pb-24 border-b border-white/5 relative overflow-hidden">
                <div className="absolute inset-0 bg-emerald-500/[0.03] blur-[120px] scale-150 -translate-y-1/2 rounded-full pointer-events-none"></div>
                <div className="max-container relative z-10 text-center space-y-8">
                    <div className="badge mx-auto">Technical Architecture</div>
                    <h1 className="text-5xl md:text-7xl tracking-tighter max-w-5xl mx-auto">
                        Product Suite — <span className="text-emerald-500">Source Verified</span>
                    </h1>
                    <p className="max-w-2xl mx-auto text-xl text-zinc-400">
                        Every feature on this page maps directly to production code. No vaporware.
                    </p>
                </div>
            </section>

            {/* ── PILLAR 1: SOMBRA ── */}
            <section className="section-padding bg-zinc-950/50">
                <div className="max-container space-y-20">
                    <div className="max-w-3xl">
                        <div className="flex items-center gap-3 text-emerald-500 mb-6">
                            <Network className="w-6 h-6" />
                            <span className="text-[10px] font-bold uppercase tracking-widest">Product 01</span>
                        </div>
                        <h2 className="text-4xl md:text-5xl mb-4">Ocultar Sombra</h2>
                        <h3 className="text-xl text-zinc-400 mb-6">The Agentic Privacy Gateway</h3>
                        <p className="text-lg text-zinc-500 leading-relaxed max-w-2xl">
                            Sombra sits between your infrastructure and every AI provider. It intercepts, 
                            sanitizes, routes, and re-hydrates — ensuring only redacted prompts are ever 
                            transmitted outbound, and only restored responses are delivered inbound.
                        </p>
                    </div>

                    {/* Multi-Model Router */}
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
                        <div className="space-y-6">
                            <h4 className="text-[10px] font-bold uppercase tracking-widest text-emerald-500">Multi-Model Router</h4>
                            <p className="text-zinc-400 leading-relaxed">
                                A pluggable <code className="text-emerald-500/70 bg-zinc-900 px-1.5 py-0.5 rounded text-xs">ModelAdapter</code> interface 
                                dispatches sanitized prompts to any registered backend. Zero-Egress domain 
                                validation is enforced per-adapter — any unregistered domain is blocked before transmission.
                            </p>
                            <p className="text-xs text-zinc-600 font-mono">
                                <span className="text-emerald-500/50">Also acts as a drop-in OpenAI-compatible proxy</span> — 
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
                                        <div className="text-[10px] font-mono text-zinc-500 bg-zinc-900 px-3 py-1 rounded-full border border-white/5">
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
                            <h4 className="text-[10px] font-bold uppercase tracking-widest text-emerald-500 mb-2">Data Source Connectors</h4>
                            <p className="text-zinc-500 text-sm">Each connector declares a <code className="text-emerald-500/70 bg-zinc-900 px-1 rounded text-xs">DataPolicy</code> that enforces which PII categories are stripped, which models may receive its data, and enforces size limits.</p>
                        </div>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            {CONNECTORS.map(c => (
                                <div key={c.name} className={`card bg-zinc-950 border-white/5 hover:border-${c.tagColor}-500/30 p-8 space-y-5 transition-all`}>
                                    <div className="flex items-start justify-between">
                                        <div className={`w-10 h-10 flex items-center justify-center rounded bg-${c.tagColor}-500/10 text-${c.tagColor}-500`}>
                                            {c.icon}
                                        </div>
                                        <span className={`text-[9px] font-bold uppercase tracking-widest px-3 py-1 rounded-full border border-${c.tagColor}-500/20 text-${c.tagColor}-400 bg-${c.tagColor}-500/5`}>
                                            {c.tag}
                                        </span>
                                    </div>
                                    <div>
                                        <h3 className="text-base font-bold mb-2">{c.name}</h3>
                                        <p className="text-sm text-zinc-500 leading-relaxed">{c.desc}</p>
                                    </div>
                                    <div className="text-[9px] font-mono text-zinc-700 pt-2 border-t border-white/5">{c.source}</div>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            </section>

            {/* ── PILLAR 2: REFINERY ── */}
            <section className="section-padding border-y border-white/5">
                <div className="max-container space-y-16">
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-20 items-start">
                        <div className="space-y-6">
                            <div className="flex items-center gap-3 text-sky-500">
                                <Cpu className="w-6 h-6" />
                                <span className="text-[10px] font-bold uppercase tracking-widest">Product 02</span>
                            </div>
                            <h2 className="text-4xl md:text-5xl">Ocultar Refinery</h2>
                            <h3 className="text-xl text-zinc-400">The High-Throughput Privacy Engine</h3>
                            <p className="text-lg text-zinc-500 leading-relaxed">
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
                                        <ShieldCheck className="w-4 h-4 text-emerald-500 flex-shrink-0" />
                                        {f}
                                    </div>
                                ))}
                            </div>
                        </div>
                        <div className="space-y-4">
                            {REFINERY_TIERS.map(t => (
                                <div key={t.tier} className="card bg-zinc-950 border-white/5 p-6 space-y-2">
                                    <div className="flex items-center gap-3 mb-3">
                                        <span className="text-[9px] font-mono font-bold text-sky-400 bg-sky-500/10 border border-sky-500/20 px-2 py-0.5 rounded-full">{t.tier}</span>
                                        <span className="text-sm font-bold">{t.name}</span>
                                    </div>
                                    <p className="text-xs text-zinc-500 leading-relaxed">{t.desc}</p>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Code Visual */}
                    <div className="card bg-black/60 border-emerald-500/10 p-8 max-w-3xl mx-auto">
                        <div className="flex items-center justify-between mb-6 border-b border-white/5 pb-4">
                            <span className="text-[10px] font-mono text-zinc-600 uppercase tracking-widest">Refinery Flow — Fail-Closed Batch</span>
                            <Terminal className="w-4 h-4 text-emerald-500" />
                        </div>
                        <pre className="text-xs font-mono text-emerald-500/70 leading-relaxed overflow-x-auto">
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
            <section className="section-padding bg-black relative overflow-hidden">
                <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-full h-[600px] bg-emerald-500/[0.04] blur-[140px] rounded-full pointer-events-none"></div>
                <div className="max-container relative z-10 space-y-16">
                    <div className="max-w-3xl">
                        <div className="flex items-center gap-3 text-emerald-500 mb-6">
                            <Layout className="w-6 h-6" />
                            <span className="text-[10px] font-bold uppercase tracking-widest">Product 03</span>
                        </div>
                        <h2 className="text-4xl md:text-5xl mb-4">Ocultar Enterprise</h2>
                        <h3 className="text-xl text-zinc-400 mb-6">Governance, Audit, and Sovereign Operations</h3>
                        <p className="text-lg text-zinc-500 leading-relaxed">
                            Extended capabilities, license-gated at the binary level. Enterprise unlocks 
                            deep-scan NER, the SharePoint connector, structured audit logging, and 
                            SIEM-compatible syslog forwarding.
                        </p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        {ENTERPRISE_FEATURES.map(f => (
                            <div key={f.title} className="card bg-zinc-950/60 border-white/5 hover:border-emerald-500/20 p-10 space-y-5 transition-all">
                                <div className="w-12 h-12 bg-emerald-500/10 rounded flex items-center justify-center text-emerald-500">
                                    {f.icon}
                                </div>
                                <h3 className="text-lg font-bold">{f.title}</h3>
                                <p className="text-sm text-zinc-500 leading-relaxed">{f.desc}</p>
                            </div>
                        ))}
                    </div>

                    {/* Syslog SIEM callout */}
                    <div className="card bg-zinc-950 border-white/5 p-8 max-w-3xl">
                        <div className="text-[10px] font-mono text-zinc-600 uppercase tracking-widest mb-6">syslog.go — SIEM Forward Architecture</div>
                        <pre className="text-xs font-mono text-zinc-400 leading-relaxed overflow-x-auto">
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
                        <p className="text-xs text-zinc-600 mt-4 font-mono">
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
            <section className="py-16 border-t border-white/5 bg-zinc-950">
                <div className="max-container flex flex-col md:flex-row justify-between items-center gap-6 text-[9px] font-mono text-zinc-700 uppercase tracking-[0.2em]">
                    <div className="flex items-center gap-6">
                        <span className="flex items-center gap-2 text-emerald-500/40">
                            <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></div>
                            Sombra_Gateway_Active
                        </span>
                        <span>Build: 4.12.26</span>
                    </div>
                    <div className="flex gap-10">
                        <span>Features sourced from /pkg</span>
                        <span>Zero vaporware</span>
                        <span className="text-emerald-500/30">Fail_Closed_By_Default</span>
                    </div>
                </div>
            </section>
        </div>
    );
}
