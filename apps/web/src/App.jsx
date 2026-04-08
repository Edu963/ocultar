import { useState, useEffect, useRef } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import ROIDashboardCard from './components/ROIDashboardCard';
import RiskAssessmentPage from './RiskAssessmentPage';
import logo from './assets/images/image.webp';

// --- Particle Background System ---
const CanvasBackground = () => {
    const canvasRef = useRef(null);

    useEffect(() => {
        const canvas = canvasRef.current;
        const ctx = canvas.getContext('2d');
        let animationFrameId;
        let particles = [];
        const numParticles = 80;
        const connectionDistance = 120;

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
                this.vx = (Math.random() - 0.5) * 1.2;
                this.vy = (Math.random() - 0.5) * 1.2;
                this.radius = 1;
            }

            update() {
                if (this.x <= 0 || this.x >= canvas.width) this.vx *= -1;
                if (this.y <= 0 || this.y >= canvas.height) this.vy *= -1;
                this.x += this.vx;
                this.y += this.vy;
                this.x = Math.max(0, Math.min(canvas.width, this.x));
                this.y = Math.max(0, Math.min(canvas.height, this.y));
            }

            draw() {
                ctx.beginPath();
                ctx.arc(this.x, this.y, this.radius, 0, Math.PI * 2);
                ctx.fillStyle = '#000000';
                ctx.fill();
            }
        }

        for (let i = 0; i < numParticles; i++) {
            particles.push(new Particle());
        }

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
                        const opacity = (1 - distance / connectionDistance) * 0.12;
                        ctx.strokeStyle = `rgba(0, 0, 0, ${opacity})`;
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

const TerminalBlock = ({ title, content, isDark = false, className = "" }) => (
    <div
        className={`terminal-block ${className}`}
        style={isDark ? { backgroundColor: '#000000', color: '#FFFFFF' } : {}}
    >
        {title && (
            <div className={`mono-label mb-4 pb-2 border-b ${isDark ? 'border-white/20' : 'border-black/10'}`} style={isDark ? { color: '#AAAAAA', borderColor: 'rgba(255,255,255,0.2)' } : {}}>
                {title}
            </div>
        )}
        <pre
            className="font-tech text-xs md:text-sm whitespace-pre-wrap leading-relaxed"
            style={isDark ? { color: '#FFFFFF' } : { color: '#000000' }}
        >
            {content}
        </pre>
    </div>
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
        <div className="selection:bg-black selection:text-white relative bg-[#FAFAFA] min-h-screen flex flex-col">
            <CanvasBackground />

            {/* Navigation */}
            <nav className="fixed top-0 left-0 w-full z-50 p-6 flex justify-between items-center mix-blend-difference pointer-events-none">
                <div className="flex gap-4 pointer-events-auto">
                    <Link to="/" className="font-tech text-xs uppercase tracking-widest border border-black px-3 py-1 hover:bg-black hover:text-white transition-all">HOME</Link>
                    <Link to="/risk-assessment" className="font-tech text-xs uppercase tracking-widest border border-black px-3 py-1 bg-black text-white hover:bg-white hover:text-black transition-all">RISK_PILOT</Link>
                    <a href="https://ocultar.dev/docs" className="font-tech text-xs uppercase tracking-widest border border-black px-3 py-1 hover:bg-black hover:text-white transition-all">DOCS</a>
                </div>
                <div className="flex gap-4 pointer-events-auto">
                    <a href="#pilot" className="font-tech text-xs uppercase tracking-widest border border-black px-3 py-1 hover:bg-black hover:text-white transition-all">INIT_DEPLOY</a>
                    <a href="#github" className="font-tech text-xs uppercase tracking-widest border border-black px-3 py-1 hover:bg-black hover:text-white transition-all">GITHUB</a>
                </div>
            </nav>

            <div className="flex-grow pt-32 pb-20">
                <Routes>
                    <Route path="/" element={<LandingPage />} />
                    <Route path="/risk-assessment" element={<RiskAssessmentPage />} />
                </Routes>
            </div>

            {/* Common Footer */}
            <footer className="py-20 border-t border-black px-6 w-full mt-auto">
                <div className="max-w-[900px] mx-auto space-y-12">
                    <div className="flex flex-col md:flex-row justify-between items-start gap-12">
                        <div className="space-y-4">
                            <div className="font-bold text-lg uppercase tracking-widest">OCULTAR</div>
                            <div className="font-tech text-[10px] text-dim">BUILD_2026.04.07 // ENTERPRISE_PILOT_MODE</div>
                        </div>
                        <div className="grid grid-cols-2 md:grid-cols-3 gap-12 font-tech text-[10px] uppercase">
                            <div className="space-y-3">
                                <span className="text-dim">Network</span>
                                <a href="https://ocultar.dev/docs" className="block hover:line-through">Documentation</a>
                                <Link to="/risk-assessment" className="block hover:line-through">Risk Pilot</Link>
                                <a href="#faq" className="block hover:line-through">FAQ</a>
                            </div>
                            <div className="space-y-3">
                                <span className="text-dim">Systems</span>
                                <a href="/#roi" className="block hover:line-through">ROI Calculator</a>
                                <a href="/#pilot" className="block hover:line-through">Pilot Program</a>
                            </div>
                            <div className="space-y-3">
                                <span className="text-dim">Contact</span>
                                <a href="mailto:sales@ocultar.dev" className="block hover:line-through">sales@ocultar.dev</a>
                            </div>
                        </div>
                    </div>
                    <div className="pt-8 border-t border-black/10 font-tech text-[9px] text-dim text-center uppercase tracking-[0.2em]">
                        &copy; 2026 OCULTAR Security // FAIL_CLOSED OR NOTHING
                    </div>
                </div>
            </footer>
        </div>
    );
}

