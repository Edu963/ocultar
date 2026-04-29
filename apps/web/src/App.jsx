import { useState, useEffect, useRef, useCallback } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import {
    Shield, Lock, BarChart3, ChevronRight, Github, Globe,
    ArrowRight, AlertTriangle, Terminal, Check, Server, Database, Activity,
    ShieldCheck, Eye, RefreshCw, CheckCircle2, Copy, CheckCheck
} from 'lucide-react';
import RiskAssessmentPage from './RiskAssessmentPage';
import CalculatorPage from './CalculatorPage';
import SolutionsPage from './SolutionsPage';
import PrivacyPage from './PrivacyPage';
import DocsLayout from './components/DocsLayout';
import DocPage from './components/DocPage';
import ReferencePage from './components/ReferencePage';
import DocsLanding from './components/DocsLanding';
import DocsHeader from './components/DocsHeader';
import { useLocation } from 'react-router-dom';

const REQUEST_ACCESS_URL =
    'mailto:sales@ocultar.dev?subject=Access%20Request&body=Hi%2C%0A%0AWork%20email%3A%20%0ACompany%20name%3A%20%0APrimary%20use%20case%20(Healthcare%20%2F%20Finance%20%2F%20Legal%20%2F%20Government%20%2F%20Other)%3A%20';

// ── Particle Canvas ────────────────────────────────────────────────────────────
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
                ctx.fillStyle = 'rgba(249, 115, 22, 0.5)';
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
                        const opacity = (1 - distance / connectionDistance) * 0.06;
                        ctx.strokeStyle = `rgba(249, 115, 22, ${opacity})`;
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

// ── Nav ────────────────────────────────────────────────────────────────────────
const Nav = () => {
    const [scrolled, setScrolled] = useState(false);
    useEffect(() => {
        const onScroll = () => setScrolled(window.scrollY > 12);
        window.addEventListener('scroll', onScroll, { passive: true });
        return () => window.removeEventListener('scroll', onScroll);
    }, []);
    return (
        <nav
            className={`fixed top-0 left-0 w-full z-50 py-5 transition-all duration-300 ${
                scrolled
                    ? 'backdrop-blur-xl bg-zinc-950/95 border-b border-zinc-800 shadow-lg shadow-black/30'
                    : 'bg-transparent border-b border-transparent'
            }`}
        >
            <div className="max-container flex justify-between items-center">
                <Link to="/" className="flex items-center gap-3">
                    <span className="font-mono font-black text-xl tracking-widest text-white">OCULTAR</span>
                </Link>
                <div className="hidden md:flex items-center gap-10 text-[12px] font-bold uppercase tracking-widest">
                    <Link to="/" className="text-zinc-400 hover:text-white transition-colors duration-200">Platform</Link>
                    <Link to="/solutions" className="text-zinc-400 hover:text-white transition-colors duration-200">Product Suite</Link>
                    <Link to="/calculator" className="text-zinc-400 hover:text-white transition-colors duration-200">ROI Calculator</Link>
                    <a href="https://github.com/Edu963/ocultar" className="text-zinc-400 hover:text-white transition-colors duration-200">GitHub</a>
                </div>
                <a
                    href={REQUEST_ACCESS_URL}
                    className="bg-orange-500 text-white text-[11px] font-bold px-6 py-3 rounded uppercase tracking-wider hover:bg-orange-600 transition-colors"
                >
                    Request Access
                </a>
            </div>
        </nav>
    );
};

