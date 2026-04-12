import { useState, useEffect, useRef } from 'react';
import { 
  Shield, 
  Cloud, 
  Server, 
  Zap, 
  TrendingDown, 
  Github,
  ChevronRight,
  Info,
  ArrowRight,
  ExternalLink,
  Lock,
  BarChart3,
  ChevronDown
} from 'lucide-react';
import { Link } from 'react-router-dom';
import logo3 from './assets/images/logo3.jpg';

const ParticleCanvas = () => {
  const canvasRef = useRef(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    let width, height;
    let particles = [];
    const PARTICLE_COUNT = 60;
    const CONNECTION_DISTANCE = 150;

    const initCanvas = () => {
      width = canvas.width = window.innerWidth;
      height = canvas.height = window.innerHeight;
      particles = [];
      for (let i = 0; i < PARTICLE_COUNT; i++) {
        particles.push({
          x: Math.random() * width,
          y: Math.random() * height,
          vx: (Math.random() - 0.5) * 0.5,
          vy: (Math.random() - 0.5) * 0.5
        });
      }
    };

    const animate = () => {
      if (!canvasRef.current) return;
      ctx.clearRect(0, 0, width, height);
      for (let i = 0; i < particles.length; i++) {
        const p = particles[i];
        p.x += p.vx;
        p.y += p.vy;

        if (p.x < 0 || p.x > width) p.vx *= -1;
        if (p.y < 0 || p.y > height) p.vy *= -1;

        ctx.fillStyle = 'rgba(255, 255, 255, 0.15)';
        ctx.beginPath();
        ctx.arc(p.x, p.y, 1, 0, Math.PI * 2);
        ctx.fill();

        for (let j = i + 1; j < particles.length; j++) {
          const p2 = particles[j];
          const dx = p.x - p2.x;
          const dy = p.y - p2.y;
          const dist = Math.sqrt(dx * dx + dy * dy);

          if (dist < CONNECTION_DISTANCE) {
            ctx.strokeStyle = `rgba(255, 255, 255, ${0.08 * (1 - dist / CONNECTION_DISTANCE)})`;
            ctx.lineWidth = 0.5;
            ctx.beginPath();
            ctx.moveTo(p.x, p.y);
            ctx.lineTo(p2.x, p2.y);
            ctx.stroke();
          }
        }
      }
      requestAnimationFrame(animate);
    };

    window.addEventListener('resize', initCanvas);
    initCanvas();
    animate();

    return () => window.removeEventListener('resize', initCanvas);
  }, []);

  return <canvas ref={canvasRef} className="fixed top-0 left-0 w-full h-full pointer-events-none z-0 opacity-40" />;
};