function LandingPage() {
    const [apiUsage, setApiUsage] = useState(100000);
    const [piiPercent, setPiiPercent] = useState(12);
    const [riskCost, setRiskCost] = useState(150);

    const annualReqs = apiUsage * 12;
    const exposedReqs = annualReqs * (piiPercent / 100);
    const baseExposure = exposedReqs * riskCost * 0.05;
    const ocultarSavings = baseExposure * 0.99;

    const formatCurrency = (val) =>
        new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(val);

    return (
        <div className="page-wrapper">
            {/* 1. HERO */}
            <section className="min-h-[60vh] flex flex-col items-center justify-center text-center py-20 border-none">
                <div className="w-full mb-12 mix-blend-multiply opacity-90 transition-opacity hover:opacity-100">
                    <img src={logo} alt="OCULTAR LOGO" className="max-w-[600px] mx-auto w-full h-auto" style={{ mixBlendMode: 'multiply' }} />
                </div>
                <div className="max-w-3xl space-y-6">
                    <h2 className="text-2xl md:text-4xl tracking-tight uppercase font-hero font-bold">Zero Egress. Zero Cloud. Total Sovereignty.</h2>
                    <p className="font-tech text-sm md:text-lg text-dim">All data flows refined locally before external interaction.</p>
                    <Link to="/risk-assessment" className="inline-block mt-8 bg-black text-white px-8 py-4 rounded-full font-tech uppercase text-xs tracking-widest hover:scale-105 transition-all">
                        Run Free Risk Audit
                    </Link>
                </div>
            </section>

            {/* 2. ARCHITECTURE / PIPELINE */}
            <section id="how-it-works" className="py-20 space-y-12 border-b border-black">
                <div className="flex justify-between items-end border-b border-black pb-4">
                    <h2 className="text-2xl font-bold uppercase font-hero">Architecture / Pipeline</h2>
                </div>
                <div className="flex flex-col lg:flex-row items-center justify-between gap-8 py-12">
                    <div className="w-full lg:w-[30%] border border-black p-8 min-h-[300px] flex flex-col items-center justify-center text-center">
                        <span className="font-bold mb-4 text-sm">[ INPUT DATA ]</span>
                        <span className="font-tech text-xs text-dim leading-relaxed">Raw logs, Prompts, <br /> JSON</span>
                    </div>
                    <div className="text-2xl opacity-50">→</div>
                    <div className="w-full lg:w-[32%] border border-black border-dashed p-8 relative min-h-[350px] flex flex-col justify-center gap-8 bg-white/40">
                        <div className="absolute -top-3 left-1/2 -translate-x-1/2 bg-[#FAFAFA] border border-black px-3 py-0.5 text-[10px] font-bold tracking-widest uppercase">Zero-Egress Zone</div>
                        <div className="text-center space-y-2">
                            <span className="font-bold block uppercase text-xs">Live Refinery</span>
                            <span className="font-tech text-[10px] text-dim block">Regex → NLP → SLM Cascade</span>
                        </div>
                        <div className="w-16 h-px bg-black/20 mx-auto"></div>
                        <div className="text-center space-y-2">
                            <span className="font-bold block uppercase text-xs">Identity Vault</span>
                            <span className="font-tech text-[10px] text-dim block">AES-256-GCM / Deterministic</span>
                        </div>
                        <div className="w-16 h-px bg-black/20 mx-auto"></div>
                        <div className="text-center space-y-2">
                            <span className="font-bold block uppercase text-xs">Sombra Gateway</span>
                            <span className="font-tech text-[10px] text-dim block">Fail-Closed Proxy Routing</span>
                        </div>
                    </div>
                    <div className="text-2xl opacity-50">→</div>
                    <div className="w-full lg:w-[30%] border border-black p-8 min-h-[300px] flex flex-col items-center justify-center text-center">
                        <span className="font-bold mb-4 text-sm">[ EXTERNAL AI ]</span>
                        <span className="font-tech text-xs text-dim leading-relaxed">Tokenized output <br /> only</span>
                    </div>
                </div>
            </section>

            {/* 3. WHY THIS EXISTS */}
            <section className="py-20 border-l-[6px] border-black pl-8 space-y-10 border-b border-black">
                <h2 className="text-4xl md:text-6xl font-bold tracking-tighter leading-[0.9] font-hero">
                    EVERY AI REQUEST LEAKS CONTEXT. <br />LOGS. PROMPTS. METADATA.
                </h2>
                <div className="bg-black text-white px-6 py-3 inline-block font-tech text-md md:text-lg">
                    We remove the risk surface entirely.
                </div>
                <ul className="font-tech text-xs space-y-4 pt-4 uppercase tracking-tight opacity-80">
                    <li key="why1" className="flex items-center gap-3"><span>[✓]</span> NO SAAS DEPENDENCY</li>
                    <li key="why2" className="flex items-center gap-3"><span>[✓]</span> NO EXTERNAL TRUST BOUNDARY</li>
                    <li key="why3" className="flex items-center gap-3"><span>[✓]</span> NO PROBABILISTIC COMPLIANCE</li>
                </ul>
            </section>

            {/* 4. SYSTEM STREAMS */}
            <section className="py-20 space-y-12 border-b border-black">
                <div className="flex justify-between items-end border-b border-black pb-4">
                    <h2 className="text-2xl font-bold uppercase font-hero">System Streams</h2>
                </div>
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                    <div className="space-y-8">
                        <TerminalBlock
                            title="SYS::DATA_REFINEMENT_ENGINE"
                            content={`> INIT TIER_0_SHIELD ... OK\n> LOAD REGEX_HEURISTICS ... OK\n> MOUNT NLP_NER_MODEL ... OK\n> MOUNT LOCAL_SLM ... OK\n\nSTATUS:\n- Sub-ms latency enforcement\n- Obfuscation resistance: ACTIVE`}
                        />
                        <TerminalBlock
                            title="SYS::SOMBRA_GATEWAY"
                            content={`> BIND :8080 (HTTP/PROXY)\n> BIND :514 (SYSLOG)\n> ROUTING TABLE LOADED\n\nPOLICY:\n- Fail-Closed default\n- Multi-model balancing: ACTIVE`}
                        />
                    </div>
                    <div className="space-y-8">
                        <TerminalBlock
                            isDark={true}
                            title="PROOF OF TRANSFORMATION"
                            content={`# INTERCEPTED INPUT\nPAYLOAD: "User John Doe, SSN 123-45-6789 requested access."\n\n# REFINEMENT EXECUTED (2.4ms)\nMATCH: [Name] -> John Doe\nMATCH: [SSN] -> ***-**-6789\n\n# EGRESS PAYLOAD\nOUTPUT: "User [PERSON_001], SSN [TOKEN_002] requested access."\n\n# VAULT STATE (LOCAL DUCKDB)\nMAP: PERSON_001 <-> John Doe\nMAP: TOKEN_002 <-> 123-45-6789 (Encrypted)`}
                            className="min-h-[350px]"
                        />
                        <TerminalBlock
                            title="SYS::GOVERNANCE"
                            content={`AUDIT LOG TAIL:\n[2026-03-19T07:34:11Z] PII_REDACT req_id=992a action=vault\n[2026-03-19T07:34:12Z] POLICY_EVAL req_id=992b result=pass\n[2026-03-19T07:34:15Z] REGULATORY_MAP GDPR=pass HIPAA=pass`}
                        />
                    </div>
                </div>
            </section>

            {/* 5. VALUE REALIZATION */}
            <section id="roi" className="py-20 space-y-12 border-b border-black">
                <div className="flex justify-between items-end border-b border-black pb-4">
                    <h2 className="text-2xl font-bold uppercase font-hero">Value Realization</h2>
                    <span className="mono-label tracking-widest text-[9px]">Live_Telemetry</span>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                    <div className="space-y-6">
                        <p className="font-tech text-sm text-dim leading-relaxed">
                            Real-time quantification of capital retention achieved via zero-egress architecture.
                            This logic bypasses external API taxation by executing SLM-based PII refinement within your local security perimeter.
                        </p>
                        <div className="flex gap-4">
                            <div className="border border-black px-4 py-2 font-tech text-[10px] uppercase">Vector: Local_Refine</div>
                            <div className="border border-black px-4 py-2 font-tech text-[10px] uppercase">Status: Zero_Egress</div>
                        </div>
                    </div>
                    <ROIDashboardCard />
                </div>
            </section>

            {/* 6. DEPLOYMENT MODES */}
            <section className="py-20 space-y-12 border-b border-black">
                <div className="flex justify-between items-center border-b border-black pb-4">
                    <h2 className="text-2xl font-bold uppercase font-hero">Deployment Modes</h2>
                    <span className="mono-label">config_blocks.yml</span>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-0 border border-black">
                    <div key="community" className="p-8 border-r border-black space-y-4">
                        <span className="font-bold text-xs uppercase">[ Community ]</span>
                        <ul className="font-tech text-[10px] space-y-1 text-dim">
                            <li key="c1">- local binary</li><li key="c2">- duckdb</li><li key="c3">- http proxy</li>
                        </ul>
                    </div>
                    <div key="sombra" className="p-8 border-r border-black bg-black text-white space-y-4">
                        <span className="font-bold text-xs uppercase">[ Sombra ]</span>
                        <ul className="font-tech text-[10px] space-y-1 opacity-70">
                            <li key="s1">- gateway</li><li key="s2">- connectors</li><li key="s3">- advanced routing</li>
                        </ul>
                    </div>
                    <div key="enterprise" className="p-8 space-y-4">
                        <span className="font-bold text-xs uppercase">[ Enterprise ]</span>
                        <ul className="font-tech text-[10px] space-y-1 text-dim">
                            <li key="e1">- postgres vaulting</li><li key="e2">- rbac / sso</li><li key="e3">- risk matrix monitoring</li>
                        </ul>
                    </div>
                </div>
            </section>

            {/* 7. QUANTIFY RISK */}
            <section className="py-20 space-y-12 border-b border-black">
                <div className="text-center max-w-2xl mx-auto space-y-4">
                    <h2 className="text-3xl font-bold uppercase font-hero">Quantify Risk</h2>
                    <p className="text-dim font-tech text-xs">Compute estimated annual exposure based on current AI integration scale.</p>
                </div>
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-0 border border-black">
                    <div className="p-8 space-y-8 border-b lg:border-b-0 lg:border-r border-black">
                        <div className="space-y-4">
                            <label className="mono-label block text-[10px]">Monthly AI/API Requests</label>
                            <input type="number" value={apiUsage} onChange={(e) => setApiUsage(Number(e.target.value))} className="w-full border border-black p-2 font-tech text-sm bg-transparent outline-none focus:bg-white" />
                        </div>
                        <div className="space-y-4">
                            <label className="mono-label block text-[10px]">PII Exposure Probability (%)</label>
                            <input type="number" value={piiPercent} onChange={(e) => setPiiPercent(Number(e.target.value))} className="w-full border border-black p-2 font-tech text-sm bg-transparent outline-none focus:bg-white" />
                        </div>
                        <div className="space-y-4">
                            <label className="mono-label block text-[10px]">Estimated Compliance Risk Cost (Per Record)</label>
                            <input type="number" value={riskCost} onChange={(e) => setRiskCost(Number(e.target.value))} className="w-full border border-black p-2 font-tech text-sm bg-transparent outline-none focus:bg-white" />
                        </div>
                    </div>
                    <div className="bg-white p-8 flex flex-col justify-center space-y-6">
                        <div className="space-y-1">
                            <span className="mono-label text-[10px]">Estimated Annual Risk</span>
                            <div className="text-3xl font-bold tabular-nums text-red-600">{formatCurrency(baseExposure)}</div>
                        </div>
                        <div className="space-y-1">
                            <span className="mono-label text-[10px]">Projected Savings (OCULTAR)</span>
                            <div className="text-3xl font-bold tabular-nums text-green-600">{formatCurrency(ocultarSavings)}</div>
                        </div>
                    </div>
                </div>
            </section>

            {/* 8. INITIALIZE DEPLOYMENT */}
            <section id="pilot" className="py-24 border-t border-black space-y-16">
                <h2 className="text-center text-4xl font-bold uppercase tracking-tight font-hero">Initialize Deployment</h2>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                    <div className="border border-black p-8 space-y-6 bg-white/40">
                        <div className="space-y-2">
                            <h3 className="font-bold uppercase text-sm">Community / Self-Serve</h3>
                            <p className="text-[10px] text-dim leading-relaxed">Download binary. Run locally in minutes. DuckDB included.</p>
                        </div>
                        <div className="font-tech text-xs pt-4 border-t border-black/10">{`> curl -sL https://ocultar.dev/install | bash`}</div>
                    </div>
                    <div className="border border-black p-8 space-y-6 bg-white/40">
                        <div className="space-y-2">
                            <h3 className="font-bold uppercase text-sm">Enterprise Pilot</h3>
                            <p className="text-[10px] text-dim leading-relaxed">Deploy in your infra in &lt;48h. Measure ROI + risk reduction.</p>
                        </div>
                        <div className="font-tech text-xs pt-4 border-t border-black/10">{`> request_deployment --guided`}</div>
                    </div>
                </div>
            </section>

            {/* 9. DEVELOPER ENTRY */}
            <section className="py-20 space-y-8">
                <div className="flex justify-between items-end border-b border-black pb-4">
                    <h2 className="text-2xl font-bold uppercase font-hero">Developer Entry</h2>
                    <span className="mono-label">api::v1</span>
                </div>
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                    <TerminalBlock
                        title="Proxy Integration"
                        isDark={true}
                        content={`curl -x http://localhost:8080 \\\n  -d "My name is John Doe" \\\n  https://api.openai.com/v1/completions`}
                    />
                    <TerminalBlock
                        title="Response Header"
                        content={`HTTP/1.1 200 OK\nX-Ocultar-Refined: true\nX-Ocultar-Latency: 2.1ms\n\n{"refined_payload": "My name is [PERSON_1]"}`}
                    />
                </div>
            </section>
        </div>
    );
}
