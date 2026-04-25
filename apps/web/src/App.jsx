import { useState, useEffect, useRef } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import {
    Shield, Lock, BarChart3, ChevronRight, Github, Globe,
    ArrowRight, AlertTriangle, Terminal, Check, Server, Database, Activity,
    ShieldCheck, Eye, RefreshCw, CheckCircle2
} from 'lucide-react';
import RiskAssessmentPage from './RiskAssessmentPage';
import CalculatorPage from './CalculatorPage';
import SolutionsPage from './SolutionsPage';
import DocsLayout from './components/DocsLayout';
import DocPage from './components/DocPage';
import ReferencePage from './components/ReferencePage';
import DocsLanding from './components/DocsLanding';
import DocsHeader from './components/DocsHeader';
import { useLocation } from 'react-router-dom';

const DEMO_URL =
    'mailto:sales@ocultar.dev?subject=Demo%20Request&body=Hi%2C%20I%27d%20like%20to%20book%20a%2030-minute%20demo%20of%20OCULTAR.';

// ── OCULTAR Wordmark ──────────────────────────────────────────────────────────
const OcultarWordmark = () => (
    <div className="font-mono font-black text-4xl tracking-widest text-gray-900 mb-6">
        OCULTAR
    </div>
);

// ── Particle Background (preserved exactly) ───────────────────────────────────
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
                ctx.fillStyle = 'rgba(8, 145, 178, 0.55)';
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
                        ctx.strokeStyle = `rgba(8, 145, 178, ${opacity})`;
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

// ── Nav ───────────────────────────────────────────────────────────────────────
const Nav = () => (
    <nav className="fixed top-0 left-0 w-full z-50 py-5 transition-all duration-300">
        <div className="max-container flex justify-between items-center">
            <Link to="/" className="flex items-center gap-3 group">
                <span className="font-mono font-black text-xl tracking-widest text-gray-900">OCULTAR</span>
            </Link>
            <div className="hidden md:flex items-center gap-10 text-[12px] font-bold uppercase tracking-widest">
                <Link to="/" className="text-gray-500 hover:text-gray-900 transition-colors">Platform</Link>
                <Link to="/solutions" className="text-gray-500 hover:text-gray-900 transition-colors">Product Suite</Link>
                <Link to="/calculator" className="text-gray-500 hover:text-gray-900 transition-colors">ROI Calculator</Link>
            </div>
            <a
                href={DEMO_URL}
                className="bg-cyan-600 text-white text-[11px] font-bold px-6 py-3 rounded uppercase tracking-wider hover:bg-cyan-700 transition-colors"
            >
                Book a Demo
            </a>
        </div>
    </nav>
);

// ── Footer ────────────────────────────────────────────────────────────────────
const Footer = () => (
    <footer className="bg-gray-50 border-t border-gray-100 py-20">
        <div className="max-container">
            <div className="grid grid-cols-1 md:grid-cols-4 gap-12">
                <div className="col-span-1 md:col-span-2 space-y-6">
                    <span className="font-mono font-black text-xl tracking-widest text-gray-900">OCULTAR</span>
                    <p className="max-w-sm text-sm text-gray-500">
                        Zero-Egress architecture for the modern AI stack. Secure your technical perimeter by default.
                    </p>
                    <div className="flex gap-4">
                        <a href="https://github.com/Edu963/ocultar" className="text-gray-400 hover:text-cyan-600 transition-colors">
                            <Github className="w-5 h-5" />
                        </a>
                        <a href="#" className="text-gray-400 hover:text-cyan-600 transition-colors">
                            <Globe className="w-5 h-5" />
                        </a>
                    </div>
                </div>
                <div className="space-y-4">
                    <h4 className="text-[10px] font-bold uppercase tracking-widest text-cyan-600">Product</h4>
                    <ul className="space-y-2 text-sm text-gray-500">
                        <li><Link to="/solutions" className="hover:text-gray-900 transition-colors">Product Suite</Link></li>
                        <li><Link to="/calculator" className="hover:text-gray-900 transition-colors">ROI Calculator</Link></li>
                    </ul>
                </div>
                <div className="space-y-4">
                    <h4 className="text-[10px] font-bold uppercase tracking-widest text-cyan-600">Contact</h4>
                    <ul className="space-y-2 text-sm text-gray-500">
                        <li><a href={DEMO_URL} className="hover:text-gray-900">Book a Demo</a></li>
                        <li><a href="mailto:engineering@ocultar.dev" className="hover:text-gray-900">Engineering</a></li>
                    </ul>
                </div>
            </div>
            <div className="mt-20 pt-8 border-t border-gray-100 text-[10px] text-gray-400 font-mono flex justify-between items-center uppercase tracking-widest">
                <div>&copy; {new Date().getFullYear()} OCULTAR SECURITY</div>
                <div className="flex gap-6">
                    <span>BUILD: 4.12.2026</span>
                    <span className="text-cyan-600/40">LOCAL_ONLY_MODE</span>
                </div>
            </div>
        </div>
    </footer>
);

