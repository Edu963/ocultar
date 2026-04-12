import { useState, useEffect } from 'react';
import { NavLink, Outlet, useLocation, Link, useNavigate } from 'react-router-dom';
import { 
    Search, ChevronRight, BookOpen, GraduationCap, 
    Lightbulb, Terminal, ChevronDown, List, ArrowLeft,
    Home, Shield
} from 'lucide-react';

const sections = [
    {
        title: "Tutorials",
        icon: <GraduationCap className="w-4 h-4" />,
        items: [
            { title: "Quickstart", path: "tutorials/quickstart" },
            { title: "First Pipeline", path: "tutorials/first-pipeline" }
        ]
    },
    {
        title: "How-to Guides",
        icon: <BookOpen className="w-4 h-4" />,
        items: [
            { title: "Slack Integration", path: "guides/slack" },
            { title: "PII Rules", path: "guides/pii-rules" },
            { title: "Sombra Proxy", path: "guides/sombra" }
        ]
    },
    {
        title: "Explanation",
        icon: <Lightbulb className="w-4 h-4" />,
        items: [
            { title: "Zero-Egress", path: "explanation/zero-egress" },
            { title: "Refinery Tiers", path: "explanation/refinery-tiers" },
            { title: "Vault Architecture", path: "explanation/vault" }
        ]
    },
    {
        title: "Reference",
        icon: <Terminal className="w-4 h-4" />,
        items: [
            { title: "API Reference", path: "reference/api" },
            { title: "Config Schema", path: "reference/config" }
        ]
    }
];

const Breadcrumbs = () => {
    const location = useLocation();
    const pathnames = location.pathname.split('/').filter((x) => x);

    if (pathnames.length === 1 && pathnames[0] === 'docs') return null;

    return (
        <nav className="flex items-center gap-2 text-[10px] font-bold uppercase tracking-widest text-zinc-500 mb-8 px-1">
            <Link to="/docs" className="hover:text-cyan-500 flex items-center gap-1">
                <Home className="w-3 h-3" /> DOCS
            </Link>
            {pathnames.slice(1).map((name, index) => {
                const routeTo = `/${pathnames.slice(0, index + 2).join('/')}`;
                const isLast = index === pathnames.slice(1).length - 1;
                return (
                    <div key={name} className="flex items-center gap-2">
                        <ChevronRight className="w-3 h-3 text-zinc-800" />
                        {isLast ? (
                            <span className="text-cyan-500/80">{name.replace(/-/g, ' ')}</span>
                        ) : (
                            <Link to={routeTo} className="hover:text-cyan-500">
                                {name.replace(/-/g, ' ')}
                            </Link>
                        )}
                    </div>
                );
            })}
        </nav>
    );
};