// ── Footer ────────────────────────────────────────────────────────────────────
const Footer = () => (
    <footer className="bg-[#0A0A0F] border-t border-zinc-800 py-20">
        <div className="max-container">
            <div className="grid grid-cols-1 md:grid-cols-4 gap-12">
                <div className="col-span-1 md:col-span-2 space-y-6">
                    <span className="font-mono font-black text-xl tracking-widest text-white">OCULTAR</span>
                    <p className="max-w-sm text-sm text-zinc-500">
                        Zero-egress PII refinery for enterprise AI. PII never leaves your infrastructure — technically impossible, not contractually forbidden.
                    </p>
                    <div className="flex gap-4">
                        <a href="https://github.com/Edu963/ocultar" className="text-zinc-600 hover:text-orange-500 transition-colors">
                            <Github className="w-5 h-5" />
                        </a>
                    </div>
                </div>
                <div className="space-y-4">
                    <h4 className="text-[10px] font-bold uppercase tracking-widest text-orange-500">Product</h4>
                    <ul className="space-y-2 text-sm text-zinc-500">
                        <li><Link to="/solutions" className="hover:text-zinc-300 transition-colors">Product Suite</Link></li>
                        <li><Link to="/calculator" className="hover:text-zinc-300 transition-colors">ROI Calculator</Link></li>
                        <li><a href="https://github.com/Edu963/ocultar" className="hover:text-zinc-300 transition-colors">GitHub (Apache 2.0)</a></li>
                    </ul>
                </div>
                <div className="space-y-4">
                    <h4 className="text-[10px] font-bold uppercase tracking-widest text-orange-500">Contact</h4>
                    <ul className="space-y-2 text-sm text-zinc-500">
                        <li><a href={REQUEST_ACCESS_URL} className="hover:text-zinc-300 transition-colors">Request Access</a></li>
                        <li><a href="mailto:engineering@ocultar.dev" className="hover:text-zinc-300 transition-colors">Engineering</a></li>
                    </ul>
                </div>
            </div>
            <div className="mt-20 pt-8 border-t border-zinc-800 text-[10px] text-zinc-700 font-mono flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 uppercase tracking-widest">
                <div>&copy; {new Date().getFullYear()} OCULTAR SECURITY</div>
                <div className="flex gap-6 items-center">
                    <Link to="/privacy" className="hover:text-zinc-500 transition-colors">Privacy Policy</Link>
                    <span className="text-orange-500/40">ZERO_EGRESS_GUARANTEED</span>
                </div>
            </div>
        </div>
    </footer>
);

// ── 1. Hero ───────────────────────────────────────────────────────────────────
const Hero = () => (
    <section className="bg-[#0A0A0F] section-padding flex flex-col items-center relative overflow-hidden">
        <div className="absolute inset-0 bg-[linear-gradient(rgba(249,115,22,0.025)_1px,transparent_1px),linear-gradient(90deg,rgba(249,115,22,0.025)_1px,transparent_1px)] bg-[size:64px_64px]" />
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_80%_50%_at_50%_-10%,rgba(249,115,22,0.06),transparent)]" />
        <div className="max-container text-center relative z-10">
            <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full border border-orange-500/20 bg-orange-500/5 mb-10">
                <span className="relative flex h-2 w-2">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-orange-400 opacity-75"></span>
                    <span className="relative inline-flex rounded-full h-2 w-2 bg-orange-500"></span>
                </span>
                <span className="text-orange-400 text-xs font-mono font-bold uppercase tracking-widest">
                    Zero-Egress · On-Premise · Enterprise
                </span>
            </div>

            <h1 className="max-w-5xl mx-auto mb-8 text-white text-center">
                Sending customer data to OpenAI is a GDPR violation by default.{' '}
                <span className="text-orange-500">Most teams haven't noticed.</span>
            </h1>

            <div className="max-w-5xl mx-auto text-left">
                <p className="text-xl text-zinc-400 mb-10">
                    Every time your team sends a message to OpenAI, patient records, credit card numbers,
                    and customer names leave your network. Permanently.
                </p>
                <p className="text-center text-5xl md:text-6xl font-black text-orange-500 tracking-tight mt-2">
                    OCULTAR stops that.
                </p>
            </div>
        </div>
    </section>
);

// ── 2. Security Badges ────────────────────────────────────────────────────────
const SecurityBadges = () => (
    <div className="bg-[#0A0A0F] border-t border-zinc-800/60 py-8">
        <div className="max-container">
            <div className="flex flex-wrap justify-center gap-3">
                {[
                    'AES-256-GCM Encrypted',
                    'HKDF-SHA256 Key Derivation',
                    'Zero-Egress Guaranteed',
                    'SIEM-Ready Audit Logs',
                    'Fail-Closed by Design',
                    'Ed25519-Signed Chain',
                    'SOC2 Ready',
                ].map(badge => (
                    <span
                        key={badge}
                        className="px-4 py-1.5 rounded-full border border-zinc-800 bg-zinc-900/60 text-zinc-400 text-[10px] font-mono font-bold uppercase tracking-wider"
                    >
                        {badge}
                    </span>
                ))}
            </div>
        </div>
    </div>
);

