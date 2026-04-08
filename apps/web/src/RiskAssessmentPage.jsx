import { useState } from 'react';
import { Shield, Upload, FileText, AlertTriangle, ArrowRight, Loader2, CheckCircle2, ChevronRight } from 'lucide-react';

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
        <div className="max-container pb-20 pt-10">
            
            {/* --- HEADER --- */}
            <div className="max-w-3xl mx-auto text-center space-y-6 mb-16">
                <div className="badge animate-in fade-in slide-in-from-bottom-2 duration-500">
                    <Shield className="w-3 h-3 mr-2" /> Pilot Assessment v3.1
                </div>
                <h1 className="text-primary">Test Your Data Risk in 60 Seconds</h1>
                <p className="text-lg text-secondary">
                    Upload a sample dataset and see your compliance exposure before using AI tools. 
                    Stateless processing. No data stored.
                </p>
            </div>

            {/* --- STEPS --- */}
            <div className="max-w-4xl mx-auto">
                
                {/* STEP 1: LANDING */}
                {step === 1 && (
                    <div className="space-y-12 animate-in fade-in duration-700">
                        <div className="feature-grid">
                            {[
                                { id: "feat-upload", icon: <Upload className="w-5 h-5" />, title: "Secure Upload", desc: "Process small CSV or JSON samples locally." },
                                { id: "feat-audit", icon: <Shield className="w-5 h-5" />, title: "Instant Audit", desc: "Identify PII exposure and K-Anonymity scores." },
                                { id: "feat-report", icon: <FileText className="w-5 h-5" />, title: "Risk Report", desc: "Get defensible Value-at-Risk (VaR) estimations." }
                            ].map((item) => (
                                <div key={item.id} className="card bg-secondary/30 space-y-4">
                                    <div className="w-10 h-10 bg-white border border-color rounded-lg flex items-center justify-center shadow-sm">{item.icon}</div>
                                    <h3 className="text-lg font-bold">{item.title}</h3>
                                    <p className="text-sm">{item.desc}</p>
                                </div>
                            ))}
                        </div>
                        <div className="flex justify-center">
                            <button 
                                onClick={handleStart}
                                className="btn btn-primary px-12 py-5 text-lg font-bold shadow-xl shadow-black/10 group"
                            >
                                Start Free Assessment <ArrowRight className="ml-3 w-5 h-5 group-hover:translate-x-1 transition-transform" />
                            </button>
                        </div>
                    </div>
                )}

                {/* STEP 2: INPUT */}
                {step === 2 && (
                    <div className="card max-w-2xl mx-auto space-y-8 animate-in fade-in duration-500 shadow-xl shadow-black/5">
                        <div className="flex gap-2 p-1 bg-secondary rounded-xl">
                            <button 
                                onClick={() => setInputType('paste')}
                                className={`flex-1 py-3 text-sm font-medium rounded-lg transition-all ${inputType === 'paste' ? 'bg-white shadow-sm text-black' : 'text-secondary hover:text-black'}`}
                            >
                                Paste Data
                            </button>
                            <button 
                                onClick={() => setInputType('upload')}
                                className={`flex-1 py-3 text-sm font-medium rounded-lg transition-all ${inputType === 'upload' ? 'bg-white shadow-sm text-black' : 'text-secondary hover:text-black'}`}
                            >
                                File Upload
                            </button>
                        </div>

                        <div className="min-h-[300px]">
                            {inputType === 'paste' ? (
                                <div className="space-y-4">
                                    <textarea 
                                        value={pastedData}
                                        onChange={(e) => setPastedData(e.target.value)}
                                        className="w-full h-64 p-6 bg-secondary border border-color rounded-2xl font-mono text-sm focus:border-black outline-none transition-colors"
                                        placeholder='Paste JSON or CSV here...'
                                    />
                                    <p className="text-xs text-muted font-medium italic">Pre-filled with a sample employee dataset for demonstration.</p>
                                </div>
                            ) : (
                                <div className="relative border-2 border-dashed border-color rounded-2xl h-64 flex flex-col items-center justify-center p-8 text-center gap-4 bg-secondary/30 hover:bg-secondary/50 transition-colors cursor-pointer group">
                                    <Upload className="w-10 h-10 text-muted group-hover:text-black transition-colors" />
                                    <div className="space-y-1">
                                        <p className="font-bold">Drop your file here or <span className="underline">browse</span></p>
                                        <p className="text-xs text-muted">Supports .csv, .json, .txt (Max 100KB)</p>
                                    </div>
                                    <input 
                                        type="file" 
                                        onChange={(e) => setFile(e.target.files[0])}
                                        className="absolute inset-0 opacity-0 cursor-pointer z-10"
                                        accept=".csv,.json,.txt"
                                    />
                                    {file && <div className="absolute inset-0 bg-white/95 rounded-2xl flex items-center justify-center gap-2 z-20">
                                        <FileText className="w-5 h-5" />
                                        <span className="font-bold">{file.name}</span>
                                        <button onClick={(e) => {e.stopPropagation(); setFile(null)}} className="text-xs text-red-500 font-bold ml-2">Remove</button>
                                    </div>}
                                </div>
                            )}
                        </div>

                        <div className="p-4 bg-yellow-50 border border-yellow-100 rounded-xl flex gap-3">
                            <AlertTriangle className="w-5 h-5 text-yellow-600 shrink-0" />
                            <p className="text-xs text-yellow-800">
                                <strong>Privacy Notice:</strong> Do not upload sensitive production data. This assessment is processed entirely in-memory for demonstration.
                            </p>
                        </div>

                        <button 
                            onClick={handleProcess}
                            disabled={loading || (inputType === 'upload' && !file)}
                            className="w-full btn btn-primary py-4 text-base font-bold shadow-lg shadow-black/10 disabled:opacity-50"
                        >
                            {loading ? <Loader2 className="w-5 h-5 animate-spin" /> : 'Run Risk Analysis'}
                        </button>
                    </div>
                )}

                {/* STEP 3: LEAD GATE */}
                {step === 3 && (
                    <div className="card max-w-md mx-auto space-y-8 py-10 animate-in fade-in duration-500 shadow-2xl shadow-black/5">
                        <div className="text-center space-y-2">
                            <div className="text-red-600 font-bold text-xs uppercase tracking-widest inline-flex items-center gap-2">
                                <div className="w-2 h-2 bg-red-600 rounded-full animate-pulse" />
                                Risk Detected
                            </div>
                            <h2 className="text-2xl font-bold">Unlock Analysis</h2>
                            <p className="text-sm text-secondary">Your report is ready. Enter your work details to see the exposure values.</p>
                        </div>

                        <form onSubmit={handleUnlock} className="space-y-5">
                            <div className="space-y-2">
                                <label className="text-xs font-bold uppercase tracking-widest text-muted">Work Email</label>
                                <input 
                                    type="email" required
                                    value={email} onChange={(e) => setEmail(e.target.value)}
                                    className="w-full p-4 bg-secondary border border-color rounded-xl outline-none focus:border-black transition-colors"
                                    placeholder="you@company.com"
                                />
                            </div>
                            <div className="space-y-2">
                                <label className="text-xs font-bold uppercase tracking-widest text-muted">Company Name</label>
                                <input 
                                    type="text" required
                                    value={company} onChange={(e) => setCompany(e.target.value)}
                                    className="w-full p-4 bg-secondary border border-color rounded-xl outline-none focus:border-black transition-colors"
                                    placeholder="Acme Corp"
                                />
                            </div>
                            <button 
                                className="w-full btn btn-primary py-4 text-base font-bold"
                                disabled={loading}
                            >
                                {loading ? <Loader2 className="w-5 h-5 animate-spin" /> : 'Unlock Detailed Report'}
                            </button>
                        </form>
                    </div>
                )}

                {/* STEP 4: REPORT PREVIEW */}
                {step === 4 && report && (
                    <div className="space-y-12 animate-in fade-in duration-500">
                        {/* Summary Card */}
                        <div className="card bg-black text-white p-10 flex flex-col md:flex-row justify-between items-center gap-8 overflow-hidden relative">
                             <Shield className="absolute -right-10 -bottom-10 w-64 h-64 text-white/5 rotate-12" />
                            <div className="relative z-10 space-y-2">
                                <div className="text-xs font-bold uppercase tracking-widest text-white/60">Risk Profile</div>
                                <div className="text-5xl font-bold tracking-tighter">{report.overall_risk_level}</div>
                            </div>
                            <div className="relative z-10 text-center md:text-right space-y-4">
                                <div className={`inline-flex items-center gap-2 px-4 py-2 rounded-full text-xs font-bold uppercase ${report.is_gdpr_pseudonymized ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
                                    {report.is_gdpr_pseudonymized ? <CheckCircle2 className="w-4 h-4" /> : <AlertTriangle className="w-4 h-4" />}
                                    {report.is_gdpr_pseudonymized ? 'Heuristic Pseudonymization Identified' : 'High Re-identification Risk'}
                                </div>
                                <div className="text-sm text-white/60 font-medium">Compliance Review: {report.is_gdpr_pseudonymized ? 'Compliant' : 'Violation Likely'}</div>
                            </div>
                        </div>

                        {/* Top Metrics Grid */}
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                            <div className="card space-y-6">
                                <div className="text-xs font-bold uppercase tracking-widest text-muted">Est. Financial Exposure</div>
                                <div className="text-4xl font-bold text-red-600 tracking-tighter">
                                    €{Math.round(report.financial_exposure.var_min_eur).toLocaleString()} – €{Math.round(report.financial_exposure.var_max_eur).toLocaleString()}
                                </div>
                                <p className="text-sm">Based on industry Simulation anchors and sensitivity scores.</p>
                            </div>
                            
                            <div className="card space-y-6">
                                <div className="text-xs font-bold uppercase tracking-widest text-muted">AI Readiness</div>
                                <div className="flex items-center gap-3">
                                    <div className={`p-2 rounded-lg ${report.ai_readiness.status === 'ALLOW' ? 'bg-green-100 text-green-600' : 'bg-red-100 text-red-600'}`}>
                                        <Shield className="w-6 h-6" />
                                    </div>
                                    <span className="font-bold text-2xl tracking-tight">{report.ai_readiness.status}</span>
                                </div>
                                <p className="text-sm font-medium">{report.ai_readiness.recommendation}</p>
                            </div>
                        </div>

                        {/* Full Report Details */}
                        {report.full && (
                            <div className="space-y-16 py-10 border-t border-color mt-20">
                                {/* 1. SCORECARD */}
                                <div className="space-y-8">
                                    <h2 className="text-2xl font-bold flex items-center gap-3">
                                        <BarChart3 className="w-6 h-6" /> 
                                        Risk Scorecard v{report.full.Meta.MethodologyVersion}
                                    </h2>
                                    <div className="feature-grid">
                                        {[
                                            { label: "Identifiability", data: report.full.Risk.identifiability_risk },
                                            { label: "Financial Sensitivity", data: report.full.Risk.financial_sensitivity },
                                            { label: "Re-id Risk", data: report.full.Risk.reidentification_risk },
                                            { label: "Compliance Readiness", data: report.full.Risk.compliance_readiness }
                                        ].map((stat) => (
                                            <div key={stat.label} className="card bg-secondary/30 flex flex-col gap-4">
                                                <span className="text-xs font-bold uppercase tracking-widest text-muted">{stat.label}</span>
                                                <div className="text-3xl font-bold">{stat.data.score.toFixed(1)}<span className="text-lg opacity-40">/10</span></div>
                                                <div className={`text-[10px] font-bold uppercase px-3 py-1 rounded-full w-fit ${stat.data.score > 7 ? 'bg-red-100 text-red-600' : stat.data.score > 4 ? 'bg-yellow-100 text-yellow-600' : 'bg-green-100 text-green-600'}`}>
                                                    {stat.data.label}
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                </div>

                                {/* 2. TRANSFORMATION TABLE */}
                                <div className="space-y-8">
                                    <h2 className="text-2xl font-bold">Ocultar Transformation Simulation</h2>
                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-0 border border-color rounded-2xl overflow-hidden">
                                        {[report.full.Before, report.full.After].map((scenario, idx) => (
                                            <div key={scenario.Label} className={`p-10 space-y-6 ${idx === 0 ? 'bg-white border-r border-color' : 'bg-secondary/20'}`}>
                                                <div className="flex justify-between items-center mb-4">
                                                    <span className={`text-xs font-bold uppercase tracking-[0.2em] ${idx === 0 ? 'text-red-500' : 'text-green-600'}`}>{scenario.Label}</span>
                                                    <span className="badge">{scenario.RiskLevel}</span>
                                                </div>
                                                <div className="space-y-4">
                                                    <div className="flex justify-between text-sm py-2 border-b border-color/40">
                                                        <span className="text-muted font-medium">Exposure Range</span>
                                                        <span className="font-bold">{scenario.VaRRange}</span>
                                                    </div>
                                                    <div className="flex justify-between text-sm py-2 border-b border-color/40">
                                                        <span className="text-muted font-medium">AI Status</span>
                                                        <span className="font-bold uppercase">{scenario.AIStatus}</span>
                                                    </div>
                                                </div>
                                                <p className="text-sm text-secondary leading-relaxed pt-2">{scenario.Description}</p>
                                            </div>
                                        ))}
                                    </div>
                                </div>

                                {/* 3. REMEDIATION */}
                                <div className="card border-yellow-200 bg-yellow-50/50 p-10 space-y-6">
                                    <div className="flex items-center gap-3 text-yellow-800">
                                        <div className="p-3 bg-yellow-100 rounded-xl">
                                            <AlertTriangle className="w-6 h-6" />
                                        </div>
                                        <h3 className="text-xl font-bold">Immediate Remediation Plan</h3>
                                    </div>
                                    <div className="text-sm text-yellow-900 leading-relaxed font-medium bg-white/50 p-6 rounded-xl border border-yellow-100">
                                        {report.full.Risk.recommendation}
                                    </div>
                                </div>

                                {/* 4. FINAL ACTIONS */}
                                <div className="bg-black text-white p-12 rounded-3xl text-center space-y-8 relative overflow-hidden">
                                    <div className="absolute top-0 right-0 p-12 opacity-10">
                                        <Shield className="w-64 h-64" />
                                    </div>
                                    <div className="relative z-10 space-y-4">
                                        <h2 className="text-3xl font-bold tracking-tight">Activate High-Fidelity Protection</h2>
                                        <p className="text-white/60 max-w-xl mx-auto text-sm font-medium">
                                            Your full compliance report, re-identification heatmaps, and structured remediation plan are ready for download.
                                        </p>
                                        <div className="flex flex-col md:flex-row gap-4 justify-center pt-8">
                                            {report.report_id && (
                                                <a href={`/api/pilot/report?id=${report.report_id}`} target="_blank" rel="noopener noreferrer" className="btn btn-secondary border-white/20 text-white px-8 py-4 text-sm font-bold hover:bg-white/10">
                                                    Download PDF Report
                                                </a>
                                            )}
                                            <a href="mailto:sales@ocultar.dev?subject=Enterprise%20Trial%20Request" className="btn bg-white text-black px-8 py-4 text-sm font-bold shadow-[0_0_40px_rgba(255,255,255,0.2)] hover:scale-105">
                                                Request Guided Pilot
                                            </a>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
}
