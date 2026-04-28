import { Link } from 'react-router-dom';
import { Search, Github, ExternalLink, Shield, Bell } from 'lucide-react';

export default function DocsHeader() {
    return (
        <header className="fixed top-0 left-0 right-0 z-50 h-16 bg-zinc-950/90 backdrop-blur-xl border-b border-white/5 flex items-center justify-between px-6">
            <div className="flex items-center gap-8">
                <Link to="/" className="flex items-center gap-2 group">
                    <img src="/logo3.jpg" alt="OCULTAR" className="h-6 object-contain" />
                    <span className="text-[10px] font-bold text-zinc-500 uppercase tracking-widest border-l border-white/10 pl-3">Refinery Docs</span>
                </Link>
                
                {/* Desktop Search Placeholder */}
                <div className="hidden md:flex items-center gap-3 bg-white/5 border border-white/10 rounded-full px-4 py-1.5 w-64 group hover:border-cyan-500/30 transition-all cursor-pointer">
                    <Search className="w-3.5 h-3.5 text-zinc-500" />
                    <span className="text-[11px] text-zinc-500 flex-grow">Search...</span>
                    <kbd className="text-[9px] font-mono text-zinc-700 bg-white/5 px-1.5 py-0.5 rounded border border-white/5">⌘K</kbd>
                </div>
            </div>

            <div className="flex items-center gap-6">
                <div className="hidden lg:flex items-center gap-6 text-[10px] font-bold uppercase tracking-widest text-zinc-500">
                    <a href="#" className="hover:text-white transition-colors">Changelog</a>
                    <a href="https://github.com/Edu963/ocultar" className="flex items-center gap-1 hover:text-white transition-colors">Github <ExternalLink className="w-2.5 h-2.5" /></a>
                </div>
                
                <div className="h-4 w-px bg-white/10 hidden lg:block"></div>
                
                <div className="flex items-center gap-3">
                    <button className="p-2 text-zinc-500 hover:text-white transition-colors">
                        <Bell className="w-4 h-4" />
                    </button>
                    <Link to="/risk-assessment" className="bg-cyan-500 text-black text-[10px] font-bold px-4 py-2 rounded uppercase tracking-wider hover:bg-cyan-400 transition-all">
                        Get Started
                    </Link>
                </div>
            </div>
        </header>
    );
}
