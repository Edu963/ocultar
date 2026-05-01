import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Shield, ArrowLeft, BarChart3, Box, ChevronRight } from 'lucide-react';

import { toast } from 'sonner';

type ProviderId = 'gcp' | 'aws' | 'azure';

const PROVIDERS: Record<ProviderId, { name: string; processing: number; egress: number }> = {
  gcp:   { name: 'Google Cloud DLP',  processing: 5.00,    egress: 0.10 },
  aws:   { name: 'AWS Comprehend',    processing: 1000.00, egress: 0.10 },
  azure: { name: 'Azure AI Language', processing: 1.50,    egress: 0.10 },
};

const GB_PER_TB = 1000;
const OCULTAR_MONTHLY = 2075;
const LOCAL_COMPUTE_PER_TB = 20;

const fmt = (val: number) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(val);
const fmtRate = (val: number) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(val);

export default function Calculator() {
  const [provider, setProvider] = useState<ProviderId>('gcp');
  const [volume, setVolume] = useState(10);
  const [discount, setDiscount] = useState(0);

  const selectedProvider       = PROVIDERS[provider];
  const listPricePerGB         = selectedProvider.processing + selectedProvider.egress;
  const totalGB                = volume * GB_PER_TB;
  const totalCloudGrossMonthly = totalGB * listPricePerGB;
  const discountMultiplier     = discount / 100;
  const discountValueMonthly   = totalCloudGrossMonthly * discountMultiplier;
  const totalCloudNetMonthly   = totalCloudGrossMonthly - discountValueMonthly;
  const effectiveRate          = listPricePerGB * (1 - discountMultiplier);
  const localComputeMonthly    = volume * LOCAL_COMPUTE_PER_TB;
  const totalOcultarMonthly    = OCULTAR_MONTHLY + localComputeMonthly;
  const totalCloudAnnual       = totalCloudNetMonthly * 12;
  const totalOcultarAnnual     = totalOcultarMonthly * 12;
  const annualSavings          = Math.max(0, totalCloudAnnual - totalOcultarAnnual);
  const maxCost                = Math.max(totalCloudNetMonthly, totalOcultarMonthly, 1);
  const cloudBarPct            = (totalCloudNetMonthly / maxCost) * 100;
  const ocultarBarPct          = (totalOcultarMonthly / maxCost) * 100;
  const savingsPct             = totalCloudNetMonthly > 0
    ? Math.round((1 - totalOcultarMonthly / totalCloudNetMonthly) * 100)
    : 0;

  return (
    <div className="min-h-screen bg-[#050505]">
      <div className="max-container py-12 md:py-16">

        <Link
          to="/"
          className="inline-flex items-center gap-2 text-xs font-mono font-bold uppercase tracking-widest text-slate-600 hover:text-slate-300 transition-colors mb-10"
        >
          <ArrowLeft className="w-3 h-3" /> Platform
        </Link>

        <header className="mb-12 md:mb-16 text-center flex flex-col items-center gap-4">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-emerald-500/10 ring-1 ring-emerald-500/20 text-xs font-mono font-semibold text-emerald-400 uppercase tracking-widest">
            <BarChart3 className="w-3 h-3" /> Cost Analysis
          </div>
          <h1 className="text-white tracking-tight text-balance text-4xl font-bold">Compliance & ROI Forecast</h1>
          <p className="text-slate-400 max-w-2xl text-lg leading-relaxed">
            Quantify the financial variance between legacy cloud DLP and OCULTAR's EU-sovereign runtime.
            Factor in regulatory risk mitigation alongside infrastructure savings.
          </p>
        </header>

        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 lg:gap-10">

          {/* Inputs */}
          <div className="lg:col-span-5 flex flex-col gap-6">
            <div className="bg-[#0A0A0C] ring-1 ring-white/5 rounded-xl p-6 md:p-8 flex flex-col gap-8 shadow-xl">
              <p className="text-xs font-mono font-semibold uppercase tracking-widest text-emerald-500">Parameters</p>

              <div className="flex flex-col gap-3">
                <label className="text-xs font-mono text-slate-500 uppercase tracking-widest">Cloud Infrastructure</label>
                <select
                  value={provider}
                  onChange={(e) => setProvider(e.target.value as ProviderId)}
                  className="w-full bg-[#111114] ring-1 ring-white/10 text-white font-medium rounded-lg p-4 focus:outline-none focus:ring-emerald-500/40 transition-all appearance-none cursor-pointer"
                >
                  <option value="gcp">Google Cloud DLP ($5.00/GB)</option>
                  <option value="aws">AWS Comprehend (~$1,000/GB)</option>
                  <option value="azure">Azure AI Language (~$1.50/GB)</option>
                </select>
              </div>

              <div className="flex flex-col gap-3">
                <label className="text-xs font-mono text-slate-500 uppercase tracking-widest">Monthly Throughput</label>
                <div className="flex items-center gap-3">
                  <input
                    type="range" min="1" max="250"
                    value={volume}
                    onChange={(e) => setVolume(parseInt(e.target.value))}
                    className="flex-1 h-1.5 bg-white/10 rounded-lg appearance-none cursor-pointer accent-emerald-500"
                  />
                  <div className="flex items-center gap-1.5 bg-[#111114] ring-1 ring-white/10 rounded-lg px-3 py-2 shrink-0">
                    <input
                      type="number" min="1" max="250"
                      value={volume}
                      onChange={(e) => setVolume(Math.min(250, Math.max(1, parseInt(e.target.value) || 1)))}
                      className="w-10 bg-transparent text-white font-mono font-bold text-sm text-right focus:outline-none"
                    />
                    <span className="text-xs font-mono text-slate-500">TB</span>
                  </div>
                </div>
                <div className="flex justify-between text-xs font-mono text-slate-700"><span>1 TB</span><span>250 TB</span></div>
              </div>

              <div className="flex flex-col gap-3">
                <label className="text-xs font-mono text-slate-500 uppercase tracking-widest">Enterprise Discount</label>
                <div className="flex items-center gap-3">
                  <input
                    type="range" min="0" max="80" step="5"
                    value={discount}
                    onChange={(e) => setDiscount(parseInt(e.target.value))}
                    className="flex-1 h-1.5 bg-white/10 rounded-lg appearance-none cursor-pointer accent-emerald-500"
                  />
                  <div className="flex items-center gap-1.5 bg-[#111114] ring-1 ring-white/10 rounded-lg px-3 py-2 shrink-0">
                    <input
                      type="number" min="0" max="80" step="5"
                      value={discount}
                      onChange={(e) => setDiscount(Math.min(80, Math.max(0, parseInt(e.target.value) || 0)))}
                      className="w-10 bg-transparent text-white font-mono font-bold text-sm text-right focus:outline-none"
                    />
                    <span className="text-xs font-mono text-slate-500">%</span>
                  </div>
                </div>
                <div className="flex justify-between text-xs font-mono text-slate-700"><span>0%</span><span>80%</span></div>
              </div>
            </div>

            <div className="bg-[#0A0A0C] ring-1 ring-white/5 rounded-xl p-6 flex flex-col gap-4 shadow-xl">
              <p className="text-xs font-mono text-slate-500 uppercase tracking-widest border-b border-white/5 pb-4">Rate Summary</p>
              <div className="flex justify-between text-sm">
                <span className="text-slate-500">List Price (Processing + Egress)</span>
                <span className="text-slate-600 line-through font-mono">{fmtRate(listPricePerGB)} / GB</span>
              </div>
              <div className="flex justify-between font-bold">
                <span className="text-white">Effective Cloud Rate</span>
                <span className="text-emerald-400 font-mono">{fmtRate(effectiveRate)} / GB</span>
              </div>
            </div>
          </div>

          {/* Outputs */}
          <div className="lg:col-span-7 flex flex-col gap-6">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">

              <div className="bg-[#0A0A0C] ring-1 ring-white/5 rounded-xl p-6 flex flex-col gap-4 shadow-xl">
                <div className="flex justify-between items-center">
                  <span className="text-xs font-mono text-slate-500 uppercase tracking-widest">Legacy Stack</span>
                  <Box className="w-4 h-4 text-slate-700" />
                </div>
                <div className="text-base font-bold text-white">{selectedProvider.name}</div>
                <div className="flex flex-col gap-2 text-sm text-slate-500 font-mono">
                  <div className="flex justify-between"><span>Compute</span><span>{fmt(totalCloudGrossMonthly)}</span></div>
                  <div className="flex justify-between"><span>Discount</span><span className="text-rose-400">−{fmt(discountValueMonthly)}</span></div>
                </div>
                <div className="pt-4 border-t border-white/5">
                  <div className="text-xs font-mono text-slate-600 uppercase tracking-widest mb-2">Monthly Cost</div>
                  <div className="text-3xl font-bold font-mono text-white">{fmt(totalCloudNetMonthly)}</div>
                </div>
              </div>

              <div className="bg-[#0A0A0C] ring-1 ring-emerald-500/20 rounded-xl p-6 flex flex-col gap-4 shadow-xl shadow-emerald-500/5">
                <div className="flex justify-between items-center">
                  <span className="text-xs font-mono text-emerald-400 uppercase tracking-widest">Zero-Egress</span>
                  <Shield className="w-4 h-4 text-emerald-500" />
                </div>
                <div className="text-base font-bold text-white">OCULTAR Enterprise</div>
                <div className="flex flex-col gap-1.5 text-sm text-slate-500 font-mono">
                  <div className="flex justify-between"><span>Enterprise License</span><span className="text-emerald-400">€2,075</span></div>
                  <div className="text-xs text-slate-700 leading-relaxed py-1">
                    Proxy · Sombra · Refinery · Vault · Connectors · SIEM<br />
                    Annual: €24,900 · fixed regardless of volume
                  </div>
                  <div className="flex justify-between"><span>Local Compute</span><span>{fmt(localComputeMonthly)}</span></div>
                </div>
                <div className="pt-4 border-t border-white/5">
                  <div className="text-xs font-mono text-slate-600 uppercase tracking-widest mb-2">Monthly Equivalent</div>
                  <div className="text-3xl font-bold font-mono text-emerald-400">{fmt(totalOcultarMonthly)}</div>
                </div>
              </div>
            </div>

            <div className="bg-[#0A0A0C] ring-1 ring-white/5 rounded-xl p-6 flex flex-col gap-5 shadow-xl">
              <p className="text-xs font-mono text-slate-500 uppercase tracking-widest">Monthly Cost Comparison</p>
              <div className="flex flex-col gap-4">
                <div className="flex flex-col gap-2">
                  <div className="flex justify-between text-xs font-mono">
                    <span className="text-slate-500">{selectedProvider.name}</span>
                    <span className="text-slate-300 font-bold">{fmt(totalCloudNetMonthly)}</span>
                  </div>
                  <div className="h-3 bg-white/5 rounded-full overflow-hidden">
                    <div className="h-full bg-slate-500/70 rounded-full transition-all duration-500 ease-out" style={{ width: `${cloudBarPct}%` }} />
                  </div>
                </div>
                <div className="flex flex-col gap-2">
                  <div className="flex justify-between text-xs font-mono">
                    <span className="text-emerald-500">OCULTAR Enterprise</span>
                    <span className="text-emerald-400 font-bold">{fmt(totalOcultarMonthly)}</span>
                  </div>
                  <div className="h-3 bg-white/5 rounded-full overflow-hidden">
                    <div className="h-full bg-emerald-500 rounded-full transition-all duration-500 ease-out shadow-[0_0_8px_rgba(16,185,129,0.5)]" style={{ width: `${ocultarBarPct}%` }} />
                  </div>
                </div>
              </div>
              {savingsPct > 0 && (
                <p className="text-xs font-mono text-slate-600">
                  OCULTAR costs <span className="text-emerald-400 font-semibold">{savingsPct}% less</span> per month at this volume.
                </p>
              )}
            </div>

            <div className="bg-[#0A0A0C] ring-1 ring-white/5 rounded-xl p-8 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-6 shadow-xl">
              <div className="flex flex-col gap-1">
                <div className="text-lg font-bold text-white tracking-tight">Capital Retention</div>
                <div className="text-xs font-mono text-slate-500 uppercase tracking-widest">Projected 12-Month Savings</div>
              </div>
              <div className="text-right flex flex-col gap-1">
                <div className="text-5xl md:text-6xl font-black text-emerald-400 font-mono tracking-tighter drop-shadow-[0_0_15px_rgba(16,185,129,0.5)]">
                  {fmt(annualSavings)}
                </div>
                <div className="text-xs font-mono text-slate-600">per year</div>
              </div>
            </div>

            <div className="bg-[#0A0A0C] ring-1 ring-emerald-500/20 rounded-xl p-8 md:p-10 text-center flex flex-col items-center gap-5 shadow-xl shadow-emerald-500/5">
              <div className="text-xl font-bold text-white tracking-tight text-balance">
                Eliminate regulatory risk.<br />Unblock AI adoption.
              </div>
              <p className="text-slate-500 text-sm max-w-md leading-relaxed">
                Don't let a 6-month security review kill your AI roadmap. OCULTAR provides
                the "Privacy-First" green light that DPOs and regulators demand.
                Operational in under 60 minutes.
              </p>
              <button
                onClick={() => {
                  navigator.clipboard.writeText('sales@ocultar.dev');
                  toast.success('Email copied to clipboard', {
                    description: 'Reach out to sales@ocultar.dev to request access.',
                  });
                }}
                className="inline-flex items-center gap-3 bg-emerald-500 hover:bg-emerald-400 text-black font-bold px-8 py-4 rounded-lg text-sm uppercase tracking-widest transition-all duration-200 hover:-translate-y-0.5 hover:shadow-xl hover:shadow-emerald-500/20"
              >
                Request Access
                <ChevronRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
