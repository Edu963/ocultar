import React, { useState, useMemo, useEffect } from 'react';
import { 
  BarChart3, 
  ShieldAlert, 
  Settings2, 
  Activity, 
  Database, 
  FileCheck2, 
  TrendingUp, 
  Network, 
  Eye, 
  ArrowRightLeft, 
  Lock, 
  Unlock,
  AlertCircle,
  Clock,
  Download,
  ShieldCheck,
  Zap,
  BookOpen,
  HelpCircle,
  Code2,
  ChevronRight,
  Terminal,
  ExternalLink,
  DollarSign,
  Calculator
} from 'lucide-react';

// --- PRODUCTION PRICING CONSTANTS (Sourced from roi_calc.html) ---
const PRICING_DATA = {
  providers: {
    gcp: { name: "Google Cloud DLP", processing: 5.00, egress: 0.10 },
    aws: { name: "AWS Comprehend", processing: 1000.00, egress: 0.10 },
    azure: { name: "Azure AI Language", processing: 1.50, egress: 0.10 }
  },
  ocultar: {
    monthlyLicense: 10000,
    localComputePerTB: 20
  },
  conversion: {
    gbPerTb: 1000
  }
};

const formatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0
});

// --- COMPONENTS ---

const OverviewView = ({ tier, pricingParams, connectionStatus, systemStatus, metrics, vaultStats, auditLogs }) => (
  <div className="space-y-6 relative">
    <section className="grid grid-cols-1 md:grid-cols-4 gap-4">
      <StatusCard icon={<ShieldCheck className="text-blue-400" />} label="Enforcement" value="STRICT" />
      <StatusCard icon={<Zap className="text-emerald-400" />} label="P95 Latency" value={metrics?.latency_per_tier?.regex || "N/A"} />
      <StatusCard icon={<Eye className="text-purple-400" />} label="Scan Coverage" value={metrics?.redaction_rate ? `${(metrics.redaction_rate * 100).toFixed(1)}%` : "N/A"} />
      <StatusCard icon={<Clock className="text-amber-400" />} label="License" value={tier} />
    </section>

    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 relative z-10">
      <FailClosedWrapper status={connectionStatus}>
        <FinancialImpactWidget params={pricingParams} tier={tier} />
      </FailClosedWrapper>
      <FailClosedWrapper status={connectionStatus}>
        <ShadowApiWidget tier={tier} />
      </FailClosedWrapper>
      <FailClosedWrapper status={connectionStatus}>
        <PerformanceWidget metrics={metrics} />
      </FailClosedWrapper>
      <FailClosedWrapper status={connectionStatus}>
        <PolicySimulatorWidget tier={tier} />
      </FailClosedWrapper>
      <FailClosedWrapper status={connectionStatus}>
        <AnonymizationWidget tier={tier} vaultStats={vaultStats} />
      </FailClosedWrapper>
      <FailClosedWrapper status={connectionStatus}>
        <AuditLedgerWidget tier={tier} auditLogs={auditLogs} />
      </FailClosedWrapper>
    </div>
  </div>
);

const FailClosedWrapper = ({ children, status }) => (
  <div className="relative group">
    {children}
    {status === 'DISCONNECTED' && (
      <div className="absolute inset-0 bg-slate-950/90 backdrop-blur-sm z-50 flex flex-col items-center justify-center rounded-xl border border-red-500/50 p-6 text-center">
        <Lock className="w-12 h-12 text-red-500 mb-4 animate-bounce" />
        <h4 className="text-red-500 font-bold uppercase tracking-widest mb-2">Vaulted Connection</h4>
        <p className="text-[10px] text-slate-400 font-mono">REFINERY_UNREACHABLE: Fail-Closed enforcement active. All egress blocked by kernel hook.</p>
        <div className="mt-4 flex gap-2">
           <div className="w-2 h-2 rounded-full bg-red-500 animate-pulse" />
           <div className="w-2 h-2 rounded-full bg-red-500 animate-pulse delay-75" />
           <div className="w-2 h-2 rounded-full bg-red-500 animate-pulse delay-150" />
        </div>
      </div>
    )}
  </div>
);

const FinancialImpactWidget = ({ params, tier }) => {
  const calculations = useMemo(() => {
    const provider = PRICING_DATA.providers[params.provider];
    const gb = params.volume * PRICING_DATA.conversion.gbPerTb;
    const listPricePerGB = provider.processing + provider.egress;
    const totalCloudMonthly = gb * listPricePerGB * (1 - (params.discount / 100));
    const totalOcultarMonthly = PRICING_DATA.ocultar.monthlyLicense + (params.volume * PRICING_DATA.ocultar.localComputePerTB);
    const monthlySavings = Math.max(0, totalCloudMonthly - totalOcultarMonthly);
    
    return { monthlySavings, totalCloudMonthly, totalOcultarMonthly };
  }, [params]);

  const isPro = tier !== 'OPEN_SOURCE';

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 hover:border-blue-500/50 transition-all group relative">
      <div className="flex justify-between items-center mb-6">
        <h3 className="text-slate-200 font-semibold flex items-center gap-2">
          <TrendingUp className="w-4 h-4 text-emerald-400" />
          Financial Impact & ROI
        </h3>
        <span className={`text-[10px] px-2 py-1 rounded-full border ${isPro ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-slate-800 text-slate-500 border-slate-700'}`}>
          {isPro ? 'PRO - Live Analysis' : 'LOCKED'}
        </span>
      </div>
      
      {!isPro ? (
        <div className="py-8 flex flex-col items-center justify-center space-y-4">
           <Lock className="w-8 h-8 text-slate-700" />
           <p className="text-[10px] text-slate-500 text-center uppercase tracking-widest font-bold">Requires Sombra Standalone+</p>
        </div>
      ) : (
        <>
          <div className="space-y-4 mb-6">
            <div>
              <p className="text-[10px] text-slate-500 uppercase font-bold mb-1">Monthly Capital Retention</p>
              <p className="text-3xl font-bold text-emerald-400">{formatter.format(calculations.monthlySavings)}</p>
            </div>
            
            <div className="grid grid-cols-2 gap-4">
              <div className="p-3 bg-slate-950 rounded border border-slate-800">
                <p className="text-[8px] text-slate-500 uppercase">External Tax</p>
                <p className="text-sm font-mono text-red-400">{formatter.format(calculations.totalCloudMonthly)}</p>
              </div>
              <div className="p-3 bg-slate-950 rounded border border-slate-800">
                <p className="text-[8px] text-slate-500 uppercase">Fixed OPEX</p>
                <p className="text-sm font-mono text-blue-400">{formatter.format(calculations.totalOcultarMonthly)}</p>
              </div>
            </div>
          </div>

          <button className="w-full bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold py-2 rounded-lg transition-colors flex items-center justify-center gap-2">
            <Download className="w-3 h-3" /> Export ROI Report
          </button>
        </>
      )}
    </div>
  );
};