export default function DocsLayout() {
    const [searchQuery, setSearchQuery] = useState('');
    const [isSidebarOpen, setIsSidebarOpen] = useState(true);
    const [headings, setHeadings] = useState([]);
    const location = useLocation();
    const navigate = useNavigate();

    // Close sidebar on mobile when navigating
    useEffect(() => {
        if (window.innerWidth < 768) {
            setIsSidebarOpen(false);
        }
    }, [location.pathname]);

    // Handle headings from child pages
    const handleHeadingsExtracted = (newHeadings) => {
        setHeadings(newHeadings);
    };

    // Flatten items for pagination
    const allItems = sections.flatMap(s => s.items);
    const currentIndex = allItems.findIndex(item => location.pathname.includes(item.path));
    const prevItem = currentIndex > 0 ? allItems[currentIndex - 1] : null;
    const nextItem = currentIndex < allItems.length - 1 ? allItems[currentIndex + 1] : null;

    // Filtered sections based on search query
    const filteredSections = sections.map(section => ({
        ...section,
        items: section.items.filter(item => 
            item.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
            section.title.toLowerCase().includes(searchQuery.toLowerCase())
        )
    })).filter(section => section.items.length > 0);

    return (
        <div className="flex min-h-screen bg-zinc-950 text-slate-300 font-sans selection:bg-cyan-500/30">
            {/* Sidebar (Left) */}
            <aside 
                className={`fixed inset-y-0 left-0 z-40 w-72 bg-[#0c0c0e]/80 backdrop-blur-xl border-r border-white/5 transition-transform duration-300 transform ${isSidebarOpen ? 'translate-x-0' : '-translate-x-full'} md:translate-x-0 h-screen top-0 pt-16`}
            >
                <div className="flex flex-col h-full">
                    <div className="px-6 mb-8 mt-8">
                        <div className="relative group">
                            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500 group-focus-within:text-cyan-500 transition-colors" />
                            <input 
                                type="text" 
                                placeholder="Search documentation..." 
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                                className="w-full bg-zinc-900/50 border border-white/5 rounded-lg py-2 pl-10 pr-4 text-xs focus:outline-none focus:border-cyan-500/50 focus:ring-1 focus:ring-cyan-500/20 transition-all"
                            />
                        </div>
                    </div>

                    <nav className="flex-grow overflow-y-auto px-4 space-y-8 custom-scrollbar">
                        {filteredSections.map(section => (
                            <div key={section.title} className="space-y-2">
                                <h4 className="flex items-center gap-2 px-2 text-[10px] font-bold text-zinc-500 uppercase tracking-[0.2em] mb-3">
                                    {section.icon}
                                    {section.title}
                                </h4>
                                <ul className="space-y-1">
                                    {section.items.map(item => (
                                        <li key={item.path}>
                                            <NavLink 
                                                to={`/docs/${item.path}`}
                                                className={({ isActive }) => `
                                                    flex items-center justify-between px-3 py-2 rounded-lg text-sm transition-all group
                                                    ${isActive 
                                                        ? 'bg-cyan-500/10 text-cyan-400 font-medium' 
                                                        : 'hover:bg-white/5 text-slate-400 hover:text-white'}
                                                `}
                                            >
                                                {item.title}
                                                <ChevronRight className={`w-3 h-3 opacity-0 group-hover:opacity-100 transition-opacity`} />
                                            </NavLink>
                                        </li>
                                    ))}
                                </ul>
                            </div>
                        ))}
                    </nav>

                    <div className="p-6 border-t border-white/5">
                        <div className="flex items-center justify-between text-[10px] font-mono text-zinc-600 uppercase tracking-widest">
                            <span>v4.1.0-RC</span>
                            <span className="text-cyan-500/50">Enterprise</span>
                        </div>
                    </div>
                </div>
            </aside>

            {/* Main Content (Center) */}
            <main className={`flex-grow transition-all duration-300 ${isSidebarOpen ? 'md:pl-72' : 'pl-0'} md:pr-72 pt-16`}>
                <div className="max-w-4xl mx-auto px-6 py-12 md:px-12 md:py-16 animate-fade-in min-h-[calc(100vh-64px)]">
                    <Breadcrumbs />
                    <Outlet context={{ onHeadingsExtracted: handleHeadingsExtracted }} />
                    
                    {currentIndex !== -1 && (
                        <div className="mt-20 pt-10 border-t border-white/5 flex justify-between items-center gap-4">
                            {prevItem ? (
                                <Link to={`/docs/${prevItem.path}`} className="group flex flex-col items-start gap-2">
                                    <span className="text-[10px] font-bold text-zinc-600 uppercase tracking-widest flex items-center gap-2">
                                        <ArrowLeft className="w-3 h-3" /> Previous
                                    </span>
                                    <span className="text-sm font-bold text-slate-400 group-hover:text-white transition-colors">
                                        {prevItem.title}
                                    </span>
                                </Link>
                            ) : <div></div>}
                            
                            {nextItem ? (
                                <Link to={`/docs/${nextItem.path}`} className="group flex flex-col items-end gap-2 text-right">
                                    <span className="text-[10px] font-bold text-zinc-600 uppercase tracking-widest flex items-center gap-2">
                                        Next <ChevronRight className="w-3 h-3" />
                                    </span>
                                    <span className="text-sm font-bold text-cyan-500 group-hover:text-cyan-400 transition-colors">
                                        {nextItem.title}
                                    </span>
                                </Link>
                            ) : <div></div>}
                        </div>
                    )}
                </div>
            </main>

            {/* TOC (Right Sidebar) */}
            <aside className="fixed inset-y-0 right-0 z-30 w-72 h-screen pt-32 px-8 overflow-y-auto hidden lg:block border-l border-white/5">
                <div className="space-y-6">
                    <h4 className="flex items-center gap-2 text-[10px] font-bold text-zinc-500 uppercase tracking-widest mb-4">
                        <List className="w-3 h-3" /> On this page
                    </h4>
                    <ul className="space-y-3">
                        {headings.map((h, i) => (
                            <li key={i} style={{ paddingLeft: `${(h.level - 2) * 1}rem` }}>
                                <a 
                                    href={`#${h.id}`}
                                    className="text-[11px] text-slate-500 hover:text-cyan-500 transition-colors block leading-snug"
                                >
                                    {h.text}
                                </a>
                            </li>
                        ))}
                    </ul>
                </div>
            </aside>

            {/* Mobile Toggle */}
            <button 
                onClick={() => setIsSidebarOpen(!isSidebarOpen)}
                className="fixed bottom-6 right-6 z-50 p-4 bg-cyan-500 text-black rounded-full shadow-lg md:hidden hover:scale-110 transition-transform active:scale-95"
            >
                {isSidebarOpen ? <ArrowLeft className="w-6 h-6" /> : <List className="w-6 h-6" />}
            </button>
        </div>
    );
}

const ShieldIcon = ({ className }) => (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={className}>
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
    </svg>
);
