import React, { useState, useEffect, useCallback } from 'react';
import { 
  Activity, 
  Shield, 
  Zap, 
  Terminal, 
  Settings, 
  Database, 
  Lock,
  ArrowRight,
  Eye,
  FileText,
  X,
  Trash2,
  Plus,
  ClipboardCopy,
  CheckSquare,
  Square,
  ChevronDown,
  ChevronRight,
  BarChart2,
  Rocket,
  Check,
  AlertTriangle,
  BookOpen,
  Play
} from 'lucide-react';

const OCULTAR_API = '/api';

// ─── Pilot Data (sourced from docs/pilot/) ─────────────────────────────────

const PILOT_CHECKLIST = [
  { id: 'nda',       label: 'Execute NDA and Pilot Agreement' },
  { id: 'license',   label: 'Retrieve active OCU_LICENSE_KEY (Tier: enterprise) via keygen.go' },
  { id: 'volume',    label: 'Confirm expected daily API volume for connection pool sizing' },
  { id: 'docker',    label: 'Customer provisions Docker + Compose on target host' },
  { id: 'entities',  label: 'Ensure configs/protected_entities.json exists (fail-closed guarantee)' },
  { id: 'archive',   label: 'Share ocultar-enterprise.tar.gz with customer\'s technical lead' },
  { id: 'network',   label: 'Configure network: proxy binds to ${OCU_PROXY_PORT:-8081}' },
  { id: 'env',       label: 'Configure .env with OCU_MASTER_KEY, OCU_SALT, PROXY_TARGET, LICENSE_KEY' },
  { id: 'launch',    label: 'Launch: docker compose up -d — wait for [+] All pre-flight checks passed!' },
  { id: 'smoketest', label: 'Run automated smoke test: bash scripts/smoke_test.sh' },
  { id: 'verify',    label: 'Verify SIEM audit log is writing (Enterprise license required)' },
  { id: 'rehydrate', label: 'Customer verifies rehydrated responses (original PII restored)' },
];

const PILOT_COMMANDS = [
  {
    id: 'build',
    label: 'Step 1 — Build Enterprise Archive',
    description: 'Compile and package the Enterprise distribution from source.',
    command: 'bash tools/scripts/scripts/build_release.sh',
  },
  {
    id: 'launch',
    label: 'Step 2 — Launch Enterprise Proxy',
    description: 'Start the Docker cluster (downloads SLM model ~900 MB on first run).',
    command: 'docker compose up -d && docker compose logs -f | grep -m1 "pre-flight"',
  },
  {
    id: 'test',
    label: 'Step 3 — Send PII Test Payload',
    description: 'Verify the proxy intercepts and redacts sensitive data end-to-end.',
    command: `curl -s -X POST http://localhost:8081/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"My name is Sarah Connor and my IBAN is DE89370400440532013000"}]}'`,
  },
  {
    id: 'verify',
    label: 'Step 4 — Verify Redaction',
    description: 'Confirm PII was caught in the proxy logs.',
    command: 'docker logs ocultar-proxy | grep "redacted: true"',
  },
  {
    id: 'report',
    label: 'Step 5 — Generate Risk Report',
    description: 'Produce the financial exposure report for the pilot presentation.',
    command: 'go run services/refinery/cmd/riskreport/main.go -dataset datasets/leaky_demo.json -output pilot_risk_report.md -html pilot_risk_report.html',
  },
];

const PLAYBOOK_STEPS = [
  {
    id: 'prep',
    title: '1. Pre-Meeting Preparation',
    icon: <BookOpen className="w-4 h-4" />,
    points: [
      'Obtain technical buy-in to conduct a "black box" test.',
      'If the prospect won\'t share their own data, use datasets/leaky_demo.json to prove the concept.',
      'Ensure the OCULTAR Enterprise Proxy is running locally or accessible via the cloud gateway.',
    ],
  },
  {
    id: 'baseline',
    title: '2. The Baseline Test (The "Leaky" State)',
    icon: <AlertTriangle className="w-4 h-4" />,
    points: [
      'Run the dataset through their existing stack (regex filters, generic LLM redaction, or standard API logging).',
      'Goal: Show that standard regex fails on unstructured contexts, and LLMs hallucinate or skip PII boundaries.',
    ],
  },
  {
    id: 'ocultar',
    title: '3. The OCULTAR Test (The Proof)',
    icon: <Shield className="w-4 h-4" />,
    points: [
      'Process the dataset using the OCULTAR proxy.',
      'Point the Risk Report CLI at the proxy\'s output to quantify what was intercepted.',
      'Send a test PII payload to Port 8081 and show real-time redaction in the logs.',
    ],
  },
  {
    id: 'present',
    title: '4. The Presentation (The Close)',
    icon: <BarChart2 className="w-4 h-4" />,
    points: [
      'Open the generated pilot_risk_report.html.',
      'Focus on Financial Exposure range — ask them to project across their daily API volume.',
      'Highlight: Deterministic redaction (no LLM prompts for safety), Zero-Egress (never leaves VPC), Stateless horizontal scaling.',
      '"We just prevented €X–€Y in regulatory exposure in 5 milliseconds with zero egress. Let\'s talk pilot integration."',
    ],
  },
];

// ───────────────────────────────────────────────────────────────────────────

