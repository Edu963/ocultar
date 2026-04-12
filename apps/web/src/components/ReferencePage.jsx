import { useState } from 'react';
import { useParams } from 'react-router-dom';
import manifest from '../docs/docs-manifest.json';
import { Box, Code2, Database, Shield } from 'lucide-react';

export default function ReferencePage() {
    const { type } = useParams(); // 'api' or 'config'
    
    // Filter structs based on package/type
    const structs = type === 'config' 
        ? manifest.structs.filter(s => s.package === 'config')
        : manifest.structs.filter(s => s.package === 'audit' || s.package === 'license');

    return (
        <div className="animate-fade-in-up space-y-16">
            <header className="space-y-4">
                <div className="flex items-center gap-3">
                    <div className="p-2 bg-cyan-500/10 rounded-lg text-cyan-500">
                        {type === 'config' ? <Database className="w-6 h-6" /> : <Code2 className="w-6 h-6" />}
                    </div>
                    <h1 className="text-3xl font-bold text-white tracking-tight uppercase">
                        {type === 'config' ? 'Configuration Schema' : 'API Reference'}
                    </h1>
                </div>
                <p className="text-slate-500 max-w-2xl leading-relaxed">
                    Automatically generated from OCULTAR source code. These definitions are the definitive source of truth for the latest build (v4.1.0-RC).
                </p>
            </header>

            <div className="space-y-20">
                {structs.map(st => (
                    <section key={st.name} className="scroll-mt-24" id={st.name}>
                        <div className="flex items-center gap-4 mb-6">
                            <h2 className="text-xl font-bold text-white font-mono">{st.name}</h2>
                            <span className="px-2 py-1 bg-zinc-900 border border-white/5 rounded text-[10px] font-mono text-zinc-500 uppercase">
                                pkg: {st.package}
                            </span>
                        </div>
                        {st.doc && <p className="text-slate-400 text-sm mb-6 leading-relaxed italic">{st.doc}</p>}
                        
                        <div className="bg-zinc-950 border border-white/5 rounded-xl overflow-hidden shadow-2xl">
                            <table className="w-full text-left border-collapse">
                                <thead>
                                    <tr className="bg-white/5 border-b border-white/5">
                                        <th className="px-6 py-4 text-[10px] font-bold text-zinc-500 uppercase tracking-widest">Field</th>
                                        <th className="px-6 py-4 text-[10px] font-bold text-zinc-500 uppercase tracking-widest">Type</th>
                                        <th className="px-6 py-4 text-[10px] font-bold text-zinc-500 uppercase tracking-widest">Description</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-white/5 font-mono text-xs">
                                    {st.fields.map(field => (
                                        <tr key={field.name} className="hover:bg-white/[0.02] transition-colors">
                                            <td className="px-6 py-4">
                                                <span className="text-cyan-400 font-bold">{field.name}</span>
                                                {field.tag && (
                                                    <div className="text-[10px] text-zinc-600 mt-1">{field.tag}</div>
                                                )}
                                            </td>
                                            <td className="px-6 py-4">
                                                <span className="text-purple-400">{field.type}</span>
                                            </td>
                                            <td className="px-6 py-4 text-slate-500 font-sans leading-relaxed">
                                                {field.comment || '-'}
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </section>
                ))}
            </div>

            <div className="mt-20 p-8 bg-zinc-950 border border-cyan-500/20 rounded-2xl flex items-center gap-6">
                <Box className="w-10 h-10 text-cyan-500 shrink-0" />
                <div>
                    <h4 className="text-white font-bold mb-1">Looking for endpoint documentation?</h4>
                    <p className="text-xs text-slate-500 leading-relaxed">
                        API request/response examples are available in individual guides for 
                        <a href="/docs/guides/sombra" className="text-cyan-500 ml-1 hover:underline">Proxy Mode</a> and 
                        <a href="/docs/guides/api-ref" className="text-cyan-500 ml-1 hover:underline">Batch Operations</a>.
                    </p>
                </div>
            </div>
        </div>
    );
}
