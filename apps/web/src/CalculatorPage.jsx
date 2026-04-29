import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Shield, ArrowLeft, BarChart3, Box } from 'lucide-react';

const REQUEST_ACCESS_URL =
    'mailto:sales@ocultar.dev?subject=Access%20Request&body=Hi%2C%0A%0AWork%20email%3A%20%0ACompany%20name%3A%20%0APrimary%20use%20case%20(Healthcare%20%2F%20Finance%20%2F%20Legal%20%2F%20Government%20%2F%20Other)%3A%20';

const CalculatorPage = () => {
    const [provider, setProvider] = useState('gcp');
    const [volume, setVolume] = useState(10);
    const [discount, setDiscount] = useState(0);

    const GB_PER_TB = 1000;
    const OCULTAR_MONTHLY = 2075; // €24,900/yr ÷ 12
    const LOCAL_COMPUTE_PER_TB = 20;

    const providers = {
        gcp:   { name: 'Google Cloud DLP',    processing: 5.00,    egress: 0.10 },
        aws:   { name: 'AWS Comprehend',       processing: 1000.00, egress: 0.10 },
        azure: { name: 'Azure AI Language',    processing: 1.50,    egress: 0.10 },
    };

    const selectedProvider = providers[provider];
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

    const fmt = (val) =>
        new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(val);
    const fmtRate = (val) =>
        new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(val);

    return (
        <div className="max-container py-12 md:py-16">

            {/* Back link */}
            <Link
                to="/"
                className="inline-flex items-center gap-2 text-[10px] font-bold uppercase tracking-widest text-zinc-500 hover:text-zinc-300 transition-colors mb-10"
            >
                <ArrowLeft className="w-3 h-3" /> Platform
            </Link>

            {/* Header */}
            <header className="mb-12 md:mb-16 text-center flex flex-col items-center">
                <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-orange-500/10 border border-orange-500/20 text-[10px] font-bold text-orange-400 uppercase tracking-widest mb-6">
                    <BarChart3 className="w-3 h-3" /> Cost Analysis
                </div>
                <h1 className="text-white mb-4">ROI Forecast Engine</h1>
                <p className="text-zinc-400 max-w-2xl text-lg leading-relaxed mx-auto">
                    Quantify the financial variance between external AI pipeline costs and OCULTAR's zero-egress local processing.
                </p>
            </header>

            {/* Grid */}
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 lg:gap-10">

                {/* ── Left: Inputs ── */}
                <div className="lg:col-span-5 space-y-6">

                    {/* Parameters */}
                    <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6 md:p-8 space-y-8">
                        <p className="text-[10px] uppercase tracking-widest text-orange-400 font-mono font-bold">Parameters</p>

                        {/* Provider */}
                        <div className="space-y-3">
                            <label className="text-[10px] font-bold text-zinc-500 uppercase tracking-widest block">
                                Cloud Infrastructure
                            </label>
                            <select
                                value={provider}
                                onChange={(e) => setProvider(e.target.value)}
                                className="w-full bg-zinc-800 border border-zinc-700 text-white font-medium rounded-lg p-4 focus:outline-none focus:border-orange-500 transition-colors appearance-none cursor-pointer"
                            >
                                <option value="gcp">Google Cloud DLP ($5.00/GB)</option>
                                <option value="aws">AWS Comprehend (~$1,000/GB)</option>
                                <option value="azure">Azure AI Language (~$1.50/GB)</option>
                            </select>
                        </div>

                        {/* Volume */}
                        <div className="space-y-4">
                            <div className="flex justify-between items-end">
                                <label className="text-[10px] font-bold text-zinc-500 uppercase tracking-widest">Monthly Throughput</label>
                                <span className="text-2xl font-bold font-mono text-white">{volume} TB</span>
                            </div>
                            <input
                                type="range" min="1" max="250"
                                value={volume} onChange={(e) => setVolume(parseInt(e.target.value))}
                                className="w-full h-1.5 bg-zinc-700 rounded-lg appearance-none cursor-pointer accent-orange-500"
                            />
                            <div className="flex justify-between text-[10px] font-mono text-zinc-600">
                                <span>1 TB</span><span>250 TB</span>
                            </div>
                        </div>

                        {/* Discount */}
                        <div className="space-y-4">
                            <div className="flex justify-between items-end">
                                <label className="text-[10px] font-bold text-zinc-500 uppercase tracking-widest">Enterprise Discount</label>
                                <span className="text-2xl font-bold font-mono text-white">{discount}%</span>
                            </div>
                            <input
                                type="range" min="0" max="80" step="5"
                                value={discount} onChange={(e) => setDiscount(parseInt(e.target.value))}
                                className="w-full h-1.5 bg-zinc-700 rounded-lg appearance-none cursor-pointer accent-orange-500"
                            />
                            <div className="flex justify-between text-[10px] font-mono text-zinc-600">
                                <span>0%</span><span>80%</span>
                            </div>
                        </div>
                    </div>

                    {/* Rate Summary */}
                    <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6 space-y-4">
                        <div className="text-[10px] font-bold text-zinc-500 uppercase tracking-widest border-b border-zinc-800 pb-3">
                            Rate Summary
                        </div>
                        <div className="flex justify-between text-sm">
                            <span className="text-zinc-500">List Price (Processing + Egress)</span>
                            <span className="text-zinc-600 line-through font-mono">{fmtRate(listPricePerGB)} / GB</span>
                        </div>
                        <div className="flex justify-between font-bold">
                            <span className="text-white">Effective Cloud Rate</span>
                            <span className="text-orange-400 font-mono">{fmtRate(effectiveRate)} / GB</span>
                        </div>
                    </div>
                </div>

                {/* ── Right: Outputs ── */}
                <div className="lg:col-span-7 space-y-6">

                    {/* Cost comparison */}
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">

                        {/* Legacy */}
                        <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6 space-y-4">
                            <div className="flex justify-between items-center">
                                <span className="text-[10px] font-bold text-zinc-500 uppercase tracking-widest">Legacy Stack</span>
                                <Box className="w-4 h-4 text-zinc-700" />
                            </div>
                            <div className="text-base font-bold text-white">{selectedProvider.name}</div>
                            <div className="space-y-2 text-sm text-zinc-500 font-mono">
                                <div className="flex justify-between">
                                    <span>Compute</span>
                                    <span>{fmt(totalCloudGrossMonthly)}</span>
                                </div>
                                <div className="flex justify-between">
                                    <span>Discount</span>
                                    <span className="text-rose-400">−{fmt(discountValueMonthly)}</span>
                                </div>
                            </div>
                            <div className="pt-4 border-t border-zinc-800">
                                <div className="text-[10px] uppercase font-bold text-zinc-600 mb-2">Monthly Cost</div>
                                <div className="text-3xl font-bold font-mono text-white">{fmt(totalCloudNetMonthly)}</div>
                            </div>
                        </div>

                        {/* OCULTAR */}
                        <div className="bg-zinc-900 border border-orange-500/30 rounded-xl p-6 space-y-4">
                            <div className="flex justify-between items-center">
                                <span className="text-[10px] font-bold text-orange-400 uppercase tracking-widest">Zero-Egress</span>
                                <Shield className="w-4 h-4 text-orange-500" />
                            </div>
                            <div className="text-base font-bold text-white">OCULTAR Enterprise</div>
                            <div className="space-y-1.5 text-sm text-zinc-500 font-mono">
                                <div className="flex justify-between">
                                    <span>Enterprise License</span>
                                    <span className="text-orange-400">€2,075</span>
                                </div>
                                <div className="text-[9px] text-zinc-600 leading-relaxed py-1">
                                    Proxy · Sombra · Refinery · Vault · Connectors · SIEM<br />
                                    Annual: €24,900 · fixed regardless of volume
                                </div>
                                <div className="flex justify-between">
                                    <span>Local Compute</span>
                                    <span>{fmt(localComputeMonthly)}</span>
                                </div>
                            </div>
                            <div className="pt-4 border-t border-zinc-800">
                                <div className="text-[10px] uppercase font-bold text-zinc-600 mb-2">Monthly Equivalent</div>
                                <div className="text-3xl font-bold font-mono text-orange-400">{fmt(totalOcultarMonthly)}</div>
                            </div>
                        </div>
                    </div>

                    {/* Annual savings */}
                    <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-8 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-6">
                        <div>
                            <div className="text-lg font-bold text-white mb-1">Capital Retention</div>
                            <div className="text-[10px] text-zinc-600 uppercase tracking-widest font-mono">
                                Projected 12-Month Savings
                            </div>
                        </div>
                        <div className="text-right">
                            <div className="text-4xl font-black text-orange-500 font-mono tracking-tighter">
                                {fmt(annualSavings)}
                            </div>
                            <div className="text-[10px] text-zinc-600 font-mono mt-1">per year</div>
                        </div>
                    </div>

                    {/* CTA */}
                    <div className="bg-zinc-900 border border-orange-500/20 rounded-xl p-8 md:p-10 text-center space-y-5">
                        <div className="text-xl font-bold text-white">
                            Neutralize regulatory risk.<br />Terminate external egress taxation.
                        </div>
                        <p className="text-zinc-500 text-sm max-w-sm mx-auto leading-relaxed">
                            No shared demo environment. No data sent anywhere. A local deployment in your infrastructure — running in 30 minutes.
                        </p>
                        <a
                            href={REQUEST_ACCESS_URL}
                            className="inline-flex items-center gap-2 bg-orange-500 hover:bg-orange-600 text-white font-bold px-8 py-4 rounded text-sm uppercase tracking-widest transition-colors"
                        >
                            Request Access
                        </a>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default CalculatorPage;
