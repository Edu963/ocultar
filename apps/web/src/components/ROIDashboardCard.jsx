import React, { useState, useEffect } from 'react';
import { Shield, Zap, TrendingUp, Activity } from 'lucide-react';

const ROIDashboardCard = () => {
    const [data, setData] = useState({ savings: 1435500, vault_entries: 42080, status: 'offline', name: 'Standard Refinery' });
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchROI = async () => {
            try {
                // Use absolute URL to ensure connection to the Mock-API on :8080
                const response = await fetch('http://localhost:8080/api/roi');
                if (response.ok) {
                    const result = await response.json();
                    setData({
                        savings: result.savings || 1435500,
                        vault_entries: result.vault_entries || 42080,
                        status: 'online',
                        name: result.name || 'Standard Refinery'
                    });
                }
            } catch (error) {
                console.error("[SYS] ROI_SYNC_FAILURE:", error);
                setData(prev => ({ ...prev, status: 'error' }));
            } finally {
                setLoading(false);
            }
        };

        fetchROI();
        const interval = setInterval(fetchROI, 30000); // 30s heartbeat
        return () => clearInterval(interval);
    }, []);

    const formatCurrency = (val) =>
        new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: 'USD',
            maximumFractionDigits: 0
        }).format(val);

    return (
        <div className="card space-y-6 relative overflow-hidden group border-white/10 bg-black/60 backdrop-blur-3xl shadow-2xl">
            {/* Status Indicator */}
            <div className="absolute top-6 right-6 flex items-center gap-2">
                <div className={`w-2 h-2 rounded-full ${data.status === 'online' ? 'bg-cyan-500 animate-pulse' : 'bg-yellow-500'}`}></div>
                <span className="text-[10px] font-bold uppercase tracking-widest text-slate-500">
                    {data.status === 'online' ? 'Live Telemetry' : 'Offline Mode'}
                </span>
            </div>

            <div className="space-y-2">
                <h3 className="text-[10px] font-bold uppercase text-cyan-500 tracking-[0.3em] flex items-center gap-2">
                    <Activity className="w-3 h-3" /> {loading ? "Syncing..." : data.name}
                </h3>
                <div className="text-5xl font-bold tracking-tighter tabular-nums text-white group-hover:text-cyan-400 transition-colors">
                    {formatCurrency(data.savings)}
                </div>
                <p className="text-sm font-medium text-slate-400">Capital retention via zero-egress architecture.</p>
            </div>

            <div className="grid grid-cols-2 gap-4 pt-6 border-t border-white/5">
                <div className="space-y-1">
                    <span className="text-[10px] font-bold uppercase text-slate-500 block tracking-widest">Vault Interactions</span>
                    <div className="flex items-center gap-2 text-white">
                        <Shield className="w-4 h-4 text-cyan-500" />
                        <span className="font-bold text-lg tabular-nums">
                            {data.vault_entries.toLocaleString()}
                        </span>
                    </div>
                </div>
                <div className="space-y-1">
                    <span className="text-[10px] font-bold uppercase text-slate-500 block tracking-widest">Efficiency</span>
                    <div className="flex items-center gap-2 text-cyan-400">
                        <TrendingUp className="w-4 h-4" />
                        <span className="font-bold text-lg">99.9%</span>
                    </div>
                </div>
            </div>

            <div className="bg-cyan-500/5 p-4 rounded-xl text-xs font-medium text-slate-400 leading-relaxed border border-cyan-500/10 flex gap-3">
                 <Zap className="w-4 h-4 text-cyan-500 shrink-0 mt-0.5" />
                 <div>
                    Native SLM refinement and deterministic vaulting have neutralized {data.vault_entries.toLocaleString()} egress vectors this cycle.
                 </div>
            </div>
        </div>
    );
};

export default ROIDashboardCard;