// ── 1. Hero ───────────────────────────────────────────────────────────────────
const Hero = () => (
    <section className="section-padding flex flex-col items-center">
        <div className="max-container text-center relative z-10">

            <div className="badge mb-10">
                <span className="relative flex h-2 w-2">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-cyan-400 opacity-75"></span>
                    <span className="relative inline-flex rounded-full h-2 w-2 bg-cyan-500"></span>
                </span>
                Zero-Egress · On-Premise · GDPR Compliant by Architecture
            </div>
            <h1 className="max-w-5xl mx-auto mb-8">
                Your team uses AI. Your customers' data shouldn't leave{' '}
                <span className="text-cyan-500">your building.</span>
            </h1>
            <p className="max-w-2xl mx-auto text-xl text-slate-400 mb-12">
                <span className="text-cyan-500">OCULTAR</span> is a zero-egress PII refinery that runs entirely
                in your infrastructure. No shared servers. No data exposure. GDPR compliant by architecture.
            </p>
            <a href={DEMO_URL} className="btn btn-primary px-12 py-5 text-lg group inline-flex items-center">
                Book a 30-minute Demo
                <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform ml-2" />
            </a>
        </div>
    </section>
);

// ── Trust Strip (preserved, sits between hero and demo) ───────────────────────
const TrustStrip = () => (
    <div className="bg-gray-50 border-y border-gray-100 py-10">
        <div className="max-container grid grid-cols-2 md:grid-cols-4 gap-8">
            {[
                { icon: <Lock className="w-5 h-5" />, title: 'Zero-Egress', sub: 'Data never leaves your VPN' },
                { icon: <ShieldCheck className="w-5 h-5" />, title: 'Fail-Closed', sub: 'Blocks unsafe requests' },
                { icon: <Terminal className="w-5 h-5" />, title: 'Local Analysis', sub: 'No external scanning APIs' },
                { icon: <Database className="w-5 h-5" />, title: 'Deterministic', sub: 'Vaulted token consistency' },
            ].map(item => (
                <div key={item.title} className="flex flex-col items-center md:items-start gap-2 text-center md:text-left">
                    <div className="text-cyan-600">{item.icon}</div>
                    <div className="text-[11px] font-bold text-gray-900 uppercase tracking-wider">{item.title}</div>
                    <div className="text-[10px] text-gray-500 uppercase">{item.sub}</div>
                </div>
            ))}
        </div>
    </div>
);