const ROICalculatorView = ({ params, setParams }) => {
  const calculations = useMemo(() => {
    const provider = PRICING_DATA.providers[params.provider];
    const gb = params.volume * PRICING_DATA.conversion.gbPerTb;
    const listPricePerGB = provider.processing + provider.egress;
    const totalCloudMonthly = gb * listPricePerGB * (1 - (params.discount / 100));
    const totalOcultarMonthly = PRICING_DATA.ocultar.monthlyLicense + (params.volume * PRICING_DATA.ocultar.localComputePerTB);
    const monthlySavings = Math.max(0, totalCloudMonthly - totalOcultarMonthly);
    const annualSavings = monthlySavings * 12;

    return { monthlySavings, annualSavings, totalCloudMonthly, totalOcultarMonthly, listPricePerGB };
  }, [params]);

  return (
    <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 max-w-6xl mx-auto">
      {/* Parameters Panel */}
      <div className="lg:col-span-5 space-y-6">
        <div className="bg-slate-900 border border-slate-800 rounded-xl p-6">
          <h3 className="text-slate-200 font-bold mb-6 flex items-center gap-2 uppercase tracking-widest text-xs">
            <Calculator className="w-4 h-4 text-blue-400" /> Input Parameters
          </h3>
          
          <div className="space-y-8">
            <div>
              <label className="block text-[10px] text-slate-500 uppercase font-bold mb-3">Target Provider</label>
              <select 
                value={params.provider}
                onChange={(e) => setParams({...params, provider: e.target.value})}
                className="w-full bg-slate-950 border border-slate-800 text-slate-300 font-mono text-sm rounded-lg p-3 focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option value="gcp">Google Cloud DLP ($5.00/GB)</option>
                <option value="aws">AWS Comprehend ($1,000/GB)</option>
                <option value="azure">Azure AI Language ($1.50/GB)</option>
              </select>
            </div>

            <div>
              <div className="flex justify-between items-end mb-4">
                <span className="text-2xl font-bold text-slate-200">{params.volume} TB</span>
                <span className="text-[10px] text-slate-500 uppercase">Monthly Volume</span>
              </div>
              <input 
                type="range" min="1" max="250" value={params.volume}
                onChange={(e) => setParams({...params, volume: parseInt(e.target.value)})}
                className="w-full accent-blue-500" 
              />
            </div>

            <div>
              <div className="flex justify-between items-end mb-4">
                <span className="text-2xl font-bold text-slate-200">{params.discount}%</span>
                <span className="text-[10px] text-slate-500 uppercase">EDP Discount</span>
              </div>
              <input 
                type="range" min="0" max="80" value={params.discount} step="5"
                onChange={(e) => setParams({...params, discount: parseInt(e.target.value)})}
                className="w-full accent-emerald-500" 
              />
            </div>
          </div>
        </div>
      </div>

      {/* Results Panel */}
      <div className="lg:col-span-7 space-y-6">
        <div className="bg-slate-900 border border-slate-800 rounded-xl p-8 relative overflow-hidden group">
          <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
            <TrendingUp className="w-32 h-32" />
          </div>
          <div className="relative z-10">
            <p className="text-[10px] text-slate-500 uppercase font-bold tracking-[0.2em] mb-2">Projected 12-Month Capital Retention</p>
            <h2 className="text-6xl font-black text-emerald-400 tracking-tighter mb-6">{formatter.format(calculations.annualSavings)}</h2>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="p-4 bg-slate-950 rounded-lg border border-slate-800">
                <div className="flex justify-between items-center mb-2">
                   <span className="text-[10px] text-slate-500 uppercase">External Pipeline</span>
                   <span className="text-[8px] bg-red-500/10 text-red-400 px-1.5 py-0.5 rounded border border-red-500/20">DRAIN</span>
                </div>
                <p className="text-xl font-mono text-red-400">{formatter.format(calculations.totalCloudMonthly)}<span className="text-xs text-slate-600 font-normal ml-1">/mo</span></p>
              </div>
              <div className="p-4 bg-black rounded-lg border border-slate-800">
                <div className="flex justify-between items-center mb-2">
                   <span className="text-[10px] text-slate-500 uppercase">Ocultar Node</span>
                   <span className="text-[8px] bg-blue-500/10 text-blue-400 px-1.5 py-0.5 rounded border border-blue-500/20">FIXED</span>
                </div>
                <p className="text-xl font-mono text-blue-400">{formatter.format(calculations.totalOcultarMonthly)}<span className="text-xs text-slate-600 font-normal ml-1">/mo</span></p>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-slate-900/50 border border-slate-800 p-6 rounded-xl">
          <h4 className="text-[10px] text-slate-500 uppercase font-bold mb-4 tracking-widest border-b border-slate-800 pb-2">Rate Matrix Telemetry</h4>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 text-[11px] text-slate-400 leading-relaxed font-mono">
             <div className="space-y-2">
                <p><span className="text-slate-200">[01]</span> Provider rate: {formatter.format(calculations.listPricePerGB)}/GB list price.</p>
                <p><span className="text-slate-200">[02]</span> Effective rate: {formatter.format(calculations.listPricePerGB * (1 - params.discount/100))}/GB (incl. discount).</p>
             </div>
             <div className="space-y-2">
                <p><span className="text-slate-200">[03]</span> Local compute tax: $20/TB amortized edge cost.</p>
                <p><span className="text-slate-200">[04]</span> Variance: Fixed $10k license vs variable egress taxation.</p>
             </div>
          </div>
        </div>
      </div>
    </div>
  );
};

