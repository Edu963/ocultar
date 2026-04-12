import { useState, useEffect, useRef } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { 
    Shield, Lock, Zap, BarChart3, ChevronRight, Github, ExternalLink, Globe, 
    ArrowRight, AlertTriangle, Code2, Terminal, Copy, Check, Info, Server, 
    Database, Activity, Cpu, Box, ShieldCheck, ChevronDown
} from 'lucide-react';
import RiskAssessmentPage from './RiskAssessmentPage';
import CalculatorPage from './CalculatorPage';
import SolutionsPage from './SolutionsPage';

// --- Subtle Particle Background ---
const CanvasBackground = () => {
    const canvasRef = useRef(null);
    useEffect(() => {
        const canvas = canvasRef.current;
        const ctx = canvas.getContext('2d');
        let animationFrameId;
        let particles = [];
        const numParticles = 30;
        const connectionDistance = 200;
        const resizeCanvas = () => {
            canvas.width = window.innerWidth;
            canvas.height = window.innerHeight;
        };
        window.addEventListener('resize', resizeCanvas);
        resizeCanvas();
        class Particle {
            constructor() {
                this.x = Math.random() * canvas.width;
                this.y = Math.random() * canvas.height;
                this.vx = (Math.random() - 0.5) * 0.2;
                this.vy = (Math.random() - 0.5) * 0.2;
                this.radius = Math.random() * 1.2;
            }
            update() {
                if (this.x <= 0 || this.x >= canvas.width) this.vx *= -1;
                if (this.y <= 0 || this.y >= canvas.height) this.vy *= -1;
                this.x += this.vx;
                this.y += this.vy;
            }
            draw() {
                ctx.beginPath();
                ctx.arc(this.x, this.y, this.radius, 0, Math.PI * 2);
                ctx.fillStyle = 'rgba(16, 185, 129, 0.15)';
                ctx.fill();
            }
        }
        for (let i = 0; i < numParticles; i++) particles.push(new Particle());
        const animate = () => {
            ctx.clearRect(0, 0, canvas.width, canvas.height);
            for (let i = 0; i < particles.length; i++) {
                particles[i].update();
                particles[i].draw();
                for (let j = i + 1; j < particles.length; j++) {
                    const dx = particles[i].x - particles[j].x;
                    const dy = particles[i].y - particles[j].y;
                    const distance = Math.sqrt(dx * dx + dy * dy);
                    if (distance < connectionDistance) {
                        ctx.beginPath();
                        ctx.moveTo(particles[i].x, particles[i].y);
                        ctx.lineTo(particles[j].x, particles[j].y);
                        const opacity = (1 - distance / connectionDistance) * 0.08;
                        ctx.strokeStyle = `rgba(16, 185, 129, ${opacity})`;
                        ctx.lineWidth = 0.5;
                        ctx.stroke();
                    }
                }
            }
            animationFrameId = requestAnimationFrame(animate);
        };
        animate();
        return () => {
            window.removeEventListener('resize', resizeCanvas);
            cancelAnimationFrame(animationFrameId);
        };
    }, []);
    return <canvas ref={canvasRef} id="zero-egress-canvas" />;
};

// --- Components ---

