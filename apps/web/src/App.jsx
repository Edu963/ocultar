import { useState, useEffect, useRef } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { Shield, Lock, Zap, BarChart3, ChevronRight, Github, ExternalLink, Globe, ArrowRight, AlertTriangle, Code2, Terminal } from 'lucide-react';
import ROIDashboardCard from './components/ROIDashboardCard';
import RiskAssessmentPage from './RiskAssessmentPage';
import CalculatorPage from './CalculatorPage';
import TerminalLog from './components/TerminalLog';
import ComplianceBar from './components/ComplianceBar';
import ArchitectureDiagram from './components/ArchitectureDiagram';

// --- Subtle Particle Background ---
const CanvasBackground = () => {
    const canvasRef = useRef(null);

    useEffect(() => {
        const canvas = canvasRef.current;
        const ctx = canvas.getContext('2d');
        let animationFrameId;
        let particles = [];
        const numParticles = 30; // Fewer for more focus
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
                this.radius = Math.random() * 1.5;
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
                ctx.fillStyle = 'rgba(16, 185, 129, 0.2)';
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
                        const opacity = (1 - distance / connectionDistance) * 0.1;
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

            {/* Navigation */}
            <nav className="fixed top-0 left-0 w-full z-50 py-4 transition-all duration-300">
                <div className="max-container flex justify-between items-center">
                    <Link to="/" className="font-heading font-bold text-2xl tracking-tighter flex items-center gap-2 group">
                        <div className="w-10 h-10 bg-emerald-500 rounded-sm flex items-center justify-center group-hover:bg-emerald-400 transition-colors">
                            <Shield className="text-black w-6 h-6" strokeWidth={2.5} />
                        </div>
                        <span className="text-white">OCULTAR</span>
                    </Link>
                    <div className="hidden md:flex items-center gap-10 text-[13px] font-bold uppercase tracking-widest">
                        <Link to="/" className="hover:text-emerald-400 transition-colors text-slate-400">Home</Link>
                        <Link to="/risk-assessment" className="hover:text-emerald-400 transition-colors text-slate-400">Security Audit</Link>
                        <Link to="/calculator" className="hover:text-emerald-400 transition-colors text-slate-400">ROI Forecast</Link>
                        <a href="https://ocultar.dev/docs" className="hover:text-emerald-400 transition-colors text-slate-400 flex items-center gap-1">
                            Docs <ExternalLink className="w-3 h-3" />
                        </a>
                    </div>
                    <div className="flex items-center gap-4">
                        <Link to="/risk-assessment" className="btn btn-primary text-xs px-6 py-3">
                            Start Pilot
                        </Link>
                    </div>
                </div>
            </nav>

            <div className="flex-grow pt-20">
                <Routes>
                    <Route path="/" element={<LandingPage />} />
                    <Route path="/risk-assessment" element={<RiskAssessmentPage />} />
                    <Route path="/calculator" element={<CalculatorPage />} />
                </Routes>
            </div>

            {/* Footer */}
            <footer className="bg-secondary border-t border-white/5 py-24 relative overflow-hidden">
                <div className="max-container relative z-10">
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-16">
                        <div className="col-span-1 md:col-span-2 space-y-8">
                            <div className="font-heading font-bold text-2xl tracking-tighter text-white">OCULTAR</div>
                            <p className="max-w-xs text-sm text-slate-400 leading-relaxed">
                                The Switzerland of Data. Zero-Egress architecture for the modern AI stack. 
                                Secure your perimeter. Fail-closed by design.
                            </p>
                            <div className="flex gap-6">
                                <a href="https://github.com/Edu963/ocultar" className="text-slate-500 hover:text-emerald-400 transition-colors">
                                    <Github className="w-6 h-6" />
                                </a>
                                <a href="#" className="text-slate-500 hover:text-emerald-400 transition-colors">
                                    <Globe className="w-6 h-6" />
                                </a>
                            </div>
                        </div>
                        <div className="space-y-6">
                            <h4 className="text-[10px] font-bold uppercase tracking-[0.3em] text-emerald-500">Security Stack</h4>
                            <ul className="space-y-3 text-sm font-medium">
                                <li><Link to="/risk-assessment" className="text-slate-400 hover:text-white transition-colors">Risk Assessment</Link></li>
                                <li><Link to="/calculator" className="text-slate-400 hover:text-white transition-colors">ROI Calculator</Link></li>
                                <li><a href="https://ocultar.dev/docs" className="text-slate-400 hover:text-white transition-colors">Developer Docs</a></li>
                            </ul>
                        </div>
                        <div className="space-y-6">
                            <h4 className="text-[10px] font-bold uppercase tracking-[0.3em] text-emerald-500">Enterprise</h4>
                            <ul className="space-y-3 text-sm font-medium">
                                <li><a href="#" className="text-slate-400 hover:text-white transition-colors">Privacy Policy</a></li>
                                <li><a href="#" className="text-slate-400 hover:text-white transition-colors">Terms of Service</a></li>
                                <li><a href="mailto:sales@ocultar.dev" className="text-slate-400 hover:text-white transition-colors">Contact Engineering</a></li>
                            </ul>
                        </div>
                    </div>
                    <div className="mt-24 pt-12 border-t border-white/5 flex flex-col md:row justify-between items-center gap-6 text-[10px] text-slate-600 font-mono tracking-widest uppercase">
                        <div>&copy; {new Date().getFullYear()} OCULTAR SECURITY. ZERO-EGRESS GUARANTEED.</div>
                        <div className="flex gap-8">
                            <span className="flex items-center gap-2"><span className="w-1.5 h-1.5 bg-emerald-500/50 rounded-full animate-pulse"></span> SYSTEM: NOMINAL</span>
                            <span>BUILD: 2026.04.12.LEGAL</span>
                        </div>
                    </div>
                </div>
            </footer>
        </div>
    );
}

function LandingPage() {
    return (
        <div className="animate-fade-in-up">
            {/* Hero Section */}
            <section className="section-padding flex flex-col items-center">
                <div className="max-container text-center space-y-10 relative">
                    <div className="badge animate-in slide-in-from-bottom-4 duration-700">
                        <span className="relative flex h-2 w-2">
                            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                            <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
                        </span>
                        Enterprise-Grade Zero-Egress Proxy
                    </div>
                    
                    <h1 className="max-w-5xl mx-auto tracking-tighter">
                        Stop Leaking <span className="text-emerald-500">PII</span> to the Cloud.<br/>
                        Start Using AI Safely.
                    </h1>
                    
                    <p className="max-w-2xl mx-auto text-xl text-slate-400 font-medium leading-relaxed">
                        The only <span className="text-white border-b border-emerald-500/30">Fail-Closed PII Refinery</span>. 
                        Intercept, Vault, and Redact sensitive data locally before it ever reaches an LLM provider.
                    </p>

                    <div className="flex flex-col sm:flex-row gap-6 justify-center pt-6">
                        <Link to="/risk-assessment" className="btn btn-primary px-10 py-5 text-lg gap-2 group">
                            Start 10-Minute Pilot <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
                        </Link>
                        <a href="https://github.com/Edu963/ocultar" className="btn btn-secondary px-10 py-5 text-lg gap-3">
                            <Github className="w-5 h-5" /> View on GitHub
                        </a>
                    </div>
                    
                    {/* Live Terminal Feed */}
                    <TerminalLog />
                </div>
            </section>

            {/* Compliance Scrolling Bar */}
            <ComplianceBar />

            {/* How it Works: Architecture Diagram */}
            <section className="section-padding bg-secondary relative overflow-hidden">
                <div className="absolute top-0 left-0 w-full h-px bg-gradient-to-r from-transparent via-white/5 to-transparent"></div>
                <div className="max-container">
                    <div className="text-center space-y-4 mb-20">
                        <h2 className="text-gradient">Engineered for Sovereignty</h2>
                        <p className="text-slate-400 max-w-2xl mx-auto">
                            Ocultar acts as a transparent refinery between your application and the cloud providers.
                        </p>
                    </div>
                    
                    <ArchitectureDiagram />
                </div>
                <div className="absolute bottom-0 left-0 w-full h-px bg-gradient-to-r from-transparent via-white/5 to-transparent"></div>
            </section>

            {/* Core Capabilities */}
            <section className="section-padding relative">
                <div className="max-container">
                    <div className="feature-grid">
                        <div className="card space-y-6 group">
                            <div className="w-12 h-12 bg-emerald-500/10 border border-emerald-500/20 rounded-lg flex items-center justify-center group-hover:bg-emerald-500 group-hover:text-black transition-all">
                                <Lock className="w-6 h-6" />
                            </div>
                            <h3 className="text-xl italic">Zero-Egress Processing</h3>
                            <p className="text-sm text-slate-400 leading-relaxed">
                                Processes all PII redaction within your local security perimeter. 
                                Raw data never crosses a network boundary to a 3rd party API.
                            </p>
                        </div>
                        
                        <div className="card space-y-6 group">
                            <div className="w-12 h-12 bg-sky-500/10 border border-sky-500/20 rounded-lg flex items-center justify-center group-hover:bg-sky-500 group-hover:text-black transition-all">
                                <Zap className="w-6 h-6" />
                            </div>
                            <h3 className="text-xl italic">Multi-Layer Detection</h3>
                            <p className="text-sm text-slate-400 leading-relaxed">
                                Leveraging high-speed deterministic pipelines paired with deep-scan Local SLMs 
                                for contextual entity recognition.
                            </p>
                        </div>
                        
                        <div className="card space-y-6 group border-red-500/10 hover:border-red-500/40">
                            <div className="w-12 h-12 bg-red-500/10 border border-red-500/20 rounded-lg flex items-center justify-center group-hover:bg-red-500 group-hover:text-black transition-all">
                                <AlertTriangle className="w-6 h-6" />
                            </div>
                            <h3 className="text-xl italic">Fail-Closed Guarantee</h3>
                            <p className="text-sm text-slate-400 leading-relaxed">
                                We'd rather break the app than break the law. If deep scanning hits a resource limit, 
                                egress is automatically blocked.
                            </p>
                        </div>
                        
                        <div className="card space-y-6 group">
                            <div className="w-12 h-12 bg-purple-500/10 border border-purple-500/20 rounded-lg flex items-center justify-center group-hover:bg-purple-500 group-hover:text-black transition-all">
                                <Code2 className="w-6 h-6" />
                            </div>
                            <h3 className="text-xl italic">Developer-First DX</h3>
                            <p className="text-sm text-slate-400 leading-relaxed">
                                Invisible to your devs. Lives in your Docker Compose. Simply update your LLM base URL 
                                and gain instant compliance.
                            </p>
                        </div>
                    </div>
                </div>
            </section>

            {/* ROI Cost Section */}
            <section className="section-padding bg-[#080809] border-y border-white/5">
                <div className="max-container">
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-20 items-center">
                        <div className="space-y-10">
                            <div className="space-y-4">
                                <div className="text-emerald-500 font-mono text-[10px] tracking-[0.4em] uppercase font-bold">Risk Quantification</div>
                                <h2 className="text-white">Quantify Your Exposure</h2>
                                <p className="text-slate-400 leading-relaxed">
                                    Use our ROI engine to estimate your annual financial risk based on PII exposure 
                                    probability and industry compliance benchmarks.
                                </p>
                            </div>
                            
                            <div className="space-y-8 p-8 bg-white/[0.02] border border-white/5 rounded-2xl">
                                <div className="grid grid-cols-2 gap-8">
                                    <div className="space-y-2">
                                        <div className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Est. Risk (Annual)</div>
                                        <div className="text-3xl font-bold text-red-500">$1,450,000</div>
                                    </div>
                                    <div className="space-y-2 text-right">
                                        <div className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Ocultar Savings</div>
                                        <div className="text-3xl font-bold text-emerald-500">$1,435,500</div>
                                    </div>
                                </div>
                                <Link to="/calculator" className="btn btn-secondary w-full py-4 uppercase text-[10px] tracking-[0.2em] font-bold">
                                    Open Detailed ROI Engine
                                </Link>
                            </div>
                        </div>
                        
                        <div className="relative group">
                             <div className="absolute -inset-4 bg-emerald-500/5 blur-3xl opacity-0 group-hover:opacity-100 transition-opacity"></div>
                             <ROIDashboardCard className="relative z-10 scale-105" />
                        </div>
                    </div>
                </div>
            </section>

            {/* Final CTA */}
            <section className="section-padding relative overflow-hidden">
                <div className="absolute inset-0 bg-emerald-500/[0.02] scale-150 rotate-12 blur-3xl"></div>
                <div className="max-container text-center space-y-10 relative z-10">
                    <h2 className="text-white tracking-tight">Ready for Technical Sovereignty?</h2>
                    <p className="text-xl text-slate-400 max-w-2xl mx-auto leading-relaxed">
                        Deploy Ocultar in your infrastructure in minutes. Download our open-core binary 
                        or request a guided enterprise pilot with founder-level support.
                    </p>
                    <div className="flex flex-col md:flex-row gap-6 justify-center pt-8">
                        <button className="btn btn-primary px-16 py-6 text-xl font-bold shadow-[0_0_50px_rgba(16,185,129,0.3)]">
                            Initialize Deployment
                        </button>
                        <Link to="/risk-assessment" className="btn btn-secondary px-16 py-6 text-xl font-bold">
                            Request Audit
                        </Link>
                    </div>
                    <div className="pt-12 flex justify-center items-center gap-10 opacity-30 grayscale grayscale-100">
                         {/* These would be logos */}
                         <div className="font-heading font-black italic text-xl">SOUVERAIN</div>
                         <div className="font-mono font-bold text-lg tracking-tighter">PROTO_GUARDv2</div>
                         <div className="font-heading font-bold text-xl uppercase">Refinery_Core</div>
                    </div>
                </div>
            </section>
        </div>
    );
}

