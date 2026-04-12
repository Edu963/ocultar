import { useState, useEffect, useRef } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { Shield, Lock, Zap, BarChart3, ChevronRight, Github, ExternalLink, Globe } from 'lucide-react';
import ROIDashboardCard from './components/ROIDashboardCard';
import RiskAssessmentPage from './RiskAssessmentPage';
import CalculatorPage from './CalculatorPage';
import logo from './assets/images/image.webp';

// --- Subtle Particle Background ---
const CanvasBackground = () => {
    const canvasRef = useRef(null);

    useEffect(() => {
        const canvas = canvasRef.current;
        const ctx = canvas.getContext('2d');
        let animationFrameId;
        let particles = [];
        const numParticles = 40; // Fewer particles for minimalist look
        const connectionDistance = 150;

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
                this.vx = (Math.random() - 0.5) * 0.4; // Slower
                this.vy = (Math.random() - 0.5) * 0.4;
                this.radius = 1;
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
                ctx.fillStyle = 'rgba(0, 0, 0, 0.1)';
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
                        const opacity = (1 - distance / connectionDistance) * 0.05;
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
        <div className="min-h-screen flex flex-col selection:bg-black selection:text-white">
            <CanvasBackground />

            {/* Navigation */}
            <nav className="fixed top-0 left-0 w-full z-50 py-4">
                <div className="max-container flex justify-between items-center">
                    <Link to="/" className="font-heading font-bold text-xl tracking-tighter flex items-center gap-2">
                        <div className="w-8 h-8 bg-black rounded-lg flex items-center justify-center">
                            <Shield className="text-white w-5 h-5" />
                        </div>
                        OCULTAR
                    </Link>
                    <div className="hidden md:flex items-center gap-8 text-sm font-medium">
                        <Link to="/" className="hover:text-black transition-colors text-secondary">Home</Link>
                        <Link to="/risk-assessment" className="hover:text-black transition-colors text-secondary">Risk Pilot</Link>
                        <Link to="/calculator" className="hover:text-black transition-colors text-secondary">ROI</Link>
                        <a href="https://ocultar.dev/docs" className="hover:text-black transition-colors text-secondary flex items-center gap-1">
                            Docs <ExternalLink className="w-3 h-3" />
                        </a>
                    </div>
                    <div className="flex items-center gap-4">
                        <Link to="/risk-assessment" className="btn btn-primary text-sm px-6">
                            Start Audit
                        </Link>
                    </div>
                </div>
            </nav>

            <div className="flex-grow">
                <Routes>
                    <Route path="/" element={<LandingPage />} />
                    <Route path="/risk-assessment" element={<RiskAssessmentPage />} />
                    <Route path="/calculator" element={<CalculatorPage />} />
                </Routes>
            </div>

            {/* Footer */}
            <footer className="bg-secondary border-t border-color py-20">
                <div className="max-container">
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-12">
                        <div className="col-span-1 md:col-span-2 space-y-6">
                            <div className="font-heading font-bold text-xl tracking-tighter">OCULTAR</div>
                            <p className="max-w-xs text-sm">
                                Zero-Egress data protection for the modern AI stack. 
                                Secure your PII locally. Redact before you egress.
                            </p>
                            <div className="flex gap-4">
                                <a href="https://github.com/Edu963/ocultar" className="text-muted hover:text-black transition-colors">
                                    <Github className="w-5 h-5" />
                                </a>
                                <a href="#" className="text-muted hover:text-black transition-colors">
                                    <Globe className="w-5 h-5" />
                                </a>
                            </div>
                        </div>
                        <div className="space-y-4">
                            <h4 className="text-xs font-bold uppercase tracking-widest text-primary">Product</h4>
                            <ul className="space-y-2 text-sm">
                                <li><Link to="/risk-assessment" className="hover:text-black transition-colors">Risk Pilot</Link></li>
                                <li><Link to="/calculator" className="hover:text-black transition-colors">ROI Calculator</Link></li>
                                <li><a href="https://ocultar.dev/docs" className="hover:text-black transition-colors">Documentation</a></li>
                            </ul>
                        </div>
                        <div className="space-y-4">
                            <h4 className="text-xs font-bold uppercase tracking-widest text-primary">Company</h4>
                            <ul className="space-y-2 text-sm">
                                <li><a href="#" className="hover:text-black transition-colors">Privacy Policy</a></li>
                                <li><a href="#" className="hover:text-black transition-colors">Terms of Service</a></li>
                                <li><a href="mailto:sales@ocultar.dev" className="hover:text-black transition-colors">Contact Sales</a></li>
                            </ul>
                        </div>
                    </div>
                    <div className="mt-20 pt-8 border-t border-color flex flex-col md:row justify-between items-center gap-4 text-xs text-muted font-medium">
                        <div>&copy; 2026 OCULTAR Security. All rights reserved.</div>
                        <div className="flex gap-6">
                            <span>FAIL_CLOSED_OR_NOTHING</span>
                            <span>BUILD_2026.04.08</span>
                        </div>
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
        <div className="animate-in fade-in duration-1000">
            {/* Hero Section */}
            <section className="section-padding flex flex-col items-center text-center">
                <div className="max-container space-y-8">
                    <div className="badge animate-in slide-in-from-bottom-2 duration-500">
                        ✨ Secure AI Infrastructure
                    </div>
                    <h1 className="max-w-4xl mx-auto text-gradient">
                        Secure AI with Zero-Egress PII Redaction
                    </h1>
                    <p className="max-w-2xl mx-auto text-lg md:text-xl font-medium leading-relaxed">
                        Ocultar intercepts and redacts sensitive data locally before it leaves your infrastructure. 
                        No cloud dependencies. No data leaks. Fail-closed by design.
                    </p>
                    <div className="flex flex-col md:flex-row gap-4 justify-center pt-4">
                        <Link to="/risk-assessment" className="btn btn-primary px-8 py-4 text-base gap-2">
                            Run a Free Audit <ChevronRight className="w-4 h-4" />
                        </Link>
                        <Link to="/calculator" className="btn btn-secondary px-8 py-4 text-base">
                            View ROI Calculator
                        </Link>
                    </div>
                </div>
                
                {/* Visual Placeholder (Could be an image or a simplified diagram) */}
                <div className="mt-20 max-container w-full">
                    <div className="card bg-secondary/50 border-dashed border-2 flex items-center justify-center py-20 opacity-80">
                         <div className="flex flex-col md:flex-row items-center gap-8 text-sm font-bold uppercase tracking-widest text-muted">
                            <div className="flex flex-col items-center gap-4">
                                <div className="p-6 bg-white border border-color rounded-2xl shadow-sm text-primary">Raw Input</div>
                                <span>Sensitive JSON</span>
                            </div>
                            <div className="text-2xl">→</div>
                            <div className="flex flex-col items-center gap-4">
                                <div className="p-8 bg-black text-white rounded-3xl shadow-xl flex items-center gap-3">
                                    <Shield className="w-6 h-6 text-green-400" /> REFINERY
                                </div>
                                <span className="text-primary">Local SLM Processing</span>
                            </div>
                            <div className="text-2xl">→</div>
                            <div className="flex flex-col items-center gap-4 text-muted">
                                <div className="p-6 bg-white border border-color rounded-2xl shadow-sm">Vaulted Output</div>
                                <span>Tokenized Data</span>
                            </div>
                         </div>
                    </div>
                </div>
            </section>

            {/* Core Capabilities */}
            <section className="section-padding bg-secondary border-y border-color">
                <div className="max-container space-y-16">
                    <div className="text-center space-y-4">
                        <h2 className="text-primary font-bold">Engineered for Sovereignty</h2>
                        <p className="max-w-2xl mx-auto text-lg">Every capability maps directly to a hardened security requirement in our core engine.</p>
                    </div>
                    
                    <div className="feature-grid">
                        <div className="card space-y-6">
                            <div className="w-12 h-12 bg-black/5 rounded-xl flex items-center justify-center">
                                <Lock className="w-6 h-6" />
                            </div>
                            <h3 className="text-xl">Zero-Egress Processing</h3>
                            <p className="text-sm">Processes all PII redaction within your local security perimeter. Your raw data never crosses a network boundary to a 3rd party API.</p>
                        </div>
                        
                        <div className="card space-y-6">
                            <div className="w-12 h-12 bg-black/5 rounded-xl flex items-center justify-center">
                                <Zap className="w-6 h-6" />
                            </div>
                            <h3 className="text-xl">Multi-Layer Detection</h3>
                            <p className="text-sm">Leverages a high-speed deterministic regex pipeline paired with a deep-scan Local SLM for contextual semantic entity recognition.</p>
                        </div>
                        
                        <div className="card space-y-6">
                            <div className="w-12 h-12 bg-black/5 rounded-xl flex items-center justify-center">
                                <Shield className="w-6 h-6" />
                            </div>
                            <h3 className="text-xl">Fail-Closed Security</h3>
                            <p className="text-sm">Guaranteed protection. If the refinery engine or proxy encounter an error, all data egress is automatically blocked to prevent leaks.</p>
                        </div>
                        
                        <div className="card space-y-6">
                            <div className="w-12 h-12 bg-black/5 rounded-xl flex items-center justify-center">
                                <BarChart3 className="w-6 h-6" />
                            </div>
                            <h3 className="text-xl">High-Density Scaling</h3>
                            <p className="text-sm">Built for enterprise workloads. A parallelized refinery engine capable of processing gigabyte-scale data ingestion without bottlenecking.</p>
                        </div>
                    </div>
                </div>
            </section>

            {/* ROI Calculator Section */}
            <section id="roi" className="section-padding">
                <div className="max-container">
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
                        <div className="space-y-8">
                            <h2 className="text-primary">Quantify Your Exposure</h2>
                            <p className="text-lg">
                                Use our calculator to estimate your annual financial risk based on PII exposure probability and industry compliance benchmarks.
                            </p>
                            
                            <div className="space-y-6">
                                <div className="space-y-2">
                                    <label className="text-xs font-bold uppercase tracking-widest text-muted">Monthly AI/API Requests</label>
                                    <input 
                                        type="number" 
                                        value={apiUsage} 
                                        onChange={(e) => setApiUsage(Number(e.target.value))} 
                                        className="w-full bg-secondary border border-color p-4 rounded-xl outline-none focus:border-black transition-colors"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-xs font-bold uppercase tracking-widest text-muted">PII Exposure Probability (%)</label>
                                    <input 
                                        type="number" 
                                        value={piiPercent} 
                                        onChange={(e) => setPiiPercent(Number(e.target.value))} 
                                        className="w-full bg-secondary border border-color p-4 rounded-xl outline-none focus:border-black transition-colors"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-xs font-bold uppercase tracking-widest text-muted">Est. Risk Cost (Per Record)</label>
                                    <input 
                                        type="number" 
                                        value={riskCost} 
                                        onChange={(e) => setRiskCost(Number(e.target.value))} 
                                        className="w-full bg-secondary border border-color p-4 rounded-xl outline-none focus:border-black transition-colors"
                                    />
                                </div>
                            </div>
                            
                            <div className="grid grid-cols-2 gap-4">
                                <div className="card bg-secondary/30 p-6">
                                    <div className="text-xs font-bold text-muted uppercase mb-1">Annual Risk</div>
                                    <div className="text-2xl font-bold text-red-600">{formatCurrency(baseExposure)}</div>
                                </div>
                                <div className="card bg-secondary/30 p-6">
                                    <div className="text-xs font-bold text-muted uppercase mb-1">OCULTAR Savings</div>
                                    <div className="text-2xl font-bold text-green-600">{formatCurrency(ocultarSavings)}</div>
                                </div>
                            </div>
                        </div>
                        
                        <div className="space-y-6">
                            <div className="text-xs font-bold uppercase tracking-[0.2em] text-muted mb-4">Real-Time Verification</div>
                            <ROIDashboardCard />
                        </div>
                    </div>
                </div>
            </section>

            {/* Final CTA */}
            <section className="section-padding bg-black text-white">
                <div className="max-container text-center space-y-8">
                    <h2 className="text-white">Ready for Technical Sovereignty?</h2>
                    <p className="text-lg opacity-80 max-w-2xl mx-auto">
                        Deploy Ocultar in your infrastructure in minutes. Download our open-core binary or request a guided enterprise pilot.
                    </p>
                    <div className="flex flex-col md:flex-row gap-4 justify-center pt-8">
                        <button className="btn bg-white text-black px-12 py-5 text-lg font-bold">
                            Initialize Deployment
                        </button>
                        <Link to="/risk-assessment" className="btn btn-secondary border-white/20 hover:border-white text-white px-12 py-5 text-lg">
                            Run Pilot Test
                        </Link>
                    </div>
                </div>
            </section>
        </div>
    );
}