const Nav = () => (
    <nav className="fixed top-0 left-0 w-full z-50 py-5 transition-all duration-300">
        <div className="max-container flex justify-between items-center">
            <Link to="/" className="flex items-center gap-3 group">
                <div className="w-9 h-9 bg-emerald-500 rounded flex items-center justify-center">
                    <Shield className="text-black w-5 h-5" strokeWidth={3} />
                </div>
                <span className="text-white font-bold text-xl tracking-tighter">OCULTAR</span>
            </Link>
            <div className="hidden md:flex items-center gap-10 text-[12px] font-bold uppercase tracking-widest">
                <Link to="/" className="text-zinc-500 hover:text-white transition-colors">Platform</Link>
                <Link to="/solutions" className="text-zinc-500 hover:text-white transition-colors">Solutions</Link>
                <Link to="/risk-assessment" className="text-zinc-500 hover:text-white transition-colors">Risk Audit</Link>
                <Link to="/calculator" className="text-zinc-500 hover:text-white transition-colors">ROI Engine</Link>
                <a href="#" className="text-zinc-500 hover:text-white transition-colors flex items-center gap-1">Docs <ExternalLink className="w-3 h-3" /></a>
            </div>
            <Link to="/risk-assessment" className="bg-emerald-500 text-black text-[11px] font-bold px-6 py-3 rounded uppercase tracking-wider hover:bg-emerald-400 transition-colors">
                Start Pilot
            </Link>
        </div>
    </nav>
);

const Footer = () => (
    <footer className="bg-zinc-950 border-t border-white/5 py-20">
        <div className="max-container">
            <div className="grid grid-cols-1 md:grid-cols-4 gap-12">
                <div className="col-span-1 md:col-span-2 space-y-6">
                    <div className="font-bold text-xl tracking-tighter text-white">OCULTAR</div>
                    <p className="max-w-sm text-sm text-slate-500">
                        Zero-Egress architecture for the modern AI stack. Secure your technical perimeter by default.
                    </p>
                    <div className="flex gap-4">
                        <a href="https://github.com/Edu963/ocultar" className="text-slate-600 hover:text-emerald-500 transition-colors">
                            <Github className="w-5 h-5" />
                        </a>
                        <a href="#" className="text-slate-600 hover:text-emerald-500 transition-colors">
                            <Globe className="w-5 h-5" />
                        </a>
                    </div>
                </div>
                <div className="space-y-4">
                    <h4 className="text-[10px] font-bold uppercase tracking-widest text-emerald-500">Infrastructure</h4>
                    <ul className="space-y-2 text-sm text-slate-500">
                        <li><Link to="/risk-assessment" className="hover:text-white transition-colors">Refinery Engine</Link></li>
                        <li><Link to="/calculator" className="hover:text-white transition-colors">ROI Forecast</Link></li>
                        <li><a href="#" className="hover:text-white transition-colors">Status</a></li>
                    </ul>
                </div>
                <div className="space-y-4">
                    <h4 className="text-[10px] font-bold uppercase tracking-widest text-emerald-500">Contact</h4>
                    <ul className="space-y-2 text-sm text-slate-500">
                        <li><a href="#" className="hover:text-white">Sales</a></li>
                        <li><a href="mailto:engineering@ocultar.dev" className="hover:text-white">Engineering</a></li>
                    </ul>
                </div>
            </div>
            <div className="mt-20 pt-8 border-t border-white/5 text-[10px] text-zinc-600 font-mono flex justify-between items-center uppercase tracking-widest">
                <div>&copy; {new Date().getFullYear()} OCULTAR SECURITY</div>
                <div className="flex gap-6">
                    <span>BUILD: 4.12.2026</span>
                    <span className="text-emerald-500/50">LOCAL_ONLY_MODE</span>
                </div>
            </div>
        </div>
    </footer>
);

// --- Landing Page Sections ---

