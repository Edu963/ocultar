import { useState, useEffect } from 'react';
import { 
  Shield, 
  Cloud, 
  Server, 
  Info, 
  ArrowRight, 
  TrendingDown, 
  Tractor, 
  Dna,
  Database,
  Lock,
  Zap,
  CheckCircle2
} from 'lucide-react';

const CalculatorPage = () => {
  const [volume, setVolume] = useState(250); // TB
  const [discount, setDiscount] = useState(80); // %

  const GB_PER_TB = 1000;
  const INSPECTION_FEE = 3.00;
  const TRANSFORMATION_FEE = 2.00;
  const EGRESS_FEE = 0.10;
  const TOTAL_CLOUD_FEE = INSPECTION_FEE + TRANSFORMATION_FEE + EGRESS_FEE;
  
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
    <div className="min-h-screen bg-[#050811] text-slate-300 font-sans selection:bg-cyan-500/30 selection:text-cyan-400 overflow-x-hidden pb-20 pt-24">
      {/* Background Glow */}
      <div className="fixed top-0 right-0 w-[500px] h-[500px] bg-cyan-500/10 blur-[120px] rounded-full -translate-y-1/2 translate-x-1/2 pointer-events-none z-0"></div>
      
      <div className="max-w-7xl mx-auto px-6 relative z-10">
        {/* Header */}
        <div className="mb-12 border-b border-white/10 pb-8 flex flex-col lg:row justify-between items-start gap-6">
          <div className="space-y-4">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-white flex items-center justify-center rounded-sm">
                <Shield className="text-black w-6 h-6" strokeWidth={2.5} />
              </div>
              <div className="flex flex-col">
                 <span className="font-bold tracking-tighter text-white text-2xl uppercase">OCULTAR <span className="text-slate-500 font-light">| Data Refinery</span></span>
                 <span className="text-xs font-mono text-cyan-500 tracking-widest uppercase mt-0.5">Zero-Egress Infrastructure</span>
              </div>
            </div>
            <h1 className="text-3xl font-bold text-white tracking-tight">Zero-Egress Egress & Processing ROI Calculator</h1>
            <p className="text-slate-400 max-w-2xl text-sm leading-relaxed">
              Calculate the hidden "Cloud Tax" of sending sensitive data to external AI and DLP providers. 
              OCULTAR's flat-fee local processing eliminates per-byte inspection and egress fees.
            </p>
          </div>
          <div className="lg:text-right">
             <div className="bg-white/5 border border-white/10 p-4 rounded-lg inline-block backdrop-blur-md">
                <div className="text-[10px] font-mono text-slate-500 uppercase tracking-widest mb-1">Pricing Model</div>
                <div className="text-cyan-400 font-mono font-bold">FLAT_FEE_PILOT_V1</div>
             </div>
          </div>
        </div>

        {/* Main Grid */}
        <div className="grid lg:grid-cols-12 gap-8 items-start">
          
          {/* Card 1: Workload Parameters */}
          <div className="lg:col-span-4 bg-[#0a0f1d] border border-white/10 rounded-2xl p-8 shadow-2xl relative overflow-hidden group">
            <div className="absolute top-0 left-0 w-full h-1 bg-cyan-500/50"></div>
            <div className="flex items-center gap-3 mb-8">
              <Zap className="w-5 h-5 text-cyan-400" />
              <h3 className="font-bold text-white uppercase tracking-widest text-sm">Workload Parameters</h3>
            </div>

            <div className="space-y-10">
              {/* Volume Slider */}
              <div className="space-y-4">
                <div className="flex justify-between items-baseline">
                  <label className="text-xs font-bold text-slate-500 uppercase tracking-wider">Monthly Data Processed for PII</label>
                  <span className="text-3xl font-bold text-white font-mono">{volume} TB</span>
                </div>
                <input 
                  type="range" 
                  min="1" 
                  max="250" 
                  value={volume} 
                  onChange={(e) => setVolume(parseInt(e.target.value))}
                  className="w-full h-1.5 bg-white/10 rounded-lg appearance-none cursor-pointer accent-cyan-500"
                />
                <div className="flex justify-between text-[10px] font-mono text-slate-500 uppercase">
                  <span>1 TB</span>
                  <span>250 TB</span>
                </div>
              </div>

              {/* Discount Slider */}
              <div className="space-y-4">
                <div className="flex justify-between items-baseline">
                  <label className="text-xs font-bold text-slate-500 uppercase tracking-wider">Enterprise Cloud Discount</label>
                  <span className="text-3xl font-bold text-white font-mono">{discount}%</span>
                </div>
                <input 
                  type="range" 
                  min="0" 
                  max="80" 
                  step="5"
                  value={discount} 
                  onChange={(e) => setDiscount(parseInt(e.target.value))}
                  className="w-full h-1.5 bg-white/10 rounded-lg appearance-none cursor-pointer accent-cyan-500"
                />
                <div className="flex justify-between text-[10px] font-mono text-slate-500 uppercase">
                  <span>0%</span>
                  <span>80%</span>
                </div>
              </div>

              {/* Assumptions Table */}
              <div className="bg-black/40 border border-white/5 rounded-xl p-5 space-y-4">
                <h4 className="text-[10px] font-bold text-cyan-500/70 uppercase tracking-[0.2em] mb-2">Cloud Provider Assumptions</h4>
                <div className="space-y-3">
                  <div className="flex justify-between text-xs">
                    <span className="text-slate-400">Inspection Fee (Avg)</span>
                    <span className="text-white font-mono">$3.00 / GB</span>
                  </div>
                  <div className="flex justify-between text-xs">
                    <span className="text-slate-400">Transformation Fee (Avg)</span>
                    <span className="text-white font-mono">$2.00 / GB</span>
                  </div>
                  <div className="flex justify-between text-xs">
                    <span className="text-slate-400">Egress Fee (Avg)</span>
                    <span className="text-white font-mono">$0.10 / GB</span>
                  </div>
                  <div className="pt-3 border-t border-white/10 flex justify-between">
                    <span className="text-xs font-bold text-slate-300">Total Cloud Cost</span>
                    <span className="text-sm font-bold text-red-500 font-mono">$5.10 / GB</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Card 2: Cloud Status Quo */}
          <div className="lg:col-span-4 bg-[#0a0f1d] border border-red-500/20 rounded-2xl p-8 shadow-2xl relative flex flex-col justify-between">
            <div className="absolute top-0 left-0 w-full h-1 bg-red-500/50"></div>
            <div>
              <div className="flex items-center justify-between mb-8">
                <div className="flex items-center gap-3">
                  <Cloud className="w-5 h-5 text-red-400" />
                  <h3 className="font-bold text-white uppercase tracking-widest text-sm">Cloud Status Quo</h3>
                </div>
                <div className="animate-pulse flex items-center gap-2">
                  <div className="w-2 h-2 bg-red-500 rounded-full"></div>
                  <span className="text-[10px] font-mono text-red-500 uppercase tracking-widest font-bold">SaaS_Leakage</span>
                </div>
              </div>
              <p className="text-xs text-slate-500 font-mono mb-8 uppercase tracking-widest">AWS Comprehend / GCP DLP Pipeline</p>
              
              <div className="space-y-4">
                <div className="flex justify-between items-center group/item hover:bg-white/5 p-2 -mx-2 rounded-lg transition-colors">
                  <span className="text-xs text-slate-400 group-hover/item:text-slate-300 transition-colors">Processing Tax</span>
                  <span className="text-sm font-bold text-slate-300 font-mono">{formatCurrency(cloudProcessingCost)}</span>
                </div>
                <div className="flex justify-between items-center group/item hover:bg-white/5 p-2 -mx-2 rounded-lg transition-colors">
                  <span className="text-xs text-slate-400 group-hover/item:text-slate-300 transition-colors">Egress Tax</span>
                  <span className="text-sm font-bold text-slate-300 font-mono">{formatCurrency(cloudEgressCost)}</span>
                </div>
                <div className="flex justify-between items-center group/item hover:bg-emerald-500/5 p-2 -mx-2 rounded-lg transition-colors">
                  <span className="text-xs text-slate-400 group-hover/item:text-emerald-400 transition-colors">Enterprise Discount</span>
                  <span className="text-sm font-bold text-emerald-400 font-mono">-{formatCurrency(cloudDiscountAmount)}</span>
                </div>
                <div className="flex justify-between items-center p-2 -mx-2">
                  <span className="text-xs text-slate-400">Latency (Round Trip)</span>
                  <span className="text-xs font-bold text-red-400 font-mono uppercase tracking-widest">100ms - 500ms</span>
                </div>
              </div>
            </div>

            <div className="mt-12 pt-8 border-t border-white/10">
              <div className="text-[10px] font-bold text-slate-500 uppercase tracking-[0.2em] mb-2">Estimated Monthly Cost</div>
              <div className="text-4xl font-bold text-red-500 tracking-tighter font-mono">{formatCurrency(netCloudMonthly)}</div>
            </div>
          </div>

          {/* Card 3: Ocultar Zero-Egress */}
          <div className="lg:col-span-4 bg-gradient-to-br from-[#0c1429] to-[#0a0f1d] border border-cyan-500/30 rounded-2xl p-8 shadow-[0_0_50px_rgba(34,211,238,0.1)] relative flex flex-col justify-between overflow-hidden">
            <div className="absolute top-0 left-0 w-full h-1 bg-cyan-400"></div>
            <div className="absolute -top-10 -right-10 w-40 h-40 bg-cyan-500/5 blur-[40px] rounded-full pointer-events-none"></div>
            
            <div>
              <div className="flex items-center justify-between mb-8">
                <div className="flex items-center gap-3">
                  <Server className="w-5 h-5 text-cyan-400" />
                  <h3 className="font-bold text-white uppercase tracking-widest text-sm">OCULTAR Zero-Egress</h3>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 bg-emerald-500 rounded-full"></div>
                  <span className="text-[10px] font-mono text-emerald-500 uppercase tracking-widest font-bold">Hardened</span>
                </div>
              </div>
              <p className="text-xs text-slate-500 font-mono mb-8 uppercase tracking-widest">Local SLM Enterprise License</p>
              
              <div className="space-y-4">
                <div className="flex justify-between items-center group/item hover:bg-white/5 p-2 -mx-2 rounded-lg transition-colors">
                  <span className="text-xs text-slate-400 group-hover/item:text-slate-300 transition-colors">Processing Tax</span>
                  <span className="text-sm font-bold text-cyan-400 font-mono">$0 (Included)</span>
                </div>
                <div className="flex justify-between items-center group/item hover:bg-white/5 p-2 -mx-2 rounded-lg transition-colors">
                  <span className="text-xs text-slate-400 group-hover/item:text-slate-300 transition-colors">Egress Tax</span>
                  <span className="text-sm font-bold text-cyan-400 font-mono">$0 (Local)</span>
                </div>
                <div className="flex justify-between items-center group/item hover:bg-white/5 p-2 -mx-2 rounded-lg transition-colors">
                  <span className="text-xs text-slate-400 group-hover/item:text-slate-300 transition-colors">Latency (P50)</span>
                  <div className="flex items-center gap-1.5">
                    <Zap className="w-3 h-3 text-emerald-400" />
                    <span className="text-sm font-bold text-emerald-400 font-mono">0.92ms</span>
                  </div>
                </div>
              </div>
            </div>

            <div className="mt-12 pt-8 border-t border-white/10">
              <div className="text-[10px] font-bold text-slate-500 uppercase tracking-[0.2em] mb-2">Fixed Monthly License</div>
              <div className="text-4xl font-bold text-white tracking-tighter font-mono">{formatCurrency(OCULTAR_LICENSE)}</div>
            </div>
          </div>

          {/* Card 4: Savings Summary (Desktop/Large Only) */}
          <div className="lg:col-span-12 flex flex-col md:row items-center justify-between gap-8 p-10 mt-8 bg-gradient-to-r from-emerald-500/10 to-transparent border border-emerald-500/20 rounded-2xl relative overflow-hidden group">
            <div className="absolute top-0 right-0 p-8 opacity-5 group-hover:opacity-10 transition-opacity">
              <TrendingDown className="w-40 h-40 -rotate-12" />
            </div>
            <div className="z-10 text-center md:text-left">
               <h4 className="text-2xl font-bold text-white mb-2">Annual Projected Savings</h4>
               <p className="text-slate-400 max-w-lg text-sm">
                 Eliminate the per-byte "Data Tax" and secure your TCO. By removing egress fees and 3rd party processing fees, your ROI scales exponentially with volume.
               </p>
            </div>
            <div className="z-10 text-center md:text-right">
               <div className="text-[10px] font-bold text-emerald-500 uppercase tracking-[0.3em] mb-4">Capital Retention Plan</div>
               <div className="text-6xl md:text-7xl font-bold text-emerald-400 tracking-tighter font-mono">
                 {formatCurrency(annualSavings)}
               </div>
               <div className="flex items-center gap-2 justify-center md:justify-end mt-4 text-emerald-500 font-mono text-xs uppercase tracking-widest font-bold">
                  <CheckCircle2 className="w-4 h-4" /> Optimized Architecture
               </div>
            </div>
          </div>

        </div>

        {/* Footer Citations */}
        <div className="mt-20 pt-10 border-t border-white/10 space-y-10">
          <div className="flex items-center gap-4">
            <div className="bg-white/5 border border-white/10 px-4 py-2 rounded-sm text-[10px] font-mono text-slate-500 uppercase tracking-widest font-bold">
              Data Citations (FY 2025/2026)
            </div>
            <div className="h-[1px] flex-1 bg-white/5"></div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-10">
            <div className="space-y-6">
              <div className="space-y-2">
                <span className="text-[10px] font-mono text-cyan-500 uppercase font-bold tracking-widest">[01] AWS Official Pricing</span>
                <p className="text-[11px] leading-relaxed text-slate-500 italic">
                  "Amazon Comprehend Pricing - Detect PII: $0.0001 per 100 character unit." 
                  Effectively <span className="text-slate-300 font-bold">~$1,000 per GB</span> for dense text workloads.
                </p>
              </div>
              <div className="space-y-2">
                <span className="text-[10px] font-mono text-cyan-500 uppercase font-bold tracking-widest">[02] Google Cloud Documentation</span>
                <p className="text-[11px] leading-relaxed text-slate-500 italic">
                  "Sensitive Data Protection pricing - Content inspection/transformation method: 
                  Starting at <span className="text-slate-300 font-bold">$3.00/GB</span> for inspection and 
                  <span className="text-slate-300 font-bold">$2.00/GB</span> for transformation."
                </p>
              </div>
            </div>

            <div className="space-y-6">
              <div className="space-y-2">
                <span className="text-[10px] font-mono text-cyan-500 uppercase font-bold tracking-widest">[03] Akave / HostDime Cloud Analysis</span>
                <p className="text-[11px] leading-relaxed text-slate-500 italic">
                  Standard public cloud egress rates range from $0.08 to $0.12 per GB. 
                  "Cloud Egress Fees Explained: Why AWS & Azure Cost 5–6x More Than Advertised."
                </p>
              </div>
              <div className="space-y-2">
                <span className="text-[10px] font-mono text-cyan-500 uppercase font-bold tracking-widest">[04] IBM Newsroom</span>
                <p className="text-[11px] leading-relaxed text-slate-500 italic">
                  "2025 Cost of a Data Breach Report: U.S. breach costs rise to <span className="text-slate-300 font-bold">$10.22 million</span>."
                </p>
              </div>
            </div>

            <div className="space-y-6">
              <div className="space-y-2">
                <span className="text-[10px] font-mono text-cyan-500 uppercase font-bold tracking-widest">[05] IBM / Varonis Security Report</span>
                <p className="text-[11px] leading-relaxed text-slate-500 italic">
                  "72% of all data breaches now involve data stored in the cloud."
                </p>
              </div>
              <div className="space-y-2">
                <span className="text-[10px] font-mono text-cyan-500 uppercase font-bold tracking-widest">[06] Secureframe Analysis</span>
                <p className="text-[11px] leading-relaxed text-slate-500 italic">
                   "Every compromised customer PII record costs an organization an average of 
                   <span className="text-slate-300 font-bold">$160</span> in legal, notification, and recovery costs."
                </p>
              </div>
            </div>
          </div>

          <div className="text-[9px] text-center text-slate-700 font-mono uppercase tracking-[0.2em] pt-8">
            * CALCULATOR ASSUMES BLENDED CLOUD RATE OF $5.10/GB (INSPECTION + TRANSFORMATION + EGRESS). OCULTAR LICENSE IS ILLUSTRATIVE. 1 TB = 1,000 GB.
          </div>
        </div>
      </div>
    </div>
  );
};

export default CalculatorPage;
