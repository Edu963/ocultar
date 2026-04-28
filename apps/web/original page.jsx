import React, { useState, useEffect } from 'react';
import { 
  Shield, 
  Terminal, 
  Zap, 
  Lock, 
  ArrowRight, 
  CheckCircle2, 
  XCircle, 
  Cpu, 
  Globe, 
  Database, 
  Code2, 
  AlertTriangle,
  ChevronRight,
  Github,
  RefreshCw,
  Shuffle,
  Layers
} from 'lucide-react';

const App = () => {
  const [isProxyActive, setIsProxyActive] = useState(true);

  // Smooth scroll helper
  const scrollTo = (id) => {
    const element = document.getElementById(id);
    element?.scrollIntoView({ behavior: 'smooth' });
  };

  return (
    <div className="min-h-screen bg-[#050505] text-slate-300 font-sans selection:bg-emerald-500/30 selection:text-emerald-400">
      {/* Navigation */}
      <nav className="fixed top-0 w-full z-50 border-b border-white/5 bg-[#050505]/80 backdrop-blur-md">
        <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 bg-emerald-500 flex items-center justify-center rounded-sm">
              <Shield className="text-black w-5 h-5" strokeWidth={2.5} />
            </div>
            <span className="font-mono font-bold tracking-tighter text-white text-xl">SOUVERAIN</span>
          </div>
          <div className="hidden md:flex items-center gap-8 text-sm font-medium">
            <button onClick={() => scrollTo('features')} className="hover:text-emerald-400 transition-colors">Technology</button>
            <button onClick={() => scrollTo('sombra')} className="hover:text-emerald-400 transition-colors">Sombra</button>
            <button onClick={() => scrollTo('comparison')} className="hover:text-emerald-400 transition-colors">The Moat</button>
            <button onClick={() => scrollTo('pricing')} className="hover:text-emerald-400 transition-colors">Pricing</button>
            <button 
              onClick={() => scrollTo('pricing')}
              className="bg-emerald-500 text-black px-4 py-2 rounded-sm font-bold hover:bg-emerald-400 transition-all active:scale-95"
            >
              Start Pilot
            </button>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="relative pt-32 pb-20 px-6 overflow-hidden">
        <div className="absolute inset-0 z-0 opacity-20 pointer-events-none" 
          style={{ backgroundImage: 'radial-gradient(#10b981 0.5px, transparent 0.5px)', backgroundSize: '24px 24px' }}>
        </div>
        
        <div className="max-w-7xl mx-auto relative z-10 grid lg:grid-cols-2 gap-12 items-center">
          <div className="space-y-8">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-emerald-500/30 bg-emerald-500/5 text-emerald-400 text-xs font-mono">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
              </span>
              v2.4.0 Engine Live: 100% Zero-Egress Guaranteed
            </div>
            
            <h1 className="text-5xl lg:text-7xl font-bold text-white tracking-tight leading-[1.1]">
              Stop Leaking PII to the Cloud. <br/>
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-emerald-400 to-cyan-500">
                Start Using AI Safely.
              </span>
            </h1>
            
            <p className="text-xl text-slate-400 max-w-xl leading-relaxed">
              The only <span className="text-white border-b border-emerald-500/50">Zero-Egress PII Refinery</span>. 
              Intercept, Vault, and Redact sensitive data locally before it ever reaches an LLM provider.
            </p>
            
            <div className="flex flex-col sm:flex-row gap-4">
              <button className="flex items-center justify-center gap-2 bg-emerald-500 text-black px-8 py-4 rounded-sm font-bold text-lg hover:bg-emerald-400 transition-all shadow-[0_0_20px_rgba(16,185,129,0.2)]">
                Start 10-Minute Pilot <ArrowRight className="w-5 h-5" />
              </button>
              <button className="flex items-center justify-center gap-2 bg-white/5 border border-white/10 text-white px-8 py-4 rounded-sm font-bold text-lg hover:bg-white/10 transition-all">
                <Github className="w-5 h-5" /> View on GitHub
              </button>
            </div>
          </div>

          <div className="relative group">
            <div className="absolute -inset-1 bg-gradient-to-r from-emerald-500 to-cyan-500 rounded-lg blur opacity-25 group-hover:opacity-40 transition duration-1000"></div>
            <div className="relative bg-[#0d0d0f] border border-white/10 rounded-lg overflow-hidden shadow-2xl">
              <div className="flex items-center gap-2 px-4 py-3 bg-white/5 border-b border-white/10">
                <div className="flex gap-1.5">
                  <div className="w-3 h-3 rounded-full bg-red-500/20"></div>
                  <div className="w-3 h-3 rounded-full bg-yellow-500/20"></div>
                  <div className="w-3 h-3 rounded-full bg-emerald-500/20"></div>
                </div>
                <div className="text-[10px] font-mono text-slate-500 uppercase tracking-widest ml-4">Local Proxy Log — v2.4</div>
              </div>
              <div className="p-6 font-mono text-sm leading-relaxed">
                <div className="text-emerald-500">➜ /usr/local/souverain --start-proxy</div>
                <div className="text-slate-500">Initialising local SLM deep scan... [OK]</div>
                <div className="text-slate-500">Listening on http://localhost:8080</div>
                <div className="mt-4 text-slate-400">
                  <span className="text-blue-400">[INCOMING]</span> POST /v1/chat/completions <br/>
                  <span className="text-yellow-400">[PII_DETECTED]</span> "John Doe" (NAME) <br/>
                  <span className="text-yellow-400">[PII_DETECTED]</span> "john.doe@gmail.com" (EMAIL)
                </div>
                <div className="mt-4 text-emerald-400">
                  <span className="text-emerald-500">[REDACTING]</span> Tokenizing payload... <br/>
                  <span className="text-emerald-500">[FORWARDING]</span> Clean data sent to api.openai.com
                </div>
                <div className="mt-2 flex gap-1">
                  <span className="animate-pulse bg-emerald-500 w-2 h-5 inline-block"></span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Compliance Badge Bar */}
      <div className="border-y border-white/5 bg-white/2 py-8 overflow-hidden whitespace-nowrap">
        <div className="max-w-7xl mx-auto flex justify-around items-center opacity-40 grayscale hover:grayscale-0 transition-all gap-12 px-6">
          {['GDPR COMPLIANT', 'EU AI ACT READY', 'HIPAA SECURE', 'NIS2 CERTIFIED', 'BSI C5 ALIGNED'].map((text) => (
            <div key={text} className="flex items-center gap-2 font-mono text-sm tracking-tighter text-white">
              <Shield className="w-4 h-4 text-emerald-500" />
              {text}
            </div>
          ))}
        </div>
      </div>

      {/* From Code to Compliance Section */}
      <section id="features" className="py-24 px-6 bg-[#080809]">
        <div className="max-w-7xl mx-auto">
          <div className="text-center space-y-4 mb-20">
            <h2 className="text-3xl font-bold text-white uppercase tracking-wider font-mono">From Code to Compliance</h2>
            <p className="text-slate-400 max-w-2xl mx-auto">See how Souverain acts as a transparent refinery between your application and the cloud.</p>
          </div>

          <div className="grid lg:grid-cols-5 gap-8 items-center">
            <div className="lg:col-span-2 space-y-12">
              <div className="space-y-4">
                <div className="w-10 h-10 bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-500 font-mono font-bold">01</div>
                <h3 className="text-xl font-bold text-white">One Line of Config</h3>
                <p className="text-slate-400 leading-relaxed">
                  Simply update your LLM base URL to point to <code className="text-emerald-400 bg-emerald-400/5 px-1">localhost:8080</code>. 
                  No code logic changes, no library rewrites.
                </p>
              </div>
              <div className="space-y-4">
                <div className="w-10 h-10 bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-500 font-mono font-bold">02</div>
                <h3 className="text-xl font-bold text-white">Transparent Interception</h3>
                <p className="text-slate-400 leading-relaxed">
                  Our Proxy intercepts the stream, replaces names/emails with encrypted tokens, and passes clean data to providers.
                </p>
              </div>
            </div>

            <div className="lg:col-span-3 bg-black/40 border border-white/5 rounded-2xl p-8 relative overflow-hidden">
               <div className="absolute top-4 right-4 flex gap-4">
                  <button onClick={() => setIsProxyActive(false)} className={`px-3 py-1 text-xs font-mono transition-colors ${!isProxyActive ? 'text-red-400 border border-red-400/30' : 'text-slate-500'}`}>RAW FLOW</button>
                  <button onClick={() => setIsProxyActive(true)} className={`px-3 py-1 text-xs font-mono transition-colors ${isProxyActive ? 'text-emerald-400 border border-emerald-400/30' : 'text-slate-500'}`}>SOUVERAIN</button>
               </div>
               <div className="flex flex-col items-center gap-12 mt-8">
                  <div className="flex flex-col md:flex-row items-center justify-between w-full max-w-2xl gap-8">
                    <div className="text-center space-y-2">
                      <div className="w-16 h-16 bg-white/5 border border-white/10 flex items-center justify-center rounded-lg"><Database className="w-8 h-8 text-slate-400" /></div>
                      <span className="text-[10px] font-mono text-slate-500 uppercase">Your App Data</span>
                    </div>
                    <div className="flex-1 flex items-center justify-center relative">
                      <div className={`h-1 w-full transition-colors duration-500 ${isProxyActive ? 'bg-emerald-500/30' : 'bg-red-500/30'}`}></div>
                      {isProxyActive ? (
                        <div className="absolute inset-0 flex items-center justify-center">
                          <div className="bg-emerald-500 border-4 border-[#0a0a0b] p-3 rounded-full shadow-[0_0_30px_rgba(16,185,129,0.4)] animate-pulse">
                            <Shield className="w-8 h-8 text-black" />
                          </div>
                        </div>
                      ) : (
                        <div className="absolute inset-0 flex items-center justify-center"><XCircle className="w-8 h-8 text-red-500" /></div>
                      )}
                    </div>
                    <div className="text-center space-y-2">
                      <div className="w-16 h-16 bg-white/5 border border-white/10 flex items-center justify-center rounded-lg"><Globe className="w-8 h-8 text-slate-400" /></div>
                      <span className="text-[10px] font-mono text-slate-500 uppercase">Cloud LLM</span>
                    </div>
                  </div>
                  <div className="w-full bg-white/2 rounded-lg p-6 font-mono text-xs">
                    <div className="grid grid-cols-2 gap-8">
                       <div className="space-y-2 border-r border-white/10 pr-4">
                          <div className="text-slate-500 mb-2">INPUT DATA:</div>
                          <div className="text-slate-300">"Invite <span className={isProxyActive ? 'text-emerald-400' : 'text-red-400 font-bold'}>Sarah Jenkins</span> to the board meeting at <span className={isProxyActive ? 'text-emerald-400' : 'text-red-400 font-bold'}>sarah@corp.com</span>"</div>
                       </div>
                       <div className="space-y-2 pl-4">
                          <div className="text-slate-500 mb-2">EGRESS DATA:</div>
                          <div className="text-slate-300">
                            {isProxyActive ? (
                              <span>"Invite <span className="text-emerald-400">[SVN_TK_9122]</span> to the board meeting at <span className="text-emerald-400">[SVN_TK_9123]</span>"</span>
                            ) : (
                              <span className="text-red-400 font-bold underline decoration-red-500">PII LEAKED: "Sarah Jenkins", "sarah@corp.com"</span>
                            )}
                          </div>
                       </div>
                    </div>
                  </div>
               </div>
            </div>
          </div>
        </div>
      </section>

      {/* Sombra Power-Up Section */}
      <section id="sombra" className="py-24 px-6 bg-[#050505] relative border-y border-white/5">
        <div className="max-w-7xl mx-auto">
          <div className="flex flex-col md:flex-row items-end justify-between mb-16 gap-6">
            <div className="space-y-4">
              <div className="text-emerald-500 font-mono text-xs tracking-widest">[ ENGINE_EXT_SOMBRA_V1 ]</div>
              <h2 className="text-4xl md:text-5xl font-bold text-white tracking-tight">
                Sombra: The Brain of your <br/> Privacy Infrastructure.
              </h2>
            </div>
            <p className="text-slate-400 max-w-sm text-sm font-mono leading-relaxed border-l border-emerald-500/30 pl-6">
              Sombra isn't just a filter; it's a Go-based gateway that makes your local environment 
              more capable than the cloud providers themselves.
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-8">
            {/* Feature 1 */}
            <div className="group p-8 bg-white/2 border border-white/5 rounded-sm hover:border-emerald-500/50 transition-all">
              <div className="w-12 h-12 bg-emerald-500/10 flex items-center justify-center mb-6 text-emerald-500">
                <Shuffle className="w-6 h-6" />
              </div>
              <h3 className="text-xl font-bold text-white mb-3">Multi-Model Routing</h3>
              <p className="text-slate-400 text-sm leading-relaxed mb-6">
                Dynamic prompt steering. Automatically route sensitive payloads to local Llama 3/Mistral, 
                while clean data flows to GPT-4 or Claude 3.5.
              </p>
              <div className="flex gap-2 opacity-50">
                <span className="text-[10px] font-mono bg-white/10 px-2 py-0.5">LOCAL_LLM</span>
                <span className="text-[10px] font-mono bg-white/10 px-2 py-0.5">GPT_4O</span>
              </div>
            </div>

            {/* Feature 2: The "Magic" */}
            <div className="relative group p-8 bg-emerald-500 text-black rounded-sm shadow-[0_0_30px_rgba(16,185,129,0.1)] transition-all overflow-hidden">
              <div className="absolute top-0 right-0 p-4 opacity-20"><RefreshCw className="w-20 h-20 rotate-12" /></div>
              <div className="relative z-10">
                <div className="w-12 h-12 bg-black/10 flex items-center justify-center mb-6">
                  <RefreshCw className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold mb-3 italic">Live Re-Hydration</h3>
                <p className="text-black/80 text-sm leading-relaxed mb-6 font-medium">
                  The AI gets tokens, the user gets reality. Mask data before egress; 
                  Sombra automatically swaps real PII back in for the local end-user. 
                  Zero AI confusion.
                </p>
                <div className="text-[10px] font-mono font-bold uppercase tracking-widest border-b border-black/20 pb-1 inline-block">
                  "Transparent to the User"
                </div>
              </div>
            </div>

            {/* Feature 3 */}
            <div className="group p-8 bg-white/2 border border-white/5 rounded-sm hover:border-emerald-500/50 transition-all">
              <div className="w-12 h-12 bg-emerald-500/10 flex items-center justify-center mb-6 text-emerald-500">
                <Layers className="w-6 h-6" />
              </div>
              <h3 className="text-xl font-bold text-white mb-3">Protocol Agnostic</h3>
              <p className="text-slate-400 text-sm leading-relaxed mb-6">
                Not just REST. Secure WebSockets, gRPC, and local file streams through a 
                single, low-latency Go binary that outperforms traditional proxies.
              </p>
              <div className="flex gap-2">
                <span className="text-[10px] font-mono text-emerald-500">&lt;1ms LATENCY</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* "Fail-Closed" Guarantee */}
      <section className="py-24 px-6 relative bg-red-950/10 border-y border-red-500/10">
        <div className="max-w-4xl mx-auto text-center space-y-8">
          <div className="inline-flex items-center gap-2 text-red-400 font-mono text-sm">
            <AlertTriangle className="w-5 h-5" />
            SECURITY PROTOCOL 403-HARD
          </div>
          <h2 className="text-4xl md:text-5xl font-bold text-white tracking-tight">
            We’d Rather Break the App <br/> than Break the Law.
          </h2>
          <p className="text-xl text-slate-400 leading-relaxed">
            Unlike competitors who "best effort" redaction, SOUVERAIN is <span className="text-white font-bold italic">Fail-Closed</span>. 
            If our Tier 2 Deep Scan hits a resource limit or a processing error, we return a 403. 
            Raw PII never, ever leaves your perimeter. This is your 
            <span className="text-white border-b-2 border-red-500 mx-1">GDPR Article 28 insurance policy</span>.
          </p>
        </div>
      </section>

      {/* Comparison Table */}
      <section id="comparison" className="py-24 px-6 bg-[#050505]">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-16">
            <h2 className="text-3xl font-bold text-white mb-4">The Compliance Gap</h2>
            <p className="text-slate-400">Why enterprise security teams are moving away from SaaS PII cleaners.</p>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="border-b border-white/10">
                  <th className="py-6 px-4 font-mono text-sm text-slate-500">FEATURE</th>
                  <th className="py-6 px-4 text-emerald-400 font-bold text-lg">SOUVERAIN</th>
                  <th className="py-6 px-4 text-slate-500 font-medium">LEGACY GIANTS</th>
                </tr>
              </thead>
              <tbody className="text-sm">
                {[
                  { feature: 'Data Egress', sv: '100% Local / Zero Egress', legacy: 'Often SaaS / Cloud-Hybrid' },
                  { feature: 'Setup Time', sv: '10 Minutes (Docker)', legacy: '3-6 Months (Consulting)' },
                  { feature: 'Pricing', sv: 'Flat-Fee (€50k)', legacy: 'Usage-Based "Tax"' },
                  { feature: 'Security Model', sv: 'Fail-Closed (Return 403)', legacy: 'Fail-Open / Log Error' },
                  { feature: 'Vaulting', sv: 'Local DuckDB Vault', legacy: 'Centralized DB' }
                ].map((row, idx) => (
                  <tr key={idx} className="border-b border-white/5 hover:bg-white/[0.02] transition-colors">
                    <td className="py-6 px-4 text-slate-400 font-mono uppercase tracking-widest text-[11px]">{row.feature}</td>
                    <td className="py-6 px-4 text-white font-bold flex items-center gap-2">
                      <CheckCircle2 className="w-4 h-4 text-emerald-500" /> {row.sv}
                    </td>
                    <td className="py-6 px-4 text-slate-600">{row.legacy}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </section>

      {/* Developer Experience */}
      <section className="py-24 px-6 bg-[#080809]">
        <div className="max-w-7xl mx-auto grid lg:grid-cols-2 gap-16">
          <div className="space-y-8">
            <h2 className="text-4xl font-bold text-white">Developer-First DX</h2>
            <p className="text-slate-400 text-lg">
              We built Souverain to be invisible to your devs. It lives in your CI/CD, 
              your CLI, and your Docker Compose.
            </p>
            <div className="space-y-6">
              {[
                { icon: <Code2 />, title: "Git Pre-commit Hooks", desc: "Block local files containing PII before they even hit your git history." },
                { icon: <Terminal />, title: "Power CLI", desc: "Bulk CSV/JSON cleaning with high-concurrency DuckDB processing." },
                { icon: <Database />, title: "DuckDB Vault", desc: "Encrypted local key-value store for re-identifying data for your eyes only." }
              ].map((item, idx) => (
                <div key={idx} className="flex gap-4">
                  <div className="text-emerald-500 mt-1">{item.icon}</div>
                  <div>
                    <h4 className="text-white font-bold mb-1">{item.title}</h4>
                    <p className="text-sm text-slate-500">{item.desc}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
          <div className="bg-[#050505] p-6 rounded-xl border border-white/5 shadow-2xl relative">
             <div className="flex justify-between items-center mb-4">
                <span className="text-[11px] font-mono text-slate-500">deployment.yaml</span>
                <span className="text-emerald-500 text-[10px] font-mono">READY TO DEPLOY</span>
             </div>
             <pre className="text-sm font-mono text-slate-400 overflow-x-auto">
               <code>{`services:
  llm-proxy:
    image: ghcr.io/souverain/proxy:v2.4
    environment:
      - MODE=fail_closed
      - VAULT_BACKEND=duckdb
      - TIER2_SCAN=enabled
    ports:
      - "8080:8080"
    volumes:
      - ./vault:/data/vault
      
  app:
    image: your-app:latest
    environment:
      - OPENAI_BASE_URL=http://llm-proxy:8080/v1`}
               </code>
             </pre>
          </div>
        </div>
      </section>

      {/* Pricing */}
      <section id="pricing" className="py-32 px-6 bg-[#050505]">
        <div className="max-w-4xl mx-auto">
          <div className="text-center mb-16 space-y-4">
            <h2 className="text-4xl font-bold text-white">The Enterprise Pilot</h2>
            <p className="text-slate-400">Lock in your PII perimeter with founder-level support.</p>
          </div>
          <div className="relative group">
            <div className="absolute -inset-1 bg-gradient-to-r from-emerald-500 to-cyan-500 rounded-2xl blur opacity-20 group-hover:opacity-40 transition duration-1000"></div>
            <div className="relative bg-[#0d0d0f] border border-emerald-500/20 rounded-2xl p-10 md:p-16 flex flex-col md:flex-row gap-12 items-center">
              <div className="flex-1 space-y-6">
                <div className="space-y-2">
                  <h3 className="text-2xl font-bold text-white">Annual Pilot License</h3>
                  <div className="flex items-baseline gap-2">
                    <span className="text-5xl font-bold text-white">€4,167</span>
                    <span className="text-slate-500">/ month</span>
                  </div>
                  <div className="text-slate-500 text-sm italic underline">Billed annually (€50k total)</div>
                </div>
                <div className="space-y-4 pt-4">
                  {[
                    "Tier 2 SLM Deep Scan Engine",
                    "PostgreSQL HA Vault Support",
                    "Infinite Seat License",
                    "Foundational Compliance Audit Logs",
                    "Direct Founder Slack Channel"
                  ].map((feature, idx) => (
                    <div key={idx} className="flex items-center gap-3 text-slate-300">
                      <CheckCircle2 className="w-5 h-5 text-emerald-500" />
                      {feature}
                    </div>
                  ))}
                </div>
              </div>
              <div className="w-full md:w-auto">
                <button className="w-full bg-emerald-500 text-black px-12 py-6 rounded-sm font-bold text-xl hover:bg-emerald-400 transition-all shadow-[0_0_40px_rgba(16,185,129,0.3)]">
                  Start My Pilot
                </button>
                <p className="text-center mt-4 text-xs font-mono text-slate-600 uppercase tracking-widest">5 of 5 Slots Remaining for Q1</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="py-16 px-6 border-t border-white/5 bg-[#050505]">
        <div className="max-w-7xl mx-auto flex flex-col md:flex-row justify-between items-center gap-8">
          <div className="flex items-center gap-2">
            <div className="w-6 h-6 bg-slate-800 flex items-center justify-center rounded-sm"><Shield className="text-emerald-500 w-4 h-4" /></div>
            <span className="font-mono font-bold tracking-tighter text-white">SOUVERAIN</span>
          </div>
          <div className="text-slate-600 font-mono text-sm tracking-widest uppercase">"The Switzerland of Data"</div>
          <div className="flex gap-6 text-slate-500 text-sm font-mono">
             <button className="hover:text-emerald-400 transition-colors">STATUS: OK</button>
             <button className="hover:text-emerald-400 transition-colors">SECURITY</button>
             <button className="hover:text-emerald-400 transition-colors">LEGAL</button>
          </div>
        </div>
        <div className="text-center mt-12 text-[10px] text-slate-800 font-mono uppercase tracking-[0.2em]">
           &copy; {new Date().getFullYear()} Souverain Technologies. All rights reserved. Zero-Egress Guaranteed.
        </div>
      </footer>
    </div>
  );
};

export default App;
