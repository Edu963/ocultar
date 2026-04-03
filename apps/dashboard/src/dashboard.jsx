import React, { useState, useEffect } from 'react';
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
  X
} from 'lucide-react';

const OCULTAR_API = 'http://localhost:8080/api';

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
        requests: Math.floor(data.requests_per_second * 3600), // Simulated total for the hour
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
    if (activeTab === 'config') {
      fetchConfig();
    }
  }, [activeTab]);

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
    setModalData(null); // Clear previous data
    try {
      const res = await fetch(`${OCULTAR_API}${endpoint}`);
      const data = await res.json();
      setModalData(data);
    } catch (e) {
      setModalData({ error: 'Failed to retrieve data.' });
    }
  };

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
        
        <nav className="flex bg-white p-1 rounded-lg border border-slate-200 shadow-sm">
            {['overview', 'config', 'logs'].map(tab => (
                <button 
                    key={tab}
                    onClick={() => setActiveTab(tab)}
                    className={`px-4 py-1.5 rounded text-xs font-bold uppercase tracking-wider transition-all ${activeTab === tab ? 'bg-blue-600 text-white shadow-md' : 'text-slate-500 hover:text-slate-800'}`}
                >
                    {tab}
                </button>
            ))}
        </nav>
      </header>

      {/* Main Grid */}
      <main className="max-w-7xl mx-auto space-y-6">
        
        {/* Metrics Row */}
        <section className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            <MetricCard icon={<Activity className="w-4 h-4 text-blue-500" />} label="Requests / Hour" value={metrics.requests.toLocaleString()} />
            <MetricCard icon={<Shield className="w-4 h-4 text-emerald-500" />} label="PII Detections" value={metrics.detections.toLocaleString()} />
            <MetricCard icon={<Zap className="w-4 h-4 text-amber-500" />} label="Refinery Latency" value={metrics.latency} />
            <MetricCard icon={<Terminal className="w-4 h-4 text-purple-500" />} label="Worker Queue" value={metrics.queue} />
        </section>

        {activeTab === 'overview' && (
          <div className="grid lg:grid-cols-12 gap-6 animate-in fade-in duration-500">
              {/* Left Column: Actions & Controls */}
              <div className="lg:col-span-4 space-y-6">
                  {/* Quick Action: Refinery Test */}
                  <div className="bg-white border border-slate-200 shadow-sm rounded-xl p-5">
                      <h3 className="text-xs font-bold text-slate-500 uppercase tracking-widest mb-4 flex items-center gap-2">
                          <ArrowRight className="w-4 h-4 text-blue-500" /> Live Refinery Test
                      </h3>
                      <textarea 
                          value={testInput}
                          onChange={e => setTestInput(e.target.value)}
                          placeholder="Paste logs or JSON to test redaction..."
                          className="w-full h-32 bg-slate-50 border border-slate-200 rounded-lg p-3 text-xs font-mono mb-3 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 resize-none transition-colors"
                      />
                      <button 
                          onClick={handleTestRefinery}
                          disabled={isRefining || !testInput}
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

                  {/* System Settings Entry Points */}
                  <div className="bg-white border border-slate-200 shadow-sm rounded-xl p-5">
                      <h3 className="text-xs font-bold text-slate-500 uppercase tracking-widest mb-4 flex items-center gap-2">
                          <Settings className="w-4 h-4 text-slate-500" /> System Control
                      </h3>
                      <div className="space-y-2">
                          <OpButton 
                            icon={<Lock className="w-4 h-4" />} 
                            label="Regex Enforcement Rules" 
                            onClick={() => handleOpenModal('Regex Rules', '/config/regex')}
                          />
                          <OpButton 
                            icon={<Database className="w-4 h-4" />} 
                            label="Identity Dictionaries" 
                            onClick={() => handleOpenModal('Identity Dictionaries', '/config/dictionary')}
                          />
                          <OpButton 
                            icon={<Eye className="w-4 h-4" />} 
                            label="Risk Compliance Radar" 
                            onClick={() => handleOpenModal('Risk Radar Data', '/audit/risk')}
                          />
                      </div>
                  </div>
              </div>

              {/* Right Column: Live Audit Logs */}
              <div className="lg:col-span-8 flex flex-col h-full bg-white border border-slate-200 shadow-sm rounded-xl overflow-hidden">
                  <div className="flex justify-between items-center px-5 py-4 border-b border-slate-100 bg-slate-50/50">
                      <h3 className="text-xs font-bold text-slate-600 uppercase tracking-widest flex items-center gap-2">
                          <FileText className="w-4 h-4 text-blue-500" /> Secure Audit Trail
                      </h3>
                      <div className="flex gap-2">
                          <span className="text-[10px] text-emerald-600 bg-emerald-100 px-2 py-0.5 rounded border border-emerald-200 font-mono">ED25519_SIGNED</span>
                      </div>
                  </div>
                  <div className="flex-grow p-4 font-mono text-[10.5px] space-y-1 overflow-y-auto min-h-[400px] bg-slate-50 border-b border-slate-100">
                      {auditLogs.length > 0 ? auditLogs.map((log, i) => (
                        <LogLine 
                          key={i}
                          time={new Date(log.timestamp).toLocaleTimeString()} 
                          level={log.result === 'SUCCESS' ? 'INFO' : 'WARN'} 
                          msg={`${log.action} [${log.compliance_mapping || 'N/A'}] - ${log.notes || 'No details'}`}
                          color={log.result === 'SUCCESS' ? 'text-slate-600' : 'text-amber-600'}
                        />
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

        {activeTab === 'config' && (
           <div className="bg-white border border-slate-200 shadow-sm rounded-xl p-8 animate-in slide-in-from-top-4 duration-300">
              <div className="flex items-center gap-3 mb-6 pb-4 border-b border-slate-100">
                 <Settings className="w-6 h-6 text-slate-700" />
                 <h2 className="text-xl font-bold text-slate-900">Configuration Plane</h2>
              </div>
              <div className="h-[500px] bg-slate-50 border border-slate-200 rounded p-4 overflow-auto font-mono text-xs">
                 {!configData ? (
                   <div className="text-slate-500 animate-pulse">Fetching latest runtime configuration...</div>
                 ) : (
                   <pre className="text-slate-800 whitespace-pre-wrap">
                      {JSON.stringify(configData, null, 2)}
                   </pre>
                 )}
              </div>
           </div>
        )}

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

      {/* Modal for viewing placeholder data */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/40 backdrop-blur-sm animate-in fade-in duration-200">
          <div className="bg-white border border-slate-200 shadow-2xl rounded-xl w-full max-w-3xl max-h-[80vh] flex flex-col overflow-hidden animate-in zoom-in-95 duration-200">
            <div className="flex justify-between items-center px-6 py-4 border-b border-slate-100 bg-slate-50">
              <h2 className="text-sm font-bold text-slate-800 uppercase tracking-widest">{modalTitle}</h2>
              <button 
                onClick={() => setIsModalOpen(false)}
                className="p-1 hover:bg-slate-200 rounded transition-colors text-slate-500"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="p-6 overflow-auto bg-white flex-grow">
              {!modalData ? (
                <div className="text-slate-500 text-sm animate-pulse flex items-center gap-2">
                  <div className="w-4 h-4 block border-2 border-slate-300 border-t-blue-600 rounded-full animate-spin"></div>
                  Retrieving policy data...
                </div>
              ) : (
                <pre className="text-xs font-mono text-slate-700 whitespace-pre-wrap bg-slate-50 p-4 border border-slate-200 rounded">
                  {JSON.stringify(modalData, null, 2)}
                </pre>
              )}
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

const MetricCard = ({ icon, label, value }) => (
  <div className="bg-white border border-slate-200 shadow-sm p-5 rounded-xl hover:shadow-md transition-all group flex flex-col justify-between">
    <div className="flex items-center gap-3 mb-4">
        <div className="w-8 h-8 rounded bg-slate-50 border border-slate-100 flex items-center justify-center opacity-80 group-hover:opacity-100 group-hover:bg-slate-100 transition-all transform duration-300">
          {icon}
        </div>
        <span className="text-[10px] text-slate-500 uppercase font-bold tracking-widest leading-none">{label}</span>
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
