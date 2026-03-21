import React, { useState, useEffect } from 'react';

const ROIDashboardCard = () => {
    const [data, setData] = useState({ savings: 0, vault_entries: 0, status: 'offline', name: 'Standard_Mode' });
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchROI = async () => {
            try {
                // Use absolute URL to ensure connection to the Mock-API on :8080
                const response = await fetch('http://localhost:8080/api/roi');
                if (response.ok) {
                    const result = await response.json();
                    setData({
                        savings: result.savings || 0,
                        vault_entries: result.vault_entries || 0,
                        status: 'online',
                        name: result.name || 'Standard_Mode'
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
        <div className="border border-black bg-white p-6 space-y-4 relative overflow-hidden group">
            {/* Status Indicator */}
            <div className="absolute top-4 right-4 flex items-center gap-2">
                <div className={`w-1.5 h-1.5 rounded-full ${data.status === 'online' ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`}></div>
                <span className="font-tech text-[8px] uppercase tracking-widest text-dim">
                    {data.status === 'online' ? 'Live_Sync' : 'API_Offline'}
                </span>
            </div>

            <div className="space-y-1">
                <h3 className="font-tech text-[10px] uppercase text-dim tracking-widest italic">{loading ? "Capital_Retention_Metric" : data.name}</h3>
                <div className="text-4xl font-bold tracking-tighter tabular-nums">
                    {loading ? "---" : formatCurrency(data.savings)}
                </div>
            </div>

            <div className="grid grid-cols-2 gap-4 border-t border-black/10 pt-4">
                <div className="space-y-1">
                    <span className="font-tech text-[9px] uppercase text-dim block">Vault_Interactions</span>
                    <span className="font-bold text-sm tabular-nums">
                        {loading ? "..." : data.vault_entries.toLocaleString()}
                    </span>
                </div>
                <div className="space-y-1 text-right">
                    <span className="font-tech text-[9px] uppercase text-dim block">Efficiency_Rate</span>
                    <span className="font-bold text-sm text-green-600">99.98%</span>
                </div>
            </div>

            <div className="bg-black/5 p-3 font-tech text-[9px] text-dim leading-relaxed border-l-2 border-black">
                [SYS_MSG]: Native SLM refinement and deterministic vaulting have neutralized {data.vault_entries.toLocaleString()} external egress vectors this cycle.
            </div>
        </div>
    );
};

export default ROIDashboardCard;
