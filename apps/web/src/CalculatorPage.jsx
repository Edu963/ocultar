import { useState } from 'react';
import { Link } from 'react-router-dom';
import logo3 from './assets/images/logo3.jpg';

const CalculatorPage = () => {
    const [provider, setProvider] = useState('gcp');
    const [volume, setVolume] = useState(10);
    const [discount, setDiscount] = useState(0);

    const GB_PER_TB = 1000;
    const OCULTAR_MONTHLY = 10000;
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
        <div className="min-h-screen font-sans antialiased relative overflow-x-hidden selection:bg-white selection:text-black">

            <div className="relative z-10 max-w-6xl mx-auto p-6 md:p-12">

                {/* Header */}
                <header className="mb-12 border-b border-white pb-8 flex flex-col md:flex-row justify-between items-start gap-8 p-6">
                    <div className="max-w-2xl order-2 md:order-1">
                        <div className="flex items-baseline gap-3 mb-4">
                            <h1 className="text-3xl md:text-5xl font-bold tracking-tight uppercase">Data Refinery</h1>
                            <span className="font-mono text-gray-500">v_1.0.0</span>
                        </div>
                        <h2 className="font-mono text-sm uppercase tracking-widest border-b border-white inline-block mb-4 pb-1">ROI Calculator Module</h2>
                        <p className="font-mono text-sm text-gray-400 leading-relaxed max-w-xl">
                            [SYS_MSG]: Compute the mathematical variance between external AI pipeline taxation and OCULTAR's zero-egress local processing framework.
                        </p>
                    </div>
                    <div className="flex-shrink-0 w-full md:w-auto flex justify-start md:justify-end order-1 md:order-2">
                        <img
                            src={logo3}
                            alt="OCULTAR Logo"
                            className="h-8 md:h-10 w-auto object-contain invert"
                        />
                    </div>
                </header>

                {/* Main Layout Grid */}
                <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">

                    {/* Left Col: Parameters */}
                    <div className="lg:col-span-5 border border-white p-8">
                        <div className="flex items-center gap-2 mb-8 border-b border-white pb-4">
                            <div className="w-2 h-2 bg-white"></div>
                            <h3 className="font-mono text-sm font-bold uppercase tracking-widest">Input Parameters</h3>
                        </div>

                        {/* Cloud Provider Selection */}
                        <div className="mb-10">
                            <label className="block font-mono text-xs uppercase text-gray-400 mb-3">Target Cloud Architecture</label>
                            <select
                                value={provider}
                                onChange={(e) => setProvider(e.target.value)}
                                className="w-full bg-transparent border border-white text-white font-mono text-sm rounded-none p-4 focus:outline-none focus:ring-1 focus:ring-white cursor-pointer appearance-none"
                                style={{
                                    backgroundImage: `url("data:image/svg+xml;charset=UTF-8,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='white' stroke-width='2' stroke-linecap='square' stroke-linejoin='miter'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E")`,
                                    backgroundRepeat: 'no-repeat',
                                    backgroundPosition: 'right 1rem center',
                                    backgroundSize: '1em'
                                }}
                            >
                                <option value="gcp" className="bg-black text-white">Google Cloud DLP ($5.00 / GB)</option>
                                <option value="aws" className="bg-black text-white">AWS Comprehend (~$1,000 / GB)</option>
                                <option value="azure" className="bg-black text-white">Azure AI Language (~$1.50 / GB)</option>
                            </select>
                        </div>

                        {/* Volume Slider */}
                        <div className="mb-10">
                            <label className="block font-mono text-xs uppercase text-gray-400 mb-3">Data Throughput (Monthly)</label>
                            <div className="flex justify-between items-end mb-4 border-b border-white pb-2">
                                <span className="font-mono text-2xl font-bold">{volume} TB</span>
                                <span className="font-mono text-xs text-gray-400">Volume</span>
                            </div>
                            <input
                                type="range"
                                min="1" max="250"
                                value={volume}
                                onChange={(e) => setVolume(parseInt(e.target.value))}
                                className="w-full"
                            />
                            <div className="flex justify-between font-mono text-[10px] text-gray-400 mt-3">
                                <span>1 TB</span>
                                <span>250 TB</span>
                            </div>
                        </div>

                        {/* Discount Slider */}
                        <div className="mb-10">
                            <label className="block font-mono text-xs uppercase text-gray-400 mb-3">Negotiated EDP Discount</label>
                            <div className="flex justify-between items-end mb-4 border-b border-white pb-2">
                                <span className="font-mono text-2xl font-bold">{discount}%</span>
                                <span className="font-mono text-xs text-gray-400">Variance</span>
                            </div>
                            <input
                                type="range"
                                min="0" max="80" step="5"
                                value={discount}
                                onChange={(e) => setDiscount(parseInt(e.target.value))}
                                className="w-full"
                            />
                            <div className="flex justify-between font-mono text-[10px] text-gray-400 mt-3">
                                <span>BASE [0%]</span>
                                <span>MAX [80%]</span>
                            </div>
                        </div>

                        {/* Effective Rate Output */}
                        <div className="font-mono text-xs p-5 border border-white">
                            <h4 className="font-bold uppercase border-b border-white pb-2 mb-3">Rate Matrix</h4>
                            <div className="flex justify-between mb-2">
                                <span className="text-gray-400">List Price (Scrub + Egress)</span>
                                <span className="line-through">{formatRate(listPricePerGB)} / GB</span>
                            </div>
                            <div className="flex justify-between pt-2">
                                <span className="font-bold">Effective Rate</span>
                                <span className="font-bold">{formatRate(effectiveRate)} / GB</span>
                            </div>
                        </div>
                    </div>

                    {/* Right Col: Computation Results */}
                    <div className="lg:col-span-7 grid grid-cols-1 md:grid-cols-2 gap-8">

                        {/* Cloud Pipeline Node */}
                        <div className="border border-white p-8 flex flex-col justify-between">
                            <div>
                                <div className="flex items-center justify-between mb-6 border-b border-white pb-4">
                                    <h3 className="font-sans text-xl font-bold uppercase">{selectedProvider.name}</h3>
                                    <span className="font-mono text-xs border border-white px-2 py-1">API_EGRESS</span>
                                </div>

                                <div className="font-mono text-sm space-y-4 mb-8">
                                    <div className="flex justify-between border-b border-gray-800 pb-2">
                                        <span className="text-gray-400">Compute Overhead</span>
                                        <span className="line-through">{formatCurrency(totalCloudGrossMonthly)}</span>
                                    </div>
                                    <div className="flex justify-between border-b border-gray-800 pb-2">
                                        <span className="text-gray-400">Discount Applied</span>
                                        <span>-{formatCurrency(discountValueMonthly)}</span>
                                    </div>
                                    <div className="flex justify-between pt-2">
                                        <span className="text-gray-400">Network Latency</span>
                                        <span>100-500ms</span>
                                    </div>
                                </div>
                            </div>

                            <div className="pt-6 border-t border-white">
                                <span className="block font-mono text-xs text-gray-400 uppercase mb-2">Net Financial Drain (Mo)</span>
                                <span className="font-sans text-4xl font-bold tracking-tight">{formatCurrency(totalCloudNetMonthly)}</span>
                            </div>
                        </div>

                        {/* Ocultar Node (Inverted for Dark Mode) */}
                        <div className="border border-white bg-white text-black p-8 flex flex-col justify-between relative">
                            <div>
                                <div className="flex items-center justify-between mb-6 border-b border-black/30 pb-4">
                                    <h3 className="font-sans text-xl font-bold uppercase">Zero-Egress SLM</h3>
                                    <span className="font-mono text-xs border border-black px-2 py-1 font-bold">LOCAL_NODE</span>
                                </div>

                                <div className="font-mono text-sm space-y-4 mb-8">
                                    <div className="flex justify-between border-b border-black/20 pb-2">
                                        <span className="text-gray-600">Software License</span>
                                        <span className="font-bold">$10,000 / mo</span>
                                    </div>
                                    <div className="flex justify-between border-b border-black/20 pb-2">
                                        <span className="text-gray-600">Hardware Dep.</span>
                                        <span>+{formatCurrency(localComputeMonthly)}</span>
                                    </div>
                                    <div className="flex justify-between pt-2">
                                        <span className="text-gray-600">P50 Latency</span>
                                        <span className="font-bold">0.92ms</span>
                                    </div>
                                </div>
                            </div>

                            <div className="pt-6 border-t border-black/30">
                                <span className="block font-mono text-xs text-gray-600 uppercase mb-2 font-bold">Fixed Resource Cost (Mo)</span>
                                <span className="font-sans text-4xl font-bold tracking-tight">{formatCurrency(totalOcultarMonthly)}</span>
                            </div>
                        </div>

                        {/* Differential Banner */}
                        <div className="md:col-span-2 border border-white p-8 flex flex-col sm:flex-row items-center justify-between">
                            <div>
                                <h4 className="font-sans text-xl font-bold uppercase mb-2">Total System Variance</h4>
                                <p className="font-mono text-xs text-gray-400 uppercase tracking-widest">Projected 12-Month Capital Retention</p>
                            </div>
                            <div className="mt-6 sm:mt-0 text-right">
                                <span className="block font-sans text-5xl font-bold tracking-tighter text-white">
                                    {formatCurrency(annualSavings)}
                                </span>
                            </div>
                        </div>

                        {/* CTA Section */}
                        <div className="md:col-span-2 mt-4 text-center border border-white p-10">
                            <h2 className="font-sans text-2xl font-bold uppercase mb-4 tracking-tight">Execute Architecture Upgrade</h2>
                            <p className="font-mono text-sm text-gray-400 max-w-2xl mx-auto mb-8 leading-relaxed">
                                Initiate deployment sequence. Neutralize regulatory risk and terminate external egress taxation via localized execution.
                            </p>

                            <div className="flex flex-col sm:flex-row items-center justify-center gap-4 font-mono text-xs uppercase tracking-widest font-bold">
                                <a href="mailto:sales@ocultar.io?subject=Enterprise%20Pilot%20Request" className="w-full sm:w-auto px-8 py-4 border border-white bg-transparent text-white hover:bg-white hover:text-black transition-colors">
                                    [ Init Pilot Protocol ]
                                </a>
                                <button onClick={() => window.open('https://github.com/Edu963/ocultar', '_blank')} className="w-full sm:w-auto px-8 py-4 border border-white bg-transparent text-white hover:bg-white hover:text-black transition-colors">
                                    [ Audit Source Code ]
                                </button>
                            </div>
                        </div>

                    </div>
                </div>

                {/* Back Link */}
                <div className="mt-16 text-center">
                    <Link to="/" className="text-gray-500 hover:text-white transition-colors text-xs font-mono font-bold uppercase tracking-widest">
                        [ Back to Terminal ]
                    </Link>
                </div>

            </div>
        </div>
    );
};

export default CalculatorPage;
