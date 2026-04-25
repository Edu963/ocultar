import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Shield, ArrowLeft, Cpu, BarChart3, Lock, ExternalLink, Box } from 'lucide-react';

const CalculatorPage = () => {
    const [provider, setProvider] = useState('gcp');
    const [volume, setVolume] = useState(10);
    const [discount, setDiscount] = useState(0);

    const GB_PER_TB = 1000;
    const OCULTAR_MONTHLY = 2075; // €24,900/yr annual license ÷ 12
    const LOCAL_COMPUTE_PER_TB = 20;

    const providers = {
        gcp: { name: "Google Cloud DLP", processing: 5.00, egress: 0.10 },
        aws: { name: "AWS Comprehend", processing: 1000.00, egress: 0.10 },
        azure: { name: "Azure AI Language", processing: 1.50, egress: 0.10 }
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
        new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(val);

    const formatRate = (val) =>
        new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(val);

    return (
        <div className="min-h-screen font-sans antialiased bg-white text-gray-900 selection:bg-cyan-600 selection:text-white pb-20">
            
            {/* Nav */}
            <nav className="fixed top-0 left-0 w-full z-50 py-5">
                <div className="max-container flex justify-between items-center">
                    <Link to="/" className="flex items-center gap-3 group">
                        <span className="font-mono font-black text-xl tracking-widest text-gray-900">OCULTAR</span>
                    </Link>
                    <Link to="/" className="text-[11px] font-bold uppercase tracking-widest text-gray-500 hover:text-gray-900 transition-colors flex items-center gap-2">
                        <ArrowLeft className="w-3 h-3" /> Back to Platform
                    </Link>
                </div>
            </nav>

            <div className="max-container pt-32">
                <header className="mb-16">
                    <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-cyan-500/5 border border-cyan-500/10 text-[10px] font-bold text-cyan-500 uppercase tracking-widest mb-6">
                        <BarChart3 className="w-3 h-3" /> Cost Analysis Module
                    </div>
                    <h1 className="mb-4">ROI Forecast Engine</h1>
                    <p className="text-gray-600 max-w-2xl text-lg">
                        Quantify the financial variance between external AI pipeline taxing and OCULTAR's zero-egress local processing.
                    </p>
                </header>

                <div className="grid grid-cols-1 lg:grid-cols-12 gap-12">
                    {/* Left: Inputs */}
                    <div className="lg:col-span-5 space-y-10">
                        <div className="card space-y-8">
                            <h3 className="text-sm uppercase tracking-widest text-cyan-600 font-mono">Parameters</h3>
                            
                            <div className="space-y-3">
                                <label className="text-[10px] font-bold text-gray-500 uppercase tracking-widest block">Cloud Infrastructure</label>
                                <select
                                    value={provider}
                                    onChange={(e) => setProvider(e.target.value)}
                                    className="w-full bg-gray-50 border border-gray-200 text-gray-900 font-medium rounded-lg p-4 focus:outline-none focus:border-cyan-500 transition-colors appearance-none"
                                >
                                    <option value="gcp">Google Cloud DLP ($5.00/GB)</option>
                                    <option value="aws">AWS Comprehend (~$1,000/GB)</option>
                                    <option value="azure">Azure AI Language (~$1.50/GB)</option>
                                </select>
                            </div>

                            <div className="space-y-5">
                                <div className="flex justify-between items-end">
                                    <label className="text-[10px] font-bold text-gray-500 uppercase tracking-widest block">Monthly Throughput</label>
                                    <span className="text-2xl font-bold font-mono text-gray-900">{volume} TB</span>
                                </div>
                                <input
                                    type="range" min="1" max="250"
                                    value={volume} onChange={(e) => setVolume(parseInt(e.target.value))}
                                    className="w-full h-1.5 bg-gray-200 rounded-lg appearance-none cursor-pointer accent-cyan-600"
                                />
                            </div>

                            <div className="space-y-5">
                                <div className="flex justify-between items-end">
                                    <label className="text-[10px] font-bold text-gray-500 uppercase tracking-widest block">Enterprise Discount</label>
                                    <span className="text-2xl font-bold font-mono text-gray-900">{discount}%</span>
                                </div>
                                <input
                                    type="range" min="0" max="80" step="5"
                                    value={discount} onChange={(e) => setDiscount(parseInt(e.target.value))}
                                    className="w-full h-1.5 bg-gray-200 rounded-lg appearance-none cursor-pointer accent-cyan-600"
                                />
                            </div>
                        </div>

                        <div className="p-6 rounded-xl border border-gray-200 bg-gray-50 space-y-4">
                            <div className="text-[10px] font-bold text-gray-500 uppercase tracking-widest border-b border-gray-200 pb-2">Rate Summary</div>
                            <div className="flex justify-between text-sm">
                                <span className="text-gray-500">List Price (Processing + Egress)</span>
                                <span className="text-gray-400 line-through">{formatRate(listPricePerGB)} / GB</span>
                            </div>
                            <div className="flex justify-between text-base font-bold">
                                <span>Effective Cloud Rate</span>
                                <span className="text-cyan-600">{formatRate(effectiveRate)} / GB</span>
                            </div>
                        </div>
                    </div>

                    {/* Right: Outputs */}
                    <div className="lg:col-span-7 space-y-8">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <div className="card flex flex-col justify-between h-[340px]">
                                <div className="space-y-4">
                                    <div className="flex justify-between items-center">
                                        <div className="text-[10px] font-bold text-gray-500 uppercase tracking-widest">Legacy Stack</div>
                                        <Box className="w-4 h-4 text-gray-300" />
                                    </div>
                                    <h3 className="text-xl">{selectedProvider.name}</h3>
                                    <div className="space-y-2 text-sm text-gray-500 font-mono">
                                        <div className="flex justify-between">
                                            <span>Compute</span>
                                            <span>{formatCurrency(totalCloudGrossMonthly)}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span>Discount</span>
                                            <span className="text-rose-500">-{formatCurrency(discountValueMonthly)}</span>
                                        </div>
                                    </div>
                                </div>
                                <div className="pt-6 border-t border-gray-100">
                                    <div className="text-[10px] uppercase font-bold text-gray-400 mb-1">Monthly Cost</div>
                                    <div className="text-4xl font-bold font-mono text-gray-900">{formatCurrency(totalCloudNetMonthly)}</div>
                                </div>
                            </div>

                            <div className="card bg-cyan-50 border-cyan-200 flex flex-col justify-between h-[340px]">
                                <div className="space-y-4">
                                    <div className="flex justify-between items-center">
                                        <div className="text-[10px] font-bold text-cyan-700 uppercase tracking-widest">Zero-Egress</div>
                                        <Shield className="w-4 h-4 text-cyan-600" />
                                    </div>
                                    <h3 className="text-xl">OCULTAR Enterprise</h3>
                                    <div className="space-y-1 text-sm text-cyan-700/60 font-mono">
                                        <div className="flex justify-between">
                                            <span>Enterprise License (full stack)</span>
                                            <span className="text-cyan-700">€2,075</span>
                                        </div>
                                        <div className="text-[9px] text-cyan-600/50 leading-relaxed pb-1">
                                            Proxy · Sombra Gateway · Refinery · Vault · Connectors · SIEM<br/>
                                            Annual: €24,900 · fixed regardless of volume
                                        </div>
                                        <div className="flex justify-between pt-1">
                                            <span>Local Compute</span>
                                            <span>{formatCurrency(localComputeMonthly)}</span>
                                        </div>
                                    </div>
                                </div>
                                <div className="pt-6 border-t border-cyan-200">
                                    <div className="text-[10px] uppercase font-bold text-cyan-600/60 mb-1">Monthly Equivalent</div>
                                    <div className="text-4xl font-bold font-mono text-cyan-700">{formatCurrency(totalOcultarMonthly)}</div>
                                </div>
                            </div>
                        </div>

                        <div className="card flex items-center justify-between py-10">
                            <div>
                                <h3 className="text-3xl mb-1">Capital Retention</h3>
                                <p className="text-[10px] text-gray-400 uppercase tracking-widest font-mono">Projected 12-Month Financial Savings</p>
                            </div>
                            <div className="text-right">
                                <div className="text-6xl font-black text-transparent bg-clip-text bg-gradient-to-b from-gray-900 to-gray-400 tracking-tighter">
                                    {formatCurrency(annualSavings)}
                                </div>
                            </div>
                        </div>

                        <div className="card bg-cyan-600 text-white p-12 text-center space-y-6 border-0">
                            <h2 className="text-3xl text-white lowercase italic font-light tracking-tight">Neutralize regulatory risk and terminate external egress taxation.</h2>
                            <div className="flex justify-center gap-4">
                                <button className="bg-white text-cyan-700 px-8 py-4 rounded-lg font-bold text-sm uppercase tracking-widest hover:bg-cyan-50 transition-colors">
                                    Initialize Pilot Protocol
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default CalculatorPage;