const BeforeAfter = () => {
    const before = JSON.stringify({
        "user_id": 4392,
        "name": "Arjun Patel",
        "email": "apatel@enterprise.com",
        "ssn": "928-11-2039",
        "query": "Analyze Patel's payroll for Q3"
    }, null, 2);

    const after = JSON.stringify({
        "user_id": 4392,
        "name": "[PERSON_X9282]",
        "email": "[EMAIL_K2394]",
        "ssn": "[SSN_M9283]",
        "query": "Analyze [PERSON_X9282]'s payroll for Q3"
    }, null, 2);

    return (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-px bg-white/5 border border-white/10 rounded-xl overflow-hidden mt-16 max-w-4xl mx-auto shadow-2xl shadow-black">
            <div className="bg-zinc-950 p-6 space-y-4">
                <div className="flex items-center gap-2 text-[10px] font-bold text-zinc-500 uppercase tracking-widest">
                    <div className="w-1.5 h-1.5 rounded-full bg-rose-500"></div>
                    Raw PII (Internal)
                </div>
                <pre className="text-xs font-mono text-zinc-400 overflow-x-auto">
                    <code>{before}</code>
                </pre>
            </div>
            <div className="bg-[#080809] p-6 space-y-4 border-l border-white/5">
                <div className="flex items-center gap-2 text-[10px] font-bold text-emerald-500 uppercase tracking-widest">
                    <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></div>
                    Refined (Egress)
                </div>
                <pre className="text-xs font-mono text-emerald-500/80 overflow-x-auto">
                    <code>{after}</code>
                </pre>
            </div>
        </div>
    );
};

const EnterpriseCTA = () => {
    const [copied, setCopied] = useState(false);
    const command = "docker-compose -f docker-compose.community.yml up -d";

    const copyToClipboard = () => {
        navigator.clipboard.writeText(command);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <div className="mt-12 flex flex-col items-center gap-6 w-full max-w-xl mx-auto">
            <div className="w-full bg-zinc-950 border border-white/5 rounded-xl p-1 overflow-hidden flex items-center gap-4 group">
                <div className="bg-zinc-900 px-4 py-3 rounded-l-lg border-r border-white/5 flex items-center gap-2">
                    <Terminal className="w-4 h-4 text-emerald-500" />
                    <span className="text-[10px] font-mono font-bold text-zinc-500 uppercase tracking-widest">Local Setup</span>
                </div>
                <code className="text-[11px] font-mono text-zinc-400 flex-grow truncate px-2">
                    {command}
                </code>
                <button 
                    onClick={copyToClipboard}
                    className="p-3 hover:text-emerald-500 transition-colors"
                    title="Copy to clipboard"
                >
                    {copied ? <Check className="w-4 h-4 text-emerald-500" /> : <Copy className="w-4 h-4" />}
                </button>
            </div>
            <div className="flex flex-col items-center gap-3">
                <div className="text-[10px] text-zinc-600 font-mono uppercase tracking-widest flex items-center gap-2">
                    <ShieldCheck className="w-3 h-3 text-emerald-500/50" />
                    Local-First Architecture — Zero External Dependencies
                </div>
                <Link to="/risk-assessment" className="text-emerald-500 hover:text-emerald-400 text-[10px] font-bold uppercase tracking-widest flex items-center gap-1 transition-all">
                    Request Full Enterprise Cluster Spec <ChevronRight className="w-3 h-3" />
                </Link>
            </div>
        </div>
    );
};

const TrustStrip = () => (
    <div className="bg-zinc-950 border-y border-white/5 py-10">
        <div className="max-container grid grid-cols-2 md:grid-cols-4 gap-8">
            <div className="flex flex-col items-center md:items-start gap-2 text-center md:text-left">
                <div className="text-emerald-500"><Lock className="w-5 h-5" /></div>
                <div className="text-[11px] font-bold text-white uppercase tracking-wider">Zero-Egress</div>
                <div className="text-[10px] text-slate-500 uppercase">Data never leaves your VPN</div>
            </div>
            <div className="flex flex-col items-center md:items-start gap-2 text-center md:text-left">
                <div className="text-emerald-500"><ShieldCheck className="w-5 h-5" /></div>
                <div className="text-[11px] font-bold text-white uppercase tracking-wider">Fail-Closed</div>
                <div className="text-[10px] text-slate-500 uppercase">Blocks unsafe requests</div>
            </div>
            <div className="flex flex-col items-center md:items-start gap-2 text-center md:text-left">
                <div className="text-emerald-500"><Terminal className="w-5 h-5" /></div>
                <div className="text-[11px] font-bold text-white uppercase tracking-wider">Local Analysis</div>
                <div className="text-[10px] text-slate-500 uppercase">No external scanning APIs</div>
            </div>
            <div className="flex flex-col items-center md:items-start gap-2 text-center md:text-left">
                <div className="text-emerald-500"><Database className="w-5 h-5" /></div>
                <div className="text-[11px] font-bold text-white uppercase tracking-wider">Deterministic</div>
                <div className="text-[10px] text-slate-500 uppercase">Vaulted token consistency</div>
            </div>
        </div>
    </div>
);

