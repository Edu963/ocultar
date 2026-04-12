import { Shield, ArrowRight, Database, Globe, Lock } from 'lucide-react';

const ArchitectureDiagram = () => {
    return (
        <div className="w-full max-w-4xl mx-auto py-20 px-6 mt-10">
            <div className="flex flex-col md:flex-row items-center justify-between gap-12 relative">
                {/* Source User */}
                <div className="flex flex-col items-center gap-4 z-10 w-full md:w-auto">
                    <div className="w-20 h-20 bg-white/5 border border-white/10 rounded-2xl flex items-center justify-center group hover:border-sky-500/50 transition-colors duration-500">
                        <Database className="w-10 h-10 text-slate-400 group-hover:text-sky-400 transition-colors" />
                    </div>
                    <div className="text-center">
                        <div className="text-xs font-mono text-slate-500 uppercase tracking-widest mb-1">Source</div>
                        <div className="text-sm font-bold text-white">Sensitive PII Data</div>
                    </div>
                </div>

                {/* Connecting Line 1 */}
                <div className="hidden md:flex flex-1 items-center justify-center px-4 relative">
                    <div className="w-full h-px bg-gradient-to-r from-sky-500/20 via-sky-500 to-transparent relative">
                         <div className="absolute top-1/2 left-0 w-2 h-2 bg-sky-400 rounded-full blur-[2px] animate-flow-right"></div>
                    </div>
                </div>

                {/* Ocultar Bridge */}
                <div className="relative group z-20 w-full md:w-auto">
                    <div className="absolute -inset-4 bg-emerald-500/10 rounded-[2rem] blur-xl opacity-0 group-hover:opacity-100 transition-all duration-700"></div>
                    <div className="relative bg-[#0d0d0f] border-2 border-emerald-500/30 p-8 rounded-3xl shadow-[0_0_40px_rgba(16,185,129,0.1)] group-hover:border-emerald-500/60 transition-all duration-500">
                        <div className="flex flex-col items-center gap-4">
                            <div className="w-16 h-16 bg-black border border-emerald-500/50 rounded-xl flex items-center justify-center shadow-[0_0_20px_rgba(16,185,129,0.3)]">
                                <Shield className="w-8 h-8 text-emerald-400 animate-pulse" />
                            </div>
                            <div className="text-center">
                                <div className="text-xs font-mono text-emerald-400 font-bold uppercase tracking-[0.2em] mb-1">Zero-Egress Bridge</div>
                                <div className="text-lg font-bold text-white tracking-tight">OCULTAR REFINERY</div>
                            </div>
                        </div>
                        
                        {/* Status Indicator */}
                        <div className="mt-6 flex items-center justify-center gap-3">
                            <div className="flex gap-1">
                                <span className="w-1 h-3 bg-emerald-500/20 rounded-full"></span>
                                <span className="w-1 h-3 bg-emerald-500/40 rounded-full"></span>
                                <span className="w-1 h-3 bg-emerald-500/60 rounded-full"></span>
                                <span className="w-1 h-5 bg-emerald-500 rounded-full -mt-1"></span>
                            </div>
                            <span className="text-[10px] font-mono text-slate-500 uppercase tracking-widest">Active Interception</span>
                        </div>
                    </div>
                    
                    {/* Security Shield Overlay */}
                    <div className="absolute -top-4 -right-4 w-10 h-10 bg-black border border-emerald-500/30 rounded-full flex items-center justify-center">
                        <Lock className="w-4 h-4 text-emerald-500" />
                    </div>
                </div>

                {/* Connecting Line 2 */}
                <div className="hidden md:flex flex-1 items-center justify-center px-4 relative">
                    <div className="w-full h-px bg-gradient-to-r from-emerald-500 to-transparent relative opacity-40">
                         <div className="absolute top-1/2 left-0 w-2 h-2 bg-emerald-400 rounded-full blur-[2px] animate-flow-right delay-700"></div>
                    </div>
                </div>

                {/* Cloud Provider */}
                <div className="flex flex-col items-center gap-4 z-10 w-full md:w-auto">
                    <div className="w-20 h-20 bg-white/5 border border-white/10 rounded-2xl flex items-center justify-center group hover:border-emerald-500/50 transition-colors duration-500">
                        <Globe className="w-10 h-10 text-slate-500 group-hover:text-emerald-400 transition-colors" />
                    </div>
                    <div className="text-center">
                        <div className="text-xs font-mono text-slate-500 uppercase tracking-widest mb-1">Destination</div>
                        <div className="text-sm font-bold text-slate-400">Cloud LLM (Clean)</div>
                    </div>
                </div>
            </div>

            {/* Labels */}
            <div className="mt-16 grid grid-cols-1 md:grid-cols-3 gap-8">
                <div className="p-4 rounded-xl border border-white/5 bg-white/[0.02]">
                    <div className="text-xs font-mono text-sky-400 mb-2">01. INGEST</div>
                    <div className="text-sm text-slate-400">Raw application payloads containing Names, PII, and Secrets enter the local perimeter.</div>
                </div>
                <div className="p-4 rounded-xl border border-emerald-500/10 bg-emerald-500/[0.02]">
                    <div className="text-xs font-mono text-emerald-400 mb-2">02. REFINERY</div>
                    <div className="text-sm text-slate-400">Local SLM deep scan tokenizes sensitive data. No data leaves your infrastructure raw.</div>
                </div>
                <div className="p-4 rounded-xl border border-white/5 bg-white/[0.02]">
                    <div className="text-xs font-mono text-slate-500 mb-2">03. EGRESS</div>
                    <div className="text-sm text-slate-400">Clean, compliant data is safely forwarded to cloud providers for AI processing.</div>
                </div>
            </div>

            <style dangerouslySetInnerHTML={{ __html: `
                @keyframes flow-right {
                    0% { left: 0%; opacity: 0; }
                    20% { opacity: 1; }
                    80% { opacity: 1; }
                    100% { left: 100%; opacity: 0; }
                }
                .animate-flow-right {
                    animation: flow-right 3s linear infinite;
                }
            `}} />
        </div>
    );
};

export default ArchitectureDiagram;
