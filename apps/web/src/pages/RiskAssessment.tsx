import { useState, FormEvent } from 'react';
import { Link } from 'react-router-dom';
import {
  Shield, Upload, FileText, AlertTriangle, ArrowRight,
  Loader2, CheckCircle2, BarChart3, Activity, Lock, ArrowLeft
} from 'lucide-react';
import { toast } from 'sonner';

interface RegulatoryFinding {
  attribute: string;
  regulation: string;
  article: string;
  severity: 'HIGH' | 'MEDIUM' | 'LOW';
}

interface RiskScore { score: number; label: string; }

interface FullReport {
  Meta: { MethodologyVersion: string };
  Risk: {
    identifiability_risk: RiskScore;
    financial_sensitivity: RiskScore;
    reidentification_risk: RiskScore;
    compliance_readiness: RiskScore;
    regulatory_findings: RegulatoryFinding[];
    recommendation: string;
  };
  Before: { Label: string; RiskLevel: string; VaRRange: string; AIStatus: string; Description: string };
  After:  { Label: string; RiskLevel: string; VaRRange: string; AIStatus: string; Description: string };
}

interface Report {
  overall_risk_level: string;
  is_gdpr_pseudonymized: boolean;
  financial_exposure: { var_min_eur: number; var_max_eur: number };
  ai_readiness: { status: string; recommendation: string };
  report_id?: string;
  full?: FullReport;
}