// ── 2. Interactive Demo ───────────────────────────────────────────────────────
const DemoSection = () => {
    const [phase, setPhase] = useState(0);

    const phases = [
        {
            label: 'DETECT',
            desc: 'PII identified in the request payload',
            text: 'Hello, my name is John Doe, card 4111-1111-1111-1111',
            dotColor: 'bg-rose-500',
            textColor: 'text-rose-300',
            notice: null,
        },
        {
            label: 'VAULT',
            desc: 'PII encrypted locally. Tokens replace plaintext.',
            text: 'Hello, my name is [PERSON_a1b2c3], card [CREDIT_CARD_f4e3d2]',
            dotColor: 'bg-cyan-500',
            textColor: 'text-cyan-300',
            notice: {
                icon: <Lock className="w-4 h-4 text-cyan-500 shrink-0" />,
                color: 'bg-cyan-500/5 border-cyan-500/20',
                textColor: 'text-cyan-400',
                msg: 'AES-256-GCM encrypted · stored in local vault · master key never leaves process memory',
            },
        },
        {
            label: 'RESTORE',
            desc: 'Tokens re-hydrated for the authorized caller on the response path',
            text: 'Hello, my name is John Doe, card 4111-1111-1111-1111',
            dotColor: 'bg-emerald-500',
            textColor: 'text-emerald-300',
            notice: {
                icon: <CheckCircle2 className="w-4 h-4 text-emerald-500 shrink-0" />,
                color: 'bg-emerald-500/5 border-emerald-500/20',
                textColor: 'text-emerald-400',
                msg: 'Vault lookup · HKDF key derivation · plaintext restored for authorized caller only',
            },
        },
    ];

    const current = phases[phase];

    return (
        <section className="section-padding bg-gray-50 border-t border-gray-100">
            <div className="max-container">
                <div className="max-w-xl mb-16">
                    <h4 className="text-cyan-600 text-[10px] uppercase font-mono tracking-[0.4em] mb-4">Live Demo</h4>
                    <h2 className="mb-6">
                        See <span className="text-cyan-600">OCULTAR</span> intercept a request
                    </h2>
                    <p className="text-gray-600">
                        Every request passes through three stages before reaching the AI provider.
                        Nothing sensitive crosses the boundary.
                    </p>
                </div>

                {/* Phase tabs */}
                <div className="flex flex-col sm:flex-row gap-3 mb-10 max-w-lg">
                    {phases.map((p, i) => (
                        <button
                            key={p.label}
                            onClick={() => setPhase(i)}
                            className={`flex-1 px-4 py-3 text-[11px] font-bold uppercase tracking-widest border rounded-lg transition-all flex items-center justify-center gap-2 ${
                                phase === i
                                    ? 'bg-cyan-600 text-white border-cyan-600'
                                    : 'text-gray-500 border-gray-200 hover:border-gray-400 hover:text-gray-900'
                            }`}
                        >
                            <span className={`w-1.5 h-1.5 rounded-full ${phase === i ? 'bg-black' : p.dotColor}`} />
                            {i + 1}. {p.label}
                        </button>
                    ))}
                </div>

                {/* Terminal */}
                <div className="max-w-4xl bg-zinc-950 border border-white/10 rounded-2xl overflow-hidden shadow-2xl shadow-black">
                    <div className="flex items-center gap-3 px-6 py-4 border-b border-white/5 bg-zinc-900/50">
                        <div className="flex gap-1.5">
                            <div className="w-2.5 h-2.5 rounded-full bg-rose-500/40" />
                            <div className="w-2.5 h-2.5 rounded-full bg-amber-500/40" />
                            <div className="w-2.5 h-2.5 rounded-full bg-emerald-500/40" />
                        </div>
                        <div className="ml-2 flex items-center gap-3">
                            <div className={`w-1.5 h-1.5 rounded-full ${current.dotColor} animate-pulse`} />
                            <span className={`text-[10px] font-bold uppercase tracking-[0.3em] ${current.textColor}`}>
                                {current.label}
                            </span>
                            <span className="text-zinc-600 text-[10px]">—</span>
                            <span className="text-[10px] text-zinc-500">{current.desc}</span>
                        </div>
                    </div>

                    <div className="p-8 space-y-6">
                        <div className="space-y-2">
                            <div className="text-[10px] font-mono text-zinc-600 uppercase tracking-widest">
                                → POST /v1/chat/completions
                            </div>
                            <div className="bg-black/40 rounded-xl p-6 border border-white/5 min-h-[80px] flex items-center">
                                <pre className={`text-sm font-mono leading-relaxed transition-colors duration-300 ${current.textColor} whitespace-pre-wrap`}>
                                    {current.text}
                                </pre>
                            </div>
                        </div>
                        {current.notice && (
                            <div className={`p-4 border rounded-lg flex items-start gap-3 ${current.notice.color}`}>
                                {current.notice.icon}
                                <span className={`text-[11px] font-mono ${current.notice.textColor}`}>
                                    {current.notice.msg}
                                </span>
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </section>
    );
};

// ── 3. How It Works (5 steps horizontal) ─────────────────────────────────────
const HowItWorks = () => {
    const steps = [
        { num: '01', title: 'Detect', desc: 'OCULTAR scans every request in real time', icon: <Eye className="w-5 h-5" /> },
        { num: '02', title: 'Vault',  desc: 'PII is encrypted and stored locally, never transmitted', icon: <Lock className="w-5 h-5" /> },
        { num: '03', title: 'Redact', desc: 'Tokens replace sensitive data before it reaches the AI', icon: <Shield className="w-5 h-5" /> },
        { num: '04', title: 'Route',  desc: 'The AI receives clean, safe, useful context', icon: <ArrowRight className="w-5 h-5" /> },
        { num: '05', title: 'Restore',desc: 'The response is re-hydrated before your user sees it', icon: <RefreshCw className="w-5 h-5" /> },
    ];

    return (
        <section className="section-padding">
            <div className="max-container">
                <div className="max-w-xl mb-20">
                    <h4 className="text-cyan-600 text-[10px] uppercase font-mono tracking-[0.4em] mb-4">Architecture</h4>
                    <h2>
                        How <span className="text-cyan-600">OCULTAR</span> works
                    </h2>
                </div>

                <div className="relative">
                    {/* Connector line — desktop only */}
                    <div className="hidden md:block absolute top-10 left-[10%] right-[10%] h-px bg-gradient-to-r from-transparent via-gray-200 to-transparent" />

                    <div className="grid grid-cols-1 md:grid-cols-5 gap-10 md:gap-4">
                        {steps.map((step, i) => (
                            <div key={step.num} className="flex flex-col items-start md:items-center text-left md:text-center gap-4 relative">
                                <div className="w-20 h-20 bg-white border border-gray-200 rounded-2xl flex items-center justify-center text-cyan-600 relative z-10 hover:border-cyan-500/40 transition-colors shadow-sm">
                                    {step.icon}
                                    <div className="absolute -top-2 -right-2 w-5 h-5 bg-gray-100 border border-gray-200 rounded-full flex items-center justify-center">
                                        <span className="text-[8px] font-mono font-bold text-gray-400">{step.num}</span>
                                    </div>
                                </div>
                                <div>
                                    <div className="text-sm font-bold text-gray-900 mb-1">{step.title}</div>
                                    <p className="text-xs text-gray-500 leading-relaxed">{step.desc}</p>
                                </div>
                                {i < steps.length - 1 && (
                                    <ArrowRight className="hidden md:block absolute top-10 -right-5 w-4 h-4 text-gray-300 z-20 -translate-y-1/2" />
                                )}
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </section>
    );
};

// ── 4. Who It's For ───────────────────────────────────────────────────────────
const WhoItsFor = () => {
    const personas = [
        {
            role: 'The CISO',
            icon: <ShieldCheck className="w-6 h-6" />,
            body: 'You need AI adoption without compliance exposure. OCULTAR makes GDPR compliance architectural, not procedural.',
            tag: 'GDPR · HIPAA · CNIL',
        },
        {
            role: 'The CTO',
            icon: <Server className="w-6 h-6" />,
            body: 'You need to ship AI features without a 6-month security review. OCULTAR deploys in your infrastructure in under an hour.',
            tag: 'On-Premise · Docker · <1h Deploy',
        },
        {
            role: 'The Engineering Lead',
            icon: <Terminal className="w-6 h-6" />,
            body: 'You need a proxy that works with any upstream AI — OpenAI, Gemini, Mistral, whatever comes next.',
            tag: 'Any LLM Provider · Drop-in Proxy',
        },
    ];

    return (
        <section className="section-padding bg-gray-50 border-t border-gray-100">
            <div className="max-container">
                <div className="max-w-xl mb-16">
                    <h4 className="text-cyan-600 text-[10px] uppercase font-mono tracking-[0.4em] mb-4">Who It's For</h4>
                    <h2>
                        Built for the people who <span className="text-cyan-600">own the risk</span>
                    </h2>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                    {personas.map(p => (
                        <div key={p.role} className="card group space-y-6 hover:border-cyan-500/40 transition-all">
                            <div className="flex items-center gap-4">
                                <div className="w-12 h-12 bg-cyan-50 border border-cyan-200 rounded-xl flex items-center justify-center text-cyan-600 group-hover:bg-cyan-600 group-hover:text-white transition-all">
                                    {p.icon}
                                </div>
                                <div className="text-sm font-bold text-gray-900">{p.role}</div>
                            </div>
                            <p className="text-sm text-gray-600 leading-relaxed">{p.body}</p>
                            <div className="text-[10px] font-mono font-bold text-cyan-600/60 uppercase tracking-widest border-t border-gray-100 pt-4">
                                {p.tag}
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </section>
    );
};

// ── 5. Why Not SaaS ───────────────────────────────────────────────────────────
const WhyNotSaaS = () => (
    <section className="section-padding border-y border-gray-100">
        <div className="max-container max-w-3xl">
            <h4 className="text-gray-500 text-[10px] uppercase font-mono tracking-[0.4em] mb-6">Zero-Egress</h4>
            <h2 className="mb-16">
                Why does it run in <span className="text-cyan-600">your infrastructure</span>?
            </h2>
            <ul className="space-y-10">
                {[
                    'SaaS DLP means trusting a third party with your most sensitive data.',
                    'A breach at the vendor level exposes your customers — not just you.',
                    'OCULTAR makes data leakage technically impossible, not just contractually forbidden.',
                ].map((point, i) => (
                    <li key={i} className="flex gap-6 items-start">
                        <div className="mt-2 w-1.5 h-1.5 rounded-full bg-cyan-600 shrink-0" />
                        <p className="text-lg text-gray-700 leading-relaxed">{point}</p>
                    </li>
                ))}
            </ul>
            <div className="mt-16 pt-10 border-t border-gray-100">
                <p className="text-xl font-bold text-gray-900 tracking-tight">
                    Zero-egress is not a feature. It is the architecture.
                </p>
            </div>
        </div>
    </section>
);

// ── 6. ROI Calculator ─────────────────────────────────────────────────────────
const ROISection = () => {
    const [volume, setVolume] = useState(5);
    const cloudCost = volume * 15000;
    const ocultarCost = 24900;
    const savings = Math.max(0, cloudCost - ocultarCost);

    return (
        <section className="section-padding bg-gray-50 border-b border-gray-100 relative overflow-hidden">
            <div className="max-container">
                <div className="max-w-xl mb-16">
                    <h4 className="text-cyan-600 text-[10px] uppercase font-mono tracking-[0.4em] mb-4">Cost Analysis</h4>
                    <h2 className="mb-4">What are you paying to send PII to the cloud?</h2>
                    <p className="text-gray-600">
                        Compare <span className="text-cyan-600">OCULTAR</span>'s annual license against what AWS
                        Comprehend or Google Cloud DLP costs at your volume.
                    </p>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-2 gap-20 items-center">
                    <div className="space-y-8">
                        <div className="space-y-6 bg-white p-8 rounded-2xl border border-gray-200 shadow-sm">
                            <div className="space-y-4">
                                <label className="text-[10px] font-bold text-gray-500 uppercase tracking-widest block">
                                    Monthly Token Volume (Millions)
                                </label>
                                <input
                                    type="range" min="1" max="100" step="1"
                                    value={volume} onChange={e => setVolume(parseInt(e.target.value))}
                                    className="w-full h-1.5 bg-gray-200 rounded-lg appearance-none cursor-pointer accent-cyan-600"
                                />
                                <div className="flex justify-between text-[11px] font-mono text-gray-500">
                                    <span>1M</span>
                                    <span className="text-gray-900 font-bold">{volume}M tokens/mo</span>
                                    <span>100M+</span>
                                </div>
                            </div>

                            <div className="space-y-3 pt-6 border-t border-gray-100">
                                <div className="flex justify-between items-center">
                                    <span className="text-[10px] font-bold text-gray-500 uppercase tracking-widest">Cloud DLP / yr</span>
                                    <span className="text-xl font-bold text-rose-500 font-mono">${cloudCost.toLocaleString()}</span>
                                </div>
                                <div className="flex justify-between items-center">
                                    <span className="text-[10px] font-bold text-gray-500 uppercase tracking-widest">OCULTAR Enterprise / yr</span>
                                    <span className="text-xl font-bold text-gray-900 font-mono">€{ocultarCost.toLocaleString()}</span>
                                </div>
                                <div className="flex justify-between items-center pt-3 border-t border-gray-100">
                                    <span className="text-[10px] font-bold text-cyan-600/60 uppercase tracking-widest">You save (approx.)</span>
                                    <span className="text-xl font-bold text-cyan-600 font-mono">~${savings.toLocaleString()}</span>
                                </div>
                                <p className="text-[10px] text-gray-400 font-mono pt-1">
                                    Fixed annual license · price does not scale with volume
                                </p>
                            </div>
                        </div>

                        <Link
                            to="/calculator"
                            className="inline-flex items-center gap-2 text-[11px] font-bold text-cyan-600 hover:text-cyan-700 uppercase tracking-widest transition-colors group"
                        >
                            Open full ROI calculator
                            <ChevronRight className="w-3 h-3 group-hover:translate-x-1 transition-transform" />
                        </Link>
                    </div>

                    <div className="relative">
                        <div className="absolute inset-0 bg-cyan-500/5 blur-3xl rounded-full scale-150" />
                        <div className="bg-white border border-gray-200 rounded-2xl p-8 relative shadow-lg space-y-6">
                            <Activity className="text-cyan-600 w-8 h-8" />
                            <h3>Zero-Egress SLA</h3>
                            <div className="space-y-3">
                                {[
                                    { label: 'Enforcement Rate', value: '100.00%' },
                                    { label: 'Latency Overhead', value: '< 5ms' },
                                    { label: 'Egress Blocks', value: 'Real-time' },
                                    { label: 'Vault Encryption', value: 'AES-256-GCM' },
                                ].map(m => (
                                    <div key={m.label} className="flex justify-between items-center py-2 border-b border-gray-100">
                                        <span className="text-[11px] text-gray-500 uppercase">{m.label}</span>
                                        <span className="text-xs font-mono font-bold text-cyan-600">{m.value}</span>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </section>
    );
};

// ── 7. GDPR Exposure Calculator ───────────────────────────────────────────────
const GDPRSection = () => (
    <section className="section-padding bg-white border-b border-gray-100">
        <div className="max-container">
            <div className="max-w-xl mb-16">
                <h4 className="text-cyan-600 text-[10px] uppercase font-mono tracking-[0.4em] mb-4">Regulatory Exposure</h4>
                <h2 className="mb-4">
                    Calculate your <span className="text-cyan-600">GDPR exposure</span> in 2 minutes
                </h2>
                <p className="text-gray-600">
                    Regulators calculate fines based on data volume and industry. So should you.
                </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-14">
                {[
                    {
                        icon: <Shield className="w-6 h-6" />,
                        title: 'Secure Analysis',
                        desc: 'Process a sample dataset locally within your perimeter. No data stored.',
                    },
                    {
                        icon: <BarChart3 className="w-6 h-6" />,
                        title: 'Instant Risk Score',
                        desc: 'PII exposure and K-Anonymity scores across 40+ entity types.',
                    },
                    {
                        icon: <AlertTriangle className="w-6 h-6" />,
                        title: 'Financial Exposure (VaR)',
                        desc: 'Defensible Value-at-Risk estimations for your legal team.',
                    },
                ].map(item => (
                    <div key={item.title} className="card space-y-4 hover:border-cyan-500/40 transition-colors">
                        <div className="w-12 h-12 bg-gray-100 border border-gray-200 rounded-xl flex items-center justify-center text-cyan-600">
                            {item.icon}
                        </div>
                        <h3 className="text-xl font-bold">{item.title}</h3>
                        <p className="text-sm text-gray-600 leading-relaxed">{item.desc}</p>
                    </div>
                ))}
            </div>

            <div className="flex flex-col sm:flex-row items-start sm:items-center gap-6">
                <Link to="/risk-assessment" className="btn btn-primary px-10 py-4 group inline-flex items-center w-full sm:w-auto justify-center">
                    Start Free Audit
                    <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform ml-2" />
                </Link>
                <span className="text-[10px] text-gray-400 font-mono uppercase tracking-widest">
                    Stateless · No data stored · Zero-Egress guaranteed
                </span>
            </div>
        </div>
    </section>
);

// ── 8. Pricing ────────────────────────────────────────────────────────────────
const Pricing = () => {
    const tiers = [
        {
            name: 'COMMUNITY',
            price: 'Free',
            sub: 'Apache 2.0 · Self-hosted',
            accent: false,
            cta: { label: 'View on GitHub', href: 'https://github.com/Edu963/ocultar', external: true },
            features: [
                'Reverse proxy with fail-closed enforcement',
                'Detection tiers 0.1 – 1.5 (regex, rules, phone, address, greeting)',
                'AES-256-GCM tokenization · DuckDB local vault',
                'Dry-run mode + PII risk report generation',
                'Prometheus metrics · Docker Compose startup',
                'Risk assessment engine (k-anonymity, VaR)',
            ],
        },
        {
            name: 'ENTERPRISE',
            price: '€24,900',
            sub: '/year · annual license',
            accent: true,
            badge: 'Full Stack',
            cta: { label: 'Book a Demo', href: DEMO_URL, external: false },
            features: [
                'Everything in Community',
                'AI NER (Tier 2) — local inference, zero data egress',
                'PostgreSQL vault — multi-server HA deployments',
                'Structured audit log — GDPR Article 5(2) compliant, fail-fatal',
                'Custom detection rules via config.yaml',
                'SharePoint + Slack connectors',
                'CRM / LDAP live identity sync → Tier 0 dictionary',
                'Syslog UDP + gRPC interceptor for SIEM pipelines',
            ],
        },
    ];

    return (
        <section className="section-padding border-t border-gray-100">
            <div className="max-container">
                <div className="max-w-xl mb-6">
                    <h4 className="text-cyan-600 text-[10px] uppercase font-mono tracking-[0.4em] mb-4">Pricing</h4>
                    <h2>
                        Simple, predictable <span className="text-cyan-600">licensing</span>
                    </h2>
                </div>

                {/* License clarity banner */}
                <div className="bg-cyan-50 border border-cyan-100 rounded-xl px-6 py-4 max-w-4xl mb-10 flex flex-col sm:flex-row sm:items-center gap-3">
                    <span className="text-[10px] font-bold uppercase tracking-widest text-cyan-700 shrink-0">One license. Full stack.</span>
                    <div className="flex flex-wrap gap-1.5">
                        {['Reverse Proxy', 'Sombra Gateway', 'Refinery (all tiers + AI NER)', 'AES-256 Vault', 'CLI Tools', 'Slack Connector', 'SharePoint Connector', 'SIEM Syslog Forwarder'].map(item => (
                            <span key={item} className="text-[10px] font-medium px-2.5 py-1 rounded-full bg-white border border-cyan-200 text-cyan-700">{item}</span>
                        ))}
                    </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-4xl mb-10">
                    {tiers.map(tier => (
                        <div
                            key={tier.name}
                            className={`card flex flex-col space-y-8 ${
                                tier.accent ? 'border-cyan-500/30 bg-cyan-50/50' : ''
                            }`}
                        >
                            {tier.badge && (
                                <div className="text-[10px] font-bold text-cyan-600 uppercase tracking-widest">
                                    {tier.badge}
                                </div>
                            )}
                            <div>
                                <div className="text-[10px] font-bold text-gray-500 uppercase tracking-widest mb-3">
                                    {tier.name}
                                </div>
                                <div className="flex items-end gap-2">
                                    <span className={`text-5xl font-black tracking-tighter ${tier.accent ? 'text-cyan-600' : 'text-gray-900'}`}>
                                        {tier.price}
                                    </span>
                                </div>
                                <div className="text-[10px] text-gray-400 font-mono uppercase tracking-widest mt-1">
                                    {tier.sub}
                                </div>
                            </div>

                            <ul className="space-y-3 flex-grow">
                                {tier.features.map((f, i) => (
                                    <li key={i} className="flex items-start gap-3 text-sm text-gray-600">
                                        <Check className="w-4 h-4 text-cyan-600 shrink-0 mt-0.5" />
                                        {f}
                                    </li>
                                ))}
                            </ul>

                            <a
                                href={tier.cta.href}
                                target={tier.cta.external ? '_blank' : undefined}
                                rel={tier.cta.external ? 'noopener noreferrer' : undefined}
                                className={`btn w-full py-4 text-center ${tier.accent ? 'btn-primary' : 'btn-secondary'}`}
                            >
                                {tier.cta.label}
                            </a>
                        </div>
                    ))}
                </div>

                <p className="text-[11px] text-gray-400 font-mono">
                    Deployed entirely in your infrastructure. We never have access to your data — technically
                    impossible. Zero-egress by design.
                </p>
            </div>
        </section>
    );
};

// ── 9. Final CTA ──────────────────────────────────────────────────────────────
const FinalCTA = () => (
    <section className="section-padding bg-gray-50 border-t border-gray-100">
        <div className="max-container text-center relative z-10 space-y-8">
            <h2 className="max-w-3xl mx-auto">
                See it running in your environment in 30 minutes.
            </h2>
            <p className="text-gray-600 max-w-lg mx-auto">
                No shared demo environment. No data sent anywhere. Just a call and a local deployment.
            </p>
            <a
                href={DEMO_URL}
                className="btn btn-primary px-16 py-6 text-xl inline-flex items-center group"
            >
                Book a Demo
                <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform ml-2" />
            </a>
        </div>
    </section>
);

// ── Landing Page ──────────────────────────────────────────────────────────────
function LandingPage() {
    return (
        <div className="animate-fade-in-up">
            <Hero />
            <TrustStrip />
            <DemoSection />
            <HowItWorks />
            <WhoItsFor />
            <WhyNotSaaS />
            <ROISection />

            <Pricing />
            <FinalCTA />
        </div>
    );
}

// ── Router Shell ──────────────────────────────────────────────────────────────
export default function App() {
    return (
        <Router>
            <AppContent />
        </Router>
    );
}

function AppContent() {
    const location = useLocation();
    const isDocs = location.pathname.startsWith('/docs');

    return (
        <div className="min-h-screen flex flex-col selection:bg-cyan-500 selection:text-black">
            <CanvasBackground />
            {isDocs ? <DocsHeader /> : <Nav />}
            <div className={`flex-grow ${isDocs ? '' : 'pt-32'}`}>
                <Routes>
                    <Route path="/" element={<LandingPage />} />
                    <Route path="/risk-assessment" element={<RiskAssessmentPage />} />
                    <Route path="/calculator" element={<CalculatorPage />} />
                    <Route path="/solutions" element={<SolutionsPage />} />
                    <Route path="/docs" element={<DocsLayout />}>
                        <Route index element={<DocsLanding />} />
                        <Route path="reference/:type" element={<ReferencePage />} />
                        <Route path="*" element={<DocPage />} />
                    </Route>
                </Routes>
            </div>
            <Footer />
        </div>
    );
}
