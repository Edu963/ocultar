import { useState } from 'react';
import { 
  Shield, 
  Cloud, 
  Server, 
  Zap, 
  TrendingDown, 
  CheckCircle2,
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
import shutterDark from './assets/images/logoshutterdark.png';

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
        name: "Google Cloud", 
        service: "DLP API",
        processing: 5.00, 
        egress: 0.10,
        id: "GCP_DLP_V3",
        color: "blue"
    },
    aws: { 
        name: "AWS", 
        service: "Comprehend",
        processing: 1000.00, 
        egress: 0.10,
        id: "AWS_COMP_V2",
        color: "orange"
    },
    azure: { 
        name: "Azure", 
        service: "AI Language",
        processing: 1.50, 
        egress: 0.10,
        id: "AZURE_AI_V1",
        color: "sky"
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
    <div className="min-h-screen bg-[#050505] text-slate-300 font-sans selection:bg-emerald-500/30 selection:text-emerald-400 overflow-x-hidden">
      
      {/* Sticky Navbar */}
      <nav className="fixed top-0 left-0 right-0 z-50 bg-[#050505]/80 backdrop-blur-xl border-b border-white/5 py-4">
        <div className="max-w-7xl mx-auto px-6 flex justify-between items-center">
          <Link to="/" className="flex items-center gap-3 group">
            <div className="relative">
              <img src={logo3} alt="OCULTAR" className="h-8 w-auto object-contain brightness-125 transition-transform group-hover:scale-105" />
            </div>
            <div className="h-4 w-px bg-white/10 hidden sm:block"></div>
            <span className="hidden sm:block font-bold tracking-tight text-white/90 text-sm uppercase">ROI Calculator</span>
          </Link>
          
          <div className="flex items-center gap-6">
             <Link to="/" className="text-xs font-medium text-slate-400 hover:text-white transition-colors">Platform</Link>
             <a href="https://github.com/Edu963/ocultar" target="_blank" rel="noreferrer" className="text-slate-400 hover:text-white transition-colors">
                <Github size={18} />
             </a>
             <a href="mailto:sales@ocultar.io" className="px-4 py-2 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-[10px] font-bold uppercase tracking-widest hover:bg-emerald-500 hover:text-black transition-all rounded-sm">
                Start Pilot
             </a>
          </div>
        </div>
      </nav>

      {/* Hero Section - The North Star */}
      <section className="pt-32 pb-16 relative overflow-hidden">
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full bg-[radial-gradient(circle_at_center,rgba(16,185,129,0.05)_0%,transparent_70%)] pointer-events-none"></div>
        <div className="max-w-6xl mx-auto px-6 relative z-10 text-center">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 mb-6 group cursor-default">
                <TrendingDown className="w-3.5 h-3.5 text-emerald-400 group-hover:animate-bounce" />
                <span className="text-[10px] font-bold uppercase tracking-[0.2em] text-emerald-400">Projected Performance Variance</span>
            </div>
            <h1 className="text-4xl md:text-6xl lg:text-7xl font-bold tracking-tight text-white mb-6">
                Save <span className="text-emerald-400">{formatCurrency(annualSavings)}</span> / Year
            </h1>
            <p className="text-slate-400 max-w-2xl mx-auto text-lg leading-relaxed">
                Compute the financial arbitrage between external AI pipeline tax and OCULTAR's zero-egress local processing framework.
            </p>
        </div>
      </section>

      <div className="max-w-7xl mx-auto px-6 pb-32">
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
            
            {/* Left: Input Console */}
            <div className="lg:col-span-4 sticky top-24 space-y-6">
                <div className="glass-premium p-8 rounded-xl border-white/5 space-y-8">
                    <div className="flex items-center gap-3 pb-4 border-b border-white/5">
                        <BarChart3 className="w-4 h-4 text-emerald-500" />
                        <h3 className="text-sm font-bold uppercase tracking-wider text-white">Input Console</h3>
                    </div>

                    {/* Provider Selection */}
                    <div className="space-y-4">
                        <label className="text-[10px] uppercase tracking-widest font-bold text-slate-500">Target Ecosystem</label>
                        <div className="grid grid-cols-1 gap-2">
                            {Object.entries(providers).map(([key, p]) => (
                                <button
                                    key={key}
                                    onClick={() => setProvider(key)}
                                    className={`flex items-center justify-between p-4 border transition-all rounded-lg group ${
                                        provider === key 
                                        ? 'bg-emerald-500/10 border-emerald-500/40 text-white' 
                                        : 'bg-white/[0.02] border-white/5 text-slate-500 hover:border-white/20 hover:text-slate-300'
                                    }`}
                                >
                                    <div className="flex items-center gap-4">
                                        <div className={`p-2 rounded-md ${provider === key ? 'bg-emerald-500/20' : 'bg-white/5'}`}>
                                            <Cloud className={`w-4 h-4 ${provider === key ? 'text-emerald-400' : 'text-slate-500'}`} />
                                        </div>
                                        <div className="text-left">
                                            <div className="text-sm font-bold tracking-tight">{p.name}</div>
                                            <div className="text-[10px] opacity-60 uppercase tracking-tighter">{p.service}</div>
                                        </div>
                                    </div>
                                    <div className={`w-2 h-2 rounded-full ${provider === key ? 'bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]' : 'bg-white/10'}`}></div>
                                </button>
                            ))}
                        </div>
                    </div>

                    {/* Volume Slider */}
                    <div className="space-y-6 pt-4 border-t border-white/5">
                        <div className="flex justify-between items-end">
                            <label className="text-[10px] uppercase tracking-widest font-bold text-slate-500">Monthly Volume</label>
                            <span className="text-xl font-bold text-white leading-none">{volume} <span className="text-xs text-slate-500 font-medium">TB</span></span>
                        </div>
                        <input 
                            type="range" 
                            min="1" 
                            max="500" 
                            value={volume} 
                            onChange={(e) => setVolume(parseInt(e.target.value))}
                            className="w-full h-1.5 bg-white/10 rounded-lg appearance-none cursor-pointer accent-emerald-500"
                        />
                        <div className="flex justify-between text-[9px] text-slate-600 font-bold uppercase tracking-widest">
                            <span>1 TB</span>
                            <span>500 TB</span>
                        </div>
                    </div>

                    {/* EDP Discount Slider */}
                    <div className="space-y-6 pt-4 border-t border-white/5">
                        <div className="flex justify-between items-end">
                            <label className="text-[10px] uppercase tracking-widest font-bold text-slate-500">Cloud Discount (EDP)</label>
                            <span className="text-xl font-bold text-white leading-none">{discount}%</span>
                        </div>
                        <input 
                            type="range" 
                            min="0" 
                            max="80" 
                            step="5"
                            value={discount} 
                            onChange={(e) => setDiscount(parseInt(e.target.value))}
                            className="w-full h-1.5 bg-white/10 rounded-lg appearance-none cursor-pointer accent-emerald-500"
                        />
                        <div className="flex justify-between text-[9px] text-slate-600 font-bold uppercase tracking-widest">
                            <span>Base Price</span>
                            <span>Max Savings</span>
                        </div>
                    </div>

                    {/* Effective Rate Card */}
                    <div className="p-4 rounded-xl bg-black/40 border border-white/5 space-y-3">
                        <div className="flex justify-between text-[10px] font-bold uppercase tracking-widest text-slate-500">
                            <span>List Rate</span>
                            <span className="line-through">{formatRate(listPricePerGB)} / GB</span>
                        </div>
                        <div className="flex justify-between items-center pt-2 border-t border-white/5">
                            <span className="text-xs font-bold text-white uppercase tracking-wider">Effective Rate</span>
                            <span className="text-emerald-400 font-bold">{formatRate(effectiveRate)} / GB</span>
                        </div>
                    </div>
                </div>
            </div>

            {/* Right: Analysis Dashboard */}
            <div className="lg:col-span-8 space-y-8">
                
                {/* Comparison Card Grid */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    
                    {/* Cloud Node */}
                    <div className="glass border-white/5 p-8 rounded-2xl relative overflow-hidden group">
                        <div className="absolute top-0 right-0 p-6 opacity-[0.03] group-hover:opacity-[0.05] transition-opacity">
                            <Cloud className="w-24 h-24" />
                        </div>
                        <div className="space-y-6 relative z-10">
                            <div className="flex items-center justify-between">
                                <span className="px-3 py-1 rounded-full bg-rose-500/10 border border-rose-500/20 text-[10px] font-bold text-rose-400 uppercase tracking-widest italic">External Leakage</span>
                                <span className="text-xs font-mono text-slate-500 tracking-tighter">Egress Tax Active</span>
                            </div>
                            <div>
                                <h4 className="text-lg font-bold text-white mb-1">{selectedProvider.name} Cost</h4>
                                <p className="text-xs text-slate-500">Standard PII Inspection + Transit</p>
                            </div>
                            <div className="py-4 border-y border-white/5 space-y-2">
                                <div className="flex justify-between text-sm">
                                    <span className="text-slate-500">Gross Monthly</span>
                                    <span className="text-slate-400">{formatCurrency(totalCloudGrossMonthly)}</span>
                                </div>
                                <div className="flex justify-between text-sm">
                                    <span className="text-slate-500">EDP Credit (Net)</span>
                                    <span className="text-rose-400">-{formatCurrency(discountValueMonthly)}</span>
                                </div>
                            </div>
                            <div className="pt-2">
                                <div className="text-[10px] font-bold uppercase tracking-widest text-slate-500 mb-1">Monthly Financial Drain</div>
                                <div className="text-4xl font-bold text-rose-500 tracking-tight">{formatCurrency(totalCloudNetMonthly)}</div>
                            </div>
                        </div>
                    </div>

                    {/* Ocultar Node */}
                    <div className="glass-premium border-emerald-500/20 p-8 rounded-2xl relative overflow-hidden group">
                        <div className="absolute top-0 right-0 p-6 opacity-[0.05] group-hover:opacity-[0.1] transition-opacity">
                            <Shield className="w-24 h-24 text-emerald-500" />
                        </div>
                        <div className="space-y-6 relative z-10">
                            <div className="flex items-center justify-between">
                                <span className="px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-[10px] font-bold text-emerald-400 uppercase tracking-widest italic">Zero-Egress Efficiency</span>
                                <span className="text-xs font-mono text-emerald-500 tracking-tighter">Secure Local Node</span>
                            </div>
                            <div>
                                <h4 className="text-lg font-bold text-white mb-1">OCULTAR Deployment</h4>
                                <p className="text-xs text-slate-500">In-VPC Privacy Layer (Unlimited)</p>
                            </div>
                            <div className="py-4 border-y border-white/5 space-y-2">
                                <div className="flex justify-between text-sm">
                                    <span className="text-slate-500">Software License</span>
                                    <span className="text-emerald-400 font-bold">$10,000</span>
                                </div>
                                <div className="flex justify-between text-sm">
                                    <span className="text-slate-500">Compute Overhead</span>
                                    <span className="text-emerald-400/60">+{formatCurrency(localComputeMonthly)}</span>
                                </div>
                            </div>
                            <div className="pt-2">
                                <div className="text-[10px] font-bold uppercase tracking-widest text-emerald-500 mb-1">Predictable Operating Expense</div>
                                <div className="text-4xl font-bold text-white tracking-tight">{formatCurrency(totalOcultarMonthly)}</div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Performance Metrics Row */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="bg-white/[0.02] border border-white/5 p-6 rounded-xl">
                        <div className="flex items-center gap-3 mb-4">
                            <Zap className="w-4 h-4 text-emerald-500" />
                            <span className="text-[10px] font-bold uppercase tracking-widest text-slate-500">P99 Latency</span>
                        </div>
                        <div className="flex items-baseline gap-2">
                            <span className="text-2xl font-bold text-white">0.92ms</span>
                            <span className="text-[10px] text-emerald-500 font-bold font-mono">(-99.8%)</span>
                        </div>
                    </div>
                    <div className="bg-white/[0.02] border border-white/5 p-6 rounded-xl">
                        <div className="flex items-center gap-3 mb-4">
                            <Lock className="w-4 h-4 text-emerald-500" />
                            <span className="text-[10px] font-bold uppercase tracking-widest text-slate-500">Risk Exposure</span>
                        </div>
                        <div className="flex items-baseline gap-2">
                            <span className="text-2xl font-bold text-white">Zero</span>
                            <span className="text-[10px] text-emerald-500 font-bold font-mono">PROTECTED</span>
                        </div>
                    </div>
                    <div className="bg-white/[0.02] border border-white/5 p-6 rounded-xl">
                        <div className="flex items-center gap-3 mb-4">
                            <Server className="w-4 h-4 text-emerald-500" />
                            <span className="text-[10px] font-bold uppercase tracking-widest text-slate-500">Data Egress</span>
                        </div>
                        <div className="flex items-baseline gap-2">
                            <span className="text-2xl font-bold text-white">0 GB</span>
                            <span className="text-[10px] text-emerald-500 font-bold font-mono">STAY_IN_VPC</span>
                        </div>
                    </div>
                </div>

                {/* Main CTA Section */}
                <div className="glass-premium p-12 rounded-3xl text-center space-y-8 relative overflow-hidden">
                    <div className="absolute bottom-0 left-0 w-full h-1 bg-gradient-to-r from-transparent via-emerald-500/50 to-transparent"></div>
                    <div className="relative z-10 max-w-2xl mx-auto space-y-6">
                        <h2 className="text-3xl font-bold text-white tracking-tight">Ready to terminate egress taxation?</h2>
                        <p className="text-slate-400 leading-relaxed">
                            Join security-first data teams optimizing AI pipelines with OCULTAR. Deploy a private, high-performance redaction layer in your own cloud infrastructure today.
                        </p>
                        <div className="flex flex-col sm:flex-row items-center justify-center gap-4 pt-4">
                            <a href="mailto:sales@ocultar.io?subject=Enterprise%20Pilot%20Request" className="w-full sm:w-auto px-8 py-4 bg-emerald-500 text-black font-bold text-sm rounded-lg hover:bg-emerald-400 transition-all flex items-center justify-center gap-2 group shadow-lg shadow-emerald-500/20">
                                Start Your Pilot
                                <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
                            </a>
                            <button onClick={() => window.open('https://github.com/Edu963/ocultar', '_blank')} className="w-full sm:w-auto px-8 py-4 bg-white/5 border border-white/10 text-white font-bold text-sm rounded-lg hover:bg-white/10 transition-all">
                                Run Security Audit
                            </button>
                            <Link to="/" className="w-full sm:w-auto px-8 py-4 text-slate-400 hover:text-white transition-all text-sm font-medium">
                                View Architecture
                            </Link>
                        </div>
                    </div>
                </div>

                {/* Collapsible Evidence Room */}
                <div className="border border-white/5 rounded-2xl overflow-hidden bg-black/40">
                    <button 
                        onClick={() => setShowCitations(!showCitations)}
                        className="w-full px-8 py-6 flex items-center justify-between group hover:bg-white/[0.02] transition-colors"
                    >
                        <div className="flex items-center gap-4">
                            <div className="p-2 rounded-full bg-white/5 group-hover:bg-emerald-500/10 transition-colors">
                                <Info className={`w-4 h-4 transition-colors ${showCitations ? 'text-emerald-400' : 'text-slate-500'}`} />
                            </div>
                            <div className="text-left">
                                <h4 className="text-sm font-bold text-white uppercase tracking-widest">Data Sources & Assumptions</h4>
                                <p className="text-xs text-slate-500">Methodology and Pricing References</p>
                            </div>
                        </div>
                        <ChevronDown className={`w-5 h-5 text-slate-500 transition-transform duration-300 ${showCitations ? 'rotate-180 text-white' : ''}`} />
                    </button>
                    
                    {showCitations && (
                        <div className="px-8 pb-10 pt-2 border-t border-white/5 animate-fade-in-up">
                            <div className="grid grid-cols-1 md:grid-cols-3 gap-10 text-[11px] leading-relaxed text-slate-500">
                                <div className="space-y-4">
                                    <div className="flex items-center gap-2 text-white/80 font-bold uppercase tracking-widest pb-2 border-b border-white/5">
                                        <div className="w-1.5 h-1.5 rounded-full bg-emerald-500"></div>
                                        API Synthetics
                                    </div>
                                    <p><strong className="text-slate-400">AWS Comprehend:</strong> $0.0001 per unit (100 chars). Typical PII workload averages ~$1,000/GB processed. <a href="https://aws.amazon.com/comprehend/pricing/" target="_blank" rel="noreferrer" className="text-emerald-500/70 hover:text-emerald-400 underline inline-flex items-center gap-1">Source <ExternalLink size={10}/></a></p>
                                    <p><strong className="text-slate-400">GCP DLP:</strong> $3.00/GB inspection + $2.00/GB transformation. Total list price ~$5.00/GB. <a href="https://cloud.google.com/sensitive-data-protection/pricing" target="_blank" rel="noreferrer" className="text-emerald-500/70 hover:text-emerald-400 underline inline-flex items-center gap-1">Source <ExternalLink size={10}/></a></p>
                                </div>
                                <div className="space-y-4">
                                    <div className="flex items-center gap-2 text-white/80 font-bold uppercase tracking-widest pb-2 border-b border-white/5">
                                        <div className="w-1.5 h-1.5 rounded-full bg-emerald-500"></div>
                                        Network Layers
                                    </div>
                                    <p><strong className="text-slate-400">Azure Language:</strong> Median sentence analysis pricing calculated at ~$1.50/GB processing throughput. <a href="https://azure.microsoft.com/en-us/pricing/details/cognitive-services/language-service/" target="_blank" rel="noreferrer" className="text-emerald-500/70 hover:text-emerald-400 underline inline-flex items-center gap-1">Source <ExternalLink size={10}/></a></p>
                                    <p><strong className="text-slate-400">Egress Taxation:</strong> Global cloud averages for cross-region/internet egress established at $0.08 - $0.12 per GB.</p>
                                </div>
                                <div className="space-y-4">
                                    <div className="flex items-center gap-2 text-white/80 font-bold uppercase tracking-widest pb-2 border-b border-white/5">
                                        <div className="w-1.5 h-1.5 rounded-full bg-emerald-500"></div>
                                        Security Risk
                                    </div>
                                    <p><strong className="text-slate-400">Incident Impact:</strong> $160 average cost per PII record leached. Median data breach cost at $4.45M globally. <a href="https://www.ibm.com/reports/data-breach" target="_blank" rel="noreferrer" className="text-emerald-500/70 hover:text-emerald-400 underline inline-flex items-center gap-1">IBM 2023 Report <ExternalLink size={10}/></a></p>
                                    <p><strong className="text-slate-400">Methodology:</strong> 1 TB = 1,000 GB. Calculations assume standard un-discounted API tiers prior to EDP coefficient application.</p>
                                </div>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
      </div>

      {/* Background Decorative Element */}
      <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 opacity-[0.02] pointer-events-none z-[-1]">
        <img src={shutterDark} alt="" className="w-[1200px] h-auto" />
      </div>
    </div>
  );
};

export default CalculatorPage;
