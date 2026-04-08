import { useState } from 'react';
import { Shield, Upload, FileText, AlertTriangle, ArrowRight, Loader2 } from 'lucide-react';

const TerminalBlock = ({ title, content, isDark = false, className = "" }) => (
    <div 
        className={`terminal-block ${className} p-6 border border-black/10 bg-white/40 backdrop-blur-sm`} 
        style={isDark ? { backgroundColor: '#000000', color: '#FFFFFF' } : {}}
    >
        {title && (
            <div className={`mono-label mb-4 pb-2 border-b ${isDark ? 'border-white/20' : 'border-black/10'}`} style={isDark ? { color: '#AAAAAA', borderColor: 'rgba(255,255,255,0.2)' } : {}}>
                {title}
            </div>
        )}
        <pre 
            className="font-tech text-xs md:text-sm whitespace-pre-wrap leading-relaxed"
            style={isDark ? { color: '#FFFFFF' } : { color: '#000000' }}
        >
            {content}
        </pre>
    </div>
);

export default function RiskAssessmentPage() {
    const [step, setStep] = useState(1);
    const [inputType, setInputType] = useState('paste'); // 'paste' or 'upload'
    const [pastedData, setPastedData] = useState(`[
  { "name": "John Doe", "email": "john@example.com", "region": "North", "dept": "HR", "salary": 75000 },
  { "name": "Jane Smith", "email": "jane@example.com", "region": "South", "dept": "IT", "salary": 85000 },
  { "name": "Alice Brown", "email": "alice@example.com", "region": "North", "dept": "Sales", "salary": 65000 },
  { "name": "Bob Wilson", "email": "bob@example.com", "region": "East", "dept": "Dev", "salary": 95000 }
]`);
    const [file, setFile] = useState(null);
    const [email, setEmail] = useState('');
    const [company, setCompany] = useState('');
    const [loading, setLoading] = useState(false);
    const [report, setReport] = useState(null);

    const handleStart = () => setStep(2);
    
    const handleProcess = async () => {
        setLoading(true);
        await new Promise(r => setTimeout(r, 1500));
        setStep(3);
        setLoading(false);
    };

    const handleUnlock = async (e) => {
        e.preventDefault();
        setLoading(true);
        try {
            const formData = new FormData();
            formData.append('email', email);
            formData.append('company', company);
            
            if (inputType === 'paste') {
                formData.append('dataset', new Blob([pastedData], { type: 'application/json' }), 'data.json');
            } else {
                formData.append('dataset', file);
            }

            const res = await fetch('/api/pilot-assessment', {
                method: 'POST',
                body: formData
            });
            const data = await res.json();
            if (data.status === 'success') {
                const combinedReport = {
                    ...data.report,
                    report_id: data.report_id,
                    full: data.full_report
                };
                setReport(combinedReport);
                setStep(4);
            } else {
                alert(data.error || 'Failed to analyze data');
            }
        } catch (err) {
            console.error(err);
            alert('Service unavailable');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="page-wrapper">
            
            {/* --- HEADER --- */}
            <div className="space-y-6 text-center py-16 max-w-5xl mx-auto w-full">
                <div className="inline-flex items-center gap-2 px-3 py-1 bg-black text-white text-[10px] font-tech uppercase tracking-[0.2em] mb-4">
                    <Shield className="w-3 h-3" /> Pilot Assessment v3.1
                </div>
                <h1 className="text-4xl md:text-5xl font-bold tracking-tighter uppercase font-hero leading-tight w-full max-w-4xl mx-auto">Test Your Data Risk in 60 Seconds</h1>
                <p className="text-dim max-w-3xl mx-auto font-tech text-lg md:text-xl leading-relaxed">
                    Upload a sample dataset and see your compliance exposure before using AI tools. 
                    Stateless processing. No data stored.
                </p>
            </div>

            {/* --- STEPS --- */}
            <div className="relative min-h-[400px]">
                
                {/* STEP 1: LANDING */}
                {step === 1 && (
                    <div className="flex flex-col items-center gap-12 py-12 animate-in fade-in slide-in-from-bottom-4 duration-700">
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 w-full">
                            {[
                                { id: "feat-upload", icon: <Upload className="w-5 h-5" />, title: "Secure Upload", desc: "Small CSV or JSON samples." },
                                { id: "feat-audit", icon: <Shield className="w-5 h-5" />, title: "Instant Audit", desc: "Immediate K-Anonymity score." },
                                { id: "feat-report", icon: <FileText className="w-5 h-5" />, title: "Risk Report", desc: "Defensible VaR estimations." }
                            ].map((item) => (
                                <div key={item.id} className="p-10 md:p-12 border border-black/10 rounded-xl bg-white/40 space-y-6">
                                    <div className="p-4 w-fit bg-black/5">{item.icon}</div>
                                    <h3 className="font-bold uppercase text-xl md:text-2xl font-hero">{item.title}</h3>
                                    <p className="text-base md:text-lg text-dim leading-relaxed">{item.desc}</p>
                                </div>
                            ))}
                        </div>
                        <button 
                            onClick={handleStart}
                            className="group flex items-center gap-4 bg-black text-white px-16 py-8 rounded-full font-tech uppercase text-lg md:text-xl font-bold tracking-widest hover:scale-105 transition-all shadow-xl"
                        >
                            Start Free Assessment <ArrowRight className="w-6 h-6 group-hover:translate-x-2 transition-transform" />
                        </button>
                    </div>
                )}

                {/* STEP 2: INPUT */}
                {step === 2 && (
                    <div className="space-y-8 animate-in fade-in duration-500">
                        <div className="flex gap-4 border-b border-black/10 pb-4">
                            <button 
                                onClick={() => setInputType('paste')}
                                className={`font-tech text-xs uppercase tracking-widest px-4 py-2 transition-all ${inputType === 'paste' ? 'bg-black text-white' : 'hover:bg-black/5'}`}
                            >
                                Option A: Paste Data
                            </button>
                            <button 
                                onClick={() => setInputType('upload')}
                                className={`font-tech text-xs uppercase tracking-widest px-4 py-2 transition-all ${inputType === 'upload' ? 'bg-black text-white' : 'hover:bg-black/5'}`}
                            >
                                Option B: File Upload
                            </button>
                        </div>

                        <div className="min-h-[300px]">
                            {inputType === 'paste' ? (
                                <div className="space-y-4">
                                    <textarea 
                                        value={pastedData}
                                        onChange={(e) => setPastedData(e.target.value)}
                                        className="w-full h-64 p-6 font-tech text-xs border border-black/10 bg-white/40 focus:bg-white focus:outline-none focus:ring-2 ring-black/5"
                                        placeholder='Paste JSON or CSV here...'
                                    />
                                    <p className="text-[10px] text-dim font-tech uppercase">Tip: Pre-filled with a leaky employee dataset.</p>
                                </div>
                            ) : (
                                <div className="relative border-2 border-dashed border-black/10 h-64 flex flex-col items-center justify-center p-8 text-center gap-4 bg-white/20 overflow-hidden">
                                    <Upload className="w-8 h-8 opacity-20" />
                                    <div className="space-y-1">
                                        <p className="text-sm font-bold">Drop your file here or <span className="underline cursor-pointer">browse</span></p>
                                        <p className="text-xs text-dim">Supports .csv, .json, .txt (Max 100KB)</p>
                                    </div>
                                    <input 
                                        type="file" 
                                        onChange={(e) => setFile(e.target.files[0])}
                                        className="absolute inset-0 opacity-0 cursor-pointer z-10"
                                        accept=".csv,.json,.txt"
                                    />
                                    {file && <div className="relative z-20 text-xs font-tech bg-black text-white px-3 py-1 uppercase">{file.name}</div>}
                                </div>
                            )}
                        </div>

                        <div className="bg-yellow-50 border border-yellow-200 p-4 rounded-lg flex gap-3">
                            <AlertTriangle className="w-4 h-4 text-yellow-600 shrink-0" />
                            <p className="text-[10px] text-yellow-800 uppercase tracking-tight font-tech">
                                Disclaimer: Do not upload sensitive or production data. This tool is for demonstration purposes only.
                            </p>
                        </div>

                        <button 
                            onClick={handleProcess}
                            disabled={loading || (inputType === 'upload' && !file)}
                            className="w-full bg-black text-white py-4 font-tech uppercase text-xs tracking-widest flex items-center justify-center gap-2 disabled:opacity-50"
                        >
                            {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Run Risk Analysis'}
                        </button>
                    </div>
                )}

                {/* STEP 3: LEAD GATE */}
                {step === 3 && (
                    <div className="max-w-md mx-auto space-y-8 py-12 animate-in fade-in duration-500">
                        <div className="text-center space-y-2">
                            <div className="text-red-600 font-tech text-xs uppercase tracking-[0.2em] animate-pulse">Critical Risk Detected</div>
                            <h2 className="text-2xl font-bold uppercase font-hero">Unlock Full Report</h2>
                            <p className="text-xs text-dim font-tech">Your analysis is ready. Enter your work details to see the exposure values.</p>
                        </div>

                        <form onSubmit={handleUnlock} className="space-y-4">
                            <div className="space-y-1">
                                <label className="text-[9px] uppercase font-tech text-dim">Work Email</label>
                                <input 
                                    type="email" required
                                    value={email} onChange={(e) => setEmail(e.target.value)}
                                    className="w-full p-4 border border-black/10 font-tech text-sm focus:outline-none focus:ring-2 ring-black/5 bg-white/40"
                                    placeholder="you@company.com"
                                />
                            </div>
                            <div className="space-y-1">
                                <label className="text-[9px] uppercase font-tech text-dim">Company Name</label>
                                <input 
                                    type="text"
                                    value={company} onChange={(e) => setCompany(e.target.value)}
                                    className="w-full p-4 border border-black/10 font-tech text-sm focus:outline-none focus:ring-2 ring-black/5 bg-white/40"
                                    placeholder="Acme Corp"
                                />
                            </div>
                            <button 
                                className="w-full bg-black text-white py-4 font-tech uppercase text-xs tracking-widest flex items-center justify-center gap-2 disabled:opacity-50"
                                disabled={loading}
                            >
                                {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Unlock My Results'}
                            </button>
                        </form>
                    </div>
                )}

                {/* STEP 4: REPORT PREVIEW */}
                {step === 4 && report && (
                    <div className="space-y-8 animate-in fade-in duration-500">
                        <div className="flex justify-between items-center bg-black text-white p-8">
                            <div className="space-y-1">
                                <div className="text-[10px] font-tech uppercase opacity-60">Overall Risk Level</div>
                                <div className="text-4xl font-bold uppercase tracking-tighter font-hero">{report.overall_risk_level}</div>
                            </div>
                            <div className="text-right space-y-1">
                                <div className="text-[10px] font-tech uppercase opacity-60">Compliance Status</div>
                                <div className={`text-xs font-bold px-3 py-1 rounded-full uppercase ${report.is_gdpr_pseudonymized ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
                                    {report.is_gdpr_pseudonymized ? 'Pseudonymized' : 'Non-Compliant'}
                                </div>
                            </div>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                            <div className="p-8 border border-black/10 bg-white/40 space-y-4">
                                <h3 className="font-tech text-xs uppercase text-dim">Estimated Financial Exposure</h3>
                                <div className="flex items-baseline gap-2">
                                    <span className="text-3xl font-bold text-red-600 font-hero">€{Math.round(report.financial_exposure.var_min_eur).toLocaleString()} – €{Math.round(report.financial_exposure.var_max_eur).toLocaleString()}</span>
                                </div>
                                <p className="text-[10px] text-dim leading-relaxed uppercase font-tech">
                                    Total Value at Risk based on industry metrics.
                                </p>
                            </div>
                            
                            <div className="p-8 border border-black/10 bg-white/40 space-y-4">
                                <h3 className="font-tech text-xs uppercase text-dim">AI Readiness</h3>
                                <div className="flex items-center gap-3">
                                    <div className={`p-2 ${report.ai_readiness.status === 'ALLOW' ? 'bg-green-100 text-green-600' : 'bg-red-100 text-red-600'}`}>
                                        <Shield className="w-5 h-5" />
                                    </div>
                                    <span className="font-bold uppercase text-sm font-hero">{report.ai_readiness.status}</span>
                                </div>
                                <p className="text-xs text-dim leading-relaxed">
                                    {report.ai_readiness.recommendation}
                                </p>
                            </div>
                        </div>

                        {/* --- FULL REPORT NATIVE RENDER --- */}
                        {report.full && (
                            <div className="space-y-12 animate-in fade-in slide-in-from-bottom-2 duration-1000">
                                
                                {/* 1. RISK SCORECARD */}
                                <div className="space-y-6">
                                    <h2 className="text-xl font-bold uppercase font-hero border-b border-black/10 pb-4">Risk Scorecard v{report.full.Meta.MethodologyVersion}</h2>
                                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                                        {[
                                            { label: "Identifiability", data: report.full.Risk.identifiability_risk },
                                            { label: "Financial", data: report.full.Risk.financial_sensitivity },
                                            { label: "Re-id Risk", data: report.full.Risk.reidentification_risk },
                                            { label: "Compliance", data: report.full.Risk.compliance_readiness }
                                        ].map((stat) => (
                                            <div key={stat.label} className="border border-black/10 p-6 flex flex-col gap-2">
                                                <span className="font-tech text-[10px] uppercase text-dim">{stat.label}</span>
                                                <span className="text-2xl font-bold font-hero">{stat.data.score.toFixed(1)}/10</span>
                                                <span className={`text-[9px] font-bold uppercase px-2 py-0.5 w-fit ${stat.data.score > 7 ? 'bg-red-100 text-red-600' : stat.data.score > 4 ? 'bg-yellow-100 text-yellow-600' : 'bg-green-100 text-green-600'}`}>
                                                    {stat.data.label}
                                                </span>
                                            </div>
                                        ))}
                                    </div>
                                </div>

                                {/* 2. FINANCIAL EXPOSURE BREAKDOWN */}
                                <div className="space-y-6">
                                    <h2 className="text-xl font-bold uppercase font-hero border-b border-black/10 pb-4">Exposure Components</h2>
                                    <div className="overflow-x-auto border border-black/10">
                                        <table className="w-full text-left font-tech text-[11px]">
                                            <thead>
                                                <tr className="bg-black/5 uppercase">
                                                    <th className="p-4">Component</th>
                                                    <th className="p-4">Methodology</th>
                                                    <th className="p-4 text-right">Min Est.</th>
                                                    <th className="p-4 text-right">Max Est.</th>
                                                </tr>
                                            </thead>
                                            <tbody className="divide-y divide-black/10">
                                                <tr>
                                                    <td className="p-4 font-bold">Regulatory</td>
                                                    <td className="p-4 text-dim">Risk Score Multiplier</td>
                                                    <td className="p-4 text-right">€{Math.round(report.full.Risk.financial_exposure.regulatory_exposure_min_eur).toLocaleString()}</td>
                                                    <td className="p-4 text-right">€{Math.round(report.full.Risk.financial_exposure.regulatory_exposure_max_eur).toLocaleString()}</td>
                                                </tr>
                                                <tr>
                                                    <td className="p-4 font-bold">Operational</td>
                                                    <td className="p-4 text-dim">Recovery Cost Benchmarks</td>
                                                    <td className="p-4 text-right">€{Math.round(report.full.Risk.financial_exposure.operational_cost_min_eur).toLocaleString()}</td>
                                                    <td className="p-4 text-right">€{Math.round(report.full.Risk.financial_exposure.operational_cost_max_eur).toLocaleString()}</td>
                                                </tr>
                                                <tr className="bg-black/5 font-bold">
                                                    <td className="p-4 text-xs">TOTAL VaR</td>
                                                    <td className="p-4 text-dim">Combined Aggregation</td>
                                                    <td className="p-4 text-right text-xs">€{Math.round(report.full.Risk.financial_exposure.var_min_eur).toLocaleString()}</td>
                                                    <td className="p-4 text-right text-xs text-red-600">€{Math.round(report.full.Risk.financial_exposure.var_max_eur).toLocaleString()}</td>
                                                </tr>
                                            </tbody>
                                        </table>
                                    </div>
                                </div>

                                {/* 3. SIMULATION MATRIX */}
                                <div className="space-y-6">
                                    <div className="flex justify-between items-center border-b border-black/10 pb-4">
                                        <h2 className="text-xl font-bold uppercase font-hero">Ocultar Transformation Simulation</h2>
                                        <div className="flex gap-2">
                                            <div className="bg-red-500 w-2 h-2 rounded-full animate-pulse"></div>
                                            <div className="bg-green-500 w-2 h-2 rounded-full"></div>
                                        </div>
                                    </div>
                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-0 border border-black/10">
                                        {[report.full.Before, report.full.After].map((scenario, idx) => (
                                            <div key={scenario.Label} className={`p-8 space-y-6 ${idx === 0 ? 'border-r border-black/10' : 'bg-black/5'}`}>
                                                <div className="flex justify-between items-center">
                                                    <span className={`font-tech text-xs uppercase font-bold ${idx === 0 ? 'text-red-500' : 'text-green-600'}`}>{scenario.Label}</span>
                                                    <span className="text-[10px] font-tech text-dim uppercase">{scenario.RiskLevel}</span>
                                                </div>
                                                <div className="space-y-4">
                                                    <div className="flex justify-between border-b border-black/5 pb-2">
                                                        <span className="font-tech text-[10px] text-dim uppercase">Exposure Range</span>
                                                        <span className="font-bold text-sm tracking-tight">{scenario.VaRRange}</span>
                                                    </div>
                                                    <div className="flex justify-between border-b border-black/5 pb-2">
                                                        <span className="font-tech text-[10px] text-dim uppercase">Status</span>
                                                        <span className="font-bold text-sm tracking-tight uppercase">{scenario.AIStatus}</span>
                                                    </div>
                                                </div>
                                                <p className="text-xs text-dim leading-relaxed h-12 overflow-hidden">{scenario.Description}</p>
                                            </div>
                                        ))}
                                    </div>
                                </div>

                                {/* 4. REMEDIATION PLAN */}
                                <div className="bg-yellow-50 border border-yellow-200 p-8 space-y-4">
                                    <div className="flex items-center gap-3 text-yellow-800">
                                        <AlertTriangle className="w-5 h-5" />
                                        <h3 className="font-bold uppercase font-hero">Immediate Remediation Plan</h3>
                                    </div>
                                    <div className="font-tech text-xs text-yellow-900 leading-relaxed whitespace-pre-wrap">
                                        {report.full.Risk.recommendation}
                                    </div>
                                </div>
                            </div>
                        )}

                        {/* CONVERSION LAYER */}
                        <div className="bg-black text-white p-12 text-center space-y-8 relative overflow-hidden">
                            <div className="absolute top-0 right-0 p-12 opacity-10">
                                <Shield className="w-64 h-64" />
                            </div>
                            <div className="relative z-10 space-y-4">
                                <h2 className="text-3xl md:text-5xl font-bold uppercase tracking-tighter font-hero">View Full Enterprise Audit</h2>
                                <p className="text-dim max-w-xl mx-auto font-tech text-sm">
                                    Your high-fidelity compliance mapping, re-identification probability heatmaps, and a multi-stage remediation plan have been compiled inside your secure vault.
                                </p>
                                <div className="flex flex-col md:flex-row gap-4 justify-center pt-8">
                                    {report.report_id && (
                                        <a href={`/api/pilot/report?id=${report.report_id}`} target="_blank" rel="noopener noreferrer" className="border border-white/20 bg-white/5 backdrop-blur-md px-8 py-4 rounded-full font-tech uppercase text-xs tracking-widest hover:bg-white/10 transition-all">
                                            View PDF/HTML Report
                                        </a>
                                    )}
                                    <a href="mailto:sales@ocultar.dev?subject=Enterprise%20Trial%20Request" className="bg-white text-black px-8 py-4 rounded-full font-tech uppercase text-xs tracking-widest hover:scale-105 transition-all shadow-[0_0_40px_rgba(255,255,255,0.3)]">
                                        Activate 2-Week Enterprise Trial
                                    </a>
                                </div>
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
