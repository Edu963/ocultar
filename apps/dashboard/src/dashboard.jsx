import React, { useState, useEffect, useMemo } from 'react';
import { 
  Activity, 
  Shield, 
  Zap, 
  Terminal, 
  Settings, 
  Database, 
  AlertCircle,
  CheckCircle2,
  Lock,
  ArrowRight,
  RefreshCw,
  Eye,
  FileText
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
      
      // Accumulate detections from pii_hits_per_type
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
        // The API might return either a JSON object (if input was JSON) or a string
        setTestOutput(data.refined || (typeof data === 'string' ? data : JSON.stringify(data, null, 2)));
    } catch (e) {
        setTestOutput("ERROR: Refinery unreachable.");
    } finally {
        setIsRefining(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#0a0a0c] text-slate-300 font-sans p-6 selection:bg-blue-500/30">
      {/* Top Summary Bar */}
      <header className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-8 pb-6 border-b border-white/5">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 bg-blue-600 rounded flex items-center justify-center font-bold text-white text-lg shadow-lg shadow-blue-600/20">O</div>
          <div>
            <h1 className="text-xl font-bold text-white tracking-tight">OCULTAR <span className="font-light text-slate-500 tracking-widest">CONTROL_CENTER</span></h1>
            <div className="flex gap-4 mt-1">
                <span className="text-[10px] font-mono text-slate-500 uppercase flex items-center gap-1.5">
                    <div className={`w-1.5 h-1.5 rounded-full ${systemStatus.online ? 'bg-emerald-500 animate-pulse' : 'bg-red-500'}`} />
                    {systemStatus.online ? 'System Online' : 'System Offline'}
                </span>
                <span className="text-[10px] font-mono text-slate-500 uppercase border-l border-white/10 pl-4">{systemStatus.mode} Edition</span>
                <span className="text-[10px] font-mono text-slate-500 uppercase border-l border-white/10 pl-4">Uptime: {systemStatus.uptime}</span>
            </div>
          </div>
        </div>
        
        <nav className="flex bg-white/5 p-1 rounded-lg border border-white/5">
            {['overview', 'config', 'logs'].map(tab => (
                <button 
                    key={tab}
                    onClick={() => setActiveTab(tab)}
                    className={`px-4 py-1.5 rounded text-xs font-bold uppercase tracking-wider transition-all ${activeTab === tab ? 'bg-blue-600 text-white shadow-lg' : 'text-slate-500 hover:text-slate-200'}`}
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
            <MetricCard icon={<Activity className="w-4 h-4 text-blue-400" />} label="Requests / Hour" value={metrics.requests.toLocaleString()} />
            <MetricCard icon={<Shield className="w-4 h-4 text-emerald-400" />} label="PII Detections" value={metrics.detections.toLocaleString()} />
            <MetricCard icon={<Zap className="w-4 h-4 text-amber-400" />} label="Refinery Latency" value={metrics.latency} />
            <MetricCard icon={<Terminal className="w-4 h-4 text-purple-400" />} label="Worker Queue" value={metrics.queue} />
        </section>

        {activeTab === 'overview' && (
          <div className="grid lg:grid-cols-12 gap-6 animate-in fade-in duration-500">
              {/* Left Column: Actions & Controls */}
              <div className="lg:col-span-4 space-y-6">
                  {/* Quick Action: Refinery Test */}
                  <div className="bg-white/5 border border-white/5 rounded-xl p-5">
                      <h3 className="text-xs font-bold text-slate-400 uppercase tracking-widest mb-4 flex items-center gap-2">
                          <ArrowRight className="w-4 h-4 text-blue-400" /> Live Refinery Test
                      </h3>
                      <textarea 
                          value={testInput}
                          onChange={e => setTestInput(e.target.value)}
                          placeholder="Paste logs or JSON to test redaction..."
                          className="w-full h-32 bg-black/40 border border-white/5 rounded-lg p-3 text-xs font-mono mb-3 focus:outline-none focus:border-blue-500/50 resize-none transition-colors"
                      />
                      <button 
                          onClick={handleTestRefinery}
                          disabled={isRefining || !testInput}
                          className={`w-full py-3 rounded font-bold text-xs uppercase tracking-widest transition-all ${isRefining ? 'bg-slate-800 text-slate-500' : 'bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-600/20 active:scale-[0.98]'}`}
                      >
                          {isRefining ? 'PROCESSING...' : 'EXECUTE REDACTION'}
                      </button>
                      {testOutput && (
                          <div className="mt-4 p-3 bg-black/60 border border-emerald-500/20 rounded font-mono text-[10px] text-emerald-400 break-all overflow-auto max-h-40">
                              {testOutput}
                          </div>
                      )}
                  </div>

                  {/* System Settings Entry Points */}
                  <div className="bg-white/5 border border-white/5 rounded-xl p-5">
                      <h3 className="text-xs font-bold text-slate-400 uppercase tracking-widest mb-4 flex items-center gap-2">
                          <Settings className="w-4 h-4 text-slate-500" /> System Control
                      </h3>
                      <div className="space-y-2">
                          <OpButton icon={<Lock className="w-3 h-3" />} label="Regex Enforcement Rules" />
                          <OpButton icon={<Database className="w-3 h-3" />} label="Identity Dictionaries" />
                          <OpButton icon={<Eye className="w-3 h-3" />} label="Risk Compliance Radar" />
                      </div>
                  </div>
              </div>

              {/* Right Column: Live Audit Logs */}
              <div className="lg:col-span-8 flex flex-col h-full bg-white/5 border border-white/5 rounded-xl overflow-hidden shadow-2xl">
                  <div className="flex justify-between items-center px-5 py-4 border-b border-white/5 bg-white/[0.02]">
                      <h3 className="text-xs font-bold text-slate-400 uppercase tracking-widest flex items-center gap-2">
                          <FileText className="w-4 h-4 text-blue-400" /> Secure Audit Trail
                      </h3>
                      <div className="flex gap-2">
                          <span className="text-[10px] text-emerald-500 bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20 font-mono">ED25519_SIGNED</span>
                      </div>
                  </div>
                  <div className="flex-grow p-4 font-mono text-[10px] space-y-1 overflow-y-auto min-h-[400px] bg-black/20">
                      {auditLogs.length > 0 ? auditLogs.map((log, i) => (
                        <LogLine 
                          key={i}
                          time={new Date(log.timestamp).toLocaleTimeString()} 
                          level={log.result === 'SUCCESS' ? 'INFO' : 'WARN'} 
                          msg={`${log.action} [${log.compliance_mapping || 'N/A'}] - ${log.notes || 'No details'}`}
                          color={log.result === 'SUCCESS' ? 'text-slate-400' : 'text-amber-500'}
                        />
                      )) : (
                        <div className="text-slate-600 italic">No audit records found in audit.log...</div>
                      )}
                      <div className="flex items-center gap-2 text-slate-600 animate-pulse mt-2 pt-2 border-t border-white/5">
                          <span>_</span>
                      </div>
                  </div>
                  <div className="px-5 py-3 border-t border-white/5 bg-black/40 flex justify-between items-center text-[9px] text-slate-600">
                    <span className="uppercase tracking-widest">Fail-Closed Enforcement: Enabled</span>
                    <button className="text-blue-500 hover:underline uppercase font-bold">View raw audit.log</button>
                  </div>
              </div>
          </div>
        )}

        {activeTab === 'config' && (
           <div className="bg-white/5 border border-white/5 rounded-xl p-8 text-center animate-in slide-in-from-top-4 duration-300">
              <Lock className="w-12 h-12 text-slate-700 mx-auto mb-4" />
              <h2 className="text-lg font-bold text-white mb-2">Configuration Plane</h2>
              <p className="text-xs text-slate-500 max-w-sm mx-auto">Advanced configuration is restricted to authenticated operators. Use the CLI or update <code>config.yaml</code> directly for live reloads.</p>
           </div>
        )}

        {activeTab === 'logs' && (
           <div className="bg-black/40 border border-white/5 rounded-xl p-4 font-mono text-[11px] overflow-auto h-[600px] animate-in slide-in-from-top-4 duration-300">
              <div className="text-blue-500 mb-4 border-b border-white/10 pb-2 flex justify-between">
                <span>SYSTEM_JOURNAL_V1</span>
                <span>TAIL -F /var/log/ocultar.log</span>
              </div>
              {auditLogs.map((log, i) => (
                <div key={i} className="mb-1">
                  <span className="text-slate-600">[{new Date(log.timestamp).toISOString()}]</span>{" "}
                  <span className="text-blue-400">{log.action}</span>{" "}
                  <span className="text-slate-300">{JSON.stringify(log)}</span>
                </div>
              ))}
           </div>
        )}
      </main>

      <footer className="mt-12 pt-8 border-t border-white/5 flex justify-between items-center px-4">
        <p className="text-[9px] text-slate-600 font-mono uppercase tracking-[0.3em]">
            OCULTAR Control Center // Local Operator View
        </p>
        <p className="text-[9px] text-slate-700 font-mono">
            BUILD_SHA: 82fb12a
        </p>
      </footer>
    </div>
  );
};

const MetricCard = ({ icon, label, value }) => (
  <div className="bg-white/5 border border-white/5 p-5 rounded-xl hover:bg-white/[0.08] transition-all group cursor-default">
    <div className="flex items-center gap-3 mb-3">
        <div className="w-8 h-8 rounded bg-black/20 flex items-center justify-center opacity-80 group-hover:opacity-100 transition-opacity transform group-hover:scale-110 duration-300">{icon}</div>
        <span className="text-[10px] text-slate-500 uppercase font-bold tracking-widest leading-none">{label}</span>
    </div>
    <div className="text-2xl font-bold text-white tabular-nums tracking-tight">{value}</div>
  </div>
);

const OpButton = ({ icon, label }) => (
    <button className="w-full flex items-center justify-between p-4 rounded-lg bg-black/20 border border-white/5 hover:border-blue-500/30 hover:bg-blue-600/5 group transition-all text-left">
        <span className="flex items-center gap-3 text-xs font-semibold text-slate-400 group-hover:text-slate-100">
            {icon}
            {label}
        </span>
        <ArrowRight className="w-3 h-3 text-slate-700 group-hover:text-blue-500 group-hover:translate-x-1 transition-all" />
    </button>
);

const LogLine = ({ time, level, msg, color }) => (
    <div className="flex gap-4 group hover:bg-white/[0.02] py-0.5 px-1 rounded transition-colors">
        <span className="text-slate-600 shrink-0 select-none">{time}</span>
        <span className={`shrink-0 font-bold select-none ${level === 'WARN' ? 'text-amber-500' : 'text-blue-500'}`}>{level}</span>
        <span className={`${color || 'text-slate-400'} break-all`}>{msg}</span>
    </div>
);

export default Dashboard;