const HowItWorks = () => (
    <section className="section-padding bg-zinc-950 relative overflow-hidden">
        <div className="max-container">
            <div className="max-w-xl mb-20">
                <h4 className="text-emerald-500 text-[10px] uppercase font-mono tracking-[0.4em] mb-4">Architecture</h4>
                <h2 className="text-white mb-6">Built for Sovereignty</h2>
                <p className="text-slate-400">Ocultar acts as a transparent, high-performance proxy between your infrastructure and the AI provider.</p>
            </div>

            <div className="relative flex flex-col md:flex-row items-center justify-between gap-12 max-w-5xl mx-auto">
                <div className="flex-1 text-center p-8 bg-zinc-900 border border-white/5 rounded-2xl relative">
                    <div className="absolute -top-4 left-1/2 -translate-x-1/2 bg-zinc-800 border border-white/10 p-2 rounded-lg">
                        <Server className="w-5 h-5 text-emerald-500" />
                    </div>
                    <div className="text-sm font-bold mb-2">Internal Data</div>
                    <div className="text-[10px] text-zinc-500 uppercase">CRM, Logs, DBs</div>
                </div>
                
                <div className="w-px h-12 md:w-12 md:h-px bg-zinc-800 flex items-center justify-center">
                    <ArrowRight className="w-4 h-4 text-zinc-600 hidden md:block" />
                    <ChevronDown className="w-4 h-4 text-zinc-600 md:hidden" />
                </div>

                <div className="flex-1 text-center p-10 bg-emerald-500/5 border border-emerald-500/20 rounded-2xl relative shadow-[0_0_50px_rgba(16,185,129,0.1)]">
                    <div className="absolute -top-4 left-1/2 -translate-x-1/2 bg-emerald-500 p-2 rounded-lg shadow-lg">
                        <Shield className="w-5 h-5 text-black" />
                    </div>
                    <div className="text-sm font-bold text-white mb-2">OCULTAR Refinery</div>
                    <div className="text-[10px] text-emerald-400 uppercase">Vault stays local</div>
                </div>

                <div className="w-px h-12 md:w-12 md:h-px bg-zinc-800 flex items-center justify-center">
                    <ArrowRight className="w-4 h-4 text-zinc-600 hidden md:block" />
                    <ChevronDown className="w-4 h-4 text-zinc-600 md:hidden" />
                </div>

                <div className="flex-1 text-center p-8 bg-zinc-900 border border-white/5 rounded-2xl relative">
                    <div className="absolute -top-4 left-1/2 -translate-x-1/2 bg-zinc-800 border border-white/10 p-2 rounded-lg">
                        <Box className="w-5 h-5 text-sky-500" />
                    </div>
                    <div className="text-sm font-bold mb-2">AI Provider</div>
                    <div className="text-[10px] text-zinc-500 uppercase">OpenAI, Anthropic</div>
                </div>
            </div>
        </div>
    </section>
);