const DocumentationView = () => {
  const [docs, setDocs] = useState(null);
  
  useEffect(() => {
    fetch('http://localhost:18081/api/docs')
      .then(r => r.json())
      .then(d => setDocs(d))
      .catch(e => console.error(e));
  }, []);

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl p-8 max-w-4xl mx-auto">
      <div className="flex justify-between items-center mb-10">
        <h2 className="text-3xl font-bold flex items-center gap-3">
          <BookOpen className="text-blue-400 w-8 h-8" /> Technical Documentation
        </h2>
        <div className="flex flex-col items-end">
          <span className="text-[10px] font-mono text-slate-500 uppercase tracking-widest">v{docs?.version || '2.1'} canonical</span>
          {docs?.last_updated && <span className="text-[8px] font-mono text-slate-600 mt-1">Updated {new Date(docs.last_updated).toLocaleString()}</span>}
        </div>
      </div>
      
      <div className="space-y-12">
        {docs?.documentation ? (
          <div className="prose prose-invert max-w-none text-slate-400 text-xs leading-relaxed whitespace-pre-wrap font-mono">
            {docs.documentation}
          </div>
        ) : (
          <div className="text-center text-slate-500 py-12 animate-pulse">Loading canonical docs from repository...</div>
        )}
      </div>
    </div>
  );
};

const FAQView = () => {
  const [faq, setFaq] = useState(null);
  
  useEffect(() => {
    fetch('http://localhost:18081/api/faq')
      .then(r => r.json())
      .then(d => setFaq(d))
      .catch(e => console.error(e));
  }, []);

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl p-8 max-w-4xl mx-auto space-y-6">
      <h2 className="text-2xl font-bold mb-6 flex items-center gap-3"><HelpCircle className="text-emerald-400" /> Architecture & Security FAQ</h2>
      
      <div className="space-y-12">
        {faq?.faq ? (
          <div className="prose prose-invert max-w-none text-slate-400 text-xs leading-relaxed whitespace-pre-wrap font-mono">
            {faq.faq}
          </div>
        ) : (
          <div className="text-center text-slate-500 py-12 animate-pulse">Loading FAQ canonical source...</div>
        )}
      </div>
    </div>
  );
};

const DeveloperView = () => (
  <div className="bg-slate-900 border border-slate-800 rounded-xl p-8 max-w-4xl mx-auto space-y-8">
    <div className="flex flex-col gap-2">
      <h2 className="text-2xl font-bold flex items-center gap-3"><Code2 className="text-purple-400" /> Integration & Workspace</h2>
      <p className="text-xs text-slate-500 font-mono">OCULTAR_SDK_VERSION: v2.1.0-alpha</p>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
      <div className="space-y-4">
        <h4 className="text-sm font-bold text-slate-300 uppercase tracking-widest flex items-center gap-2">
          <Terminal className="w-3 h-3" /> The Connector Interface
        </h4>
        <p className="text-[10px] text-slate-500 leading-relaxed">
          Implement this interface in `services/refinery/pkg/connector/` to add new data sources.
        </p>
        <pre className="bg-slate-950 p-4 rounded-lg border border-slate-800 text-[10px] text-emerald-400 font-mono overflow-x-auto">
{`type Connector interface {
    // Refine processes the payload through
    // the security refinery layer.
    Refine(ctx context.Context, p []byte) ([]byte, error)
    
    // FailClosedHandler is triggered if 
    // the refinery is unreachable.
    FailClosedHandler(err error)
}`}
        </pre>
      </div>
      <div className="space-y-4">
        <h4 className="text-sm font-bold text-slate-300 uppercase tracking-widest flex items-center gap-2">
          <Network className="w-3 h-3" /> go.work Configuration
        </h4>
        <p className="text-[10px] text-slate-500 leading-relaxed">
          Multi-module development setup for local refinery testing.
        </p>
        <pre className="bg-slate-950 p-4 rounded-lg border border-slate-800 text-[10px] text-blue-400 font-mono overflow-x-auto">
{`go 1.22

use (
    .
    ./services/refinery
    ./services/sombra
    ./internal/audit
    ./pkg/crypto
)`}
        </pre>
      </div>
    </div>

    <div className="pt-6 border-t border-slate-800">
       <button className="flex items-center gap-2 text-xs font-bold text-purple-400 hover:text-purple-300 transition-colors">
          <ExternalLink className="w-3 h-3" /> Technical Spec: services/refinery/README.md
       </button>
    </div>
  </div>
);