const CalculatorPage = () => {
  const [provider, setProvider] = useState('gcp');
  const [volume, setVolume] = useState(10); // TB
  const [discount, setDiscount] = useState(0); // %
  const [showCitations, setShowCitations] = useState(false);

  const GB_PER_TB = 1000;
  const OCULTAR_MONTHLY = 10000;
  const LOCAL_COMPUTE_PER_TB = 20;

  const providers = {
    gcp: { 
        name: "Google Cloud DLP", 
        processing: 5.00, 
        egress: 0.10,
        id: "GCP_DLP_V3"
    },
    aws: { 
        name: "AWS Comprehend", 
        processing: 1000.00, 
        egress: 0.10,
        id: "AWS_COMP_V2"
    },
    azure: { 
        name: "Azure AI Language", 
        processing: 1.50, 
        egress: 0.10,
        id: "AZURE_AI_V1"
    }
  };

  const selectedProvider = providers[provider];
  const listPricePerGB = selectedProvider.processing + selectedProvider.egress;
  const totalGB = volume * GB_PER_TB;
  
  const totalCloudGrossMonthly = totalGB * listPricePerGB;
  const discountMultiplier = discount / 100;
  const discountValueMonthly = totalCloudGrossMonthly * discountMultiplier;
  const totalCloudNetMonthly = totalCloudGrossMonthly - discountValueMonthly;
  
  const effectiveRate = listPricePerGB * (1 - discountMultiplier);

  const localComputeMonthly = volume * LOCAL_COMPUTE_PER_TB;
  const totalOcultarMonthly = OCULTAR_MONTHLY + localComputeMonthly;

  const totalCloudAnnual = totalCloudNetMonthly * 12;
  const totalOcultarAnnual = totalOcultarMonthly * 12;
  const annualSavings = Math.max(0, totalCloudAnnual - totalOcultarAnnual);

  const formatCurrency = (val) => 
    new Intl.NumberFormat('en-US', { 
      style: 'currency', 
      currency: 'USD', 
      maximumFractionDigits: 0 
    }).format(val);

  const formatRate = (val) => 
    new Intl.NumberFormat('en-US', { 
      style: 'currency', 
      currency: 'USD', 
      minimumFractionDigits: 2, 
      maximumFractionDigits: 2 
    }).format(val);

  return (
    <div className="min-h-screen bg-[#050505] text-[#FAFAFA] font-sans selection:bg-white selection:text-black overflow-x-hidden">
      <ParticleCanvas />

      <div className="relative z-10 max-w-6xl mx-auto p-6 md:p-12">
        
        {/* Boxed Header */}
        <header className="mb-12 border border-white/20 p-8 flex flex-col md:flex-row justify-between items-start gap-8 bg-black/40 backdrop-blur-md">
            <div className="max-w-2xl order-2 md:order-1">
                <div className="flex items-baseline gap-3 mb-4">
                    <h1 className="text-3xl md:text-5xl font-bold tracking-tight uppercase text-white">Data Refinery</h1>
                    <span className="font-mono text-white/40 text-sm">v_1.0.0</span>
                </div>
                <div className="flex items-center gap-3 mb-6">
                    <h2 className="font-mono text-xs uppercase tracking-[0.3em] border-b border-emerald-500 pb-1 text-emerald-500 font-bold">ROI Calculator Module</h2>
                </div>
                <p className="font-mono text-sm text-white/60 leading-relaxed max-w-xl">
                    <span className="text-emerald-500">[SYS_MSG]:</span> Compute the mathematical variance between external AI pipeline taxation and OCULTAR's zero-egress local processing framework.
                </p>
            </div>
            <div className="flex-shrink-0 w-full md:w-auto flex justify-start md:justify-end order-1 md:order-2">
                <img src={logo3} alt="OCULTAR" className="h-10 md:h-12 w-auto object-contain brightness-110 grayscale opacity-80" />
            </div>
        </header>

        {/* Main Grid: 5 / 7 Split */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
            
            {/* Left Col: Parameters (Boxed) */}
            <div className="lg:col-span-5 bg-black/60 backdrop-blur-md border border-white/20 p-8 space-y-10">
                <div className="flex items-center gap-3 mb-4 border-b border-white/20 pb-4">
                    <div className="w-2 h-2 bg-emerald-500"></div>
                    <h3 className="font-mono text-xs font-bold uppercase tracking-[0.3em] text-white">Input Parameters</h3>
                </div>
                
                {/* Cloud Provider Selection */}
                <div className="space-y-4">
                    <label className="block font-mono text-[10px] uppercase text-white/40 tracking-widest font-bold">Target Cloud Architecture</label>
                    <select 
                        value={provider}
                        onChange={(e) => setProvider(e.target.value)}
                        className="w-full bg-transparent border border-white/20 text-white font-mono text-sm rounded-none p-4 focus:outline-none focus:border-emerald-500 cursor-pointer appearance-none"
                        style={{
                            backgroundImage: `url("data:image/svg+xml;charset=UTF-8,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='white' stroke-width='2' stroke-linecap='square' stroke-linejoin='miter'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E")`,
                            backgroundRepeat: 'no-repeat',
                            backgroundPosition: 'right 1rem center',
                            backgroundSize: '1em'
                        }}
                    >
                        <option value="gcp">Google Cloud DLP ($5.00 / GB)</option>
                        <option value="aws">AWS Comprehend (~$1,000 / GB)</option>
                        <option value="azure">Azure AI Language (~$1.50 / GB)</option>
                    </select>
                </div>

                {/* Volume Slider */}
                <div className="space-y-6">
                    <div className="flex justify-between items-end mb-2 border-b border-white/20 pb-2">
                        <span className="font-mono text-2xl font-bold text-white">{volume} TB</span>
                        <span className="font-mono text-[10px] text-white/40 uppercase tracking-widest pb-1">Data Throughput</span>
                    </div>
                    <input 
                        type="range" 
                        min="1" 
                        max="250" 
                        value={volume} 
                        onChange={(e) => setVolume(parseInt(e.target.value))}
                        className="slider-boxed"
                    />
                    <div className="flex justify-between font-mono text-[10px] text-white/40 uppercase tracking-widest">
                        <span>1 TB</span>
                        <span>250 TB</span>
                    </div>
                </div>

                {/* Discount Slider */}
                <div className="space-y-6">
                    <div className="flex justify-between items-end mb-2 border-b border-white/20 pb-2">
                        <span className="font-mono text-2xl font-bold text-white">{discount}%</span>
                        <span className="font-mono text-[10px] text-white/40 uppercase tracking-widest pb-1">Negotiated EDP Discount</span>
                    </div>
                    <input 
                        type="range" 
                        min="0" 
                        max="80" 
                        step="5"
                        value={discount} 
                        onChange={(e) => setDiscount(parseInt(e.target.value))}
                        className="slider-boxed"
                    />
                    <div className="flex justify-between font-mono text-[10px] text-white/40 uppercase tracking-widest">
                        <span>BASE [0%]</span>
                        <span>MAX [80%]</span>
                    </div>
                </div>

                {/* Effective Rate Matrix */}
                <div className="font-mono text-[10px] p-6 border border-white/20 bg-white/[0.02] space-y-4">
                    <h4 className="font-bold uppercase tracking-[0.2em] text-emerald-500 border-b border-white/10 pb-2 mb-2">Rate Matrix</h4>
                    <div className="flex justify-between items-center">
                        <span className="text-white/40 uppercase">List Price</span>
                        <span className="line-through text-white/40 italic">{formatRate(listPricePerGB)} / GB</span>
                    </div>
                    <div className="flex justify-between items-center pt-2 border-t border-white/10">
                        <span className="font-bold text-white uppercase tracking-wider">Effective Rate</span>
                        <span className="font-bold text-emerald-400">{formatRate(effectiveRate)} / GB</span>
                    </div>
                </div>
            </div>

            {/* Right Col: Computation Results */}
            <div className="lg:col-span-7 grid grid-cols-1 md:grid-cols-2 gap-8">
                
                {/* External Pipeline Node */}
                <div className="border border-white/10 bg-black/40 backdrop-blur-md p-8 flex flex-col justify-between">
                    <div>
                        <div className="flex items-center justify-between mb-8 border-b border-white/20 pb-4">
                            <h3 className="text-xl font-bold uppercase text-white tracking-tight">{selectedProvider.name}</h3>
                            <div className="font-mono text-[9px] border border-white/20 px-2 py-1 text-white/40 uppercase tracking-widest">API_EGRESS</div>
                        </div>
                        
                        <div className="font-mono text-sm space-y-4 mb-10">
                            <div className="flex justify-between border-b border-white/5 pb-2">
                                <span className="text-white/40">Gross Monthly</span>
                                <span className="text-white/60 line-through italic">{formatCurrency(totalCloudGrossMonthly)}</span>
                            </div>
                            <div className="flex justify-between border-b border-white/5 pb-2">
                                <span className="text-white/40">Discount Factor</span>
                                <span className="text-rose-400">-{formatCurrency(discountValueMonthly)}</span>
                            </div>
                            <div className="flex justify-between pt-2">
                                <span className="text-white/40">Network Latency</span>
                                <span className="text-white/60">100-500ms</span>
                            </div>
                        </div>
                    </div>
                    
                    <div className="pt-8 border-t border-white/20">
                        <span className="block font-mono text-[10px] text-white/40 uppercase tracking-widest mb-2 font-bold">Net Financial Drain (Mo)</span>
                        <span className="text-4xl font-bold tracking-tighter text-rose-500">{formatCurrency(totalCloudNetMonthly)}</span>
                    </div>
                </div>

                {/* OCULTAR Node (Solid Contrast) */}
                <div className="border border-emerald-500/30 bg-black p-8 flex flex-col justify-between relative overflow-hidden group">
                    <div className="absolute top-0 left-0 w-1 h-full bg-emerald-500"></div>
                    <div>
                        <div className="flex items-center justify-between mb-8 border-b border-white/20 pb-4">
                            <h3 className="text-xl font-bold uppercase text-white tracking-tight">Zero-Egress SLM</h3>
                            <div className="font-mono text-[9px] bg-emerald-500 px-2 py-1 text-black uppercase tracking-widest font-bold">LOCAL_NODE</div>
                        </div>
                        
                        <div className="font-mono text-sm space-y-4 mb-10">
                            <div className="flex justify-between border-b border-white/5 pb-2">
                                <span className="text-white/40">Software License</span>
                                <span className="text-emerald-400 font-bold">$10,000 / mo</span>
                            </div>
                            <div className="flex justify-between border-b border-white/5 pb-2">
                                <span className="text-white/40">Hardware Dep.</span>
                                <span className="text-emerald-400/60">+{formatCurrency(localComputeMonthly)}</span>
                            </div>
                            <div className="flex justify-between pt-2">
                                <span className="text-white/40">P50 Latency</span>
                                <span className="text-white font-bold">0.92ms</span>
                            </div>
                        </div>
                    </div>
                    
                    <div className="pt-8 border-t border-white/20">
                        <span className="block font-mono text-[10px] text-white/40 uppercase tracking-widest mb-2 font-bold">Fixed Resource Cost (Mo)</span>
                        <span className="text-4xl font-bold tracking-tighter text-white">{formatCurrency(totalOcultarMonthly)}</span>
                    </div>
                </div>

                {/* Total System Variance Banner */}
                <div className="md:col-span-2 border border-emerald-500/20 bg-emerald-500/[0.02] p-8 flex flex-col sm:flex-row items-center justify-between group hover:border-emerald-500/40 transition-colors">
                    <div>
                        <h4 className="text-2xl font-bold uppercase text-white tracking-tight mb-2">Total System Variance</h4>
                        <p className="font-mono text-[10px] text-emerald-500 uppercase tracking-[0.2em] font-bold">Projected 12-Month Capital Retention</p>
                    </div>
                    <div className="text-right">
                        <span className="block text-6xl font-bold tracking-tighter text-emerald-400 drop-shadow-[0_0_15px_rgba(16,185,129,0.3)]">
                            {formatCurrency(annualSavings)}
                        </span>
                    </div>
                </div>

                {/* CTA Upgrade Path */}
                <div className="md:col-span-2 mt-4 text-center border border-white/20 bg-white/[0.02] p-12 space-y-8">
                    <h2 className="text-2xl font-bold uppercase text-white tracking-tight">Execute Architecture Upgrade</h2>
                    <p className="font-mono text-sm text-white/40 max-w-2xl mx-auto leading-relaxed">
                        Initiate deployment sequence. Neutralize regulatory risk and terminate external egress taxation via localized execution.
                    </p>
                    
                    <div className="flex flex-col sm:flex-row items-center justify-center gap-6 font-mono text-[10px] uppercase font-bold tracking-widest">
                        <a href="mailto:sales@ocultar.io?subject=Enterprise%20Pilot%20Request" className="w-full sm:w-auto px-10 py-4 bg-emerald-500 text-black hover:bg-white transition-all flex items-center justify-center gap-2">
                            [ Start Pilot ]
                            <ArrowRight className="w-4 h-4" />
                        </a>
                        <button onClick={() => window.open('https://github.com/Edu963/ocultar', '_blank')} className="w-full sm:w-auto px-10 py-4 border border-white/20 text-white hover:bg-white hover:text-black transition-all">
                            [ Audit Source Code ]
                        </button>
                    </div>
                </div>

                {/* Telemetry Citations (Collapsible) */}
                <div className="md:col-span-2">
                    <button 
                        onClick={() => setShowCitations(!showCitations)}
                        className="w-full px-8 py-6 flex items-center justify-between border border-white/10 bg-black/20 hover:bg-white/[0.02] transition-colors"
                    >
                        <div className="flex items-center gap-4">
                            <Info className={`w-4 h-4 ${showCitations ? 'text-emerald-400' : 'text-white/40'}`} />
                            <h4 className="text-xs font-bold text-white uppercase tracking-[0.2em]">Telemetry & Data Sources</h4>
                        </div>
                        <ChevronDown className={`w-4 h-4 text-white/40 transition-transform ${showCitations ? 'rotate-180 text-white' : ''}`} />
                    </button>
                    
                    {showCitations && (
                        <div className="p-8 border-x border-b border-white/10 grid grid-cols-1 md:grid-cols-3 gap-10 font-mono text-[10px] text-white/40 leading-relaxed animate-fade-in-up">
                            <div className="space-y-4">
                                <p><strong className="text-white/60 uppercase tracking-tighter font-bold">[01] AWS Syntax:</strong> <a href="https://aws.amazon.com/comprehend/pricing/" target="_blank" className="underline hover:text-emerald-400">Comprehend API</a> - $0.0001 per 100 character array. ~1k/GB volumetric.</p>
                                <p><strong className="text-white/60 uppercase tracking-tighter font-bold">[02] GCP Syntax:</strong> <a href="https://cloud.google.com/sensitive-data-protection/pricing" target="_blank" className="underline hover:text-emerald-400">Sensitive Data Protection API</a> - Base $3.00/GB inspect, $2.00/GB transform vector.</p>
                            </div>
                            <div className="space-y-4">
                                <p><strong className="text-white/60 uppercase tracking-tighter font-bold">[03] Azure Syntax:</strong> <a href="https://azure.microsoft.com/en-us/pricing/details/cognitive-services/language-service/" target="_blank" className="underline hover:text-emerald-400">AI Language Services</a> - Median translation ~$1.50/GB processing.</p>
                                <p><strong className="text-white/60 uppercase tracking-tighter font-bold">[04] Network Egress:</strong> Standard topology dictates $0.08 - $0.12/GB payload transit tax.</p>
                            </div>
                            <div className="space-y-4">
                                <p><strong className="text-white/60 uppercase tracking-tighter font-bold">[05] Local Compute:</strong> Amortized generic Edge/CPU nodes via Q4_K_M GGUF.</p>
                                <p><strong className="text-white/60 uppercase tracking-tighter font-bold">[06] Risk Vectors:</strong> <a href="https://www.ibm.com/reports/data-breach" target="_blank" className="underline hover:text-emerald-400">IBM Ponemon Report</a> - $10.22M median incident cost. PII at $160/record.</p>
                            </div>
                        </div>
                    )}
                </div>

            </div>
        </div>

        {/* Home/Back Link */}
        <div className="mt-12 text-center pb-12">
            <Link to="/" className="inline-flex items-center gap-2 text-white/40 hover:text-white transition-colors text-[10px] font-bold uppercase tracking-widest border-b border-white/20 pb-1">
                <ChevronRight className="w-3 h-3 rotate-180" /> Back to Terminal
            </Link>
        </div>
      </div>
    </div>
  );
};

export default CalculatorPage;