const FailClosedVisual = () => (
    <section className="section-padding border-t border-white/5">
        <div className="max-container grid grid-cols-1 md:grid-cols-2 gap-px bg-white/5 border border-white/5 rounded-3xl overflow-hidden">
            <div className="p-12 md:p-16 space-y-8 bg-zinc-950/20">
                <div className="space-y-4">
                    <h4 className="text-zinc-500 text-[10px] uppercase tracking-widest">Typical Solutions</h4>
                    <h3 className="text-zinc-400">Fail-Open Architecture</h3>
                    <p className="text-sm text-zinc-500">Traditional proxies prioritize uptime over security, allowing unscrubbed data to leak if scanning fails or times out.</p>
                </div>
                <div className="space-y-4">
                    <div className="flex items-center gap-3 text-rose-500/50">
                        <AlertTriangle className="w-4 h-4" />
                        <span className="text-[10px] font-bold uppercase tracking-wider">Potential PII Egress</span>
                    </div>
                    <div className="w-full h-1.5 bg-zinc-900 rounded-full overflow-hidden">
                        <div className="w-1/2 h-full bg-rose-500/50"></div>
                    </div>
                </div>
            </div>
            
            <div className="p-12 md:p-16 space-y-8 bg-emerald-500/[0.02]">
                <div className="space-y-4">
                    <h4 className="text-emerald-500 text-[10px] uppercase tracking-widest">OCULTAR</h4>
                    <h3 className="text-white">Fail-Closed Guarantee</h3>
                    <p className="text-sm text-slate-400">Engineering truth: we'd rather break a query than break a law. Any refinement error blocks the egress immediately.</p>
                </div>
                <div className="space-y-4">
                    <div className="flex items-center gap-3 text-emerald-500">
                        <ShieldCheck className="w-4 h-4" />
                        <span className="text-[10px] font-bold uppercase tracking-wider">100% Guaranteed Protection</span>
                    </div>
                    <div className="w-full h-1.5 bg-zinc-900 rounded-full overflow-hidden">
                        <div className="w-full h-full bg-emerald-500"></div>
                    </div>
                </div>
            </div>
        </div>
    </section>
);

const SovereignPacks = () => {
    const packs = [
        "FR_SIREN", "BR_CPF", "CL_RUT", "ES_DNI", "DE_STID", "UK_NINO", "IT_CF", "PL_PESEL", "IN_AADHAAR"
    ];
    return (
        <section className="section-padding bg-[#080809]">
            <div className="max-container text-center max-w-3xl">
                <h4 className="text-zinc-500 text-[10px] uppercase tracking-[0.3em] mb-4">Sovereign Packs</h4>
                <h2 className="text-white mb-6">Region-specific PII detection</h2>
                <div className="flex flex-wrap justify-center gap-3 mt-10">
                    {packs.map(p => (
                        <div key={p} className="px-4 py-2 bg-emerald-500/5 border border-emerald-500/10 rounded-full text-[10px] font-mono font-bold text-emerald-400 uppercase tracking-widest">
                            {p}
                        </div>
                    ))}
                    <div className="px-4 py-2 bg-white/5 border border-white/10 rounded-full text-[10px] font-mono font-bold text-zinc-500 uppercase tracking-widest">
                        + 40 MORE
                    </div>
                </div>
                <p className="text-sm text-zinc-500 mt-8">High-performance regex + validation logic built into the binary core.</p>
            </div>
        </section>
    );
};

