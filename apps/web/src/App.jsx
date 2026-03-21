import React, { useState, useEffect, useRef } from 'react';

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
                // Fail-Closed / Zero-Egress Boundary Bounce
                if (this.x <= 0 || this.x >= canvas.width) this.vx *= -1;
                if (this.y <= 0 || this.y >= canvas.height) this.vy *= -1;

                this.x += this.vx;
                this.y += this.vy;

                // Clamp to prevent getting stuck
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
                        // Subtle connections representing data flow
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

    return (
        <canvas
            ref={canvasRef}
            id="zero-egress-canvas"
        />
    );
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

const NavLink = ({ href, children }) => (
    <a
        href={href}
        className="font-tech text-xs uppercase tracking-widest border border-black px-3 py-1 hover:bg-black hover:text-white transition-all pointer-events-auto"
    >
        {children}
    </a>
);

// --- Main Application ---
export default function App() {
    // ROI Calculator State
    const [employees, setEmployees] = useState(500);
    const [apiUsage, setApiUsage] = useState(100000);
    const [piiPercent, setPiiPercent] = useState(12);
    const [riskCost, setRiskCost] = useState(150);

    // ROI Derived Computations
    const annualReqs = apiUsage * 12;
    const exposedReqs = annualReqs * (piiPercent / 100);
    const baseExposure = exposedReqs * riskCost * 0.05; 
    const ocultarSavings = baseExposure * 0.99;

    const formatCurrency = (val) =>
        new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: 'USD',
            maximumFractionDigits: 0
        }).format(val);

    return (
        <div className="selection:bg-black selection:text-white relative bg-[#FAFAFA] min-h-screen">
            <CanvasBackground />

            {/* Navigation (Minimal) */}
            <nav className="fixed top-0 left-0 w-full z-50 p-6 flex justify-between items-center mix-blend-difference pointer-events-none">
                <div className="flex gap-8 pointer-events-auto uppercase tracking-widest text-[10px] font-tech text-white">
                    <NavLink href="/roi_calc.html">ROI_CALC</NavLink>
                    <NavLink href="https://ocultar.dev/docs">DOCS</NavLink>
                    <NavLink href="#faq">FAQ</NavLink>
                </div>
                <div className="flex gap-8 pointer-events-auto">
                    <NavLink href="#pilot">INIT_DEPLOY</NavLink>
                    <NavLink href="#github">GITHUB</NavLink>
                </div>
            </nav>

            <main className="system-container">
                {/* 1. HERO */}
                <section className="min-h-screen flex flex-col items-center justify-center text-center pt-24 pb-[30vh]">
                    <div className="max-w-[800px] w-full mb-12 mix-blend-multiply opacity-90 transition-opacity hover:opacity-100">
                        <img 
                            src="/src/assets/images/image.webp" 
                            alt="OCULTAR LOGO" 
                            className="w-full h-auto"
                            style={{ mixBlendMode: 'multiply' }}
                        />
                    </div>
                    
                    <div className="max-w-3xl space-y-6">
                        <h2 className="text-2xl md:text-4xl tracking-tight">
                            Zero Egress. Zero Cloud. Total Sovereignty.
                        </h2>
                        <p className="font-tech text-sm md:text-lg text-dim">
                            All data flows refined locally before external interaction.
                        </p>
                    </div>
                </section>

                {/* 2. HOW IT WORKS (CRITICAL PIPELINE) */}
                <section id="how-it-works" className="space-y-16">
                    <div className="flex justify-between items-end border-b border-black pb-4">
                        <h2 className="text-2xl font-bold uppercase">Architecture / Pipeline</h2>
                    </div>

                    <div className="flex flex-col lg:flex-row items-center justify-between gap-8 py-12">
                        {/* INPUT BOX */}
                        <div className="w-full lg:w-[30%] border border-black p-12 min-h-[350px] flex flex-col items-center justify-center text-center">
                            <span className="font-bold text-lg mb-4">[ INPUT DATA ]</span>
                            <span className="font-tech text-xs text-dim leading-relaxed">Raw logs, Prompts, <br /> JSON</span>
                        </div>
                        
                        <div className="text-2xl opacity-50">→</div>
                        
                        {/* ZERO-EGRESS ZONE (DASHED) */}
                        <div className="w-full lg:w-[32%] border border-black border-dashed p-8 relative min-h-[400px] flex flex-col justify-center gap-12 bg-white/40">
                            <div className="absolute -top-3 left-1/2 -translate-x-1/2 bg-[#FAFAFA] border border-black px-3 py-0.5 text-[10px] font-bold tracking-widest uppercase">
                                Zero-Egress Zone
                            </div>
                            
                            <div className="text-center space-y-2">
                                <span className="font-bold block uppercase text-sm">Live Refinery</span>
                                <span className="font-tech text-[10px] text-dim block">Regex → NLP → SLM Cascade</span>
                            </div>
                            
                            <div className="w-16 h-px bg-black/20 mx-auto"></div>
                            
                            <div className="text-center space-y-2">
                                <span className="font-bold block uppercase text-sm">Identity Vault</span>
                                <span className="font-tech text-[10px] text-dim block">AES-256-GCM / Deterministic</span>
                            </div>
                            
                            <div className="w-16 h-px bg-black/20 mx-auto"></div>
                            
                            <div className="text-center space-y-2">
                                <span className="font-bold block uppercase text-sm">Sombra Gateway</span>
                                <span className="font-tech text-[10px] text-dim block">Fail-Closed Proxy Routing</span>
                            </div>
                        </div>

                        <div className="text-2xl opacity-50">→</div>

                        {/* EXTERNAL AI BOX */}
                        <div className="w-full lg:w-[30%] border border-black p-12 min-h-[350px] flex flex-col items-center justify-center text-center">
                            <span className="font-bold text-lg mb-4">[ EXTERNAL AI ]</span>
                            <span className="font-tech text-xs text-dim leading-relaxed">Tokenized output <br /> only</span>
                        </div>
                    </div>
                </section>

                {/* 5. WHY THIS EXISTS (CONTEXT LEAK) */}
                <section className="max-w-6xl border-l-[6px] border-black pl-12 py-20 space-y-10">
                    <h2 className="text-5xl md:text-7xl font-bold tracking-tighter leading-[0.9] max-w-4xl">
                        EVERY AI REQUEST LEAKS CONTEXT. <br />
                        LOGS. PROMPTS. METADATA.
                    </h2>
                    
                    <div className="bg-black text-white px-6 py-3 inline-block font-tech text-lg md:text-xl">
                        We remove the risk surface entirely.
                    </div>

                    <ul className="font-tech text-sm space-y-4 pt-4 uppercase tracking-tight opacity-80">
                        <li className="flex items-center gap-3">
                            <span>[✓]</span> NO SAAS DEPENDENCY
                        </li>
                        <li className="flex items-center gap-3">
                            <span>[✓]</span> NO EXTERNAL TRUST BOUNDARY
                        </li>
                        <li className="flex items-center gap-3">
                            <span>[✓]</span> NO PROBABILISTIC COMPLIANCE
                        </li>
                    </ul>
                </section>

                {/* 3. LIVE CAPABILITIES & 4. PROOF (SYSTEM STREAMS) */}
                <section className="space-y-12">
                    <div className="flex justify-between items-end border-b border-black pb-4">
                        <h2 className="text-2xl font-bold uppercase">System Streams</h2>
                    </div>

                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-12">
                        {/* LEFT COLUMN: BLOCKS */}
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

                        {/* RIGHT COLUMN: PROOF & GOVERNANCE */}
                        <div className="space-y-8">
                            <TerminalBlock 
                                isDark={true}
                                title="PROOF OF TRANSFORMATION"
                                content={`# INTERCEPTED INPUT\nPAYLOAD: "User John Doe, SSN 123-45-6789 requested access." \n\n# REFINEMENT EXECUTED (2.4ms)\nMATCH: [Name] -> John Doe\nMATCH: [SSN] -> ***-**-6789\n\n# EGRESS PAYLOAD\nOUTPUT: "User [PERSON_001], SSN [TOKEN_002] requested access."\n\n# VAULT STATE (LOCAL DUCKDB)\nMAP: PERSON_001 <-> John Doe\nMAP: TOKEN_002 <-> 123-45-6789 (Encrypted)`}
                                className="min-h-[350px]"
                            />
                            <TerminalBlock 
                                title="SYS::GOVERNANCE"
                                content={`AUDIT LOG TAIL:\n[2026-03-19T07:34:11Z] PII_REDACT req_id=992a action=vault\n[2026-03-19T07:34:12Z] POLICY_EVAL req_id=992b result=pass\n[2026-03-19T07:34:15Z] REGULATORY_MAP GDPR=pass HIPAA=pass`}
                            />
                        </div>
                    </div>
                </section>

                {/* 6. DEPLOYMENT MODES */}
                <section className="space-y-12">
                    <div className="flex justify-between items-center border-b border-black pb-4">
                        <h2 className="text-2xl font-bold uppercase">Deployment Modes</h2>
                        <span className="mono-label">config_blocks.yml</span>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-0 border border-black">
                        {/* COMMUNITY */}
                        <div className="p-8 border-r border-black space-y-4">
                            <span className="font-bold text-sm uppercase">[ Community ]</span>
                            <ul className="font-tech text-xs space-y-1 text-dim">
                                <li>- local binary</li>
                                <li>- duckdb</li>
                                <li>- http proxy</li>
                            </ul>
                        </div>
                        
                        {/* SOMBRA - SHADED */}
                        <div className="p-8 border-r border-black bg-black text-white space-y-4">
                            <span className="font-bold text-sm uppercase">[ Sombra ]</span>
                            <ul className="font-tech text-xs space-y-1 opacity-70">
                                <li>- gateway</li>
                                <li>- connectors</li>
                                <li>- advanced routing</li>
                            </ul>
                        </div>

                        {/* ENTERPRISE */}
                        <div className="p-8 space-y-4">
                            <span className="font-bold text-sm uppercase">[ Enterprise ]</span>
                            <ul className="font-tech text-xs space-y-1 text-dim">
                                <li>- postgres vaulting</li>
                                <li>- rbac / sso</li>
                                <li>- risk matrix monitoring</li>
                            </ul>
                        </div>
                    </div>
                </section>

                {/* 7. ROI CALCULATOR */}
                <section id="roi" className="space-y-12">
                    <div className="text-center max-w-2xl mx-auto space-y-4">
                        <h2 className="text-4xl font-bold uppercase">Quantify Risk</h2>
                        <p className="text-dim font-tech text-sm">Compute estimated annual exposure based on current AI integration scale.</p>
                    </div>

                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-0 border border-black">
                        <div className="p-8 space-y-8 border-b lg:border-b-0 lg:border-r border-black">
                            <div className="space-y-4">
                                <label className="mono-label block text-[10px]">Monthly AI/API Requests</label>
                                <input 
                                    type="number"
                                    value={apiUsage} onChange={(e) => setApiUsage(Number(e.target.value))}
                                    className="w-full border border-black p-2 font-tech text-sm bg-transparent outline-none focus:bg-white"
                                />
                                <div className="flex justify-between font-tech text-[10px] text-dim">
                                    <span>Scale: Enterprise Grade</span>
                                    <span>Default: 100k/mo</span>
                                </div>
                            </div>
                            
                            <div className="space-y-4">
                                <label className="mono-label block text-[10px]">PII Exposure Probability (%)</label>
                                <input 
                                    type="number"
                                    value={piiPercent} onChange={(e) => setPiiPercent(Number(e.target.value))}
                                    className="w-full border border-black p-2 font-tech text-sm bg-transparent outline-none focus:bg-white"
                                />
                                <div className="text-right font-tech text-[10px] text-dim">{piiPercent}% Risk Variable</div>
                            </div>

                            <div className="space-y-4">
                                <label className="mono-label block text-[10px]">Estimated Compliance Risk Cost (Per Record)</label>
                                <input 
                                    type="number"
                                    value={riskCost} onChange={(e) => setRiskCost(Number(e.target.value))}
                                    className="w-full border border-black p-2 font-tech text-sm bg-transparent outline-none focus:bg-white"
                                />
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
                            <div className="pt-4 border-t border-black/10 flex justify-between items-center">
                                <span className="font-tech text-[10px] uppercase">Risk Reduction</span>
                                <span className="font-bold underline decoration-2 underline-offset-4">99.2%</span>
                            </div>
                        </div>
                    </div>
                </section>

                {/* 8. INITIALIZE DEPLOYMENT */}
                <section id="pilot" className="py-24 border-t border-black space-y-16">
                    <h2 className="text-center text-4xl font-bold uppercase tracking-tight">Initialize Deployment</h2>
                    
                    <div className="flex flex-col md:flex-row justify-center gap-8 max-w-5xl mx-auto">
                        {/* COMMUNITY */}
                        <div className="flex-1 border border-black p-10 space-y-8 bg-white/40">
                            <div className="space-y-2">
                                <h3 className="font-bold uppercase text-sm">Community / Self-Serve</h3>
                                <p className="text-[10px] text-dim leading-relaxed">Download binary. Run locally in minutes. DuckDB included.</p>
                            </div>
                            <div className="font-tech text-xs pt-4 border-t border-black/10">
                                {`> curl -sL https://ocultar.dev/install | bash`}
                            </div>
                        </div>

                        {/* ENTERPRISE */}
                        <div className="flex-1 border border-black p-10 space-y-8 bg-white/40">
                            <div className="space-y-2">
                                <h3 className="font-bold uppercase text-sm">Enterprise Pilot</h3>
                                <p className="text-[10px] text-dim leading-relaxed">Deploy in your infra in &lt;48h. Measure ROI + risk reduction.</p>
                            </div>
                            <div className="font-tech text-xs pt-4 border-t border-black/10">
                                {`> request_deployment --guided`}
                            </div>
                        </div>
                    </div>
                </section>

                {/* 9. DEVELOPER ENTRY */}
                <section className="space-y-8">
                    <div className="flex justify-between items-end border-b border-black pb-4">
                        <h2 className="text-2xl font-bold uppercase">Developer Entry</h2>
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

                {/* Footer */}
                <footer className="py-20 border-t border-black space-y-12">
                    <div className="flex flex-col md:flex-row justify-between items-start gap-12">
                        <div className="space-y-4">
                            <div className="font-bold text-lg uppercase tracking-widest">OCULTAR</div>
                            <div className="font-tech text-[10px] text-dim">BUILD_2026.03.19 // ZERO_EGRESS_REFINERY</div>
                        </div>
                        
                        <div className="grid grid-cols-2 md:grid-cols-3 gap-12 font-tech text-[10px] uppercase">
                            <div className="space-y-3">
                                <span className="text-dim">Network</span>
                                <a href="https://ocultar.dev/docs" className="block hover:line-through">Documentation</a>
                                <a href="#github" className="block hover:line-through">GitHub</a>
                                <a href="#faq" className="block hover:line-through">FAQ</a>
                            </div>
                            <div className="space-y-3">
                                <span className="text-dim">Systems</span>
                                <a href="#roi" className="block hover:line-through">ROI Calculator</a>
                                <a href="#pilot" className="block hover:line-through">Pilot Program</a>
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
                </footer>
            </main>
        </div>
    );
}
