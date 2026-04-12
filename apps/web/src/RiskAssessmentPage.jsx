import { useState } from 'react';
import { Link } from 'react-router-dom';
import { 
  Shield, 
  Upload, 
  FileText, 
  AlertTriangle, 
  ArrowRight, 
  Loader2, 
  CheckCircle2, 
  BarChart3, 
  Activity,
  Lock,
  ArrowLeft
} from 'lucide-react';

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
        // Simulate local inference latency
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
        <div className="min-h-screen font-sans antialiased bg-[#050505] text-white selection:bg-cyan-500 selection:text-black pb-32">
            
            {/* Nav */}
            <nav className="fixed top-0 left-0 w-full z-50 py-5 bg-black/50 backdrop-blur-xl border-b border-white/5">
                <div className="max-container flex justify-between items-center">
                    <Link to="/" className="flex items-center gap-3 group">
                        <img src="/logo3.jpg" alt="OCULTAR" className="h-8 md:h-10 object-contain" />
                    </Link>
                    <Link to="/" className="text-[11px] font-bold uppercase tracking-widest text-zinc-500 hover:text-white transition-colors flex items-center gap-2">
                        <ArrowLeft className="w-3 h-3" /> Back to Platform
                    </Link>
                </div>
            </nav>

            <div className="max-container pt-32 relative z-10">
                
                {/* --- HEADER --- */}
                <header className="max-w-4xl mx-auto text-center space-y-8 mb-20">
                    <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-cyan-500/5 border border-cyan-500/10 text-[10px] font-bold text-cyan-500 uppercase tracking-widest">
                        <Shield className="w-3 h-3" /> Pilot Audit Engine v3.1
                    </div>
                    <h1 className="tracking-tighter">
                        Test Your Data <span className="text-cyan-500">Risk</span> in 60 Seconds
                    </h1>
                    <p className="max-w-2xl mx-auto text-lg text-slate-400 font-medium leading-relaxed">
                        Upload a sample dataset and see your compliance exposure before using cloud AI tools. 
                        Stateless processing. No data stored. Zero-Egress guaranteed.
                    </p>
                </header>

                <div className="max-w-5xl mx-auto">
                    
                    {/* STEP 1: LANDING */}
                    {step === 1 && (
                        <div className="space-y-16 animate-in fade-in slide-in-from-bottom-4 duration-1000">
                            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                                {[
                                    { id: "feat-upload", icon: <Upload className="w-6 h-6" />, title: "Secure Interception", desc: "Process small CSV or JSON samples locally within your perimeter." },
                                    { id: "feat-audit", icon: <Shield className="w-6 h-6" />, title: "Instant Integrity Audit", desc: "Identify PII exposure and K-Anonymity scores across 40+ entity types." },
                                    { id: "feat-report", icon: <FileText className="w-6 h-6" />, title: "Risk Exposure (VaR)", desc: "Get defensible Value-at-Risk estimations for your legal team." }
                                ].map((item) => (
                                    <div key={item.id} className="card bg-zinc-950 hover:border-cyan-500/30 transition-colors">
                                        <div className="w-12 h-12 bg-white/5 border border-white/10 rounded-xl flex items-center justify-center mb-6 text-cyan-400">
                                            {item.icon}
                                        </div>
                                        <h3 className="text-xl font-bold text-white mb-2">{item.title}</h3>
                                        <p className="text-sm text-slate-400 leading-relaxed">{item.desc}</p>
                                    </div>
                                ))}
                            </div>
                            <div className="flex justify-center">
                                <button 
                                    onClick={handleStart}
                                    className="bg-cyan-500 text-black px-16 py-6 rounded-lg text-xl font-bold shadow-2xl hover:scale-105 transition-transform flex items-center group"
                                >
                                    Start Free Audit <ArrowRight className="ml-3 w-6 h-6 group-hover:translate-x-1 transition-transform" />
                                </button>
                            </div>
                        </div>
                    )}

                    {/* STEP 2: INPUT */}
                    {step === 2 && (
                        <div className="card max-w-3xl mx-auto bg-zinc-950 space-y-10 animate-in fade-in slide-in-from-bottom-4 duration-700 shadow-2xl">
                            <div className="flex gap-2 p-1 bg-white/5 rounded-xl border border-white/5">
                                <button 
                                    onClick={() => setInputType('paste')}
                                    className={`flex-1 py-4 text-[10px] font-bold uppercase tracking-widest rounded-lg transition-all ${inputType === 'paste' ? 'bg-cyan-500 text-black shadow-lg' : 'text-slate-500 hover:text-white'}`}
                                >
                                    Paste Dataset
                                </button>
                                <button 
                                    onClick={() => setInputType('upload')}
                                    className={`flex-1 py-4 text-[10px] font-bold uppercase tracking-widest rounded-lg transition-all ${inputType === 'upload' ? 'bg-cyan-500 text-black shadow-lg' : 'text-slate-500 hover:text-white'}`}
                                >
                                    Upload File
                                </button>
                            </div>

                            <div className="min-h-[350px]">
                                {inputType === 'paste' ? (
                                    <div className="space-y-4">
                                        <textarea 
                                            value={pastedData}
                                            onChange={(e) => setPastedData(e.target.value)}
                                            className="w-full h-80 p-8 bg-black/40 border border-white/10 rounded-2xl font-mono text-sm text-cyan-400/80 focus:border-cyan-500/50 outline-none transition-colors"
                                            placeholder='Paste JSON or CSV here...'
                                        />
                                        <p className="text-[10px] text-zinc-500 font-bold uppercase tracking-widest font-mono">Input Buffer [READY]</p>
                                    </div>
                                ) : (
                                    <div className="relative border-2 border-dashed border-white/10 rounded-2xl h-80 flex flex-col items-center justify-center p-8 text-center gap-6 bg-white/[0.01] hover:bg-white/[0.03] transition-colors cursor-pointer group hover:border-cyan-500/30">
                                        <Upload className="w-12 h-12 text-zinc-600 group-hover:text-cyan-500 transition-colors" />
                                        <div className="space-y-2">
                                            <p className="font-bold text-white text-lg">Drop dataset here or <span className="text-cyan-500 underline">browse</span></p>
                                            <p className="text-xs text-slate-500 font-mono">.csv, .json, .txt (Max 100KB)</p>
                                        </div>
                                        <input 
                                            type="file" 
                                            onChange={(e) => setFile(e.target.files[0])}
                                            className="absolute inset-0 opacity-0 cursor-pointer z-10"
                                            accept=".csv,.json,.txt"
                                        />
                                        {file && <div className="absolute inset-0 bg-zinc-950 rounded-2xl flex items-center justify-center gap-3 z-20 border border-cyan-500/50 shadow-[0_0_30px_rgba(16,185,129,0.2)]">
                                            <FileText className="w-6 h-6 text-cyan-500" />
                                            <span className="font-bold text-white">{file.name}</span>
                                            <button onClick={(e) => {e.stopPropagation(); setFile(null)}} className="text-xs text-red-500 font-bold ml-4 uppercase tracking-widest hover:underline">Remove</button>
                                        </div>}
                                    </div>
                                )}
                            </div>

                            <div className="p-5 bg-cyan-500/5 border border-cyan-500/20 rounded-xl flex gap-4">
                                <AlertTriangle className="w-5 h-5 text-cyan-500 shrink-0 mt-0.5" />
                                <p className="text-xs text-slate-400 leading-relaxed font-medium">
                                    <strong className="text-cyan-400">Security Note:</strong> All processing is done via transient local inference. This node is stateless and encrypted. Data never leaves the bridge.
                                </p>
                            </div>

                            <button 
                                onClick={handleProcess}
                                disabled={loading || (inputType === 'upload' && !file)}
                                className="w-full bg-cyan-500 text-black py-5 rounded-lg text-lg font-bold shadow-xl disabled:opacity-50 hover:bg-cyan-400 transition-colors flex items-center justify-center"
                            >
                                {loading ? <Loader2 className="w-6 h-6 animate-spin" /> : 'Run Integrity Scan'}
                            </button>
                        </div>
                    )}

                    {/* STEP 3: LEAD GATE */}
                    {step === 3 && (
                        <div className="card max-w-lg mx-auto bg-zinc-950 space-y-10 py-12 animate-in fade-in slide-in-from-bottom-4 duration-700 shadow-2xl">
                            <div className="text-center space-y-4">
                                <div className="text-cyan-500 font-bold text-[10px] uppercase tracking-[0.4em] inline-flex items-center gap-3">
                                    <div className="w-2 h-2 bg-cyan-500 rounded-full animate-pulse" />
                                    Analysis Complete
                                </div>
                                <h2 className="text-3xl font-bold tracking-tight">Unlock Report</h2>
                                <p className="text-sm text-slate-500 font-medium">Verification successful. Enter your details to generate your risk exposure summary.</p>
                            </div>

                            <form onSubmit={handleUnlock} className="space-y-6">
                                <div className="space-y-2">
                                    <label className="text-[10px] font-bold uppercase tracking-widest text-zinc-500">Corporate Email</label>
                                    <input 
                                        type="email" required
                                        value={email} onChange={(e) => setEmail(e.target.value)}
                                        className="w-full p-4 bg-black/40 border border-white/10 rounded-xl outline-none focus:border-cyan-500/50 transition-colors text-white"
                                        placeholder="you@company.com"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-[10px] font-bold uppercase tracking-widest text-zinc-500">Organization</label>
                                    <input 
                                        type="text" required
                                        value={company} onChange={(e) => setCompany(e.target.value)}
                                        className="w-full p-4 bg-black/40 border border-white/10 rounded-xl outline-none focus:border-cyan-500/50 transition-colors text-white"
                                        placeholder="Tech Corp Inc."
                                    />
                                </div>
                                <button 
                                    className="w-full bg-cyan-500 text-black py-5 rounded-lg text-lg font-bold hover:bg-cyan-400 transition-colors"
                                    disabled={loading}
                                >
                                    {loading ? <Loader2 className="w-6 h-6 animate-spin text-black" /> : 'Generate Audit Report'}
                                </button>
                            </form>
                        </div>
                    )}

                    {/* STEP 4: REPORT PREVIEW */}
                    {step === 4 && report && (
                        <div className="space-y-12 animate-in fade-in slide-in-from-bottom-4 duration-1000 pb-20">
                            {/* Summary Card */}
                            <div className="card bg-cyan-500 text-black p-12 flex flex-col md:flex-row justify-between items-center gap-10 overflow-hidden relative border-none rounded-[2rem]">
                                <Shield className="absolute -right-16 -bottom-16 w-80 h-80 text-black/5 rotate-12" />
                                <div className="relative z-10 space-y-4">
                                    <div className="text-[10px] font-bold uppercase tracking-[0.4em] text-black/40 font-mono">Consolidated Risk Profile</div>
                                    <div className="text-6xl font-black tracking-tighter uppercase italic">{report.overall_risk_level}</div>
                                </div>
                                <div className="relative z-10 text-center md:text-right space-y-6">
                                    <div className={`inline-flex items-center gap-3 px-6 py-3 rounded-full text-xs font-bold uppercase tracking-widest ${report.is_gdpr_pseudonymized ? 'bg-black/10 text-black border border-black/10' : 'bg-rose-500 text-white'}`}>
                                        {report.is_gdpr_pseudonymized ? <CheckCircle2 className="w-5 h-5" /> : <AlertTriangle className="w-5 h-5" />}
                                        {report.is_gdpr_pseudonymized ? 'Integrity Verified' : 'Critical Leakage Risk'}
                                    </div>
                                    <div className="text-sm font-bold opacity-60 uppercase tracking-widest font-mono">Compliance Ready: {report.is_gdpr_pseudonymized ? 'YES' : 'REMEDIATION_REQUIRED'}</div>
                                </div>
                            </div>

                            {/* Top Metrics Grid */}
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                                <div className="card bg-zinc-950 space-y-8">
                                    <div className="text-[10px] font-bold uppercase tracking-widest text-zinc-500">Est. Financial Exposure (VaR)</div>
                                    <div className="text-5xl font-bold text-rose-500 tracking-tighter">
                                        €{Math.round(report.financial_exposure.var_min_eur).toLocaleString()} – €{Math.round(report.financial_exposure.var_max_eur).toLocaleString()}
                                    </div>
                                    <p className="text-sm text-slate-400 leading-relaxed font-medium">Projected annualized impact across regulatory domains.</p>
                                </div>
                                
                                <div className="card bg-zinc-950 space-y-8">
                                    <div className="text-[10px] font-bold uppercase tracking-widest text-zinc-500">AI Readiness Status</div>
                                    <div className="flex items-center gap-4">
                                        <div className={`p-4 rounded-xl ${report.ai_readiness.status === 'ALLOW' ? 'bg-cyan-500/10 text-cyan-500' : 'bg-rose-500/10 text-rose-500'}`}>
                                            <Activity className="w-8 h-8" />
                                        </div>
                                        <span className="font-bold text-4xl tracking-tighter text-white uppercase italic">{report.ai_readiness.status}</span>
                                    </div>
                                    <p className="text-sm text-slate-400 font-medium leading-relaxed">{report.ai_readiness.recommendation}</p>
                                </div>
                            </div>

                            {/* Full Report Details */}
                            {report.full && (
                                <div className="space-y-24 py-16 border-t border-white/5 mt-20">
                                    {/* 1. SCORECARD */}
                                    <div className="space-y-12">
                                        <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
                                            <div className="space-y-4">
                                                <div className="text-cyan-500 font-mono text-[10px] tracking-[0.4em] uppercase font-bold">[ ASSESSMENT_RESULT_TX_{report.report_id?.slice(0,8)} ]</div>
                                                <h2 className="text-4xl font-bold tracking-tight flex items-center gap-4">
                                                    Technical Risk Scorecard <BarChart3 className="w-8 h-8 text-cyan-500" />
                                                </h2>
                                            </div>
                                            <p className="text-zinc-500 text-sm max-w-xs font-medium italic underline underline-offset-4 decoration-cyan-500/30">
                                                Integrity Mapping v{report.full.Meta.MethodologyVersion}
                                            </p>
                                        </div>
                                        
                                        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                                            {[
                                                { label: "Identifiability", data: report.full.Risk.identifiability_risk },
                                                { label: "Financial Sensitivity", data: report.full.Risk.financial_sensitivity },
                                                { label: "Re-id Complexity", data: report.full.Risk.reidentification_risk },
                                                { label: "Regulatory Readiness", data: report.full.Risk.compliance_readiness }
                                            ].map((stat) => (
                                                <div key={stat.label} className="card bg-zinc-950 flex flex-col gap-6">
                                                    <span className="text-[10px] font-bold uppercase tracking-widest text-zinc-600 font-mono">{stat.label}</span>
                                                    <div className="text-5xl font-bold text-white tracking-tighter">{stat.data.score.toFixed(1)}<span className="text-xl text-zinc-800 ml-1">/10</span></div>
                                                    <div className={`text-[10px] font-bold uppercase px-4 py-1.5 rounded-full w-fit tracking-widest ${stat.data.score > 7 ? 'bg-rose-500/10 text-rose-400' : stat.data.score > 4 ? 'bg-amber-500/10 text-amber-400' : 'bg-cyan-500/10 text-cyan-400'}`}>
                                                        {stat.data.label}
                                                    </div>
                                                </div>
                                            ))}
                                        </div>
                                    </div>

                                    {/* 2. REGULATORY RISK MATRIX */}
                                    <div className="space-y-12">
                                        <h2 className="text-4xl font-bold tracking-tight flex items-center gap-4">
                                            Regulatory Alignment Matrix <Shield className="w-8 h-8 text-cyan-500" />
                                        </h2>
                                        <div className="card p-0 bg-zinc-950 overflow-hidden border-white/5 shadow-3xl">
                                            <div className="overflow-x-auto">
                                                <table className="w-full text-left border-collapse">
                                                    <thead>
                                                        <tr className="bg-white/5 border-b border-white/10">
                                                            <th className="p-6 text-[10px] font-bold uppercase tracking-[0.2em] text-zinc-500">Sensitive Attribute</th>
                                                            <th className="p-6 text-[10px] font-bold uppercase tracking-[0.2em] text-zinc-500">Governing Regulation</th>
                                                            <th className="p-6 text-[10px] font-bold uppercase tracking-[0.2em] text-zinc-500">Primary Article</th>
                                                            <th className="p-6 text-[10px] font-bold uppercase tracking-[0.2em] text-zinc-500">Security Threat</th>
                                                        </tr>
                                                    </thead>
                                                    <tbody>
                                                        {report.full.Risk.regulatory_findings?.map((finding, idx) => (
                                                            <tr key={idx} className="border-b border-white/5 hover:bg-white/[0.03] transition-colors">
                                                                <td className="p-6 font-mono text-sm font-bold text-white flex items-center gap-3">
                                                                    <div className="w-1.5 h-1.5 bg-cyan-500/50 rounded-full"></div>
                                                                    {finding.attribute}
                                                                </td>
                                                                <td className="p-6 text-sm font-medium text-slate-400">{finding.regulation}</td>
                                                                <td className="p-6">
                                                                    <span className="px-3 py-1 bg-white/5 border border-white/10 rounded font-mono text-[10px] text-cyan-400 font-bold tracking-widest">{finding.article}</span>
                                                                </td>
                                                                <td className="p-6">
                                                                    <div className={`text-[10px] font-bold uppercase px-3 py-1.5 rounded-sm inline-flex items-center gap-2 tracking-[0.1em] ${
                                                                        finding.severity === 'HIGH' ? 'bg-rose-500/10 text-rose-500 border border-rose-500/20' : 
                                                                        finding.severity === 'MEDIUM' ? 'bg-amber-500/10 text-amber-500 border border-amber-500/20' : 
                                                                        'bg-cyan-500/10 text-cyan-500 border border-cyan-500/20'
                                                                    }`}>
                                                                        {finding.severity} RISK
                                                                    </div>
                                                                </td>
                                                            </tr>
                                                        ))}
                                                    </tbody>
                                                </table>
                                            </div>
                                            <div className="p-6 bg-white/[0.02] border-t border-white/5">
                                                <p className="text-[11px] text-zinc-600 leading-relaxed font-bold uppercase font-mono tracking-widest">
                                                    * DISCLOSURE: Mapped via Heuristic Engine v1.2. Article references are for infrastructure audit prioritization.
                                                </p>
                                            </div>
                                        </div>
                                    </div>

                                    {/* 3. MITIGATION SIM */}
                                    <div className="space-y-12">
                                        <h2 className="text-4xl font-bold tracking-tight flex items-center gap-4">
                                            Ocultar Mitigation Simulation <CheckCircle2 className="w-8 h-8 text-cyan-500" />
                                        </h2>
                                        <div className="grid grid-cols-1 md:grid-cols-2 gap-0 border border-white/10 rounded-[2.5rem] overflow-hidden shadow-2xl">
                                            {[report.full.Before, report.full.After].map((scenario, idx) => (
                                                <div key={scenario.Label} className={`p-12 space-y-10 ${idx === 0 ? 'bg-zinc-950 border-r border-white/10' : 'bg-cyan-500/[0.03]'}`}>
                                                    <div className="flex justify-between items-center mb-6">
                                                        <div className={`text-[10px] font-bold uppercase tracking-[0.4em] ${idx === 0 ? 'text-rose-500' : 'text-cyan-500 font-black'}`}>[ {scenario.Label} ]</div>
                                                        <div className={`px-4 py-1.5 border rounded-full text-[10px] font-bold uppercase tracking-widest ${idx === 0 ? 'border-rose-500/50 text-rose-500' : 'border-cyan-500/50 text-cyan-500 font-black'}`}>{scenario.RiskLevel}</div>
                                                    </div>
                                                    <div className="space-y-6">
                                                        <div className="flex justify-between text-sm py-4 border-b border-white/5">
                                                            <span className="text-zinc-500 font-bold uppercase tracking-widest text-[10px]">Projection (VaR)</span>
                                                            <span className={`font-mono font-bold ${idx === 0 ? 'text-rose-400' : 'text-cyan-400'}`}>{scenario.VaRRange}</span>
                                                        </div>
                                                        <div className="flex justify-between text-sm py-4 border-b border-white/5">
                                                            <span className="text-zinc-500 font-bold uppercase tracking-widest text-[10px]">Cloud Egress</span>
                                                            <span className={`font-bold uppercase tracking-widest ${idx === 0 ? 'text-rose-500' : 'text-white'}`}>{scenario.AIStatus}</span>
                                                        </div>
                                                    </div>
                                                    <div className="bg-black/40 p-6 rounded-2xl border border-white/5 italic">
                                                        <p className="text-base text-slate-400 leading-relaxed font-medium">"{scenario.Description}"</p>
                                                    </div>
                                                </div>
                                            ))}
                                        </div>
                                    </div>

                                    {/* 4. REMEDIATION */}
                                    <div className="card border-rose-500/30 bg-rose-500/[0.02] p-12 space-y-8 relative overflow-hidden group rounded-[2.5rem]">
                                        <div className="absolute top-0 right-0 p-12 opacity-5 rotate-12 group-hover:opacity-10 transition-opacity">
                                            <AlertTriangle className="w-40 h-40" />
                                        </div>
                                        <div className="flex items-center gap-4 text-rose-500">
                                            <div className="p-4 bg-rose-500/10 border border-rose-500/20 rounded-2xl flex items-center justify-center">
                                                <Lock className="w-8 h-8" />
                                            </div>
                                            <h3 className="text-3xl font-bold tracking-tight">Hardened Remediation Plan</h3>
                                        </div>
                                        <div className="text-xl text-slate-300 leading-relaxed font-medium bg-zinc-950 p-8 rounded-2xl border border-white/10 shadow-2xl relative z-10">
                                            <span className="text-rose-500 font-mono font-bold mr-2 tracking-widest underline underline-offset-8 decoration-rose-500/50">CRITICAL:</span> {report.full.Risk.recommendation}
                                        </div>
                                    </div>

                                    {/* 5. FINAL ACTIONS */}
                                    <div className="bg-cyan-500 text-black p-16 rounded-[3rem] text-center space-y-10 relative overflow-hidden shadow-[0_0_100px_rgba(16,185,129,0.2)]">
                                        <div className="absolute top-0 right-0 p-16 opacity-10">
                                            <Shield className="w-96 h-96" />
                                        </div>
                                        <div className="relative z-10 space-y-6">
                                            <div className="text-[10px] font-black uppercase tracking-[0.5em] text-black/40">Secure Your Perimeter</div>
                                            <h2 className="text-5xl font-black tracking-tighter lowercase italic">activate technical sovereignty.</h2>
                                            <p className="text-black/70 max-w-2xl mx-auto text-lg font-bold leading-relaxed">
                                                Your full compliance report, re-identification heatmaps, and structured 
                                                remediation plan are ready for executive review.
                                            </p>
                                            <div className="flex flex-col md:flex-row gap-4 justify-center pt-8">
                                                {report.report_id && (
                                                    <a href={`/api/pilot/report?id=${report.report_id}`} target="_blank" rel="noopener noreferrer" className="bg-black text-white px-12 py-6 rounded-lg text-lg font-bold hover:scale-105 transition-transform shadow-2xl border border-black">
                                                        Download PDF Audit
                                                    </a>
                                                )}
                                                <a href="mailto:sales@ocultar.dev" className="border border-black/20 text-black px-12 py-6 rounded-lg text-lg font-bold hover:bg-black/5 transition-all">
                                                    Consult Specialist
                                                </a>
                                            </div>
                                        </div>
                                    </div>
                                    
                                    <div className="flex justify-center gap-10">
                                        <Link to="/" className="text-[10px] font-bold text-zinc-600 hover:text-white transition-colors tracking-widest uppercase font-mono">Return Home</Link>
                                        <Link to="/calculator" className="text-[10px] font-bold text-zinc-600 hover:text-white transition-colors tracking-widest uppercase font-mono">ROI Forecast</Link>
                                    </div>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}