const SimplifiedROI = () => {
    const [volume, setVolume] = useState(1); // Million tokens
    const savings = volume * 14500; // Mock calculation

    return (
        <section className="section-padding border-y border-white/5 relative overflow-hidden">
            <div className="max-container grid grid-cols-1 lg:grid-cols-2 gap-20 items-center">
                <div className="space-y-8">
                    <h4 className="text-emerald-500 text-[10px] uppercase tracking-widest">Risk Assessment</h4>
                    <h2 className="text-white">Quantify your risk in seconds</h2>
                    <p className="text-slate-400">Our engine calculates potential GDPR/CCPA liability based on volume and industry exposure probability.</p>
                    
                    <div className="space-y-6 bg-zinc-950 p-8 rounded-2xl border border-white/5">
                        <div className="space-y-4">
                            <label className="text-[10px] font-bold text-zinc-500 uppercase tracking-widest block">Monthly Token Volume (Millions)</label>
                            <input 
                                type="range" min="1" max="100" step="1"
                                value={volume} onChange={(e) => setVolume(parseInt(e.target.value))}
                                className="w-full h-1.5 bg-zinc-900 rounded-lg appearance-none cursor-pointer accent-emerald-500"
                            />
                            <div className="flex justify-between text-[11px] font-mono text-zinc-500">
                                <span>1M</span>
                                <span className="text-white">{volume}M</span>
                                <span>100M+</span>
                            </div>
                        </div>
                        
                        <div className="grid grid-cols-2 gap-6 pt-6 border-t border-white/5">
                            <div>
                                <div className="text-[10px] font-bold text-zinc-500 uppercase tracking-widest mb-1">Est. Liability</div>
                                <div className="text-2xl font-bold text-rose-500 font-mono">${(volume * 15000).toLocaleString()}</div>
                            </div>
                            <div className="text-right">
                                <div className="text-[10px] font-bold text-emerald-500/50 uppercase tracking-widest mb-1">OCULTAR Savings</div>
                                <div className="text-2xl font-bold text-emerald-500 font-mono">${savings.toLocaleString()}</div>
                            </div>
                        </div>
                        
                        <a href="/calculator.html" className="btn btn-secondary w-full py-4 text-[10px] uppercase font-bold tracking-widest">
                            Expand Full Risk Calculator
                        </a>
                    </div>
                </div>
                <div className="relative group">
                    <div className="absolute inset-0 bg-emerald-500/5 blur-3xl rounded-full scale-150"></div>
                    <div className="bg-zinc-950 border border-white/10 rounded-2xl p-8 relative shadow-2xl space-y-6">
                        <Activity className="text-emerald-500 w-8 h-8" />
                        <h3 className="text-white">Zero-Egress SLA</h3>
                        <div className="space-y-3">
                            {[
                                {label: "Enforcement Rate", value: "100.00%"},
                                {label: "Latency Overhead", value: "< 5ms"},
                                {label: "Egress Blocks", value: "Real-time"}
                            ].map(m => (
                                <div key={m.label} className="flex justify-between items-center py-2 border-b border-white/5">
                                    <span className="text-[11px] text-zinc-500 uppercase">{m.label}</span>
                                    <span className="text-xs font-mono font-bold text-emerald-400">{m.value}</span>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        </section>
    );
};

const UseCases = () => (
    <section className="section-padding">
        <div className="max-container">
            <h4 className="text-emerald-500 text-[10px] uppercase tracking-widest mb-4">Implementation</h4>
            <h2 className="text-white mb-20">Concrete Use Cases</h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                {[
                    { title: "CRM to LLM", desc: "Send customer records to external models safely by automatically tokenizing identities.", icon: <Database /> },
                    { title: "Audit Logs", desc: "Analyze sensitive debug logs without leaking internal IP addresses or system credentials.", icon: <Cpu /> },
                    { title: "AI Support", desc: "Enable contextual customer support chatbots without exposing plain-text PII to the model.", icon: <Activity /> }
                ].map(c => (
                    <div key={c.title} className="card group space-y-6">
                        <div className="w-10 h-10 bg-emerald-500/10 rounded flex items-center justify-center text-emerald-500 group-hover:bg-emerald-500 group-hover:text-black transition-all">
                            {c.icon}
                        </div>
                        <h3 className="text-xl">{c.title}</h3>
                        <p className="text-sm text-slate-500">{c.desc}</p>
                    </div>
                ))}
            </div>
        </div>
    </section>
);

const DataSources = () => (
    <section className="py-20 bg-zinc-950 border-t border-white/5">
        <div className="max-container">
            <details className="group border border-white/5 rounded-2xl p-6 bg-zinc-950 hover:bg-zinc-900/50 transition-colors">
                <summary className="flex justify-between items-center cursor-pointer list-none">
                    <div className="flex items-center gap-4">
                        <Info className="w-5 h-5 text-emerald-500" />
                        <span className="text-sm font-bold uppercase tracking-wider">Data Sources & Trust Assumptions</span>
                    </div>
                    <ChevronDown className="w-4 h-4 text-zinc-600 group-open:rotate-180 transition-transform" />
                </summary>
                <div className="mt-8 grid grid-cols-1 md:grid-cols-3 gap-8 text-xs text-zinc-500 leading-relaxed uppercase tracking-widest font-mono">
                    <div className="space-y-4">
                        <div className="text-white">Validation Models</div>
                        <div>Costing based on avg IBM/AWS security incident data (2024). Egress logic compliant with FIPS 140-2 requirements.</div>
                    </div>
                    <div className="space-y-4">
                        <div className="text-white">Regional PII</div>
                        <div>Detection logic verified against GDPR Article 4 and CCPA 1798.140 definitions.</div>
                    </div>
                    <div className="space-y-4">
                        <div className="text-white">SLA Metrics</div>
                        <div>Testing performed on standard AWS c5.large instances across 50k concurrent requests.</div>
                    </div>
                </div>
            </details>
        </div>
    </section>
);

// --- Main Application ---
export default function App() {
    return (
        <Router>
            <AppContent />
        </Router>
    );
}

function AppContent() {
    return (
        <div className="min-h-screen flex flex-col selection:bg-emerald-500 selection:text-black">
            <CanvasBackground />
            <Nav />
            <div className="flex-grow pt-32">
                <Routes>
                    <Route path="/" element={<LandingPage />} />
                    <Route path="/risk-assessment" element={<RiskAssessmentPage />} />
                    <Route path="/calculator" element={<CalculatorPage />} />
                    <Route path="/solutions" element={<SolutionsPage />} />
                </Routes>
            </div>
            <Footer />
        </div>
    );
}

function LandingPage() {
    return (
        <div className="animate-fade-in-up">
            <section className="section-padding flex flex-col items-center">
                <div className="max-container text-center relative z-10">
                    <div className="badge mb-10">
                        <span className="relative flex h-2 w-2">
                            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                            <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
                        </span>
                        Enterprise-Grade Zero-Egress Proxy
                    </div>
                    <h1 className="max-w-5xl mx-auto mb-8">
                        Use AI with <span className="text-emerald-500">sensitive data</span> — without it ever leaving your infrastructure.
                    </h1>
                    <p className="max-w-2xl mx-auto text-xl text-slate-400 mb-12">
                        OCULTAR is a built-in PII refinery that runs locally, ensuring 100% security 
                        sovereignty before your data reaches any cloud-based LLM.
                    </p>
                    <div className="flex flex-col sm:flex-row gap-5 justify-center">
                        <Link to="/risk-assessment" className="btn btn-primary px-10 py-5 text-lg group">
                            Start Pilot <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform ml-2" />
                        </Link>
                        <button className="btn btn-secondary px-10 py-5 text-lg">
                            Run Live Demo
                        </button>
                    </div>
                    <EnterpriseCTA />
                    <BeforeAfter />
                </div>
            </section>

            <TrustStrip />
            <HowItWorks />
            <FailClosedVisual />
            <SovereignPacks />
            <UseCases />
            <DataSources />

            <section className="section-padding relative overflow-hidden bg-zinc-950 border-t border-white/5">
                <div className="max-container text-center relative z-10 space-y-12">
                    <h2 className="text-white max-w-3xl mx-auto">Ready for Technical Sovereignty?</h2>
                    <div className="flex justify-center gap-6">
                        <Link to="/risk-assessment" className="btn btn-primary px-16 py-6 text-xl">
                            Initialize Deployment
                        </Link>
                    </div>
                </div>
            </section>
        </div>
    );
}