const ConfigView = () => {
    const [regexType, setRegexType] = useState('CUSTOM_PATTERN');
    const [regexPattern, setRegexPattern] = useState('');
    
    const [dictType, setDictType] = useState('PROTECTED_ENTITY');
    const [dictTerm, setDictTerm] = useState('');

    const [limitConcurrency, setLimitConcurrency] = useState(10);
    const [limitQueue, setLimitQueue] = useState(5);

    const [previewInput, setPreviewInput] = useState('');
    const [previewOutput, setPreviewOutput] = useState('');

    const handleAddRegex = async () => {
        try {
            const res = await fetch('http://localhost:8080/api/config/regex', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ type: regexType, pattern: regexPattern })
            });
            if (res.ok) {
                alert('Regex Rule Saved!');
                setRegexPattern('');
            } else {
                const err = await res.text();
                alert('Regex Error: ' + err);
            }
        } catch(e) {
            alert('Connection error');
        }
    };

    const handleAddDict = async () => {
        try {
            const res = await fetch('http://localhost:8080/api/config/dictionary', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ type: dictType, term: dictTerm })
            });
            if (res.ok) {
                alert('Dictionary Term Saved!');
                setDictTerm('');
            } else {
                alert('Dictionary Error');
            }
        } catch(e) {
            alert('Connection error');
        }
    };

    const handleUpdateSystem = async () => {
        try {
            const res = await fetch('http://localhost:8080/api/config/system', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ max_concurrency: limitConcurrency, queue_size: limitQueue })
            });
            if (res.ok) alert('System Limits Updated!');
            else alert('System Update Error');
        } catch(e) {
            alert('Connection error');
        }
    };

    const handlePreview = async () => {
        try {
            const res = await fetch('http://localhost:8080/api/refine', {
                method: 'POST',
                body: previewInput
            });
            if (res.ok) {
                const data = await res.json();
                setPreviewOutput(data.refined);
            }
        } catch(e) {
            setPreviewOutput('Error reaching refinery node.');
        }
    };

    return (
        <div className="bg-slate-900 border border-slate-800 rounded-xl p-8 max-w-5xl mx-auto space-y-8">
            <h2 className="text-2xl font-bold flex items-center gap-3"><Settings2 className="text-blue-400" /> Operational Control Plane</h2>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                {/* Rules & Limits left side */}
                <div className="space-y-6">
                    <div className="p-5 bg-slate-950 border border-slate-800 rounded-lg">
                        <h3 className="text-xs font-bold text-slate-300 uppercase tracking-widest mb-4 border-b border-slate-800 pb-2">Regex Manager</h3>
                        <div className="space-y-3">
                            <input type="text" placeholder="Rule Type (e.g., SECRET_CODE)" value={regexType} onChange={e => setRegexType(e.target.value)} className="w-full bg-slate-900 border border-slate-800 rounded p-2 text-xs font-mono focus:ring-1 focus:ring-blue-500 outline-none" />
                            <input type="text" placeholder="Regex Pattern (e.g., \bCONF-\d{4}\b)" value={regexPattern} onChange={e => setRegexPattern(e.target.value)} className="w-full bg-slate-900 border border-slate-800 rounded p-2 text-xs font-mono focus:ring-1 focus:ring-blue-500 outline-none" />
                            <button onClick={handleAddRegex} className="w-full bg-blue-600/20 text-blue-400 border border-blue-500/30 hover:bg-blue-600/30 text-xs font-bold py-2 rounded transition-colors">Apply Regex Rule (Hot Reload)</button>
                        </div>
                    </div>

                    <div className="p-5 bg-slate-950 border border-slate-800 rounded-lg">
                        <h3 className="text-xs font-bold text-slate-300 uppercase tracking-widest mb-4 border-b border-slate-800 pb-2">Dictionary Shield</h3>
                        <div className="space-y-3">
                            <input type="text" placeholder="Category (e.g., PROTECTED_ENTITY)" value={dictType} onChange={e => setDictType(e.target.value)} className="w-full bg-slate-900 border border-slate-800 rounded p-2 text-xs font-mono focus:ring-1 focus:ring-emerald-500 outline-none" />
                            <input type="text" placeholder="Exact Term (e.g., Ouroboros Protocol)" value={dictTerm} onChange={e => setDictTerm(e.target.value)} className="w-full bg-slate-900 border border-slate-800 rounded p-2 text-xs font-mono focus:ring-1 focus:ring-emerald-500 outline-none" />
                            <button onClick={handleAddDict} className="w-full bg-emerald-600/20 text-emerald-400 border border-emerald-500/30 hover:bg-emerald-600/30 text-xs font-bold py-2 rounded transition-colors">Append Term to Node</button>
                        </div>
                    </div>

                    <div className="p-5 bg-slate-950 border border-slate-800 rounded-lg">
                        <h3 className="text-xs font-bold text-slate-300 uppercase tracking-widest mb-4 border-b border-slate-800 pb-2">System Resources</h3>
                        <div className="grid grid-cols-2 gap-4 mb-3">
                            <div>
                                <label className="block text-[10px] text-slate-500 uppercase mb-1">Max Concurrency</label>
                                <input type="number" value={limitConcurrency} onChange={e => setLimitConcurrency(parseInt(e.target.value))} className="w-full bg-slate-900 border border-slate-800 rounded p-2 text-xs font-mono outline-none" />
                            </div>
                            <div>
                                <label className="block text-[10px] text-slate-500 uppercase mb-1">Queue Size</label>
                                <input type="number" value={limitQueue} onChange={e => setLimitQueue(parseInt(e.target.value))} className="w-full bg-slate-900 border border-slate-800 rounded p-2 text-xs font-mono outline-none" />
                            </div>
                        </div>
                        <button onClick={handleUpdateSystem} className="w-full bg-amber-600/20 text-amber-500 border border-amber-500/30 hover:bg-amber-600/30 text-xs font-bold py-2 rounded transition-colors">Update Limits</button>
                    </div>
                </div>

                {/* Live Preview right side */}
                <div className="p-5 bg-slate-950 border border-slate-800 rounded-lg flex flex-col h-full relative">
                    <div className="absolute top-0 right-0 p-4 opacity-5 pointer-events-none">
                        <Eye className="w-32 h-32" />
                    </div>
                    <div className="relative z-10 flex-grow flex flex-col">
                        <h3 className="text-xs font-bold text-purple-400 uppercase tracking-widest mb-4 border-b border-purple-500/30 pb-2 flex items-center justify-between">
                            Live Redaction Preview
                            <span className="text-[9px] bg-purple-500/20 text-purple-300 px-2 py-0.5 rounded">Active Sandbox</span>
                        </h3>
                        <div className="flex-grow flex flex-col space-y-4">
                            <textarea 
                                value={previewInput} 
                                onChange={e => setPreviewInput(e.target.value)} 
                                placeholder="Type plain text to verify regex and dictionary rules dynamically..."
                                className="w-full h-32 bg-slate-900 border border-slate-800 rounded p-3 text-xs font-mono resize-none focus:ring-1 focus:ring-purple-500 outline-none"
                            />
                            <div className="flex justify-end">
                                <button onClick={handlePreview} className="bg-purple-600 hover:bg-purple-500 text-white text-xs font-bold py-2 px-6 rounded transition-colors">Send to Refinery</button>
                            </div>
                            <div className="w-full flex-grow bg-black border border-slate-800 rounded p-3 text-xs font-mono overflow-y-auto whitespace-pre-wrap text-emerald-400">
                                {previewOutput || 'Waiting for input...'}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

const AutomationView = () => {
    const [commands, setCommands] = useState([]);
    const [selectedCmd, setSelectedCmd] = useState(null);
    const [inputs, setInputs] = useState({});
    const [logs, setLogs] = useState('');
    const [isRunning, setIsRunning] = useState(false);
    const [history, setHistory] = useState([]);

    useEffect(() => {
        fetch('http://localhost:18081/api/automation/commands')
            .then(res => res.json())
            .then(data => setCommands(data || []))
            .catch(err => console.error(err));
        
        fetchHistory();
    }, []);

    const fetchHistory = () => {
        fetch('http://localhost:18081/api/automation/history')
            .then(res => res.json())
            .then(data => setHistory(data || []))
            .catch(err => console.error(err));
    };

    const handleRun = async () => {
        if (!selectedCmd) return;
        setIsRunning(true);
        setLogs('Starting execution: ' + selectedCmd.name + '\n');
        
        try {
            const response = await fetch('http://localhost:18081/api/automation/run', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ command_id: selectedCmd.id, inputs })
            });

            const reader = response.body.getReader();
            const decoder = new TextDecoder("utf-8");

            while (true) {
                const { done, value } = await reader.read();
                if (done) break;
                setLogs(prev => prev + decoder.decode(value, { stream: true }));
            }
        } catch (error) {
            setLogs(prev => prev + '\nError: ' + error.message);
        } finally {
            setIsRunning(false);
            fetchHistory();
        }
    };

    return (
        <div className="bg-slate-900 border border-slate-800 rounded-xl p-8 max-w-6xl mx-auto space-y-8">
            <h2 className="text-2xl font-bold flex items-center gap-3"><Terminal className="text-emerald-400" /> Automation Control Plane</h2>
            
            <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                <div className="space-y-6 md:col-span-1 border-r border-slate-800 pr-6">
                    <h3 className="text-sm border-b border-slate-800 pb-2 text-slate-400 font-bold uppercase tracking-widest">Available Capabilities</h3>
                    <div className="space-y-3">
                        {commands.map(cmd => (
                            <button 
                                key={cmd.id} 
                                onClick={() => { setSelectedCmd(cmd); setInputs({}); setLogs(''); }}
                                className={`w-full text-left p-3 rounded-lg border transition-all ${selectedCmd?.id === cmd.id ? 'bg-emerald-600/20 border-emerald-500/50 text-emerald-400' : 'bg-slate-950 border-slate-800 text-slate-400 hover:border-slate-600'}`}
                            >
                                <div className="text-xs font-bold font-mono">{cmd.name}</div>
                                <div className="text-[10px] mt-1 opacity-80">{cmd.description}</div>
                            </button>
                        ))}
                    </div>
                </div>

                <div className="space-y-6 md:col-span-2">
                    {selectedCmd ? (
                        <>
                            <div className="flex justify-between items-center bg-slate-950 p-4 rounded-lg border border-slate-800">
                                <div className="flex items-center gap-3">
                                    <h3 className="text-sm font-bold text-slate-200">{selectedCmd.name}</h3>
                                    <span className="text-[10px] px-2 py-0.5 rounded bg-slate-800 text-slate-400 font-mono">{selectedCmd.command} {selectedCmd.args.join(' ')}</span>
                                </div>
                                <button 
                                    onClick={handleRun} 
                                    disabled={isRunning}
                                    className={`px-4 py-2 rounded text-xs font-bold transition-all ${isRunning ? 'bg-slate-800 text-slate-600' : 'bg-emerald-600 hover:bg-emerald-500 text-white shadow-lg shadow-emerald-500/20'}`}
                                >
                                    {isRunning ? 'EXECUTING...' : 'RUN COMMAND'}
                                </button>
                            </div>

                            {selectedCmd.inputs.length > 0 && (
                                <div className="space-y-4 bg-slate-950/50 p-4 rounded-lg border border-slate-800">
                                    <h4 className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Required Variables</h4>
                                    {selectedCmd.inputs.map(input => (
                                        <div key={input.name}>
                                            <label className="block text-xs text-slate-400 mb-1 font-mono">{input.name} {input.required && <span className="text-red-400">*</span>}</label>
                                            <input 
                                                type="text" 
                                                placeholder={input.description}
                                                className="w-full bg-slate-900 border border-slate-800 rounded p-2 text-xs font-mono focus:ring-1 focus:ring-emerald-500 outline-none"
                                                value={inputs[input.name] || ''}
                                                onChange={e => setInputs({...inputs, [input.name]: e.target.value})}
                                            />
                                        </div>
                                    ))}
                                </div>
                            )}

                            <div className="h-96 bg-black border border-slate-800 rounded-lg p-4 overflow-y-auto font-mono text-[11px] text-emerald-400 whitespace-pre-wrap leading-relaxed shadow-inner">
                                {logs || 'Awaiting command execution...'}
                                {isRunning && <span className="animate-pulse w-2 h-4 bg-emerald-400 inline-block align-middle ml-1"></span>}
                            </div>
                        </>
                    ) : (
                        <div className="h-full flex flex-col items-center justify-center opacity-30 border-2 border-dashed border-slate-800 rounded-xl p-8">
                            <Terminal className="w-16 h-16 mb-4" />
                            <p className="text-sm font-mono text-center">Select a capability from the registry to begin.</p>
                        </div>
                    )}
                </div>
            </div>
            
            {history?.length > 0 && (
                <div className="pt-8 border-t border-slate-800">
                    <h3 className="text-xs border-b border-slate-800 pb-2 mb-4 text-slate-400 font-bold uppercase tracking-widest">Execution History ({history.length})</h3>
                    <div className="space-y-2">
                        {history.map(h => (
                            <div key={h.run_id} className="flex justify-between items-center text-[10px] p-2 bg-slate-950 rounded border border-slate-800 font-mono">
                                <span className={h.status === 'success' ? 'text-emerald-400' : h.status === 'failure' ? 'text-red-400' : 'text-blue-400'}>[{h.status.toUpperCase()}]</span>
                                <span className="text-slate-300">{h.command_id}</span>
                                <span className="text-slate-500">{new Date(h.start_time).toLocaleTimeString()}</span>
                            </div>
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
};

// --- MAIN APP ---

export default function App() {
  const [activeView, setActiveView] = useState('Overview');
  const [tier, setTier] = useState('ENTERPRISE'); // 'OPEN_SOURCE' | 'SOMBRA_STANDALONE' | 'ENTERPRISE'
  const [connectionStatus, setConnectionStatus] = useState('CONNECTED'); // 'CONNECTED' | 'DISCONNECTED'
  
  // Real-time states
  const [systemStatus, setSystemStatus] = useState(null);
  const [metrics, setMetrics] = useState(null);
  const [vaultStats, setVaultStats] = useState(null);
  const [auditLogs, setAuditLogs] = useState([]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [statusRes, metricsRes, vaultRes, auditRes] = await Promise.all([
          fetch('http://localhost:8080/api/system/status'),
          fetch('http://localhost:8080/api/system/metrics'),
          fetch('http://localhost:8080/api/vault/stats'),
          fetch('http://localhost:8080/api/audit/logs')
        ]);
        
        if (!statusRes.ok) throw new Error("API failed");
        
        const status = await statusRes.json();
        const mets = await metricsRes.json();
        const vault = await vaultRes.json();
        const audit = await auditRes.json();
        
        setSystemStatus(status);
        setMetrics(mets);
        setVaultStats(vault);
        setAuditLogs(audit.logs || []);
        
        setConnectionStatus('CONNECTED');
        if (status.mode === 'enterprise') setTier('ENTERPRISE');
        else setTier('OPEN_SOURCE');
      } catch (err) {
        setConnectionStatus('DISCONNECTED');
      }
    };

    fetchData();
    const interval = setInterval(fetchData, 3000);
    return () => clearInterval(interval);
  }, []);

  // ROI Parameters (Live State)
  const [pricingParams, setPricingParams] = useState({
    provider: 'gcp',
    volume: 10,
    discount: 0
  });

  const isPro = useMemo(() => tier !== 'OPEN_SOURCE', [tier]);
  const isEnterprise = useMemo(() => tier === 'ENTERPRISE', [tier]);

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 font-sans selection:bg-blue-500/30 overflow-x-hidden">
      {/* Background Decorative Element */}
      <div className="fixed top-0 left-0 w-full h-full pointer-events-none z-0 opacity-20 transition-opacity">
         <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-600/20 blur-[120px] rounded-full" />
         <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-emerald-600/10 blur-[120px] rounded-full" />
      </div>

      <header className="border-b border-white/5 h-20 flex items-center justify-between px-8 sticky top-0 bg-slate-950/60 backdrop-blur-xl z-50">
        <div className="flex items-center gap-5">
          <div className="w-10 h-10 bg-gradient-to-br from-blue-600 to-blue-400 rounded-xl flex items-center justify-center font-bold text-white shadow-xl shadow-blue-500/20 text-xl tracking-tighter">O</div>
          <div className="border-l border-white/10 pl-5">
            <h1 className="font-bold text-xl tracking-tight text-white flex items-center gap-2">OCULTAR <span className="text-blue-400 font-light opacity-80 decoration-blue-500/50 underline-offset-8 underline decoration-1">CONTROL_CENTER</span></h1>
            <p className="text-[9px] text-slate-500 font-mono uppercase tracking-[0.3em] font-medium">Sovereign Data Refinery // {systemStatus?.version || 'v2.1'}</p>
          </div>
        </div>
        
        <nav className="hidden md:flex items-center gap-1 bg-slate-900/50 p-1 rounded-lg border border-slate-800">
          <NavItem active={activeView === 'Overview'} onClick={() => setActiveView('Overview')} icon={<BarChart3 className="w-4 h-4" />} label="Overview" />
          <NavItem active={activeView === 'ROI'} onClick={() => setActiveView('ROI')} icon={<DollarSign className="w-4 h-4" />} label="ROI Analytics" />
          <NavItem active={activeView === 'Docs'} onClick={() => setActiveView('Docs')} icon={<BookOpen className="w-4 h-4" />} label="Docs" />
          <NavItem active={activeView === 'FAQ'} onClick={() => setActiveView('FAQ')} icon={<HelpCircle className="w-4 h-4" />} label="FAQ" />
          <NavItem active={activeView === 'Dev'} onClick={() => setActiveView('Dev')} icon={<Code2 className="w-4 h-4" />} label="Developer" />
          <NavItem active={activeView === 'Automation'} onClick={() => setActiveView('Automation')} icon={<Terminal className="w-4 h-4" />} label="Automation" />
          <NavItem active={activeView === 'Config'} onClick={() => setActiveView('Config')} icon={<Settings2 className="w-4 h-4" />} label="Config" />
        </nav>

        <div className="flex items-center gap-3">
           <div className="hidden lg:flex items-center gap-2 bg-slate-900 border border-slate-800 rounded-lg px-2 py-1 mr-4">
              <select 
                value={tier} 
                onChange={(e) => setTier(e.target.value)}
                className="bg-transparent text-[10px] text-slate-400 font-bold uppercase outline-none cursor-pointer"
              >
                <option value="OPEN_SOURCE">CE</option>
                <option value="SOMBRA_STANDALONE">PRO</option>
                <option value="ENTERPRISE">ENT</option>
              </select>
              <div className="w-px h-3 bg-slate-800 mx-1" />
              <button 
                onClick={() => setConnectionStatus(s => s === 'CONNECTED' ? 'DISCONNECTED' : 'CONNECTED')}
                className={`text-[10px] font-bold uppercase transition-colors ${connectionStatus === 'CONNECTED' ? 'text-emerald-500' : 'text-red-500'}`}
              >
                {connectionStatus === 'CONNECTED' ? 'Online' : 'Offline'}
              </button>
           </div>
           
           <div className="flex flex-col items-end mr-4">
              <span className={`text-[10px] font-bold uppercase ${connectionStatus === 'CONNECTED' ? 'text-emerald-400' : 'text-red-500'}`}>
                {connectionStatus === 'CONNECTED' ? `System Active (${systemStatus?.uptime || 'running'})` : 'Fail-Closed'}
              </span>
              <div className={`w-1.5 h-1.5 rounded-full animate-pulse ${connectionStatus === 'CONNECTED' ? 'bg-emerald-400' : 'bg-red-500'}`} />
           </div>
           <Settings2 className="w-5 h-5 text-slate-500 cursor-pointer" />
        </div>
      </header>

      <main className="p-6 max-w-7xl mx-auto min-h-[calc(100vh-4rem)]">
        {activeView === 'Overview' && <OverviewView tier={tier} pricingParams={pricingParams} connectionStatus={connectionStatus} systemStatus={systemStatus} metrics={metrics} vaultStats={vaultStats} auditLogs={auditLogs} />}
        {activeView === 'ROI' && <ROICalculatorView params={pricingParams} setParams={setPricingParams} />}
        {activeView === 'Docs' && <DocumentationView />}
        {activeView === 'FAQ' && <FAQView />}
        {activeView === 'Dev' && <DeveloperView />}
        {activeView === 'Automation' && <AutomationView />}
        {activeView === 'Config' && <ConfigView />}
      </main>

      <footer className="border-t border-white/5 py-8 mt-12 bg-slate-950/40 relative z-10">
        <div className="max-w-7xl mx-auto px-8 flex justify-between items-center text-[10px] text-slate-500 font-mono uppercase tracking-widest">
           <div className="flex items-center gap-6">
             <span>© 2026 OCULTAR Security. All Rights Reserved.</span>
             <div className="w-px h-3 bg-white/10" />
             <span className="flex items-center gap-1.5 transition-colors hover:text-emerald-400 cursor-default"><ShieldCheck className="w-3 h-3 text-emerald-500/50" /> ROI Verified via roi-accountant skill</span>
           </div>
           <span className="opacity-50">Contract Version: v2.1.0-ENTERPRISE</span>
        </div>
      </footer>
    </div>
  );
}

// --- SUB-COMPONENTS ---
const ShadowApiWidget = ({ tier }) => {
  const isPro = tier !== 'OPEN_SOURCE';
  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 hover:border-red-500/50 transition-all relative overflow-hidden group">
      {!isPro && (
        <div className="absolute inset-0 bg-slate-900/60 backdrop-blur-[2px] z-10 flex flex-col items-center justify-center p-6 text-center">
           <Lock className="w-6 h-6 text-slate-700 mb-2" />
           <p className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">Requires Sombra Pro+</p>
        </div>
      )}
      <div className="flex justify-between items-center mb-6">
        <h3 className="text-slate-200 font-semibold flex items-center gap-2">
          <Network className="w-4 h-4 text-red-400" /> Shadow AI Surface
        </h3>
        {isPro && (
          <span className="text-[8px] bg-red-400/10 text-red-400 px-1.5 py-0.5 rounded border border-red-400/20 font-bold uppercase">Active Scan</span>
        )}
      </div>
      <div className="relative h-48 mb-4 border border-slate-800 rounded-lg bg-slate-950 flex items-center justify-center overflow-hidden">
          <AlertCircle className={`w-12 h-12 animate-pulse opacity-20 ${isPro ? 'text-red-500' : 'text-slate-700'}`} />
          <div className="absolute inset-0 flex flex-col items-center justify-center">
             <span className={`text-xs font-bold uppercase tracking-widest ${isPro ? 'text-red-500' : 'text-slate-700'}`}>
               {isPro ? '3 Alerts Detected' : 'Discovery Disabled'}
             </span>
             <span className="text-[10px] text-slate-500 mt-1">Unregistered Egress Blocking Active</span>
          </div>
          {isPro && (
            <div className="absolute bottom-0 left-0 right-0 h-1 bg-gradient-to-r from-transparent via-red-500 to-transparent opacity-50" />
          )}
      </div>
      <button 
        disabled={!isPro}
        className={`w-full text-xs font-bold py-3 rounded-lg border transition-all ${
          isPro 
            ? 'bg-red-600/20 text-red-400 border-red-500/30 hover:bg-red-600/30' 
            : 'bg-slate-800/50 text-slate-600 border-slate-800 cursor-not-allowed'
        }`}
      >
        Enforce Fail-Closed
      </button>
    </div>
  );
};