// ── 3. Pain Cards ─────────────────────────────────────────────────────────────
const PainCards = () => {
    const cards = [
        {
            label: 'Exposed',
            body: "Every API call to OpenAI, Gemini, or Claude carries your customers' real names, SSNs, and health records. One breach at your AI provider becomes your breach.",
        },
        {
            label: 'Non-Compliant',
            body: "GDPR Article 28, HIPAA's Business Associate rules, and SOC2 CC6.1 all prohibit sending PII to third parties without explicit controls. Most companies are already in violation.",
        },
        {
            label: 'Blocked',
            body: 'Security reviews kill AI projects for 6–18 months. Engineers build workarounds. Compliance teams say no. The productivity win never arrives.',
        },
        {
            label: 'Invisible',
            body: "No audit trail. No way to prove what data touched which model. No evidence for regulators. If something goes wrong, you can't prove it didn't.",
        },
    ];

    return (
        <section className="bg-zinc-950 border-t border-zinc-800 section-padding">
            <div className="max-container">
                <div className="text-center flex flex-col items-center mb-16">
                    <p className="text-zinc-500 text-[10px] font-mono font-bold uppercase tracking-[0.4em] mb-6">The Problem</p>
                    <h2
                        style={{ textAlign: 'center', maxWidth: '42rem', margin: '0 auto 4rem auto', fontSize: '2rem' }}
                        className="text-white"
                    >
                        Why most AI deployments are a compliance time bomb
                    </h2>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                    {cards.map(card => (
                        <div
                            key={card.label}
                            className="bg-zinc-900 rounded-xl p-6 space-y-4 border-x border-b border-zinc-800 border-t-2 border-t-orange-500"
                        >
                            <div className="text-sm font-black text-white uppercase tracking-wider">{card.label}</div>
                            <p className="text-sm text-zinc-400 leading-relaxed">{card.body}</p>
                        </div>
                    ))}
                </div>
            </div>
        </section>
    );
};

