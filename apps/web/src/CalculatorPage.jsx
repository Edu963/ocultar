import { useState } from 'react';
import { 
  Shield, 
  Cloud, 
  Server, 
  Zap, 
  TrendingDown, 
  CheckCircle2,
  Github,
  ChevronRight
} from 'lucide-react';
import { Link } from 'react-router-dom';

const CalculatorPage = () => {
  const [volume, setVolume] = useState(250); // TB
  const [discount, setDiscount] = useState(80); // %

  const GB_PER_TB = 1000;
  const INSPECTION_FEE = 3.00;
  const TRANSFORMATION_FEE = 2.00;
  const EGRESS_FEE = 0.10;
  
  const OCULTAR_LICENSE = 10000;

  // Calculations
  const totalGB = volume * GB_PER_TB;
  const cloudProcessingCost = totalGB * (INSPECTION_FEE + TRANSFORMATION_FEE);
  const cloudEgressCost = totalGB * EGRESS_FEE;
  const rawCloudMonthly = cloudProcessingCost + cloudEgressCost;
  
  const discountMultiplier = discount / 100;
  const cloudDiscountAmount = rawCloudMonthly * discountMultiplier;
  const netCloudMonthly = rawCloudMonthly - cloudDiscountAmount;
  
  const annualCloudCost = netCloudMonthly * 12;
  const annualOcultarCost = OCULTAR_LICENSE * 12;
  const annualSavings = Math.max(0, annualCloudCost - annualOcultarCost);

  const formatCurrency = (val) => 
    new Intl.NumberFormat('en-US', { 
      style: 'currency', 
      currency: 'USD', 
      maximumFractionDigits: 0 
    }).format(val);

  return (
    <div className="min-h-screen bg-[#050505] text-slate-300 font-sans selection:bg-emerald-500/30 selection:text-emerald-400 overflow-x-hidden pb-20 pt-24">
      {/* Background Glow */}
      <div className="fixed top-0 right-0 w-[600px] h-[600px] bg-emerald-500/5 blur-[120px] rounded-full -translate-y-1/2 translate-x-1/2 pointer-events-none z-0"></div>
      
      <div className="max-w-7xl mx-auto px-6 relative z-10">
        {/* Header */}
        <div className="mb-20 border-b border-white/5 pb-12 flex flex-col lg:flex-row justify-between items-start gap-12">
          <div className="space-y-6">
            <div className="badge">
                <span className="relative flex h-2 w-2">
                    <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
                </span>
                ROI Forecast Engine v1.2
            </div>
            <h1 className="text-5xl font-bold text-white tracking-tighter">
                Eliminate the <span className="text-emerald-500">Cloud Tax</span>.
            </h1>
            <p className="text-slate-400 max-w-2xl text-lg leading-relaxed font-medium">
              Calculate the hidden cost of sending sensitive data to external AI providers. 
              Ocultar's flat-fee local processing eliminates per-byte inspection and egress fees.
            </p>
          </div>
          <div className="lg:text-right w-full lg:w-auto">
             <div className="glass p-6 rounded-xl border border-emerald-500/20 shadow-2xl">
                <div className="text-[10px] font-mono text-slate-500 uppercase tracking-[0.3em] mb-2 font-bold">Standard Pilot Access</div>
                <div className="text-emerald-400 font-mono font-bold text-lg">FIXED_LICENSING_ENABLED</div>
                <div className="mt-4 flex gap-2">
                    <button className="btn btn-primary text-xs w-full">Request Quote</button>
                </div>
             </div>
          </div>
        </div>

        {/* Main Grid */}
        <div className="grid lg:grid-cols-12 gap-12 items-start">
          
          {/* Card 1: Workload Parameters */}
          <div className="lg:col-span-4 card border-white/10 p-8 relative overflow-hidden group">
            <div className="flex items-center gap-3 mb-10">
              <Zap className="w-5 h-5 text-emerald-400" />
              <h3 className="font-bold text-white uppercase tracking-[0.2em] text-xs">Workload Profile</h3>
            </div>

            <div className="space-y-12">
              {/* Volume Slider */}
              <div className="space-y-6">
                <div className="flex justify-between items-baseline">
                  <label className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Monthly Data (TB)</label>
                  <span className="text-4xl font-bold text-white font-heading">{volume}<span className="text-lg text-slate-600 ml-1">TB</span></span>
                </div>
                <input 
                  type="range" 
                  min="1" 
                  max="250" 
                  value={volume} 
                  onChange={(e) => setVolume(parseInt(e.target.value))}
                  className="w-full h-1 bg-white/10 rounded-lg appearance-none cursor-pointer accent-emerald-500"
                />
                <div className="flex justify-between text-[10px] font-mono text-slate-600 uppercase font-bold tracking-widest">
                  <span>1 TB</span>
                  <span>250 TB</span>
                </div>
              </div>

              {/* Discount Slider */}
              <div className="space-y-6">
                <div className="flex justify-between items-baseline">
                  <label className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Current Cloud Discount</label>
                  <span className="text-4xl font-bold text-white font-heading">{discount}<span className="text-lg text-slate-600 ml-1">%</span></span>
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
                <div className="flex justify-between text-[10px] font-mono text-slate-600 uppercase font-bold tracking-widest">
                  <span>0%</span>
                  <span>80% EDP</span>
                </div>
              </div>

              {/* Assumptions Table */}
              <div className="bg-black/40 border border-white/5 rounded-xl p-6 space-y-4">
                <h4 className="text-[10px] font-bold text-emerald-500/70 uppercase tracking-[0.2em] mb-4">SaaS DLP Assumptions</h4>
                <div className="space-y-4">
                  <div className="flex justify-between text-xs font-medium">
                    <span className="text-slate-500">Inspection Fee</span>
                    <span className="text-white font-mono">$3.00 / GB</span>
                  </div>
                  <div className="flex justify-between text-xs font-medium">
                    <span className="text-slate-500">Inbound Egress</span>
                    <span className="text-white font-mono">$0.10 / GB</span>
                  </div>
                  <div className="pt-4 border-t border-white/5 flex justify-between">
                    <span className="text-xs font-bold text-slate-400">Blended SaaS Rate</span>
                    <span className="text-sm font-bold text-red-500 font-mono">$5.10 / GB</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Card 2: Cloud Status Quo */}
          <div className="lg:col-span-4 card border-red-500/10 hover:border-red-500/30 p-8 shadow-2xl group flex flex-col justify-between">
            <div>
              <div className="flex items-center justify-between mb-10">
                <div className="flex items-center gap-3">
                  <Cloud className="w-5 h-5 text-red-400" />
                  <h3 className="font-bold text-white uppercase tracking-[0.2em] text-xs">SaaS Status Quo</h3>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-1.5 h-1.5 bg-red-500 rounded-full animate-pulse"></div>
                  <span className="text-[10px] font-mono text-red-500 uppercase tracking-widest font-bold">Variable_Cost</span>
                </div>
              </div>
              
              <div className="space-y-5">
                <div className="flex justify-between items-center p-3 hover:bg-white/[0.02] rounded-lg transition-colors border border-transparent hover:border-white/5">
                  <span className="text-xs text-slate-500 font-bold uppercase tracking-wider">Gross Inspection</span>
                  <span className="text-sm font-bold text-slate-300 font-mono">{formatCurrency(cloudProcessingCost)}</span>
                </div>
                <div className="flex justify-between items-center p-3 hover:bg-white/[0.02] rounded-lg transition-colors border border-transparent hover:border-white/5">
                  <span className="text-xs text-slate-500 font-bold uppercase tracking-wider">Egress Surcharge</span>
                  <span className="text-sm font-bold text-slate-300 font-mono">{formatCurrency(cloudEgressCost)}</span>
                </div>
                <div className="flex justify-between items-center p-3 bg-red-500/5 rounded-lg border border-red-500/10">
                  <span className="text-xs text-red-400 font-bold uppercase tracking-wider">Enterprise EDP</span>
                  <span className="text-sm font-bold text-red-400 font-mono">-{formatCurrency(cloudDiscountAmount)}</span>
                </div>
              </div>
            </div>

            <div className="mt-16 pt-8 border-t border-white/5">
              <div className="text-[10px] font-bold text-slate-500 uppercase tracking-[0.3em] mb-3">Estimated Monthly "Data Tax"</div>
              <div className="text-4xl font-bold text-red-500 tracking-tighter font-heading">{formatCurrency(netCloudMonthly)}</div>
            </div>
          </div>

          {/* Card 3: Ocultar Zero-Egress */}
          <div className="lg:col-span-4 card border-emerald-500/20 hover:border-emerald-500/50 p-8 shadow-[0_0_50px_rgba(16,185,129,0.05)] relative flex flex-col justify-between overflow-hidden bg-emerald-500/[0.02]">
            <div className="absolute -top-10 -right-10 w-40 h-40 bg-emerald-500/10 blur-[40px] rounded-full pointer-events-none"></div>
            
            <div>
              <div className="flex items-center justify-between mb-10">
                <div className="flex items-center gap-3">
                  <Server className="w-5 h-5 text-emerald-400" />
                  <h3 className="font-bold text-white uppercase tracking-[0.2em] text-xs">OCULTAR Bridged</h3>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-1.5 h-1.5 bg-emerald-500 rounded-full"></div>
                  <span className="text-[10px] font-mono text-emerald-500 uppercase tracking-widest font-bold">Hardened</span>
                </div>
              </div>
              
              <div className="space-y-5">
                <div className="flex justify-between items-center p-3 hover:bg-white/[0.02] rounded-lg border border-transparent hover:border-white/5">
                  <span className="text-xs text-slate-500 font-bold uppercase tracking-wider">Unit Processing</span>
                  <span className="text-sm font-bold text-emerald-400 font-mono">$0.00</span>
                </div>
                <div className="flex justify-between items-center p-3 hover:bg-white/[0.02] rounded-lg border border-transparent hover:border-white/5">
                  <span className="text-xs text-slate-500 font-bold uppercase tracking-wider">Egress Surcharge</span>
                  <span className="text-sm font-bold text-emerald-400 font-mono">$0.00</span>
                </div>
                <div className="flex justify-between items-center p-3 bg-emerald-500/10 rounded-lg border border-emerald-500/20">
                  <span className="text-xs text-emerald-400 font-bold uppercase tracking-wider">Local Processing</span>
                  <div className="flex items-center gap-1.5">
                    <Zap className="w-3 h-3 text-emerald-400" />
                    <span className="text-sm font-bold text-emerald-400 font-mono capitalize">Enabled</span>
                  </div>
                </div>
              </div>
            </div>

            <div className="mt-16 pt-8 border-t border-white/5">
              <div className="text-[10px] font-bold text-slate-500 uppercase tracking-[0.3em] mb-3">Fixed Platform License</div>
              <div className="text-4xl font-bold text-white tracking-tighter font-heading">{formatCurrency(OCULTAR_LICENSE)}</div>
            </div>
          </div>

          {/* Card 4: Savings Summary */}
          <div className="lg:col-span-12 flex flex-col md:flex-row items-center justify-between gap-12 p-12 mt-12 bg-emerald-500/[0.04] border border-emerald-500/20 rounded-2xl relative overflow-hidden group">
            <div className="absolute top-0 right-0 p-8 opacity-[0.03] group-hover:opacity-[0.08] transition-opacity">
              <TrendingDown className="w-64 h-64 -rotate-12" />
            </div>
            <div className="z-10 text-center md:text-left space-y-4">
               <div className="badge border-emerald-500/30 text-emerald-400">Savings Forecast Engine</div>
               <h4 className="text-4xl font-bold text-white tracking-tight">Annual Projected Retained Capital</h4>
               <p className="text-slate-500 max-w-xl text-lg font-medium">
                 By removing egress fees and per-byte "Data Taxes", your ROI scales exponentially with volume. 
                 Predictable costs for unpredictable AI workloads.
               </p>
            </div>
            <div className="z-10 text-center md:text-right">
               <div className="text-7xl md:text-8xl font-bold text-emerald-400 tracking-tighter font-heading drop-shadow-[0_0_30px_rgba(16,185,129,0.3)]">
                 {formatCurrency(annualSavings)}
               </div>
               <div className="flex items-center gap-2 justify-center md:justify-end mt-6 text-emerald-500 font-mono text-sm uppercase tracking-[0.2em] font-bold">
                  <CheckCircle2 className="w-5 h-5" /> Optimized for Volume
               </div>
            </div>
          </div>

        </div>

        {/* Footer Citations */}
        <div className="mt-32 pt-16 border-t border-white/5 text-center">
            <p className="text-[10px] font-mono text-slate-700 tracking-[0.4em] uppercase">
                Calculations based on 2026 Public Cloud DLP Benchmarks. Subject to infrastructure optimization and pilot terms.
            </p>
            <div className="mt-8 flex justify-center gap-10">
                <Link to="/" className="text-xs font-bold text-slate-500 hover:text-white transition-colors">Return Home</Link>
                <Link to="/risk-assessment" className="text-xs font-bold text-slate-500 hover:text-white transition-colors">Start Audit</Link>
                <a href="https://github.com/Edu963/ocultar" className="text-xs font-bold text-slate-500 hover:text-white transition-colors flex items-center gap-2">
                    <Github className="w-3 h-3" /> Core Engine
                </a>
            </div>
        </div>
      </div>
    </div>
  );
};

export default CalculatorPage;