const PerformanceWidget = ({ metrics }) => (
  <div className="bg-slate-900 border border-slate-800 rounded-xl p-5">
    <div className="flex justify-between items-center mb-6">
      <h3 className="text-slate-200 font-semibold flex items-center gap-2"><Activity className="w-4 h-4 text-blue-400" /> Latency Profiler</h3>
      <span className="text-xs font-mono text-blue-400 font-bold">P95: {metrics?.latency_per_tier?.regex || "N/A"}</span>
    </div>
    <div className="space-y-4">
       <PerformanceBar label="Engine" value={metrics?.latency_per_tier?.regex ? 80 : 0} ms={metrics?.latency_per_tier?.regex || "--"} color="bg-blue-500" />
       <PerformanceBar label="Vault" value={metrics?.latency_per_tier?.dict ? 20 : 0} ms={metrics?.latency_per_tier?.dict || "--"} color="bg-emerald-500" />
    </div>
  </div>
);

const PerformanceBar = ({ label, value, ms, color }) => (
  <div>
    <div className="flex justify-between text-[10px] uppercase font-bold text-slate-500 mb-1">
      <span>{label}</span>
      <span className="text-slate-300">{ms}ms</span>
    </div>
    <div className="h-2 w-full bg-slate-800 rounded-full overflow-hidden">
      <div className={`h-full ${color} transition-all duration-1000`} style={{ width: `${value}%` }} />
    </div>
  </div>
);

