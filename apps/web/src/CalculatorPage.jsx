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
  ArrowRight
} from 'lucide-react';
import { Link } from 'react-router-dom';
import logo3 from './assets/images/logo3.jpg';
import shutterDark from './assets/images/logoshutterdark.png';

const CalculatorPage = () => {
  const [provider, setProvider] = useState('gcp');
  const [volume, setVolume] = useState(10); // TB
  const [discount, setDiscount] = useState(0); // %

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
    <div className="min-h-screen bg-[#050505] text-slate-300 font-sans selection:bg-emerald-500/30 selection:text-emerald-400 overflow-x-hidden pb-20 pt-12">
      {/* Background Decorative Element */}
      <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 opacity-[0.03] pointer-events-none z-0">
        <img src={shutterDark} alt="" className="w-[800px] h-auto animate-pulse-slow" />
      </div>

      <div className="max-w-6xl mx-auto px-6 relative z-10">
        
        {/* Header Alignment with Reference */}
        <header className="mb-16 border-b border-white/10 pb-10 flex flex-col md:flex-row justify-between items-start gap-8">
            <div className="max-w-2xl order-2 md:order-1">
                <div className="flex items-baseline gap-3 mb-4">
                    <h1 className="text-3xl md:text-5xl font-bold tracking-tight uppercase text-white">Data Refinery</h1>
                    <span className="font-mono text-slate-500 text-sm">v_1.0.0</span>
                </div>
                <div className="flex items-center gap-3 mb-6">
                    <div className="h-px w-8 bg-emerald-500"></div>
                    <h2 className="font-mono text-xs uppercase tracking-[0.3em] text-emerald-500 font-bold">ROI Calculator Module</h2>
                </div>
                <p className="font-mono text-sm text-slate-400 leading-relaxed max-w-xl">
                    <span className="text-emerald-500/70">[SYS_MSG]:</span> Compute the mathematical variance between external AI pipeline taxation and OCULTAR's zero-egress local processing framework.
                </p>
            </div>
            <div className="flex-shrink-0 w-full md:w-auto flex justify-start md:justify-end order-1 md:order-2">
                <img src={logo3} alt="OCULTAR Logo" className="h-10 md:h-12 w-auto object-contain brightness-110" />
            </div>
        </header>

        {/* Main Layout Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-12 items-start">
            
            {/* Left Col: Parameters */}
            <div className="lg:col-span-5 glass border-white/10 p-10 space-y-10">
                <div className="flex items-center gap-3 mb-4 border-b border-white/10 pb-6">
                    <div className="w-2 h-2 bg-emerald-500"></div>
                    <h3 className="font-mono text-xs font-bold uppercase tracking-[0.3em] text-white">Input Parameters</h3>
                </div>
                
                {/* Cloud Provider Selection */}
                <div className="space-y-4">
                    <label className="block font-mono text-[10px] uppercase text-slate-500 tracking-widest font-bold">Target Cloud Architecture</label>
                    <div className="relative">
                        <select 
                            value={provider}
                            onChange={(e) => setProvider(e.target.value)}
                            className="w-full bg-black/50 border border-white/10 text-white font-mono text-sm rounded-none p-4 focus:outline-none focus:border-emerald-500 transition-colors appearance-none cursor-pointer"
                        >
                            <option value="gcp">Google Cloud DLP ($5.00 / GB)</option>
                            <option value="aws">AWS Comprehend (~$1,000 / GB)</option>
                            <option value="azure">Azure AI Language (~$1.50 / GB)</option>
                        </select>
                        <ChevronRight className="absolute right-4 top-1/2 -translate-y-1/2 rotate-90 w-4 h-4 text-slate-500 pointer-events-none" />
                    </div>
                </div>

                {/* Volume Slider */}
                <div className="space-y-6">
                    <div className="flex justify-between items-end mb-2 border-b border-white/10 pb-4">
                        <div className="space-y-1">
                            <span className="block font-mono text-[10px] uppercase text-slate-500 tracking-widest font-bold">Volume</span>
                            <span className="font-mono text-3xl font-bold text-white">{volume} TB</span>
                        </div>
                        <span className="font-mono text-[10px] text-slate-500 uppercase tracking-widest pb-1">Monthly Throughput</span>
                    </div>
                    <input 
                        type="range" 
                        min="1" 
                        max="250" 
                        value={volume} 
                        onChange={(e) => setVolume(parseInt(e.target.value))}
                        className="w-full h-1 bg-white/10 rounded-lg appearance-none cursor-pointer accent-emerald-500"
                    />
                    <div className="flex justify-between font-mono text-[10px] text-slate-600 uppercase tracking-widest">
                        <span>1 TB</span>
                        <span>250 TB</span>
                    </div>
                </div>

                {/* Discount Slider */}
                <div className="space-y-6">
                    <div className="flex justify-between items-end mb-2 border-b border-white/10 pb-4">
                        <div className="space-y-1">
                            <span className="block font-mono text-[10px] uppercase text-slate-500 tracking-widest font-bold">Variance</span>
                            <span className="font-mono text-3xl font-bold text-white">{discount}%</span>
                        </div>
                        <span className="font-mono text-[10px] text-slate-500 uppercase tracking-widest pb-1">EDP Discount</span>
                    </div>
                    <input 
                        type="range" 
                        min="0" 
                        max="80" 
                        step="5"
                        value={discount} 
                        onChange={(e) => setDiscount(parseInt(e.target.value))}
                        className="w-full h-1 bg-white/10 rounded-lg appearance-none cursor-pointer accent-emerald-500"
                    />
                    <div className="flex justify-between font-mono text-[10px] text-slate-600 uppercase tracking-widest">
                        <span>BASE [0%]</span>
                        <span>MAX [80%]</span>
                    </div>
                </div>

                {/* Effective Rate Output */}
                <div className="font-mono text-[10px] p-6 border border-white/10 bg-black/40 space-y-4">
                    <div className="flex items-center gap-2 border-b border-white/10 pb-3 mb-2">
                        <Info className="w-3 h-3 text-emerald-500" />
                        <h4 className="font-bold uppercase tracking-[0.2em] text-emerald-500/80">Rate Matrix</h4>
                    </div>
                    <div className="flex justify-between items-center text-xs">
                        <span className="text-slate-500 uppercase">List Price (Scrub + Egress)</span>
                        <span className="line-through text-slate-600 italic">{formatRate(listPricePerGB)} / GB</span>
                    </div>
                    <div className="flex justify-between items-center text-sm pt-2 border-t border-white/5">
                        <span className="font-bold text-white uppercase tracking-wider">Effective Rate</span>
                        <span className="font-bold text-emerald-400">{formatRate(effectiveRate)} / GB</span>
                    </div>
                </div>
            </div>

            {/* Right Col: Computation Results */}
            <div className="lg:col-span-7 space-y-8">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                    
                    {/* Cloud Pipeline Node */}
                    <div className="glass border-white/10 p-8 flex flex-col justify-between group hover:border-red-500/30 transition-colors">
                        <div>
                            <div className="flex items-center justify-between mb-8 border-b border-white/10 pb-6">
                                <h3 className="text-lg font-bold uppercase text-white tracking-tight">{selectedProvider.name}</h3>
                                <div className="font-mono text-[9px] border border-white/20 px-2 py-1 text-slate-500 uppercase tracking-widest font-bold">API_EGRESS</div>
                            </div>
                            
                            <div className="font-mono text-xs space-y-4 mb-10">
                                <div className="flex justify-between border-b border-white/5 pb-3">
                                    <span className="text-slate-500">Compute Overhead</span>
                                    <span className="text-slate-300 italic">{formatCurrency(totalCloudGrossMonthly)}</span>
                                </div>
                                <div className="flex justify-between border-b border-white/5 pb-3">
                                    <span className="text-slate-500">Discount Applied</span>
                                    <span className="text-red-500/80">-{formatCurrency(discountValueMonthly)}</span>
                                </div>
                                <div className="flex justify-between pt-2">
                                    <span className="text-slate-500">Network Latency</span>
                                    <span className="text-slate-400">100-500ms</span>
                                </div>
                            </div>
                        </div>
                        
                        <div className="pt-8 border-t border-white/10">
                            <span className="block font-mono text-[10px] text-slate-500 uppercase tracking-widest mb-3 font-bold">Net Financial Drain (Mo)</span>
                            <span className="text-4xl font-bold tracking-tighter text-red-500 drop-shadow-[0_0_15px_rgba(239,68,68,0.1)]">{formatCurrency(totalCloudNetMonthly)}</span>
                        </div>
                    </div>

                    {/* Ocultar Node */}
                    <div className="card-premium border-emerald-500/20 bg-emerald-500/[0.02] p-8 flex flex-col justify-between group relative overflow-hidden">
                        <div className="absolute top-0 right-0 p-4 opacity-10">
                            <Shield className="w-12 h-12 text-emerald-500" />
                        </div>
                        <div>
                            <div className="flex items-center justify-between mb-8 border-b border-white/10 pb-6">
                                <h3 className="text-lg font-bold uppercase text-white tracking-tight">Zero-Egress SLM</h3>
                                <div className="font-mono text-[9px] border border-emerald-500/30 px-2 py-1 text-emerald-500 uppercase tracking-widest font-bold">LOCAL_NODE</div>
                            </div>
                            
                            <div className="font-mono text-xs space-y-4 mb-10">
                                <div className="flex justify-between border-b border-white/5 pb-3">
                                    <span className="text-slate-500">Software License</span>
                                    <span className="text-emerald-400 font-bold">$10,000 / mo</span>
                                </div>
                                <div className="flex justify-between border-b border-white/5 pb-3">
                                    <span className="text-slate-500">Hardware Dep.</span>
                                    <span className="text-emerald-400/70">+{formatCurrency(localComputeMonthly)}</span>
                                </div>
                                <div className="flex justify-between pt-2">
                                    <span className="text-slate-500">P50 Latency</span>
                                    <span className="text-emerald-500 font-bold">0.92ms</span>
                                </div>
                            </div>
                        </div>
                        
                        <div className="pt-8 border-t border-white/10">
                            <span className="block font-mono text-[10px] text-slate-500 uppercase tracking-widest mb-3 font-bold">Fixed Resource Cost (Mo)</span>
                            <span className="text-4xl font-bold tracking-tighter text-white drop-shadow-[0_0_15px_rgba(255,255,255,0.1)]">{formatCurrency(totalOcultarMonthly)}</span>
                        </div>
                    </div>
                </div>

                {/* Differential Banner */}
                <div className="relative border border-white/10 bg-emerald-500/[0.03] p-10 flex flex-col sm:flex-row items-center justify-between overflow-hidden group hover:border-emerald-500/30 transition-all">
                    <div className="absolute top-0 left-0 w-full h-full bg-gradient-to-r from-emerald-500/[0.05] to-transparent pointer-events-none"></div>
                    <div className="relative z-10 space-y-2 mb-6 sm:mb-0">
                        <h4 className="text-2xl font-bold uppercase text-white tracking-tight">Total System Variance</h4>
                        <p className="font-mono text-xs text-emerald-500 uppercase tracking-[0.3em] font-bold">Projected 12-Month Capital Retention</p>
                    </div>
                    <div className="relative z-10 text-right">
                        <span className="block text-6xl font-bold tracking-tighter text-emerald-400 font-heading drop-shadow-[0_0_25px_rgba(52,211,153,0.3)]">
                            {formatCurrency(annualSavings)}
                        </span>
                    </div>
                </div>

                {/* CTA Section */}
                <div className="border border-white/10 bg-black/40 p-12 text-center space-y-8">
                    <h2 className="text-2xl md:text-3xl font-bold uppercase text-white tracking-tight">Execute Architecture Upgrade</h2>
                    <p className="font-mono text-sm text-slate-400 max-w-2xl mx-auto leading-relaxed">
                        Initiate deployment sequence. Neutralize regulatory risk and terminate external egress taxation via localized execution.
                    </p>
                    
                    <div className="flex flex-col sm:flex-row items-center justify-center gap-6">
                        <a href="mailto:sales@ocultar.io?subject=Enterprise%20Pilot%20Request" className="w-full sm:w-auto px-10 py-4 bg-emerald-500 text-black font-mono text-xs font-bold uppercase tracking-widest hover:bg-emerald-400 transition-colors flex items-center justify-center gap-2 group">
                            [ Init Pilot Protocol ]
                            <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
                        </a>
                        <a href="https://github.com/Edu963/ocultar" target="_blank" rel="noopener noreferrer" className="w-full sm:w-auto px-10 py-4 border border-white/10 text-white font-mono text-xs font-bold uppercase tracking-widest hover:bg-white/5 transition-colors">
                            [ Audit Source Code ]
                        </a>
                    </div>
                </div>
            </div>
        </div>

        {/* Citations Footer */}
        <div className="mt-24 pt-12 border-t border-white/10 font-mono text-[10px] text-slate-500">
            <h4 className="font-bold text-white mb-8 uppercase tracking-[0.4em] border-b border-white/10 inline-block pb-2">Telemetry & Data Sources</h4>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-12 text-slate-600 leading-relaxed">
                <div className="space-y-4">
                    <p><strong className="text-slate-400">[01] AWS Syntax:</strong> <a href="https://aws.amazon.com/comprehend/pricing/" target="_blank" rel="noopener noreferrer" className="underline hover:text-emerald-500 transition-colors">Comprehend API</a> - $0.0001 per 100 character array. ~1k/GB volumetric.</p>
                    <p><strong className="text-slate-400">[02] GCP Syntax:</strong> <a href="https://cloud.google.com/sensitive-data-protection/pricing" target="_blank" rel="noopener noreferrer" className="underline hover:text-emerald-500 transition-colors">Sensitive Data Protection API</a> - Base $3.00/GB inspect, $2.00/GB transform vector.</p>
                </div>
                <div className="space-y-4">
                    <p><strong className="text-slate-400">[03] Azure Syntax:</strong> <a href="https://azure.microsoft.com/en-us/pricing/details/cognitive-services/language-service/" target="_blank" rel="noopener noreferrer" className="underline hover:text-emerald-500 transition-colors">AI Language Services</a> - Median translation ~$1.50/GB processing.</p>
                    <p><strong className="text-slate-400">[04] Network Egress:</strong> Standard topology dictates $0.08 - $0.12/GB payload transit tax.</p>
                </div>
                <div className="space-y-4">
                    <p><strong className="text-slate-400">[05] Local Compute:</strong> Amortized generic Edge/CPU nodes via Q4_K_M GGUF.</p>
                    <p><strong className="text-slate-400">[06] Risk Vectors:</strong> <a href="https://www.ibm.com/reports/data-breach" target="_blank" rel="noopener noreferrer" className="underline hover:text-emerald-500 transition-colors">IBM Ponemon Report</a> - $10.22M median incident cost. PII at $160/record.</p>
                </div>
            </div>
            <div className="mt-12 pt-6 border-t border-white/5 text-center flex flex-col items-center gap-4">
                 <p className="uppercase tracking-[0.2em]">SYS_NOTE: 1 TB = 1,000 GB. Calculations assume standard un-discounted API tiers prior to EDP coefficient application.</p>
                 <Link to="/" className="flex items-center gap-2 text-emerald-500 hover:text-white transition-colors text-xs font-bold uppercase tracking-widest">
                    <ChevronRight className="w-4 h-4 rotate-180" /> [ BACK_TO_CONTROL_CENTER ]
                 </Link>
            </div>
        </div>
      </div>
    </div>
  );
};

export default CalculatorPage;