const Dashboard = () => {
  const [activeTab, setActiveTab] = useState('overview');
  const [systemStatus, setSystemStatus] = useState({ online: true, mode: 'Enterprise', version: 'v2.4.0', uptime: '0s' });
  const [metrics, setMetrics] = useState({ requests: 0, detections: 0, latency: '12ms', queue: 0 });
  const [auditLogs, setAuditLogs] = useState([]);
  const [isRefining, setIsRefining] = useState(false);
  const [testInput, setTestInput] = useState('');
  const [testOutput, setTestOutput] = useState('');
  
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [modalTitle, setModalTitle] = useState('');
  const [modalData, setModalData] = useState(null);
  
  const [configData, setConfigData] = useState(null);
  
  const [modalEndpoint, setModalEndpoint] = useState('');
  const [newRegex, setNewRegex] = useState({ type: '', pattern: '' });
  const [newDictTerm, setNewDictTerm] = useState({ type: '', term: '' });

  // Pilot tab state
  const [pilotChecklist, setPilotChecklist] = useState(() =>
    Object.fromEntries(PILOT_CHECKLIST.map(item => [item.id, false]))
  );
  const [pilotDataset, setPilotDataset] = useState('datasets/leaky_demo.json');
  const [isGeneratingReport, setIsGeneratingReport] = useState(false);
  const [pilotReport, setPilotReport] = useState(null);
  const [copiedCmd, setCopiedCmd] = useState(null);
  const [playbookOpen, setPlaybookOpen] = useState(null);

  // New Pilot History & Upload state
  const [pilotHistory, setPilotHistory] = useState([]);
  const [isUploading, setIsUploading] = useState(false);
  const [activeReportId, setActiveReportId] = useState(null);
  const [reportModalOpen, setReportModalOpen] = useState(false);

  // 1. Fetch System Status
  const fetchStatus = async () => {
    try {
      const res = await fetch(`${OCULTAR_API}/system/status`);
      const data = await res.json();
      setSystemStatus({ 
        online: true, 
        mode: data.mode === 'enterprise' ? 'Enterprise' : 'Community', 
        version: data.version,
        uptime: data.uptime
      });
      setMetrics(prev => ({ ...prev, queue: data.queue_depth || 0 }));
    } catch (e) {
      setSystemStatus(prev => ({ ...prev, online: false }));
    }
  };

  // 2. Fetch Metrics
  const fetchMetrics = async () => {
    try {
      const res = await fetch(`${OCULTAR_API}/system/metrics`);
      const data = await res.json();
      const detections = Object.values(data.pii_hits_per_type || {}).reduce((a, b) => a + b, 0);
      setMetrics({
        requests: Math.floor(data.requests_per_second * 3600),
        detections: detections,
        latency: data.latency_per_tier?.regex || '12ms',
        queue: metrics.queue
      });
    } catch (e) {
      console.error("Failed to fetch metrics", e);
    }
  };

  // 3. Fetch Audit Logs
  const fetchLogs = async () => {
    try {
      const res = await fetch(`${OCULTAR_API}/audit/logs`);
      const data = await res.json();
      setAuditLogs(data.logs || []);
    } catch (e) {
      console.error("Failed to fetch logs", e);
    }
  };

  // 4. Fetch Config if we are in the Config Tab
  const fetchConfig = async () => {
    try {
      const res = await fetch(`${OCULTAR_API}/config`);
      const data = await res.json();
      const sysRes = await fetch(`${OCULTAR_API}/config/system`);
      const sysData = await sysRes.json();
      setConfigData({ applicationConfig: data, systemSettings: sysData });
    } catch (e) {
      setConfigData({ error: 'Failed to load configuration' });
    }
  };

  useEffect(() => {
    fetchStatus();
    fetchMetrics();
    fetchLogs();
    const statusInterval = setInterval(fetchStatus, 5000);
    const metricsInterval = setInterval(fetchMetrics, 3000);
    const logsInterval = setInterval(fetchLogs, 5000);
    return () => {
      clearInterval(statusInterval);
      clearInterval(metricsInterval);
      clearInterval(logsInterval);
    };
  }, []);

  useEffect(() => {
    if (activeTab === 'config') fetchConfig();
    if (activeTab === 'pilot') fetchPilotHistory();
  }, [activeTab]);

  const fetchPilotHistory = async () => {
    try {
      const res = await fetch(`${OCULTAR_API}/pilot/history`);
      const data = await res.json();
      setPilotHistory(Array.isArray(data) ? data : []);
    } catch (e) { console.error("History fetch failed", e); }
  };

  const handleFileUpload = async (e) => {
    const file = e.target.files[0];
    if (!file) return;
    setIsUploading(true);
    const formData = new FormData();
    formData.append('dataset', file);
    try {
      const res = await fetch(`${OCULTAR_API}/pilot/upload`, {
        method: 'POST',
        body: formData
      });
      const data = await res.json();
      setPilotDataset(data.filename);
      alert(`Dataset "${data.original_name}" uploaded successfully!`);
    } catch (e) { alert("Upload failed."); }
    finally { setIsUploading(false); }
  };

  const handleTestRefinery = async () => {
    setIsRefining(true);
    setTestOutput('');
    try {
      const res = await fetch(`${OCULTAR_API}/refine`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text: testInput })
      });
      const data = await res.json();
      setTestOutput(data.refined || (typeof data === 'string' ? data : JSON.stringify(data, null, 2)));
    } catch (e) {
      setTestOutput("ERROR: Refinery unreachable.");
    } finally {
      setIsRefining(false);
    }
  };

  const handleOpenModal = async (title, endpoint) => {
    setIsModalOpen(true);
    setModalTitle(title);
    setModalEndpoint(endpoint);
    setModalData(null);
    try {
      const res = await fetch(`${OCULTAR_API}${endpoint}`);
      const data = await res.json();
      setModalData(data);
    } catch (e) {
      setModalData({ error: 'Failed to retrieve data.' });
    }
  };

  const handleDelete = async (type, endpoint) => {
    try {
      await fetch(`${OCULTAR_API}${endpoint}`, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ type })
      });
      handleOpenModal(modalTitle, modalEndpoint);
    } catch (e) {
      console.error("Delete failed", e);
    }
  };

  const handleAddRegex = async () => {
    try {
      await fetch(`${OCULTAR_API}/config/regex`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newRegex)
      });
      setNewRegex({ type: '', pattern: '' });
      handleOpenModal(modalTitle, modalEndpoint);
    } catch (e) {
      console.error("Add failed", e);
    }
  };

  const handleAddDictTerm = async () => {
    try {
      await fetch(`${OCULTAR_API}/config/dictionary`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newDictTerm)
      });
      setNewDictTerm({ ...newDictTerm, term: '' });
      handleOpenModal(modalTitle, modalEndpoint);
    } catch (e) {
      console.error("Add failed", e);
    }
  };

  // Pilot handlers
  const toggleChecklistItem = (id) => {
    setPilotChecklist(prev => ({ ...prev, [id]: !prev[id] }));
  };

  const handleCopyCommand = (id, command) => {
    navigator.clipboard.writeText(command).then(() => {
      setCopiedCmd(id);
      setTimeout(() => setCopiedCmd(null), 2000);
    });
  };

  const handleGeneratePilotReport = async () => {
    setIsGeneratingReport(true);
    setPilotReport(null);
    try {
      const res = await fetch(`${OCULTAR_API}/pilot/riskreport`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dataset_path: pilotDataset })
      });
      const data = await res.json();
      setPilotReport(data.report);
      fetchPilotHistory(); // Refresh history
      if (data.report_id) {
        setActiveReportId(data.report_id);
      }
    } catch (e) {
      setPilotReport({ error: 'Refinery offline or endpoint not yet wired.' });
    } finally {
      setIsGeneratingReport(false);
    }
  };

  const viewReport = (id) => {
    setActiveReportId(id);
    setReportModalOpen(true);
  };

  const pilotProgress = Object.values(pilotChecklist).filter(Boolean).length;

  const NAV_TABS = ['overview', 'pilot', 'config', 'logs'];

  return (
    <div className="min-h-screen bg-slate-50 text-slate-700 font-sans p-6 selection:bg-blue-500/30">
      {/* Top Summary Bar */}
      <header className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-8 pb-6 border-b border-slate-200">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 bg-blue-600 rounded flex items-center justify-center font-bold text-white text-lg shadow-lg shadow-blue-600/20">O</div>
          <div>
            <h1 className="text-xl font-bold text-slate-900 tracking-tight">OCULTAR <span className="font-light text-slate-500 tracking-widest">CONTROL_CENTER</span></h1>
            <div className="flex gap-4 mt-1">
              <span className="text-[10px] font-mono text-slate-500 uppercase flex items-center gap-1.5">
                <div className={`w-1.5 h-1.5 rounded-full ${systemStatus.online ? 'bg-emerald-500 animate-pulse' : 'bg-red-500'}`} />
                {systemStatus.online ? 'System Online' : 'System Offline'}
              </span>
              <span className="text-[10px] font-mono text-slate-500 uppercase border-l border-slate-200 pl-4">{systemStatus.mode} Edition</span>
              <span className="text-[10px] font-mono text-slate-500 uppercase border-l border-slate-200 pl-4">Uptime: {systemStatus.uptime}</span>
            </div>
          </div>
        </div>
        
        <nav className="flex bg-white p-1 rounded-lg border border-slate-200 shadow-sm gap-0.5">
          {NAV_TABS.map(tab => (
            <button 
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`px-4 py-1.5 rounded text-xs font-bold uppercase tracking-wider transition-all flex items-center gap-1.5 ${
                activeTab === tab ? 'bg-blue-600 text-white shadow-md' : 'text-slate-500 hover:text-slate-800'
              } ${tab === 'pilot' ? 'relative' : ''}`}
            >
              {tab === 'pilot' && <Rocket className="w-3 h-3" />}
              {tab}
              {tab === 'pilot' && pilotProgress > 0 && (
                <span className={`text-[9px] font-bold px-1.5 py-0.5 rounded-full ${activeTab === 'pilot' ? 'bg-white/20 text-white' : 'bg-blue-100 text-blue-700'}`}>
                  {pilotProgress}/{PILOT_CHECKLIST.length}
                </span>
              )}
            </button>
          ))}
        </nav>
      </header>

      {/* Main Grid */}
      <main className="max-w-7xl mx-auto space-y-6">
        
        {/* Metrics Row — always visible */}
        <section className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <MetricCard icon={<Activity className="w-4 h-4 text-blue-500" />} label="Requests / Hour" value={metrics.requests.toLocaleString()} description="Total throughput across all active security connectors." />
          <MetricCard icon={<Shield className="w-4 h-4 text-emerald-500" />} label="PII Detections" value={metrics.detections.toLocaleString()} description="Sensitive entities identified and tokenized by policy." />
          <MetricCard icon={<Zap className="w-4 h-4 text-amber-500" />} label="Refinery Latency" value={metrics.latency} description="Processing overhead for regex and local SLM inference." />
          <MetricCard icon={<Terminal className="w-4 h-4 text-purple-500" />} label="Worker Queue" value={metrics.queue} description="Pending redaction tasks awaiting availability in the refinery." />
        </section>

        {/* ── OVERVIEW TAB ── */}
        {activeTab === 'overview' && (
          <div className="grid lg:grid-cols-12 gap-6 animate-in fade-in duration-500">
            <div className="lg:col-span-4 space-y-6">
              <div className="bg-white border border-slate-200 shadow-sm rounded-xl p-5">
                <h3 className="text-xs font-bold text-slate-500 uppercase tracking-widest mb-4 flex items-center gap-2">
                  <ArrowRight className="w-4 h-4 text-blue-500" /> Live Refinery Test
                </h3>
                <p className="text-[10px] text-slate-400 mb-4 -mt-2">Instantly validate redaction policies against sample text or logs.</p>
                <textarea 
                  value={testInput} onChange={e => setTestInput(e.target.value)}
                  placeholder="Paste logs or JSON to test redaction..."
                  className="w-full h-32 bg-slate-50 border border-slate-200 rounded-lg p-3 text-xs font-mono mb-3 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 resize-none transition-colors"
                />
                <button 
                  onClick={handleTestRefinery} disabled={isRefining || !testInput}
                  className={`w-full py-3 rounded font-bold text-xs uppercase tracking-widest transition-all ${isRefining ? 'bg-slate-200 text-slate-500' : 'bg-blue-600 hover:bg-blue-500 text-white shadow hover:shadow-md active:scale-[0.98]'}`}
                >
                  {isRefining ? 'PROCESSING...' : 'EXECUTE REDACTION'}
                </button>
                {testOutput && (
                  <div className="mt-4 p-3 bg-emerald-50/50 border border-emerald-500/20 rounded font-mono text-[10px] text-emerald-700 break-all overflow-auto max-h-40">
                    {testOutput}
                  </div>
                )}
              </div>

              <div className="bg-white border border-slate-200 shadow-sm rounded-xl p-5">
                <h3 className="text-xs font-bold text-slate-500 uppercase tracking-widest mb-4 flex items-center gap-2">
                  <Settings className="w-4 h-4 text-slate-500" /> System Control
                </h3>
                <p className="text-[10px] text-slate-400 mb-4 -mt-2">Real-time enforcement audit and configuration management.</p>
                <div className="space-y-2">
                  <OpButton icon={<Lock className="w-4 h-4" />} label="Regex Enforcement Rules" onClick={() => handleOpenModal('Regex Rules', '/config/regex')} />
                  <OpButton icon={<Database className="w-4 h-4" />} label="Identity Dictionaries" onClick={() => handleOpenModal('Identity Dictionaries', '/config/dictionary')} />
                  <OpButton icon={<Eye className="w-4 h-4" />} label="Risk Compliance Radar" onClick={() => handleOpenModal('Risk Radar Data', '/audit/risk')} />
                  <OpButton icon={<Activity className="w-4 h-4" />} label="Canonical Entity Registry" onClick={() => handleOpenModal('Canonical Entity Mapping', '/config/mapping')} />
                </div>
              </div>
            </div>

            <div className="lg:col-span-8 flex flex-col h-full bg-white border border-slate-200 shadow-sm rounded-xl overflow-hidden">
              <div className="flex justify-between items-center px-5 py-4 border-b border-slate-100 bg-slate-50/50">
                <div>
                  <h3 className="text-xs font-bold text-slate-600 uppercase tracking-widest flex items-center gap-2">
                    <FileText className="w-4 h-4 text-blue-500" /> Secure Audit Trail
                  </h3>
                  <p className="text-[9px] text-slate-400 font-mono mt-1">Immutable journal of all redaction events and policy matches.</p>
                </div>
                <div className="flex gap-2">
                  <span className="text-[10px] text-emerald-600 bg-emerald-100 px-2 py-0.5 rounded border border-emerald-200 font-mono">ED25519_SIGNED</span>
                </div>
              </div>
              <div className="flex-grow p-4 font-mono text-[10.5px] space-y-1 overflow-y-auto min-h-[400px] bg-slate-50 border-b border-slate-100">
                {auditLogs.length > 0 ? auditLogs.map((log, i) => (
                  <LogLine key={i} time={new Date(log.timestamp).toLocaleTimeString()} level={log.result === 'SUCCESS' ? 'INFO' : 'WARN'} msg={`${log.action} [${log.compliance_mapping || 'N/A'}] - ${log.notes || 'No details'}`} color={log.result === 'SUCCESS' ? 'text-slate-600' : 'text-amber-600'} />
                )) : (
                  <div className="text-slate-500 italic px-2">No audit records found in audit.log...</div>
                )}
                <div className="flex items-center gap-2 text-slate-400 animate-pulse mt-2 pt-2 border-t border-slate-200/50">
                  <span>_</span>
                </div>
              </div>
              <div className="px-5 py-3 bg-white flex justify-between items-center text-[10px] text-slate-500 font-medium">
                <span className="uppercase tracking-widest">Fail-Closed Enforcement: Enabled</span>
                <button onClick={() => setActiveTab('logs')} className="text-blue-600 hover:text-blue-700 hover:underline uppercase font-bold">View raw audit.log</button>
              </div>
            </div>
          </div>
        )}

        {/* ── PILOT TAB ── */}
        {activeTab === 'pilot' && (
          <div className="animate-in fade-in duration-500 space-y-6">
            {/* Pilot Header Banner */}
            <div className="bg-gradient-to-r from-blue-700 to-blue-900 rounded-xl p-6 text-white flex items-start justify-between">
              <div>
                <div className="flex items-center gap-2 mb-1">
                  <Rocket className="w-5 h-5 opacity-80" />
                  <span className="text-xs font-bold uppercase tracking-widest opacity-70">Enterprise Pilot Control</span>
                </div>
                <h2 className="text-xl font-bold mb-1">OCULTAR Pilot Workspace</h2>
                <p className="text-sm opacity-70 max-w-xl">End-to-end guided workflow for delivering, verifying, and presenting the Enterprise Pilot to prospect stakeholders.</p>
              </div>
              <div className="text-right shrink-0 ml-6">
                <div className="text-3xl font-bold tabular-nums">{pilotProgress}<span className="text-lg opacity-50">/{PILOT_CHECKLIST.length}</span></div>
                <div className="text-xs opacity-60 uppercase tracking-wider mt-1">Tasks Complete</div>
                <div className="mt-2 w-32 h-1.5 bg-white/20 rounded-full overflow-hidden">
                  <div className="h-full bg-white/80 rounded-full transition-all duration-500" style={{ width: `${(pilotProgress / PILOT_CHECKLIST.length) * 100}%` }} />
                </div>
              </div>
            </div>

            <div className="grid lg:grid-cols-12 gap-6">
              {/* Left column: Checklist + Playbook */}
              <div className="lg:col-span-5 space-y-6">

                {/* Onboarding Checklist */}
                <div className="bg-white border border-slate-200 shadow-sm rounded-xl overflow-hidden">
                  <div className="px-5 py-4 border-b border-slate-100 bg-slate-50/50 flex items-center justify-between">
                    <h3 className="text-xs font-bold text-slate-600 uppercase tracking-widest flex items-center gap-2">
                      <CheckSquare className="w-4 h-4 text-blue-500" /> Onboarding Checklist
                    </h3>
                    <button onClick={() => setPilotChecklist(Object.fromEntries(PILOT_CHECKLIST.map(i => [i.id, false])))} className="text-[10px] text-slate-400 hover:text-slate-600 uppercase font-bold tracking-wider">Reset</button>
                  </div>
                  <div className="divide-y divide-slate-50">
                    {PILOT_CHECKLIST.map(item => (
                      <button
                        key={item.id}
                        onClick={() => toggleChecklistItem(item.id)}
                        className="w-full flex items-start gap-3 px-5 py-3 hover:bg-slate-50 transition-colors text-left group"
                      >
                        <span className={`mt-0.5 shrink-0 transition-colors ${pilotChecklist[item.id] ? 'text-emerald-500' : 'text-slate-300 group-hover:text-slate-400'}`}>
                          {pilotChecklist[item.id] ? <CheckSquare className="w-4 h-4" /> : <Square className="w-4 h-4" />}
                        </span>
                        <span className={`text-xs leading-relaxed transition-colors ${pilotChecklist[item.id] ? 'text-slate-400 line-through' : 'text-slate-700'}`}>
                          {item.label}
                        </span>
                      </button>
                    ))}
                  </div>
                </div>

                {/* Sales Playbook */}
                <div className="bg-white border border-slate-200 shadow-sm rounded-xl overflow-hidden">
                  <div className="px-5 py-4 border-b border-slate-100 bg-slate-50/50">
                    <h3 className="text-xs font-bold text-slate-600 uppercase tracking-widest flex items-center gap-2">
                      <BarChart2 className="w-4 h-4 text-blue-500" /> Sales Playbook
                    </h3>
                    <p className="text-[10px] text-slate-400 mt-1">Data Risk Assessment motion for converting Enterprise Pilot prospects.</p>
                  </div>
                  <div className="divide-y divide-slate-100">
                    {PLAYBOOK_STEPS.map((step, idx) => (
                      <div key={step.id}>
                        <button
                          onClick={() => setPlaybookOpen(playbookOpen === idx ? null : idx)}
                          className="w-full flex items-center justify-between px-5 py-3.5 hover:bg-slate-50 transition-colors text-left"
                        >
                          <span className="flex items-center gap-2.5 text-xs font-semibold text-slate-700">
                            <span className={`p-1 rounded ${playbookOpen === idx ? 'text-blue-600 bg-blue-50' : 'text-slate-400'}`}>{step.icon}</span>
                            {step.title}
                          </span>
                          {playbookOpen === idx
                            ? <ChevronDown className="w-4 h-4 text-slate-400 shrink-0" />
                            : <ChevronRight className="w-4 h-4 text-slate-300 shrink-0" />
                          }
                        </button>
                        {playbookOpen === idx && (
                          <div className="px-5 pb-4 bg-slate-50/50 border-t border-slate-100">
                            <ul className="space-y-2 mt-3">
                              {step.points.map((point, pIdx) => (
                                <li key={pIdx} className="flex gap-2 text-xs text-slate-600">
                                  <span className="text-blue-500 shrink-0 mt-0.5">▸</span>
                                  <span>{point}</span>
                                </li>
                              ))}
                            </ul>
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              {/* Right column: Quick Commands + Report Runner */}
              <div className="lg:col-span-7 space-y-6">

                {/* Quick Commands */}
                <div className="bg-white border border-slate-200 shadow-sm rounded-xl overflow-hidden">
                  <div className="px-5 py-4 border-b border-slate-100 bg-slate-50/50">
                    <h3 className="text-xs font-bold text-slate-600 uppercase tracking-widest flex items-center gap-2">
                      <Terminal className="w-4 h-4 text-blue-500" /> Quick Commands
                    </h3>
                    <p className="text-[10px] text-slate-400 mt-1">Copy and paste into your terminal — runs in ~/ocultar unless noted.</p>
                  </div>
                  <div className="divide-y divide-slate-100">
                    {PILOT_COMMANDS.map(cmd => (
                      <div key={cmd.id} className="px-5 py-4">
                        <div className="flex items-start justify-between gap-3 mb-2">
                          <div>
                            <div className="text-xs font-bold text-slate-800">{cmd.label}</div>
                            <div className="text-[10px] text-slate-400 mt-0.5">{cmd.description}</div>
                          </div>
                          <button
                            onClick={() => handleCopyCommand(cmd.id, cmd.command)}
                            className={`shrink-0 flex items-center gap-1.5 px-3 py-1.5 rounded text-[10px] font-bold uppercase tracking-wider transition-all ${copiedCmd === cmd.id ? 'bg-emerald-100 text-emerald-700 border border-emerald-200' : 'bg-slate-100 text-slate-600 hover:bg-blue-50 hover:text-blue-700 border border-slate-200 hover:border-blue-200'}`}
                          >
                            {copiedCmd === cmd.id ? <Check className="w-3 h-3" /> : <ClipboardCopy className="w-3 h-3" />}
                            {copiedCmd === cmd.id ? 'Copied' : 'Copy'}
                          </button>
                        </div>
                        <pre className="bg-slate-900 text-emerald-400 text-[10px] font-mono p-3 rounded-lg overflow-x-auto leading-relaxed whitespace-pre-wrap">{cmd.command}</pre>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Inline Risk Report Runner */}
                <div className="bg-white border border-slate-200 shadow-sm rounded-xl overflow-hidden">
                  <div className="px-5 py-4 border-b border-slate-100 bg-slate-50/50">
                    <h3 className="text-xs font-bold text-slate-600 uppercase tracking-widest flex items-center gap-2">
                      <BarChart2 className="w-4 h-4 text-blue-500" /> Risk Report Generator
                    </h3>
                    <p className="text-[10px] text-slate-400 mt-1">Generate an inline risk assessment with estimated VaR range for the pilot dataset.</p>
                  </div>
                  <div className="p-5">
                    <div className="flex flex-col gap-3 mb-6">
                      <div className="flex flex-col gap-1.5">
                        <label className="text-[10px] font-bold text-slate-500 uppercase tracking-wider ml-1">1. Prospect Dataset Source</label>
                        <div className="flex gap-2">
                          <input
                            type="text"
                            value={pilotDataset}
                            onChange={e => setPilotDataset(e.target.value)}
                            placeholder="Enter dataset path (e.g. datasets/leaky_demo.json)..."
                            className="flex-1 text-xs font-mono p-3 bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:border-blue-500 transition-colors"
                          />
                          <label className="shrink-0 flex items-center justify-center gap-2 px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-[11px] font-bold uppercase tracking-wider cursor-pointer shadow-md shadow-emerald-600/20 transition-all active:scale-[0.98]">
                            <Plus className="w-4 h-4" /> Upload Custom
                            <input type="file" className="hidden" accept=".json,.csv" onChange={handleFileUpload} disabled={isUploading} />
                          </label>
                        </div>
                        {isUploading && <div className="text-[10px] text-blue-600 font-bold ml-1 animate-pulse italic">Uploading prospect dataset...</div>}
                      </div>

                      <div className="flex flex-col gap-1.5">
                        <label className="text-[10px] font-bold text-slate-500 uppercase tracking-wider ml-1">2. Run Risk Engine</label>
                        <button
                          onClick={handleGeneratePilotReport}
                          disabled={isGeneratingReport || !pilotDataset}
                          className={`flex items-center justify-center gap-2 w-full py-3 rounded-lg text-xs font-bold uppercase tracking-widest transition-all ${isGeneratingReport ? 'bg-slate-200 text-slate-500' : 'bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-600/20 active:scale-[0.98]'}`}
                        >
                          {isGeneratingReport ? (
                            <>
                              <div className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                              Modeling Regulatory Exposure...
                            </>
                          ) : (
                            <>
                              <Zap className="w-4 h-4 fill-white" />
                              Generate Senior Specialist Report
                            </>
                          )}
                        </button>
                      </div>
                    </div>

                    {pilotReport && !pilotReport.error && (
                      <div className="space-y-3 animate-in fade-in duration-300">
                        {/* Overall risk banner */}
                        <div className={`rounded-lg p-4 border ${
                          pilotReport.overall_risk_level === 'CRITICAL' ? 'bg-red-50 border-red-200' :
                          pilotReport.overall_risk_level === 'HIGH' ? 'bg-orange-50 border-orange-200' :
                          pilotReport.overall_risk_level === 'MEDIUM' ? 'bg-amber-50 border-amber-200' :
                          'bg-emerald-50 border-emerald-200'
                        }`}>
                          <div className="flex items-center justify-between mb-3">
                            <div>
                              <div className="text-xs font-bold text-slate-700 uppercase tracking-wider">{pilotReport.overall_risk_level} Risk</div>
                              <div className="text-[10px] text-slate-500 mt-0.5">Score: {pilotReport.overall_risk_score?.toFixed(1)}/10 · K={pilotReport.k_anonymity} · L={pilotReport.l_diversity}</div>
                            </div>
                            <div className="text-right">
                              <div className="text-[10px] text-slate-500 uppercase font-bold">Est. Exposure Range</div>
                              <div className="text-sm font-bold text-slate-800">
                                €{Math.round(pilotReport.financial_exposure?.var_min_eur || 0).toLocaleString()} – €{Math.round(pilotReport.financial_exposure?.var_max_eur || 0).toLocaleString()}
                              </div>
                            </div>
                          </div>
                          <div className="grid grid-cols-3 gap-2 mb-4">
                            <div className="bg-white/60 rounded p-2 text-center">
                              <div className="text-[9px] text-slate-500 uppercase font-bold">Records</div>
                              <div className="text-sm font-bold text-slate-800">{pilotReport.total_records}</div>
                            </div>
                            <div className="bg-white/60 rounded p-2 text-center">
                              <div className="text-[9px] text-slate-500 uppercase font-bold">Violating</div>
                              <div className="text-sm font-bold text-red-600">{pilotReport.violating_records}</div>
                            </div>
                            <div className="bg-white/60 rounded p-2 text-center">
                              <div className="text-[9px] text-slate-500 uppercase font-bold">AI Status</div>
                              <div className={`text-xs font-bold ${pilotReport.ai_readiness?.status === 'BLOCK' ? 'text-red-600' : pilotReport.ai_readiness?.status === 'SANITIZE_FIRST' ? 'text-amber-600' : 'text-emerald-600'}`}>
                                {pilotReport.ai_readiness?.status}
                              </div>
                            </div>
                          </div>
                        </div>
                        <div className="flex gap-2">
                          <button 
                            onClick={() => viewReport(activeReportId)}
                            className="flex-1 flex items-center justify-center gap-2 py-3 bg-slate-900 text-white rounded-lg text-xs font-bold uppercase tracking-widest hover:bg-slate-800 transition-all"
                          >
                            <Eye className="w-4 h-4" /> View Detailed HTML Report
                          </button>
                        </div>
                        <div className="p-3 bg-slate-50 border border-slate-200 rounded-lg">
                          <div className="text-[10px] text-slate-500 font-bold uppercase tracking-wider mb-1">Assumptions Note</div>
                          <p className="text-[10px] text-slate-600 leading-relaxed">{pilotReport.financial_exposure?.assumptions_note}</p>
                        </div>
                      </div>
                    )}

                    {pilotReport?.error && (
                      <div className="p-4 bg-amber-50 border border-amber-200 rounded-lg">
                        <div className="flex items-start gap-2">
                          <AlertTriangle className="w-4 h-4 text-amber-600 shrink-0 mt-0.5" />
                          <div>
                            <div className="text-xs font-bold text-amber-800 mb-1">Inline Report Unavailable</div>
                            <p className="text-[10px] text-amber-700">{pilotReport.error}</p>
                            <div className="mt-2 p-2 bg-amber-100 rounded font-mono text-[10px] text-amber-800">
                              go run services/refinery/cmd/riskreport/main.go -dataset {pilotDataset} -output pilot_risk_report.md -html pilot_risk_report.html
                            </div>
                          </div>
                        </div>
                      </div>
                    )}

                    {!pilotReport && (
                      <div className="text-center py-8 text-slate-400">
                        <BarChart2 className="w-8 h-8 mx-auto mb-2 opacity-30" />
                        <p className="text-xs">Enter a dataset path and click Run to generate an inline risk assessment.</p>
                      </div>
                    )}
                  </div>
                </div>

                {/* Report History */}
                <div className="bg-white border border-slate-200 shadow-sm rounded-xl overflow-hidden mt-6">
                  <div className="px-5 py-4 border-b border-slate-100 bg-slate-50/50 flex justify-between items-center">
                    <h3 className="text-xs font-bold text-slate-600 uppercase tracking-widest flex items-center gap-2">
                       <Database className="w-4 h-4 text-blue-500" /> Recent Assessments
                    </h3>
                  </div>
                  <div className="divide-y divide-slate-100 max-h-[300px] overflow-y-auto">
                    {pilotHistory.length > 0 ? pilotHistory.map(item => (
                      <div key={item.id} className="px-5 py-3 flex items-center justify-between hover:bg-slate-50 transition-colors">
                        <div className="min-w-0 flex-1">
                          <div className="text-[10px] font-bold text-slate-800 truncate">{item.dataset_name}</div>
                          <div className="text-[9px] text-slate-400 font-mono mt-0.5">{item.timestamp} · {item.total_records} records</div>
                        </div>
                        <div className="flex items-center gap-3 shrink-0 ml-4">
                          <span className={`text-[9px] font-bold px-1.5 py-0.5 rounded ${
                            item.overall_risk === 'CRITICAL' ? 'bg-red-100 text-red-700' :
                            item.overall_risk === 'HIGH' ? 'bg-orange-100 text-orange-700' :
                            'bg-emerald-100 text-emerald-700'
                          }`}>
                            {item.overall_risk}
                          </span>
                          <button onClick={() => viewReport(item.id)} className="p-1.5 text-blue-600 hover:bg-blue-50 rounded transition-all">
                            <Eye className="w-3.5 h-3.5" />
                          </button>
                        </div>
                      </div>
                    )) : (
                      <div className="p-8 text-center text-slate-400 text-xs italic">No past reports found.</div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* ── CONFIG TAB ── */}
        {activeTab === 'config' && (
          <div className="bg-white border border-slate-200 shadow-sm rounded-xl p-8 animate-in slide-in-from-top-4 duration-300">
            <div className="flex items-center justify-between mb-6 pb-4 border-b border-slate-100">
              <div className="flex items-center gap-3">
                <Settings className="w-6 h-6 text-slate-700" />
                <div>
                  <h2 className="text-xl font-bold text-slate-900">Configuration Plane</h2>
                  <p className="text-xs text-slate-500 mt-1">Direct visibility into the OCULTAR runtime environment and security policy state.</p>
                </div>
              </div>
            </div>
            <div className="h-[500px] bg-slate-50 border border-slate-200 rounded p-4 overflow-auto font-mono text-xs">
              {!configData ? (
                <div className="text-slate-500 animate-pulse">Fetching latest runtime configuration...</div>
              ) : (
                <pre className="text-slate-800 whitespace-pre-wrap">{JSON.stringify(configData, null, 2)}</pre>
              )}
            </div>
          </div>
        )}

        {/* ── LOGS TAB ── */}
        {activeTab === 'logs' && (
          <div className="bg-white shadow-sm border border-slate-200 rounded-xl p-4 font-mono text-[11px] overflow-auto h-[600px] animate-in slide-in-from-top-4 duration-300">
            <div className="text-blue-600 font-bold mb-4 border-b border-slate-100 pb-3 flex justify-between bg-slate-50 p-2 rounded items-center">
              <span>SYSTEM_JOURNAL_V1</span>
              <span>TAIL -F /var/log/ocultar.log</span>
            </div>
            <div className="p-2 space-y-1.5">
              {auditLogs.map((log, i) => (
                <div key={i} className="flex gap-3 items-start border-l-2 border-slate-200 pl-3 hover:bg-slate-50 py-1 transition-colors">
                  <span className="text-slate-400 shrink-0">[{new Date(log.timestamp).toISOString()}]</span>
                  <span className="text-blue-600 font-bold w-32 shrink-0">{log.action}</span>
                  <span className="text-slate-700 break-all">{JSON.stringify(log)}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </main>

      {/* Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/40 backdrop-blur-sm animate-in fade-in duration-200">
          <div className="bg-white border border-slate-200 shadow-2xl rounded-xl w-full max-w-3xl max-h-[80vh] flex flex-col overflow-hidden animate-in zoom-in-95 duration-200">
            <div className="flex justify-between items-center px-6 py-4 border-b border-slate-100 bg-slate-50">
              <h2 className="text-sm font-bold text-slate-800 uppercase tracking-widest">{modalTitle}</h2>
              <button onClick={() => setIsModalOpen(false)} className="p-1 hover:bg-slate-200 rounded transition-colors text-slate-500">
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="p-0 overflow-auto bg-white flex-grow">
              {!modalData ? (
                <div className="p-6 text-slate-500 text-sm animate-pulse flex items-center gap-2">
                  <div className="w-4 h-4 block border-2 border-slate-300 border-t-blue-600 rounded-full animate-spin"></div>
                  Retrieving policy data...
                </div>
              ) : Array.isArray(modalData) ? (
                <div className="flex flex-col">
                  <div className="divide-y divide-slate-100 max-h-[400px] overflow-y-auto">
                    {modalData.map((item, idx) => (
                      <div key={idx} className="flex items-center justify-between px-6 py-3 hover:bg-slate-50 transition-colors">
                        <div className="flex flex-col gap-1">
                          <div className="flex items-center gap-2">
                            <span className="font-bold text-slate-800 text-xs">{item.type}</span>
                            {item.canonical_mapping && (
                              <span className="text-[9px] bg-blue-50 text-blue-600 px-1.5 py-0.5 rounded border border-blue-100 font-medium">Mapped: {item.canonical_mapping}</span>
                            )}
                          </div>
                          <code className="text-[10px] text-slate-500 truncate max-w-md">
                            {item.pattern || (item.terms ? item.terms.join(', ') : '')}
                          </code>
                        </div>
                        <button onClick={() => handleDelete(item.type, modalEndpoint)} className="p-1.5 text-slate-400 hover:text-red-500 hover:bg-red-50 rounded transition-all" title="Remove Rule">
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    ))}
                    {modalData.length === 0 && <div className="p-6 text-slate-400 italic text-xs">No entries configured.</div>}
                  </div>
                  <div className="p-6 bg-slate-50 border-t border-slate-100">
                    <h4 className="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-3">Add New {modalTitle.includes('Regex') ? 'Regex Rule' : 'Dictionary Term'}</h4>
                    {modalTitle.includes('Regex') ? (
                      <div className="flex gap-2">
                        <input type="text" placeholder="Type (e.g. SSN)" className="flex-1 text-[11px] p-2 border border-slate-200 rounded focus:border-blue-500 outline-none font-mono" value={newRegex.type} onChange={e => setNewRegex({...newRegex, type: e.target.value})} />
                        <input type="text" placeholder="Pattern (e.g. \b\d{3}...)" className="flex-[2] text-[11px] p-2 border border-slate-200 rounded focus:border-blue-500 outline-none font-mono" value={newRegex.pattern} onChange={e => setNewRegex({...newRegex, pattern: e.target.value})} />
                        <button onClick={handleAddRegex} className="px-3 bg-blue-600 text-white rounded hover:bg-blue-500 transition-colors"><Plus className="w-4 h-4" /></button>
                      </div>
                    ) : modalTitle.includes('Dictionaries') ? (
                      <div className="flex gap-2">
                        <input type="text" placeholder="Category (e.g. VIP)" className="flex-1 text-[11px] p-2 border border-slate-200 rounded focus:border-blue-500 outline-none font-mono" value={newDictTerm.type} onChange={e => setNewDictTerm({...newDictTerm, type: e.target.value})} />
                        <input type="text" placeholder="New Term (e.g. John Doe)" className="flex-[2] text-[11px] p-2 border border-slate-200 rounded focus:border-blue-500 outline-none font-mono" value={newDictTerm.term} onChange={e => setNewDictTerm({...newDictTerm, term: e.target.value})} />
                        <button onClick={handleAddDictTerm} className="px-3 bg-blue-600 text-white rounded hover:bg-blue-500 transition-colors"><Plus className="w-4 h-4" /></button>
                      </div>
                    ) : (
                      <div className="text-[10px] text-slate-400 italic">Advanced configuration is restricted to Read-Only in this view.</div>
                    )}
                  </div>
                </div>
              ) : (
                <div className="p-6">
                  <pre className="text-xs font-mono text-slate-700 whitespace-pre-wrap bg-slate-50 p-4 border border-slate-200 rounded">
                    {JSON.stringify(modalData, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Report Modal */}
      {reportModalOpen && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-md animate-in fade-in duration-200">
          <div className="bg-white border border-slate-200 shadow-2xl rounded-2xl w-full max-w-5xl h-[85vh] flex flex-col overflow-hidden animate-in zoom-in-95 duration-200">
            <div className="flex justify-between items-center px-6 py-4 border-b border-slate-100 bg-slate-50/80">
              <div>
                <h2 className="text-sm font-bold text-slate-800 uppercase tracking-widest flex items-center gap-2">
                  <Shield className="w-4 h-4 text-blue-600" /> Senior Risk Specialist Assessment
                </h2>
                <div className="text-[10px] text-slate-400 font-mono mt-0.5 tracking-wider uppercase">Deliverable ID: OCU-{activeReportId?.toUpperCase()}</div>
              </div>
              <button onClick={() => setReportModalOpen(false)} className="p-2 hover:bg-slate-200 rounded-full transition-colors text-slate-500">
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="flex-grow bg-slate-100 relative">
              <iframe 
                src={`http://localhost:8080/api/pilot/report?id=${activeReportId}`} 
                className="w-full h-full border-none"
                title="Risk Report Viewer"
              />
            </div>
            <div className="px-6 py-4 bg-white border-t border-slate-100 flex justify-between items-center">
               <span className="text-[11px] text-slate-400 font-medium">OCULTAR Enterprise Methodology v3.1 · Simulation Anchor Model</span>
               <button onClick={() => window.open(`http://localhost:8080/api/pilot/report?id=${activeReportId}`, '_blank')} className="text-blue-600 text-[11px] font-bold uppercase tracking-wider hover:underline flex items-center gap-1.5">
                  <Eye className="w-4 h-4" /> Open in New Tab
               </button>
            </div>
          </div>
        </div>
      )}

      <footer className="mt-12 pt-8 border-t border-slate-200 flex justify-between items-center px-4">
        <p className="text-[9px] text-slate-500 font-mono uppercase tracking-[0.3em]">
          OCULTAR Control Center // Local Operator View
        </p>
        <p className="text-[9px] text-slate-500 font-mono">
          BUILD_SHA: 82fb12a
        </p>
      </footer>
    </div>
  );
};

const MetricCard = ({ icon, label, value, description }) => (
  <div className="bg-white border border-slate-200 shadow-sm p-5 rounded-xl hover:shadow-md transition-all group flex flex-col justify-between">
    <div className="flex items-center gap-3 mb-4">
      <div className="w-8 h-8 rounded bg-slate-50 border border-slate-100 flex items-center justify-center opacity-80 group-hover:opacity-100 group-hover:bg-slate-100 transition-all transform duration-300">
        {icon}
      </div>
      <div className="flex flex-col">
        <span className="text-[10px] text-slate-500 uppercase font-bold tracking-widest leading-none">{label}</span>
        {description && <span className="text-[8.5px] text-slate-400 font-normal leading-relaxed mt-1 opacity-0 group-hover:opacity-100 transition-opacity duration-300">{description}</span>}
      </div>
    </div>
    <div className="text-2xl font-bold text-slate-900 tabular-nums tracking-tight">{value}</div>
  </div>
);

const OpButton = ({ icon, label, onClick }) => (
  <button 
    onClick={onClick}
    className="w-full flex items-center justify-between p-4 rounded-lg bg-slate-50 border border-slate-200 shadow-sm hover:border-blue-400 hover:bg-blue-50/50 group transition-all text-left active:scale-[0.99]"
  >
    <span className="flex items-center gap-3 text-xs font-semibold text-slate-700 group-hover:text-blue-700">
      <span className="text-slate-500 group-hover:text-blue-600">{icon}</span>
      {label}
    </span>
    <ArrowRight className="w-4 h-4 text-slate-400 group-hover:text-blue-600 group-hover:translate-x-1 transition-all" />
  </button>
);

const LogLine = ({ time, level, msg, color }) => (
  <div className="flex gap-4 group hover:bg-white p-1 rounded border border-transparent hover:border-slate-200 hover:shadow-sm transition-all duration-200">
    <span className="text-slate-400 shrink-0 select-none py-0.5">{time}</span>
    <span className={`shrink-0 font-bold select-none py-0.5 ${level === 'WARN' ? 'text-amber-600' : 'text-blue-600'}`}>{level}</span>
    <span className={`${color || 'text-slate-700'} break-all font-medium py-0.5`}>{msg}</span>
  </div>
);

export default Dashboard;
