import { Link } from 'react-router-dom';
import { 
    BookOpen, GraduationCap, Code2, Shield, 
    Zap, Terminal, Database, ArrowRight, ExternalLink
} from 'lucide-react';

const DocCard = ({ title, description, icon, to, color = 'cyan' }) => (
    <Link 
        to={to} 
        className="group relative bg-zinc-950 border border-white/5 p-8 rounded-2xl hover:border-cyan-500/50 transition-all hover:bg-zinc-900/50"
    >
        <div className={`w-12 h-12 bg-${color}-500/10 rounded-xl flex items-center justify-center text-${color}-500 mb-6 group-hover:scale-110 transition-transform`}>
            {icon}
        </div>
        <h3 className="text-lg font-bold text-white mb-2 group-hover:text-cyan-400 transition-colors">{title}</h3>
        <p className="text-sm text-slate-500 leading-relaxed mb-6">{description}</p>
        <div className="flex items-center gap-2 text-[10px] font-bold uppercase tracking-widest text-cyan-500 group-hover:gap-3 transition-all">
            Explore Documentation <ArrowRight className="w-3 h-3" />
        </div>
    </Link>
);

export default function DocsLanding() {
    return (
        <div className="animate-fade-in-up space-y-20 pb-20">
            {/* Hero */}
            <header className="space-y-6">
                <div className="badge mb-4">Documentation Portal</div>
                <h1 className="text-5xl font-extrabold text-white tracking-tight">
                    OCULTAR <span className="text-cyan-500">Docs</span>
                </h1>
                <p className="text-lg text-slate-400 max-w-2xl leading-relaxed">
                    Secure your AI infrastructure with the industry's first Zero-Egress refinery. Learn how to deploy, configure, and scale OCULTAR for your enterprise.
                </p>
                <div className="flex gap-4">
                    <Link to="/docs/tutorials/quickstart" className="btn btn-primary px-8 py-3">
                        Get Started
                    </Link>
                    <a href="https://github.com/Edu963/ocultar" className="btn btn-secondary px-8 py-3 flex items-center gap-2">
                        <Code2 className="w-4 h-4" /> View Source
                    </a>
                </div>
            </header>

            {/* Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <DocCard 
                    title="Quickstart Guide"
                    description="Run OCULTAR locally with Docker Compose and send your first refined prompt in under 5 minutes."
                    icon={<GraduationCap className="w-6 h-6" />}
                    to="/docs/tutorials/quickstart"
                />
                <DocCard 
                    title="API Reference"
                    description="Automated technical documentation for the Refinery endpoints, payload structures, and response schemas."
                    icon={<Code2 className="w-6 h-6" />}
                    to="/docs/reference/api"
                />
                <DocCard 
                    title="Architecture"
                    description="Deep dive into Zero-Egress logic, Sombra Proxy mechanics, and the three-tier refinement lifecycle."
                    icon={<Shield className="w-6 h-6" />}
                    to="/docs/explanation/zero-egress"
                />
                <DocCard 
                    title="Configurations"
                    description="Comprehensive guide to config.yaml, including PII detection rules, Vault backends, and regional packs."
                    icon={<Terminal className="w-6 h-6" />}
                    to="/docs/reference/config"
                />
            </div>

            {/* Community/CTA */}
            <div className="bg-cyan-500/5 border border-cyan-500/10 p-12 rounded-3xl relative overflow-hidden group">
                <div className="absolute top-0 right-0 p-12 opacity-10 group-hover:scale-110 transition-transform">
                    <Zap className="w-32 h-32 text-cyan-500" />
                </div>
                <div className="relative z-10 space-y-4 max-w-xl">
                    <h3 className="text-2xl font-bold text-white">Join the OCULTAR Community</h3>
                    <p className="text-slate-400">
                        Stay up to date with the latest security packs, regional PII updates, and enterprise features.
                    </p>
                    <div className="pt-4 flex gap-8">
                        <a href="#" className="flex items-center gap-2 text-xs font-bold uppercase tracking-widest text-cyan-500 hover:text-cyan-400">
                            Slack Community <ExternalLink className="w-3 h-3" />
                        </a>
                        <a href="#" className="flex items-center gap-2 text-xs font-bold uppercase tracking-widest text-cyan-500 hover:text-cyan-400">
                            Weekly Updates <ExternalLink className="w-3 h-3" />
                        </a>
                    </div>
                </div>
            </div>
        </div>
    );
}