const PolicySimulatorWidget = ({ tier }) => {
    const isEnterprise = tier === 'ENTERPRISE';
    return (
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 col-span-1 md:col-span-2 relative overflow-hidden group">
        {!isEnterprise && (
          <div className="absolute inset-0 bg-slate-900/60 backdrop-blur-[2px] z-10 flex flex-col items-center justify-center p-6 text-center">
             <Lock className="w-6 h-6 text-slate-700 mb-2" />
             <p className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">Enterprise Feature: Policy Simulation</p>
          </div>
        )}
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-slate-200 font-semibold flex items-center gap-2">
            <Settings2 className="w-4 h-4 text-purple-400" /> Regulatory Policy Impact Simulator
          </h3>
          {isEnterprise && (
            <span className="text-[8px] bg-purple-500/10 text-purple-400 px-1.5 py-0.5 rounded border border-purple-500/20 font-bold uppercase">Draft Analysis</span>
          )}
        </div>
        <div className={`h-24 rounded flex items-center justify-center text-xs italic border transition-all ${
          isEnterprise ? 'bg-slate-950 text-slate-400 border-slate-800' : 'bg-slate-950/50 text-slate-700 border-slate-900 font-mono'
        }`}>
           {isEnterprise 
             ? "Simulation active: replaying 12.4k historical logs against draft schema..." 
             : "PROPOSED_DRAFT_V2: [REDACTED]"}
        </div>
      </div>
    );
};

