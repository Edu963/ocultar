import { useState } from 'react';
import { Link } from 'react-router-dom';
import { 
    Shield, Lock, Zap, BarChart3, ChevronRight, Globe, Terminal, 
    ShieldCheck, Database, Activity, Cpu, Box, Info, Server, 
    Search, Network, Layout, Key
} from 'lucide-react';

export default function SolutionsPage() {
    return (
        <div className="min-h-screen bg-black text-white selection:bg-emerald-500 selection:text-black">
            {/* Hero Section */}
            <section className="pt-40 pb-20 border-b border-white/5 relative overflow-hidden">
                <div className="absolute inset-0 bg-emerald-500/[0.02] blur-3xl rounded-full scale-150 -translate-y-1/2"></div>
                <div className="max-container relative z-10 text-center">
                    <div className="badge mb-6 mx-auto">Enterprise Ecosystem</div>
                    <h1 className="text-5xl md:text-7xl mb-8 tracking-tighter">
                        Complete <span className="text-emerald-500">Security</span> Sovereignty
                    </h1>
                    <p className="max-w-2xl mx-auto text-xl text-zinc-500">
                        From edge gateways to centralized governance, Ocultar provides the infrastructure for safe AI adoption.
                    </p>
                </div>
            </section>

            {/* Product Pillar 1: Sombra */}
            <section className="section-padding bg-zinc-950/50">
                <div className="max-container grid grid-cols-1 lg:grid-cols-2 gap-20 items-center">
                    <div className="space-y-8 animate-fade-in-up">
                        <div className="flex items-center gap-3 text-emerald-500">
                            <Network className="w-6 h-6" />
                            <span className="text-[10px] font-bold uppercase tracking-widest">Product 01</span>
                        </div>
                        <h2 className="text-4xl md:text-5xl">Ocultar Sombra</h2>
                        <h3 className="text-xl text-zinc-400 font-medium">The Agentic Data Gateway</h3>
                        <p className="text-lg text-zinc-500 leading-relaxed">
                            A high-performance, edge-deployed privacy mesh that acts as the secure entry point for all AI interactions. Sombra manages the lifecycle of agentic data flows before they reach external providers.
                        </p>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 pt-6">
                            {[
                                { title: "Multi-Model Routing", desc: "Intelligent traffic directing between Local SLMs, OpenAI, and Gemini." },
                                { title: "Zero-Trust Proxy", desc: "Identity-aware gateway that protects internal infrastructure IPs." },
                                { title: "Connector Mesh", desc: "Native high-speed bridges for Slack, SharePoint, and custom DBs." },
                                { title: "Edge Deployment", desc: "Deploy as a sidecar or cluster gateway within your VPC." }
                            ].map(f => (
                                <div key={f.title} className="space-y-2">
                                    <div className="text-sm font-bold text-white flex items-center gap-2">
                                        <div className="w-1.5 h-1.5 rounded-full bg-emerald-500"></div>
                                        {f.title}
                                    </div>
                                    <p className="text-xs text-zinc-500">{f.desc}</p>
                                </div>
                            ))}
                        </div>
                    </div>
                    <div className="relative group">
                        <div className="absolute inset-0 bg-emerald-500/10 blur-3xl rounded-full scale-75 group-hover:scale-100 transition-transform duration-700"></div>
                        <div className="card border-emerald-500/20 bg-black/50 p-8 relative">
                            <div className="flex items-center justify-between mb-8 border-b border-white/5 pb-6">
                                <div className="text-[10px] font-mono text-zinc-500 uppercase tracking-widest">sombra.yaml — Runtime Config</div>
                                <Terminal className="w-4 h-4 text-emerald-500" />
                            </div>
                            <pre className="text-xs font-mono text-emerald-500/80 leading-relaxed overflow-x-auto">
                                <code>{`policy:
  strip_categories: ["SSN", "CREDENTIAL", "SECRET"]
  allowed_models: ["local-slm", "gpt-4o"]
  fail_mode: "closed"

connectors:
  - id: "slack-prod"
    type: "agentic_bridge"
    endpoint: "https://hooks.slack.com/..."`}</code>
                            </pre>
                        </div>
                    </div>
                </div>
            </section>

            {/* Product Pillar 2: Refinery */}
            <section className="section-padding border-y border-white/5">
                <div className="max-container grid grid-cols-1 lg:grid-cols-2 gap-20 items-center">
                    <div className="order-2 lg:order-1 relative">
                        <div className="card p-8 border-white/5 bg-zinc-950">
                            <div className="flex items-center justify-between mb-8 border-b border-white/5 pb-6">
                                <Activity className="w-5 h-5 text-sky-500" />
                                <div className="text-[10px] font-mono text-zinc-500 uppercase tracking-widest">Engine Performance</div>
                            </div>
                            <div className="space-y-6">
                                {[
                                    { label: "P50 Detection Latency", value: "< 0.92ms" },
                                    { label: "Concurrent Throughput", value: "50k req/s" },
                                    { label: "Detection Accuracy", value: "99.98%" },
                                    { label: "Cold Start Time", value: "140ms" }
                                ].map(m => (
                                    <div key={m.label} className="flex justify-between items-center py-2 border-b border-white/5">
                                        <span className="text-[11px] text-zinc-500 uppercase">{m.label}</span>
                                        <span className="text-xs font-mono font-bold text-sky-400">{m.value}</span>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                    <div className="order-1 lg:order-2 space-y-8 animate-fade-in-up">
                        <div className="flex items-center gap-3 text-sky-500">
                            <Cpu className="w-6 h-6" />
                            <span className="text-[10px] font-bold uppercase tracking-widest">Product 02</span>
                        </div>
                        <h2 className="text-4xl md:text-5xl">Ocultar Refinery</h2>
                        <h3 className="text-xl text-zinc-400 font-medium">The High-Performance Privacy Engine</h3>
                        <p className="text-lg text-zinc-500 leading-relaxed">
                            The core computational engine of the Ocultar ecosystem. Refinery uses deterministic logic and local vector analysis to identify and sanitize complex PII patterns in sub-millisecond speeds.
                        </p>
                        <div className="space-y-4">
                            {[
                                "30+ Built-in Regional PII Sovereignty Packs",
                                "Deterministic Tokenization for Data Consistency",
                                "Fail-Closed Execution Architecture",
                                "FIPS 140-2 Compliant Transformation Logic"
                            ].map(item => (
                                <div key={item} className="flex items-center gap-3 text-sm text-zinc-400">
                                    <ShieldCheck className="w-4 h-4 text-emerald-500" />
                                    {item}
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            </section>

            {/* Product Pillar 3: Enterprise */}
            <section className="section-padding bg-black relative overflow-hidden">
                <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-full h-[500px] bg-emerald-500/5 blur-[120px] rounded-full"></div>
                <div className="max-container text-center space-y-12 relative z-10">
                    <div className="space-y-6">
                        <div className="flex items-center justify-center gap-3 text-emerald-500">
                            <Layout className="w-6 h-6" />
                            <span className="text-[10px] font-bold uppercase tracking-widest">Product 03</span>
                        </div>
                        <h2 className="text-4xl md:text-6xl max-w-4xl mx-auto">Ocultar Enterprise</h2>
                        <p className="text-xl text-zinc-400 max-w-2xl mx-auto">
                            Centralized governance for the distributed AI stack. Managed policies, deep audit logging, and enterprise-grade integration.
                        </p>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                        {[
                            { title: "Governance Hub", desc: "Centralized policy management and model performance visibility.", icon: <Globe /> },
                            { title: "Deep Audit Logs", desc: "Identity-mapped transaction logs for compliance and forensics.", icon: <Search /> },
                            { title: "SIEM Integration", desc: "Native connectors for Splunk, Datadog, and Sentinel.", icon: <Shield /> }
                        ].map(c => (
                            <div key={c.title} className="card p-10 space-y-6 bg-zinc-950/50 border-white/5 hover:border-emerald-500/30 transition-all">
                                <div className="w-12 h-12 bg-emerald-500/10 rounded flex items-center justify-center text-emerald-500">
                                    {c.icon}
                                </div>
                                <h3 className="text-xl font-bold">{c.title}</h3>
                                <p className="text-sm text-zinc-500 leading-relaxed">{c.desc}</p>
                            </div>
                        ))}
                    </div>

                    <div className="pt-12">
                        <Link to="/risk-assessment" className="btn btn-primary px-12 py-5 text-lg">
                            Request Enterprise Access
                        </Link>
                    </div>
                </div>
            </section>

            {/* Technical Context Footer */}
            <section className="py-20 border-t border-white/5 bg-zinc-950">
                <div className="max-container">
                    <div className="flex flex-col md:flex-row justify-between items-center gap-8 text-[10px] font-mono text-zinc-600 uppercase tracking-[0.2em]">
                        <div className="flex items-center gap-6">
                            <span className="flex items-center gap-2 text-emerald-500/50">
                                <div className="w-1.5 h-1.5 rounded-full bg-emerald-500"></div>
                                Sombra_Mesh_Active
                            </span>
                            <span>Build: 4.12.26</span>
                        </div>
                        <div className="flex gap-12">
                            <span>Determined by Code</span>
                            <span>Verified by Audit</span>
                            <span className="text-emerald-500/30">Zero_Trust_Protocol_V4</span>
                        </div>
                    </div>
                </div>
            </section>
        </div>
    );
}