// ── 4. Trust Strip ────────────────────────────────────────────────────────────
const TrustStrip = () => (
    <div className="bg-zinc-950 border-t border-zinc-800 py-20">
        <div className="max-container">
            <p className="text-[10px] font-mono font-bold uppercase tracking-[0.4em] text-zinc-500 mb-12 text-center">
                Core Security Architecture
            </p>
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-6">
                {[
                    {
                        icon: <Lock className="w-6 h-6" />,
                        title: 'Zero-Egress',
                        sub: 'Data never leaves your VPC. Technically impossible, not contractually forbidden.',
                    },
                    {
                        icon: <ShieldCheck className="w-6 h-6" />,
                        title: 'Fail-Closed',
                        sub: 'Blocks on any error — never falls back to passthrough.',
                    },
                    {
                        icon: <Terminal className="w-6 h-6" />,
                        title: 'Local NER',
                        sub: 'All entity recognition runs on-premise. No external scanning APIs.',
                    },
                    {
                        icon: <Database className="w-6 h-6" />,
                        title: 'Deterministic',
                        sub: 'Same input always produces the same token. Vault lookups are consistent.',
                    },
                ].map(item => (
                    <div
                        key={item.title}
                        className="group flex flex-col gap-4 p-6 rounded-xl border border-zinc-800 bg-zinc-900/40 hover:border-zinc-700 hover:bg-zinc-900 transition-all duration-200"
                    >
                        <div className="text-orange-500 transition-transform duration-200 group-hover:scale-110 w-fit">
                            {item.icon}
                        </div>
                        <div>
                            <div className="text-sm font-bold text-white uppercase tracking-wider mb-2">{item.title}</div>
                            <div className="text-sm text-zinc-400 leading-relaxed">{item.sub}</div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    </div>
);

// ── 5. Sovereign PII Packs ────────────────────────────────────────────────────
const SovereignPacks = () => (
    <div className="bg-[#0A0A0F] border-t border-zinc-800 py-20">
        <div className="max-container">
            <div className="flex flex-col md:flex-row md:items-start gap-12">
                <div className="shrink-0 md:max-w-sm">
                    <p className="text-[10px] font-mono font-bold uppercase tracking-[0.4em] text-orange-500 mb-4">
                        Detection Coverage
                    </p>
                    <h3 className="text-2xl font-bold text-white mb-4">Global &amp; Customizable Detection</h3>
                    <p className="text-base text-zinc-400 leading-relaxed">
                        Fully customizable via dictionaries and regex rules — extend coverage to any entity type, language, or jurisdiction.
                    </p>
                </div>
                <div className="flex flex-wrap gap-2 content-start md:pt-10">
                    {[
                        'FR_NIR · French SSN',
                        'ES_DNI · Spanish ID',
                        'EU_VAT · Tax ID',
                        'IBAN · Bank Account',
                        'FR_PHONE · Validated',
                        'SSN · US Social Security',
                        '40+ entity types',
                    ].map(badge => (
                        <span
                            key={badge}
                            className="px-3 py-1.5 rounded-lg border border-zinc-700 bg-zinc-900 text-zinc-300 text-xs font-mono hover:border-zinc-600 hover:text-zinc-200 transition-colors"
                        >
                            {badge}
                        </span>
                    ))}
                </div>
            </div>
        </div>
    </div>
);

// ── 6. Code Integration ───────────────────────────────────────────────────────
const CopyButton = ({ text }) => {
    const [copied, setCopied] = useState(false);
    const copy = () => {
        navigator.clipboard?.writeText(text).catch(() => {});
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };
    return (
        <button
            onClick={copy}
            className="flex items-center gap-1.5 text-zinc-500 hover:text-zinc-300 text-[10px] font-mono transition-colors"
        >
            {copied
                ? <><CheckCheck className="w-3 h-3 text-orange-400" /><span className="text-orange-400">Copied</span></>
                : <><Copy className="w-3 h-3" />Copy</>
            }
        </button>
    );
};

const AFTER_CODE = `const response = await openai.chat.completions.create({
  baseURL: "https://your-ocultar-instance/proxy/openai",
  model: "gpt-4o",
  messages: [{ role: "user", content: userMessage }]
});`;

const CodeIntegration = () => (
    <section className="section-padding bg-zinc-950 border-t border-zinc-800">
        <div className="max-container">
            <div className="max-w-2xl mb-20 text-center mx-auto flex flex-col items-center">
                <h4 className="text-orange-500 text-[10px] uppercase font-mono tracking-[0.4em] mb-4">Implementation</h4>
                <h2 className="text-white mb-6">Drop-in replacement for the OpenAI SDK</h2>
                <p className="text-zinc-400 max-w-xl">
                    Change one line of code. Redirect your SDK to your OCULTAR instance. 
                    Local redaction and vaulting happen automatically.
                </p>
            </div>
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
                <div className="space-y-8">
                    <div>
                        <h3 className="text-2xl font-bold text-white mb-4">Ship in an afternoon.<br />Not a quarter.</h3>
                        <p className="text-zinc-400 leading-relaxed mb-6">
                            OCULTAR runs as a transparent reverse proxy. Point your existing AI calls at it.
                            No SDK. No code changes. No retraining your team.
                        </p>
                        <div className="space-y-4">
                            {[
                                'Zero code changes to application logic',
                                'Supports all OpenAI-compatible SDKs',
                                'Deploy as a sidecar or central gateway'
                            ].map(item => (
                                <div key={item} className="flex items-center gap-3">
                                    <div className="w-5 h-5 rounded-full bg-orange-500/10 border border-orange-500/20 flex items-center justify-center">
                                        <Check className="w-3 h-3 text-orange-500" />
                                    </div>
                                    <span className="text-sm text-zinc-300">{item}</span>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>

                <div className="flex flex-col gap-8">
                    {/* Before */}
                    <div>
                        <p className="text-[9px] font-mono font-bold uppercase tracking-[0.3em] text-zinc-600 mb-3 pl-1">Before</p>
                        <div className="opacity-60 rounded-xl overflow-hidden border border-zinc-700/50 shadow-lg shadow-black/40">
                            <div className="flex items-center gap-3 px-4 py-2.5 border-b border-zinc-700/50 bg-zinc-900">
                                <div className="flex gap-1.5">
                                    <div className="w-2.5 h-2.5 rounded-full bg-rose-500/40" />
                                    <div className="w-2.5 h-2.5 rounded-full bg-amber-500/40" />
                                    <div className="w-2.5 h-2.5 rounded-full bg-emerald-500/40" />
                                </div>
                                <span className="text-[10px] font-mono text-rose-400/80 uppercase tracking-widest ml-2">
                                    Before — PII reaches OpenAI
                                </span>
                            </div>
                            <pre className="p-5 text-xs font-mono text-zinc-500 leading-relaxed overflow-x-auto bg-zinc-900/60">{`const response = await openai.chat.completions.create({
  model: "gpt-4o",
  messages: [{ role: "user", content: userMessage }]
});`}</pre>
                        </div>
                    </div>

                    {/* After */}
                    <div>
                        <p className="text-[9px] font-mono font-bold uppercase tracking-[0.3em] text-orange-500/70 mb-3 pl-1">After — Zero-Egress</p>
                        <div className="rounded-xl overflow-hidden border border-orange-500/50 shadow-xl shadow-orange-500/10">
                            <div className="flex items-center justify-between px-4 py-2.5 border-b border-zinc-800 bg-zinc-900">
                                <div className="flex items-center gap-3">
                                    <div className="flex gap-1.5">
                                        <div className="w-2.5 h-2.5 rounded-full bg-rose-500/50" />
                                        <div className="w-2.5 h-2.5 rounded-full bg-amber-500/50" />
                                        <div className="w-2.5 h-2.5 rounded-full bg-emerald-500/50" />
                                    </div>
                                    <span className="text-[10px] font-mono text-orange-400 uppercase tracking-widest ml-2">
                                        After — Zero-egress guaranteed
                                    </span>
                                </div>
                                <CopyButton text={AFTER_CODE} />
                            </div>
                            <pre className="p-5 text-xs font-mono leading-relaxed overflow-x-auto bg-zinc-950">
                                <span className="text-zinc-400">{'const response = await openai.chat.completions.create({\n'}</span>
                                <span className="text-orange-400">{'  baseURL: "https://your-ocultar-instance/proxy/openai",  // ← only change\n'}</span>
                                <span className="text-zinc-400">{'  model: "gpt-4o",\n'}</span>
                                <span className="text-zinc-400">{'  messages: [{ role: "user", content: userMessage }]\n'}</span>
                                <span className="text-zinc-400">{'});'}</span>
                            </pre>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </section>
);

// ── 7. Demo Section ───────────────────────────────────────────────────────────
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
            dotColor: 'bg-orange-500',
            textColor: 'text-orange-300',
            notice: {
                icon: <Lock className="w-4 h-4 text-orange-500 shrink-0" />,
                color: 'bg-orange-500/5 border-orange-500/20',
                textColor: 'text-orange-400',
                msg: 'AES-256-GCM encrypted · stored in local vault · master key never leaves process memory',
            },
        },
        {
            label: 'RESTORE',
            desc: 'Tokens re-hydrated for the authorized caller',
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
        <section className="section-padding bg-[#0A0A0F] border-t border-zinc-800">
            <div className="max-container flex flex-col items-center">
                <div className="max-w-xl mb-16 text-center mx-auto flex flex-col items-center">
                    <h4 className="text-orange-500 text-[10px] uppercase font-mono tracking-[0.4em] mb-4">Live Demo</h4>
                    <h2 className="text-white mb-6">See OCULTAR intercept a request</h2>
                    <p className="text-zinc-400">
                        Every request passes through three stages before reaching the AI provider.
                        Nothing sensitive crosses the boundary.
                    </p>
                </div>

                <div className="flex flex-col sm:flex-row gap-3 mb-10 max-w-lg">
                    {phases.map((p, i) => (
                        <button
                            key={p.label}
                            onClick={() => setPhase(i)}
                            className={`flex-1 px-4 py-3 text-[11px] font-bold uppercase tracking-widest border rounded-lg transition-all flex items-center justify-center gap-2 ${
                                phase === i
                                    ? 'bg-orange-500 text-white border-orange-500'
                                    : 'text-zinc-500 border-zinc-700 hover:border-zinc-500 hover:text-zinc-300'
                            }`}
                        >
                            <span className={`w-1.5 h-1.5 rounded-full ${phase === i ? 'bg-white' : p.dotColor}`} />
                            {i + 1}. {p.label}
                        </button>
                    ))}
                </div>

                <div className="max-w-4xl bg-zinc-950 border border-zinc-800 rounded-2xl overflow-hidden shadow-2xl shadow-black">
                    <div className="flex items-center gap-3 px-6 py-4 border-b border-zinc-800 bg-zinc-900/50">
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
                            <div className="bg-black/40 rounded-xl p-6 border border-zinc-800 min-h-[80px] flex items-center">
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

// ── 8. Audit Log ──────────────────────────────────────────────────────────────
const AUDIT_ENTRIES = [
    { time: '09:14:32', op: 'VAULTED',    type: 'PERSON',      input: '"John Doe"',              output: '[PERSON_a1b2c3d4]',  req: '8f3a' },
    { time: '09:14:32', op: 'VAULTED',    type: 'CREDIT_CARD', input: '"4111111111111111"',       output: '[CC_7d9e1f2a]',      req: '8f3a' },
    { time: '09:14:33', op: 'FORWARDED',  type: null,          input: 'api.openai.com',           output: 'status: 200',        req: '8f3a' },
    { time: '09:14:33', op: 'REHYDRATED', type: 'PERSON',      input: '[PERSON_a1b2c3d4]',        output: '"John Doe"',         req: '8f3a' },
];

const opColor = { VAULTED: 'text-orange-400', FORWARDED: 'text-zinc-400', REHYDRATED: 'text-emerald-400' };

const AuditLog = () => (
    <section className="section-padding bg-zinc-950 border-t border-zinc-800">
        <div className="max-container">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
                <div>
                    <p className="text-orange-500 text-[10px] font-mono font-bold uppercase tracking-[0.4em] mb-4">Audit Trail</p>
                    <h2 className="text-white mb-6">
                        Every token. Every request.<br />Every decision. Logged.
                    </h2>
                    <p className="text-zinc-400 mb-8">
                        Every vaulted and matched event is written to a tamper-proof, SIEM-ready audit log.
                        When a regulator asks what data touched which model, you have the answer in seconds.
                    </p>
                    <div className="space-y-3">
                        {[
                            'Ed25519-signed — each entry cryptographically linked to the last',
                            'SHA-256 hash-chained — deletion or modification is detectable',
                            'Syslog UDP forwarder — PII-scrubbed events to your SIEM',
                            'GDPR Art. 5(2) accountability compliance out of the box',
                        ].map(point => (
                            <div key={point} className="flex items-start gap-3">
                                <Check className="w-4 h-4 text-orange-500 shrink-0 mt-0.5" />
                                <span className="text-sm text-zinc-400">{point}</span>
                            </div>
                        ))}
                    </div>
                </div>

                <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden">
                    <div className="flex items-center gap-3 px-4 py-3 border-b border-zinc-800">
                        <div className="flex gap-1.5">
                            <div className="w-2.5 h-2.5 rounded-full bg-rose-500/40" />
                            <div className="w-2.5 h-2.5 rounded-full bg-amber-500/40" />
                            <div className="w-2.5 h-2.5 rounded-full bg-emerald-500/40" />
                        </div>
                        <span className="text-[10px] font-mono text-zinc-500 uppercase tracking-widest ml-2">
                            audit.log — real-time
                        </span>
                    </div>
                    <div className="p-4 space-y-0 font-mono text-[11px]">
                        {AUDIT_ENTRIES.map((entry, i) => (
                            <div key={i} className="flex flex-wrap gap-x-2 gap-y-0.5 py-2 border-b border-zinc-800/60 last:border-0">
                                <span className="text-zinc-600">[2026-04-29 {entry.time}]</span>
                                <span className={`font-bold min-w-[80px] ${opColor[entry.op]}`}>{entry.op}</span>
                                {entry.type && <span className="text-zinc-500 min-w-[100px]">{entry.type}</span>}
                                <span className="text-zinc-300">{entry.input}</span>
                                <span className="text-zinc-600">→</span>
                                <span className="text-zinc-300">{entry.output}</span>
                                <span className="text-zinc-700 ml-auto hidden sm:block">req: {entry.req}</span>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    </section>
);

// ── 9. Agent Block ────────────────────────────────────────────────────────────
const AgentBlock = () => (
    <section className="section-padding bg-zinc-950 border-t border-zinc-800">
        <div className="max-container">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
                <div className="bg-zinc-950 rounded-xl p-8 border border-zinc-800 space-y-6">
                    <div className="grid grid-cols-2 gap-3">
                        {[
                            { label: 'MCP Tool Call', detail: 'PII stripped before tool invocation' },
                            { label: 'A2A Handoff', detail: 'Tokens persist across agent boundaries' },
                            { label: 'Workflow Step', detail: 'Every intermediate state is vaulted' },
                            { label: 'API Response', detail: 'Re-hydrated only for authorized callers' },
                        ].map(item => (
                            <div key={item.label} className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                                <div className="text-[10px] font-bold text-orange-400 uppercase tracking-wider mb-1">{item.label}</div>
                                <div className="text-[10px] text-zinc-500 leading-relaxed">{item.detail}</div>
                            </div>
                        ))}
                    </div>
                    <div className="flex flex-wrap gap-2 pt-2 border-t border-zinc-800">
                        {['MCP Compatible', 'OpenAI Agents SDK', 'LangChain', 'CrewAI', 'Custom workflows'].map(tag => (
                            <span key={tag} className="px-3 py-1 rounded-full bg-zinc-800 border border-zinc-700 text-zinc-400 text-[10px] font-mono">
                                ✓ {tag}
                            </span>
                        ))}
                    </div>
                </div>

                <div>
                    <p className="text-orange-500 text-[10px] font-mono font-bold uppercase tracking-[0.4em] mb-4">Agentic AI</p>
                    <h2 className="text-white mb-6">Built for agents.<br />Not just chat.</h2>
                    <p className="text-zinc-400 mb-6">
                        Autonomous AI agents make dozens of tool calls per workflow — each one a potential PII leak.
                        OCULTAR's Sombra gateway wraps every agent interaction, every MCP tool call, and every
                        A2A handoff in the same zero-egress privacy guarantee.
                    </p>
                    <p className="text-zinc-500 text-sm leading-relaxed">
                        The same architectural guarantee that protects chat completions extends to
                        every tool invocation, every agent handoff, and every intermediate state
                        in a multi-agent workflow.
                    </p>
                </div>
            </div>
        </div>
    </section>
);

// ── 10. Who It's For ──────────────────────────────────────────────────────────
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
        <section className="section-padding bg-zinc-950 border-t border-zinc-800">
            <div className="max-container">
                <div className="max-w-xl mb-16">
                    <h4 className="text-orange-500 text-[10px] uppercase font-mono tracking-[0.4em] mb-4">Who It's For</h4>
                    <h2 className="text-white">Built for the people who <span className="text-orange-500">own the risk</span></h2>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                    {personas.map(p => (
                        <div key={p.role} className="card group space-y-6 hover:border-orange-500/30 transition-all">
                            <div className="flex items-center gap-4">
                                <div className="w-12 h-12 bg-orange-500/10 border border-orange-500/20 rounded-xl flex items-center justify-center text-orange-500 group-hover:bg-orange-500 group-hover:text-white transition-all">
                                    {p.icon}
                                </div>
                                <div className="text-sm font-bold text-white">{p.role}</div>
                            </div>
                            <p className="text-sm text-zinc-400 leading-relaxed">{p.body}</p>
                            <div className="text-[10px] font-mono font-bold text-orange-500/60 uppercase tracking-widest border-t border-zinc-800 pt-4">
                                {p.tag}
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </section>
    );
};

// ── 11. ROI Section ───────────────────────────────────────────────────────────
const ROISection = () => {
    const [volume, setVolume] = useState(5);
    const cloudCost = volume * 15000;
    const ocultarCost = 24900;
    const savings = Math.max(0, cloudCost - ocultarCost);

    return (
        <section className="section-padding bg-zinc-950 border-b border-zinc-800 relative overflow-hidden">
            <div className="max-container flex flex-col items-center">
                <div className="max-w-xl mb-16 text-center">
                    <h4 className="text-orange-500 text-[10px] uppercase font-mono tracking-[0.4em] mb-4">Cost Analysis</h4>
                    <h2 className="text-white mb-4">What are you paying to send PII to the cloud?</h2>
                    <p className="text-zinc-400">
                        Compare OCULTAR's fixed annual license against what AWS Comprehend or Google Cloud DLP
                        costs at your volume. The price doesn't scale with traffic.
                    </p>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-2 gap-20 items-center">
                    <div className="space-y-8">
                        <div className="space-y-6 bg-zinc-900 p-8 rounded-2xl border border-zinc-800">
                            <div className="space-y-4">
                                <label className="text-[10px] font-bold text-zinc-500 uppercase tracking-widest block">
                                    Monthly Token Volume (Millions)
                                </label>
                                <input
                                    type="range" min="1" max="100" step="1"
                                    value={volume} onChange={e => setVolume(parseInt(e.target.value))}
                                    className="w-full h-1.5 bg-zinc-700 rounded-lg appearance-none cursor-pointer accent-orange-500"
                                />
                                <div className="flex justify-between text-[11px] font-mono text-zinc-500">
                                    <span>1M</span>
                                    <span className="text-white font-bold">{volume}M tokens/mo</span>
                                    <span>100M+</span>
                                </div>
                            </div>

                            <div className="space-y-3 pt-6 border-t border-zinc-800">
                                <div className="flex justify-between items-center">
                                    <span className="text-[10px] font-bold text-zinc-500 uppercase tracking-widest">Cloud DLP / yr</span>
                                    <span className="text-xl font-bold text-rose-400 font-mono">${cloudCost.toLocaleString()}</span>
                                </div>
                                <div className="flex justify-between items-center">
                                    <span className="text-[10px] font-bold text-zinc-500 uppercase tracking-widest">OCULTAR Enterprise / yr</span>
                                    <span className="text-xl font-bold text-white font-mono">€{ocultarCost.toLocaleString()}</span>
                                </div>
                                <div className="flex justify-between items-center pt-3 border-t border-zinc-800">
                                    <span className="text-[10px] font-bold text-orange-500/60 uppercase tracking-widest">You save (approx.)</span>
                                    <span className="text-xl font-bold text-orange-500 font-mono">~${savings.toLocaleString()}</span>
                                </div>
                                <p className="text-[10px] text-zinc-600 font-mono pt-1">
                                    Fixed annual license · price does not scale with volume
                                </p>
                            </div>
                        </div>

                        <Link
                            to="/calculator"
                            className="inline-flex items-center gap-2 text-[11px] font-bold text-orange-500 hover:text-orange-600 uppercase tracking-widest transition-colors group"
                        >
                            Open full ROI calculator
                            <ChevronRight className="w-3 h-3 group-hover:translate-x-1 transition-transform" />
                        </Link>
                    </div>

                    <div className="relative">
                        <div className="absolute inset-0 bg-orange-500/5 blur-3xl rounded-full scale-150" />
                        <div className="bg-zinc-900 border border-zinc-800 rounded-2xl p-8 relative space-y-6">
                            <Activity className="text-orange-500 w-8 h-8" />
                            <h3 className="text-white">Zero-Egress SLA</h3>
                            <div className="space-y-3">
                                {[
                                    { label: 'Enforcement Rate', value: '100.00%' },
                                    { label: 'Latency Overhead', value: '< 5ms' },
                                    { label: 'Egress Blocks', value: 'Real-time' },
                                    { label: 'Vault Encryption', value: 'AES-256-GCM' },
                                ].map(m => (
                                    <div key={m.label} className="flex justify-between items-center py-2 border-b border-zinc-800">
                                        <span className="text-[11px] text-zinc-500 uppercase">{m.label}</span>
                                        <span className="text-xs font-mono font-bold text-orange-500">{m.value}</span>
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

// ── 12. Final CTA ─────────────────────────────────────────────────────────────
const FinalCTA = () => (
    <section className="section-padding bg-[#0A0A0F] border-t border-zinc-800 relative overflow-hidden">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_80%_60%_at_50%_100%,rgba(249,115,22,0.05),transparent)]" />
        <div className="max-container text-center relative z-10 space-y-8">
            <h2 className="max-w-3xl mx-auto text-white">
                The conversation is free.<br />The compliance risk isn't.
            </h2>
            <p className="text-zinc-400 max-w-lg mx-auto">
                No shared demo environment. No data sent anywhere. Just a call and a local
                deployment in your infrastructure — running in 30 minutes.
            </p>
            <a
                href={REQUEST_ACCESS_URL}
                className="inline-flex items-center gap-3 bg-orange-500 hover:bg-orange-600 text-white font-bold px-16 py-6 text-xl rounded transition-all duration-200 hover:-translate-y-0.5 hover:shadow-xl hover:shadow-orange-500/20 group"
            >
                Request Access
                <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
            </a>
            <p className="text-zinc-600 text-xs font-mono uppercase tracking-widest">
                Enterprise · Not self-serve · We'll reach out within 24h
            </p>
        </div>
    </section>
);

// ── Landing Page ──────────────────────────────────────────────────────────────
function LandingPage() {
    return (
        <div className="animate-fade-in-up">
            <Hero />
            <SecurityBadges />
            <PainCards />
            <TrustStrip />
            <SovereignPacks />
            <CodeIntegration />
            <DemoSection />
            <AuditLog />
            <AgentBlock />
            <WhoItsFor />
            <ROISection />
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
        <div className="min-h-screen flex flex-col selection:bg-orange-500 selection:text-white">
            <CanvasBackground />
            {isDocs ? <DocsHeader /> : <Nav />}
            <div className={`flex-grow ${isDocs ? '' : 'pt-32'}`}>
                <Routes>
                    <Route path="/" element={<LandingPage />} />
                    <Route path="/risk-assessment" element={<RiskAssessmentPage />} />
                    <Route path="/calculator" element={<CalculatorPage />} />
                    <Route path="/solutions" element={<SolutionsPage />} />
                    <Route path="/privacy" element={<PrivacyPage />} />
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