export default function RiskAssessment() {
  const [step, setStep] = useState(1);
  const [inputType, setInputType] = useState<'paste' | 'upload'>('paste');
  const [pastedData, setPastedData] = useState(`[
  { "name": "John Doe", "email": "john@example.com", "region": "North", "dept": "HR", "salary": 75000 },
  { "name": "Jane Smith", "email": "jane@example.com", "region": "South", "dept": "IT", "salary": 85000 },
  { "name": "Alice Brown", "email": "alice@example.com", "region": "North", "dept": "Sales", "salary": 65000 },
  { "name": "Bob Wilson", "email": "bob@example.com", "region": "East", "dept": "Dev", "salary": 95000 }
]`);
  const [file, setFile] = useState<File | null>(null);
  const [email, setEmail] = useState('');
  const [company, setCompany] = useState('');
  const [loading, setLoading] = useState(false);
  const [report, setReport] = useState<Report | null>(null);

  const handleProcess = async () => {
    setLoading(true);
    await new Promise(r => setTimeout(r, 1500));
    setStep(3);
    setLoading(false);
  };

  const handleUnlock = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const formData = new FormData();
      formData.append('email', email);
      formData.append('company', company);
      if (inputType === 'paste') {
        formData.append('dataset', new Blob([pastedData], { type: 'application/json' }), 'data.json');
      } else if (file) {
        formData.append('dataset', file);
      }
      const res = await fetch('/api/pilot-assessment', { method: 'POST', body: formData });
      const data = await res.json();
      if (data.status === 'success') {
        setReport({ ...data.report, report_id: data.report_id, full: data.full_report });
        setStep(4);
      } else {
        alert(data.error || 'Failed to analyze data');
      }
    } catch {
      alert('Service unavailable');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen font-sans antialiased bg-white text-gray-900 selection:bg-cyan-500 selection:text-black pb-32">

      <nav className="fixed top-0 left-0 w-full z-50 py-5 bg-white/80 backdrop-blur-xl border-b border-gray-100">
        <div className="max-container flex justify-between items-center">
          <Link to="/" className="font-mono font-black text-xl tracking-widest text-gray-900">OCULTAR</Link>
          <Link to="/" className="text-[11px] font-bold uppercase tracking-widest text-gray-500 hover:text-gray-900 transition-colors flex items-center gap-2">
            <ArrowLeft className="w-3 h-3" /> Back to Platform
          </Link>
        </div>
      </nav>

      <div className="max-container pt-32 relative z-10">

        <header className="max-w-4xl mx-auto text-center space-y-8 mb-20">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-cyan-50 border border-cyan-100 text-[10px] font-bold text-cyan-600 uppercase tracking-widest">
            <Shield className="w-3 h-3" /> Pilot Audit Engine v3.1
          </div>
          <h1 className="tracking-tighter text-gray-900 text-5xl font-bold">
            Test Your Data <span className="text-cyan-600">Risk</span> in 60 Seconds
          </h1>
          <p className="max-w-2xl mx-auto text-lg text-gray-500 font-medium leading-relaxed">
            Upload a sample dataset and see your compliance exposure before using cloud AI tools.
            Stateless processing. No data stored. Zero-Egress guaranteed.
          </p>
        </header>

        <div className="max-w-5xl mx-auto">

          {/* STEP 1 */}
          {step === 1 && (
            <div className="space-y-16">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {[
                  { icon: <Upload className="w-6 h-6" />, title: 'Secure Interception', desc: 'Process small CSV or JSON samples locally within your perimeter.' },
                  { icon: <Shield className="w-6 h-6" />, title: 'Instant Integrity Audit', desc: 'Identify PII exposure and K-Anonymity scores across 40+ entity types.' },
                  { icon: <FileText className="w-6 h-6" />, title: 'Risk Exposure (VaR)', desc: 'Get defensible Value-at-Risk estimations for your legal team.' },
                ].map((item) => (
                  <div key={item.title} className="card bg-white border-gray-200 shadow-sm">
                    <div className="w-12 h-12 bg-gray-50 border border-gray-100 rounded-xl flex items-center justify-center mb-6 text-cyan-600">{item.icon}</div>
                    <h3 className="text-xl font-bold text-gray-900 mb-2">{item.title}</h3>
                    <p className="text-sm text-gray-500 leading-relaxed">{item.desc}</p>
                  </div>
                ))}
              </div>
              <div className="flex justify-center">
                <button
                  onClick={() => setStep(2)}
                  className="bg-cyan-600 text-white px-16 py-6 rounded-lg text-xl font-bold shadow-xl hover:bg-cyan-700 hover:scale-105 transition-all flex items-center group"
                >
                  Start Free Audit <ArrowRight className="ml-3 w-6 h-6 group-hover:translate-x-1 transition-transform" />
                </button>
              </div>
            </div>
          )}

          {/* STEP 2 */}
          {step === 2 && (
            <div className="card max-w-3xl mx-auto bg-white space-y-10 shadow-xl border-gray-200">
              <div className="flex gap-2 p-1 bg-gray-50 rounded-xl border border-gray-100">
                {(['paste', 'upload'] as const).map((t) => (
                  <button
                    key={t}
                    onClick={() => setInputType(t)}
                    className={`flex-1 py-4 text-[10px] font-bold uppercase tracking-widest rounded-lg transition-all ${inputType === t ? 'bg-white text-cyan-600 shadow-sm border border-gray-100' : 'text-gray-400 hover:text-gray-600'}`}
                  >
                    {t === 'paste' ? 'Paste Dataset' : 'Upload File'}
                  </button>
                ))}
              </div>

              <div className="min-h-[350px]">
                {inputType === 'paste' ? (
                  <div className="space-y-4">
                    <textarea
                      value={pastedData}
                      onChange={(e) => setPastedData(e.target.value)}
                      className="w-full h-80 p-8 bg-gray-50 border border-gray-200 rounded-2xl font-mono text-sm text-gray-700 focus:border-cyan-500/50 outline-none transition-colors"
                      placeholder="Paste JSON or CSV here..."
                    />
                    <p className="text-[10px] text-gray-400 font-bold uppercase tracking-widest font-mono">Input Buffer [READY]</p>
                  </div>
                ) : (
                  <div className="relative border-2 border-dashed border-gray-200 rounded-2xl h-80 flex flex-col items-center justify-center p-8 text-center gap-6 bg-gray-50/50 hover:bg-gray-50 transition-colors cursor-pointer group hover:border-cyan-500/30">
                    <Upload className="w-12 h-12 text-gray-300 group-hover:text-cyan-500 transition-colors" />
                    <div className="space-y-2">
                      <p className="font-bold text-gray-900 text-lg">Drop dataset here or <span className="text-cyan-600 underline">browse</span></p>
                      <p className="text-xs text-gray-500 font-mono">.csv, .json, .txt (Max 100KB)</p>
                    </div>
                    <input type="file" onChange={(e) => setFile(e.target.files?.[0] ?? null)} className="absolute inset-0 opacity-0 cursor-pointer z-10" accept=".csv,.json,.txt" />
                    {file && (
                      <div className="absolute inset-0 bg-white rounded-2xl flex items-center justify-center gap-3 z-20 border border-cyan-500/50 shadow-lg">
                        <FileText className="w-6 h-6 text-cyan-600" />
                        <span className="font-bold text-gray-900">{file.name}</span>
                        <button onClick={(e) => { e.stopPropagation(); setFile(null); }} className="text-xs text-red-500 font-bold ml-4 uppercase tracking-widest hover:underline">Remove</button>
                      </div>
                    )}
                  </div>
                )}
              </div>

              <div className="p-5 bg-cyan-50 border border-cyan-100 rounded-xl flex gap-4">
                <AlertTriangle className="w-5 h-5 text-cyan-600 shrink-0 mt-0.5" />
                <p className="text-xs text-gray-600 leading-relaxed font-medium">
                  <strong className="text-cyan-600">Security Note:</strong> All processing is done via transient local inference. This node is stateless and encrypted. Data never leaves the bridge.
                </p>
              </div>

              <button
                onClick={handleProcess}
                disabled={loading || (inputType === 'upload' && !file)}
                className="w-full bg-cyan-600 text-white py-5 rounded-lg text-lg font-bold shadow-lg disabled:opacity-50 hover:bg-cyan-700 transition-colors flex items-center justify-center"
              >
                {loading ? <Loader2 className="w-6 h-6 animate-spin" /> : 'Run Integrity Scan'}
              </button>
            </div>
          )}

          {/* STEP 3 — Lead gate */}
          {step === 3 && (
            <div className="card max-w-lg mx-auto bg-white space-y-10 py-12 shadow-xl border-gray-200">
              <div className="text-center space-y-4">
                <div className="text-cyan-600 font-bold text-[10px] uppercase tracking-[0.4em] inline-flex items-center gap-3">
                  <div className="w-2 h-2 bg-cyan-600 rounded-full animate-pulse" />
                  Analysis Complete
                </div>
                <h2 className="text-3xl font-bold tracking-tight text-gray-900">Unlock Report</h2>
                <p className="text-sm text-gray-500 font-medium">Verification successful. Enter your details to generate your risk exposure summary.</p>
              </div>
              <form onSubmit={handleUnlock} className="space-y-6">
                <div className="space-y-2">
                  <label className="text-[10px] font-bold uppercase tracking-widest text-gray-400">Corporate Email</label>
                  <input type="email" required value={email} onChange={(e) => setEmail(e.target.value)} className="w-full p-4 bg-gray-50 border border-gray-200 rounded-xl outline-none focus:border-cyan-500/50 transition-colors text-gray-900" placeholder="you@company.com" />
                </div>
                <div className="space-y-2">
                  <label className="text-[10px] font-bold uppercase tracking-widest text-gray-400">Organization</label>
                  <input type="text" required value={company} onChange={(e) => setCompany(e.target.value)} className="w-full p-4 bg-gray-50 border border-gray-200 rounded-xl outline-none focus:border-cyan-500/50 transition-colors text-gray-900" placeholder="Tech Corp Inc." />
                </div>
                <button className="w-full bg-cyan-600 text-white py-5 rounded-lg text-lg font-bold hover:bg-cyan-700 transition-colors flex items-center justify-center" disabled={loading}>
                  {loading ? <Loader2 className="w-6 h-6 animate-spin text-white" /> : 'Generate Audit Report'}
                </button>
              </form>
            </div>
          )}

          {/* STEP 4 — Report */}
          {step === 4 && report && (
            <div className="space-y-12 pb-20">
              <div className="card bg-cyan-600 text-white p-12 flex flex-col md:flex-row justify-between items-center gap-10 overflow-hidden relative border-none rounded-[2rem] shadow-2xl">
                <Shield className="absolute -right-16 -bottom-16 w-80 h-80 text-white/10 rotate-12" />
                <div className="relative z-10 space-y-4">
                  <div className="text-[10px] font-bold uppercase tracking-[0.4em] text-white/60 font-mono">Consolidated Risk Profile</div>
                  <div className="text-6xl font-black tracking-tighter uppercase italic">{report.overall_risk_level}</div>
                </div>
                <div className="relative z-10 text-center md:text-right space-y-6">
                  <div className={`inline-flex items-center gap-3 px-6 py-3 rounded-full text-xs font-bold uppercase tracking-widest ${report.is_gdpr_pseudonymized ? 'bg-white/10 text-white border border-white/10' : 'bg-rose-500 text-white'}`}>
                    {report.is_gdpr_pseudonymized ? <CheckCircle2 className="w-5 h-5" /> : <AlertTriangle className="w-5 h-5" />}
                    {report.is_gdpr_pseudonymized ? 'Integrity Verified' : 'Critical Leakage Risk'}
                  </div>
                  <div className="text-sm font-bold opacity-60 uppercase tracking-widest font-mono">
                    Compliance Ready: {report.is_gdpr_pseudonymized ? 'YES' : 'REMEDIATION_REQUIRED'}
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                <div className="card bg-white border-gray-100 shadow-sm space-y-8">
                  <div className="text-[10px] font-bold uppercase tracking-widest text-gray-400">Est. Financial Exposure (VaR)</div>
                  <div className="text-5xl font-bold text-rose-500 tracking-tighter">
                    €{Math.round(report.financial_exposure.var_min_eur).toLocaleString()} – €{Math.round(report.financial_exposure.var_max_eur).toLocaleString()}
                  </div>
                  <p className="text-sm text-gray-500 leading-relaxed font-medium">Projected annualized impact across regulatory domains.</p>
                </div>
                <div className="card bg-white border-gray-100 shadow-sm space-y-8">
                  <div className="text-[10px] font-bold uppercase tracking-widest text-gray-400">AI Readiness Status</div>
                  <div className="flex items-center gap-4">
                    <div className={`p-4 rounded-xl ${report.ai_readiness.status === 'ALLOW' ? 'bg-cyan-50 text-cyan-600' : 'bg-rose-50 text-rose-500'}`}>
                      <Activity className="w-8 h-8" />
                    </div>
                    <span className={`font-bold text-4xl tracking-tighter uppercase italic ${report.ai_readiness.status === 'ALLOW' ? 'text-cyan-600' : 'text-rose-500'}`}>{report.ai_readiness.status}</span>
                  </div>
                  <p className="text-sm text-gray-500 font-medium leading-relaxed">{report.ai_readiness.recommendation}</p>
                </div>
              </div>

              {report.full && (
                <div className="space-y-24 py-16 border-t border-gray-100 mt-20">
                  {/* Scorecard */}
                  <div className="space-y-12">
                    <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
                      <div className="space-y-4">
                        <div className="text-cyan-600 font-mono text-[10px] tracking-[0.4em] uppercase font-bold">[ ASSESSMENT_RESULT_TX_{report.report_id?.slice(0, 8)} ]</div>
                        <h2 className="text-4xl font-bold tracking-tight text-gray-900 flex items-center gap-4">Technical Risk Scorecard <BarChart3 className="w-8 h-8 text-cyan-600" /></h2>
                      </div>
                      <p className="text-gray-400 text-sm max-w-xs font-medium italic underline underline-offset-4 decoration-cyan-500/30">
                        Integrity Mapping v{report.full.Meta.MethodologyVersion}
                      </p>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                      {[
                        { label: 'Identifiability',      data: report.full.Risk.identifiability_risk },
                        { label: 'Financial Sensitivity', data: report.full.Risk.financial_sensitivity },
                        { label: 'Re-id Complexity',      data: report.full.Risk.reidentification_risk },
                        { label: 'Regulatory Readiness',  data: report.full.Risk.compliance_readiness },
                      ].map((stat) => (
                        <div key={stat.label} className="card bg-white border-gray-100 shadow-sm flex flex-col gap-6">
                          <span className="text-[10px] font-bold uppercase tracking-widest text-gray-400 font-mono">{stat.label}</span>
                          <div className="text-5xl font-bold text-gray-900 tracking-tighter">{stat.data.score.toFixed(1)}<span className="text-xl text-gray-200 ml-1">/10</span></div>
                          <div className={`text-[10px] font-bold uppercase px-4 py-1.5 rounded-full w-fit tracking-widest ${stat.data.score > 7 ? 'bg-rose-50 text-rose-600' : stat.data.score > 4 ? 'bg-amber-50 text-amber-600' : 'bg-cyan-50 text-cyan-600'}`}>
                            {stat.data.label}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Regulatory matrix */}
                  <div className="space-y-12">
                    <h2 className="text-4xl font-bold tracking-tight text-gray-900 flex items-center gap-4">Regulatory Alignment Matrix <Shield className="w-8 h-8 text-cyan-600" /></h2>
                    <div className="card p-0 bg-white overflow-hidden border-gray-100 shadow-xl">
                      <div className="overflow-x-auto">
                        <table className="w-full text-left border-collapse">
                          <thead>
                            <tr className="bg-gray-50 border-b border-gray-100">
                              {['Sensitive Attribute', 'Governing Regulation', 'Primary Article', 'Security Threat'].map(h => (
                                <th key={h} className="p-6 text-[10px] font-bold uppercase tracking-[0.2em] text-gray-400">{h}</th>
                              ))}
                            </tr>
                          </thead>
                          <tbody>
                            {report.full.Risk.regulatory_findings?.map((finding, idx) => (
                              <tr key={idx} className="border-b border-gray-50 hover:bg-gray-50/50 transition-colors">
                                <td className="p-6 font-mono text-sm font-bold text-gray-900 flex items-center gap-3">
                                  <div className="w-1.5 h-1.5 bg-cyan-600/50 rounded-full" />{finding.attribute}
                                </td>
                                <td className="p-6 text-sm font-medium text-gray-500">{finding.regulation}</td>
                                <td className="p-6"><span className="px-3 py-1 bg-gray-50 border border-gray-100 rounded font-mono text-[10px] text-cyan-600 font-bold tracking-widest">{finding.article}</span></td>
                                <td className="p-6">
                                  <div className={`text-[10px] font-bold uppercase px-3 py-1.5 rounded-sm inline-flex items-center gap-2 tracking-[0.1em] ${finding.severity === 'HIGH' ? 'bg-rose-50 text-rose-600 border border-rose-100' : finding.severity === 'MEDIUM' ? 'bg-amber-50 text-amber-600 border border-amber-100' : 'bg-cyan-50 text-cyan-600 border border-cyan-100'}`}>
                                    {finding.severity} RISK
                                  </div>
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                      <div className="p-6 bg-gray-50/50 border-t border-gray-100">
                        <p className="text-[11px] text-gray-400 leading-relaxed font-bold uppercase font-mono tracking-widest">
                          * DISCLOSURE: Mapped via Heuristic Engine v1.2. Article references are for infrastructure audit prioritization.
                        </p>
                      </div>
                    </div>
                  </div>

                  {/* Mitigation sim */}
                  <div className="space-y-12">
                    <h2 className="text-4xl font-bold tracking-tight text-gray-900 flex items-center gap-4">Ocultar Mitigation Simulation <CheckCircle2 className="w-8 h-8 text-cyan-600" /></h2>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-0 border border-gray-100 rounded-[2.5rem] overflow-hidden shadow-xl">
                      {[report.full.Before, report.full.After].map((scenario, idx) => (
                        <div key={scenario.Label} className={`p-12 space-y-10 ${idx === 0 ? 'bg-white border-r border-gray-100' : 'bg-cyan-50/30'}`}>
                          <div className="flex justify-between items-center mb-6">
                            <div className={`text-[10px] font-bold uppercase tracking-[0.4em] ${idx === 0 ? 'text-rose-500' : 'text-cyan-600 font-black'}`}>[ {scenario.Label} ]</div>
                            <div className={`px-4 py-1.5 border rounded-full text-[10px] font-bold uppercase tracking-widest ${idx === 0 ? 'border-rose-200 text-rose-600' : 'border-cyan-200 text-cyan-600'}`}>{scenario.RiskLevel}</div>
                          </div>
                          <div className="space-y-6">
                            <div className="flex justify-between text-sm py-4 border-b border-gray-50">
                              <span className="text-gray-400 font-bold uppercase tracking-widest text-[10px]">Projection (VaR)</span>
                              <span className={`font-mono font-bold ${idx === 0 ? 'text-rose-600' : 'text-cyan-600'}`}>{scenario.VaRRange}</span>
                            </div>
                            <div className="flex justify-between text-sm py-4 border-b border-gray-50">
                              <span className="text-gray-400 font-bold uppercase tracking-widest text-[10px]">Cloud Egress</span>
                              <span className={`font-bold uppercase tracking-widest ${idx === 0 ? 'text-rose-600' : 'text-gray-900'}`}>{scenario.AIStatus}</span>
                            </div>
                          </div>
                          <div className="bg-gray-50 p-6 rounded-2xl border border-gray-100 italic">
                            <p className="text-base text-gray-500 leading-relaxed font-medium">"{scenario.Description}"</p>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Remediation */}
                  <div className="card border-rose-100 bg-rose-50/30 p-12 space-y-8 relative overflow-hidden group rounded-[2.5rem]">
                    <div className="absolute top-0 right-0 p-12 opacity-5 rotate-12 group-hover:opacity-10 transition-opacity">
                      <AlertTriangle className="w-40 h-40 text-rose-600" />
                    </div>
                    <div className="flex items-center gap-4 text-rose-600">
                      <div className="p-4 bg-rose-100 border border-rose-200 rounded-2xl flex items-center justify-center">
                        <Lock className="w-8 h-8" />
                      </div>
                      <h3 className="text-3xl font-bold tracking-tight">Hardened Remediation Plan</h3>
                    </div>
                    <div className="text-xl text-gray-700 leading-relaxed font-medium bg-white p-8 rounded-2xl border border-gray-100 shadow-lg relative z-10">
                      <span className="text-rose-600 font-mono font-bold mr-2 tracking-widest underline underline-offset-8 decoration-rose-500/50">CRITICAL:</span> {report.full.Risk.recommendation}
                    </div>
                  </div>

                  {/* Final CTA */}
                  <div className="bg-cyan-600 text-white p-16 rounded-[3rem] text-center space-y-10 relative overflow-hidden shadow-2xl">
                    <div className="absolute top-0 right-0 p-16 opacity-10"><Shield className="w-96 h-96" /></div>
                    <div className="relative z-10 space-y-6">
                      <div className="text-[10px] font-black uppercase tracking-[0.5em] text-white/40">Secure Your Perimeter</div>
                      <h2 className="text-5xl font-black tracking-tighter lowercase italic">activate technical sovereignty.</h2>
                      <p className="text-white/80 max-w-2xl mx-auto text-lg font-bold leading-relaxed">
                        Your full compliance report, re-identification heatmaps, and structured remediation plan are ready for executive review.
                      </p>
                      <div className="flex flex-col md:flex-row gap-4 justify-center pt-8">
                        {report.report_id && (
                          <a href={`/api/pilot/report?id=${report.report_id}`} target="_blank" rel="noopener noreferrer" className="bg-white text-cyan-600 px-12 py-6 rounded-lg text-lg font-bold hover:scale-105 transition-transform shadow-2xl">
                            Download PDF Audit
                          </a>
                        )}
                        <button
                          onClick={() => {
                            navigator.clipboard.writeText('sales@ocultar.dev');
                            toast.success('Email copied to clipboard', {
                              description: 'Reach out to sales@ocultar.dev to request access.',
                            });
                          }}
                          className="border border-white/20 text-white px-12 py-6 rounded-lg text-lg font-bold hover:bg-white/5 transition-all"
                        >
                          Consult Specialist
                        </button>
                      </div>
                    </div>
                  </div>

                  <div className="flex justify-center gap-10">
                    <Link to="/" className="text-[10px] font-bold text-gray-400 hover:text-gray-900 transition-colors tracking-widest uppercase font-mono">Return Home</Link>
                    <Link to="/calculator" className="text-[10px] font-bold text-gray-400 hover:text-gray-900 transition-colors tracking-widest uppercase font-mono">ROI Forecast</Link>
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