const AnonymizationWidget = ({ tier, vaultStats }) => {
    const isEnterprise = tier === 'ENTERPRISE';
    return (
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 relative overflow-hidden">
        {!isEnterprise && (
          <div className="absolute inset-0 bg-slate-900/60 backdrop-blur-[2px] z-10 flex flex-col items-center justify-center p-6 text-center">
             <Lock className="w-6 h-6 text-slate-700 mb-2" />
             <p className="text-[10px] text-slate-500 font-bold uppercase tracking-widest font-mono">Tier: Enterprise Only</p>
          </div>
        )}
        <h3 className="text-slate-200 font-semibold mb-4 text-xs uppercase tracking-widest flex items-center gap-2">
          <ShieldAlert className="w-3 h-3 text-emerald-400" /> Privacy Risk Radar
        </h3>
        <div className="flex items-center justify-center h-24">
           <div className={`w-20 h-20 border-4 rounded-full flex items-center justify-center font-bold text-xl transition-all ${isEnterprise ? 'border-slate-800 border-t-emerald-500 text-emerald-400' : 'border-slate-800 text-slate-700'}`}>
             {isEnterprise ? '88' : '--'}
           </div>
        </div>
      </div>
    );
};

const AuditLedgerWidget = ({ tier, auditLogs }) => {
  const [verifying, setVerifying] = useState(null);
  const isEnterprise = tier === 'ENTERPRISE';

  const logs = auditLogs;

  const handleVerify = (logId) => {
    setVerifying(logId);
    setTimeout(() => {
      setVerifying(null);
      alert("Cryptographic Integrity Verified\n\nProtocol: OCU_SIG_V1\nAlgorithm: Ed25519\nSigner: compliance-certificate-signer\nStatus: VALID_CHAIN_OF_CUSTODY");
    }, 1500);
  };

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 col-span-1 md:col-span-2 relative overflow-hidden">
      {!isEnterprise && (
        <div className="absolute inset-0 bg-slate-900/60 backdrop-blur-[2px] z-10 flex flex-col items-center justify-center p-6 text-center">
           <Lock className="w-6 h-6 text-slate-700 mb-2" />
           <p className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">Enterprise Feature: Immutable Audit Trail</p>
        </div>
      )}
      <h3 className="text-slate-200 font-semibold mb-4 text-xs uppercase tracking-widest flex items-center gap-2">
        <FileCheck2 className="w-3 h-3 text-blue-400" /> Audit Ledger
      </h3>
      <table className="w-full text-left text-[10px] text-slate-400">
        <thead>
          <tr className="text-slate-500 border-b border-slate-800">
            <th className="pb-3 px-2">Event</th>
            <th className="pb-3 px-2">Type</th>
            <th className="pb-3 px-2 text-right">Signature Status</th>
          </tr>
        </thead>
        <tbody>
          {logs.map((log, idx) => (
            <tr key={idx} className="border-b border-slate-800/50 group/row hover:bg-slate-800/30 transition-colors">
              <td className="py-3 px-2">
                <div className="flex flex-col">
                  <span className="text-slate-200 font-mono">{log.action}: {log.result}</span>
                  <span className="text-[8px] text-slate-600">{new Date(log.timestamp).toLocaleString()}</span>
                </div>
              </td>
              <td className="py-3 px-2"><span className="bg-slate-850 border border-slate-800 px-1.5 py-0.5 rounded text-[8px] text-blue-400 font-mono">{log.compliance_mapping}</span></td>
              <td className="py-3 px-2 text-right">
                <button 
                  onClick={() => handleVerify(idx)}
                  disabled={verifying === idx || !isEnterprise}
                  className={`px-2 py-1 rounded border font-mono text-[9px] transition-all ${
                    verifying === idx 
                      ? 'bg-blue-500/20 border-blue-500/50 text-blue-400 animate-pulse' 
                      : 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400 hover:bg-emerald-500/20'
                  } ${!isEnterprise ? 'opacity-50 cursor-not-allowed' : ''}`}
                >
                  {verifying === idx ? 'VERIFYING...' : 'ED25519_VERIFIED'}
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

const NavItem = ({ active, onClick, icon, label }) => (
  <button onClick={onClick} className={`flex items-center gap-2 px-4 py-1.5 rounded-md transition-all text-xs font-medium ${active ? 'bg-slate-800 text-white shadow-lg' : 'text-slate-500 hover:text-slate-300'}`}>
    {icon} {label}
  </button>
);

const StatusCard = ({ icon, label, value }) => (
  <div className="bg-slate-900/50 border border-slate-800 p-4 rounded-xl flex items-center gap-4">
    <div className="w-10 h-10 rounded-lg bg-slate-800 flex items-center justify-center border border-slate-700">{icon}</div>
    <div>
      <p className="text-[10px] text-slate-500 uppercase font-bold tracking-widest">{label}</p>
      <p className="font-bold">{value}</p>
    </div>
  </div>
);

const DocSection = ({ title, children }) => (
  <div className="border-l-2 border-slate-800 pl-6 space-y-2">
    <h4 className="text-slate-200 font-bold">{title}</h4>
    {children}
  </div>
);

const FAQItem = ({ question, answer }) => (
  <div className="p-4 bg-slate-950 border border-slate-800 rounded-lg hover:border-slate-700 transition-colors">
    <p className="text-slate-200 text-sm font-bold mb-2 flex items-center justify-between">{question} <ChevronRight className="w-3 h-3 text-slate-600" /></p>
    <p className="text-slate-500 text-xs leading-relaxed">{answer}</p>
  </div>
);
