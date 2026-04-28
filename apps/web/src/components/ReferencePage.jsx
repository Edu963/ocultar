import { useState } from 'react';
import { useParams } from 'react-router-dom';
import manifest from '../docs/docs-manifest.json';
import { 
    Box, Code2, Database, Shield, ChevronDown, 
    ChevronRight, Info, Zap, Terminal 
} from 'lucide-react';

const PropertyItem = ({ name, type, comment, tag, isFirst }) => {
    const [isOpen, setIsOpen] = useState(false);
    
    // Clean up type names (hex strings to readable names if possible)
    const cleanType = type.includes('0x') ? 'Object' : type;

    return (
        <div className={`py-6 ${!isFirst ? 'border-t border-white/5' : ''}`}>
            <div className="flex flex-col md:flex-row md:items-start justify-between gap-4 group">
                <div className="space-y-1">
                    <div className="flex items-center gap-2">
                        <code className="text-sm font-bold text-cyan-400 bg-cyan-500/5 px-2 py-0.5 rounded border border-cyan-500/10 tracking-tight">
                            {name}
                        </code>
                        <span className="text-[11px] font-mono text-zinc-500">
                            {cleanType}
                        </span>
                    </div>
                    {tag && (
                        <div className="text-[10px] font-mono text-zinc-600 uppercase tracking-widest">
                            tag: {tag}
                        </div>
                    )}
                </div>
                
                <div className="flex-grow max-w-2xl">
                    <p className="text-sm text-slate-400 leading-relaxed">
                        {comment || 'No description provided for this field.'}
                    </p>
                </div>
            </div>
        </div>
    );
};

export default function ReferencePage() {
    const { type } = useParams(); // 'api' or 'config'
    
    const structs = type === 'config' 
        ? manifest.structs.filter(s => s.package === 'config')
        : manifest.structs.filter(s => s.package === 'audit' || s.package === 'license');

    const Icon = type === 'config' ? Database : Terminal;

    return (
        <div className="animate-fade-in-up space-y-16">
            <header className="space-y-4">
                <div className="flex items-center gap-3">
                    <div className="p-2 bg-cyan-500/10 rounded-lg text-cyan-500">
                        <Icon className="w-6 h-6" />
                    </div>
                    <h1 className="text-4xl font-extrabold text-white tracking-tight uppercase">
                        {type === 'config' ? 'Configuration Schema' : 'API Definitions'}
                    </h1>
                </div>
                <p className="text-slate-400 max-w-2xl leading-relaxed">
                    Technical specifications for OCULTAR {type === 'config' ? 'configuration variables' : 'API data structures'}. 
                    All fields below are verifiable via the Go source code definitions.
                </p>
            </header>

            <div className="space-y-24">
                {structs.map((st, i) => (
                    <section key={st.name} className="scroll-mt-24" id={st.name}>
                        <div className="flex items-center justify-between border-b border-white/5 pb-4 mb-8">
                            <div className="space-y-1">
                                <h2 className="text-2xl font-bold text-white tracking-tight">{st.name}</h2>
                                <div className="text-[10px] font-bold text-zinc-600 uppercase tracking-[0.2em]">
                                    Package: {st.package}
                                </div>
                            </div>
                            <div className="flex items-center gap-2 px-3 py-1 bg-zinc-900 border border-white/5 rounded text-[10px] font-bold text-cyan-500 uppercase tracking-widest">
                                <Zap className="w-3 h-3" /> Struct Definition
                            </div>
                        </div>

                        {st.doc && (
                            <div className="mb-8 p-4 bg-zinc-900/50 border-l-2 border-cyan-500/30 text-sm text-slate-400 italic leading-relaxed">
                                {st.doc}
                            </div>
                        )}

                        <div className="space-y-0">
                            {st.fields.map((field, index) => (
                                <PropertyItem 
                                    key={field.name}
                                    name={field.name}
                                    type={field.type}
                                    comment={field.comment}
                                    tag={field.tag}
                                    isFirst={index === 0}
                                />
                            ))}
                        </div>
                    </section>
                ))}
            </div>

            {/* Footer Note */}
            <div className="bg-zinc-950 border border-white/5 p-8 rounded-2xl flex md:items-center gap-6 group">
                <div className="w-12 h-12 bg-cyan-500/10 rounded-xl flex items-center justify-center text-cyan-500 group-hover:scale-110 transition-transform">
                    <Info className="w-6 h-6" />
                </div>
                <div className="space-y-1">
                    <h4 className="text-white font-bold">Automatic Versioning</h4>
                    <p className="text-sm text-slate-500 leading-relaxed">
                        These references are generated at build time. For legacy version documentation, please consult the 
                        <a href="#" className="text-cyan-500 ml-1 hover:underline">OCULTAR Archive</a>.
                    </p>
                </div>
            </div>
        </div>
    );
}
