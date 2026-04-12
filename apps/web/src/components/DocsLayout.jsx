import { useState, useEffect } from 'react';
import { NavLink, Outlet, useLocation } from 'react-router-dom';
import { 
    Search, ChevronRight, BookOpen, GraduationCap, 
    Lightbulb, Terminal, ChevronDown, List, ArrowLeft
} from 'lucide-react';
import FlexSearch from 'flexsearch';

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
            { title: "Refinery Tiers", path: "explanation/refinery-tiers" }
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

export default function DocsLayout() {
    const [searchQuery, setSearchQuery] = useState('');
    const [isSidebarOpen, setIsSidebarOpen] = useState(true);
    const location = useLocation();

    // Close sidebar on mobile when navigating
    useEffect(() => {
        if (window.innerWidth < 768) {
            setIsSidebarOpen(false);
        }
    }, [location]);

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
            {/* Sidebar */}
            <aside 
                className={`fixed inset-y-0 left-0 z-40 w-72 bg-zinc-950/80 backdrop-blur-xl border-r border-white/5 transition-transform duration-300 transform ${isSidebarOpen ? 'translate-x-0' : '-translate-x-full'} md:translate-x-0`}
            >
                <div className="flex flex-col h-full pt-20">
                    {/* Search */}
                    <div className="px-6 mb-8">
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

                    {/* Navigation */}
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

                    {/* Version footer */}
                    <div className="p-6 border-t border-white/5">
                        <div className="flex items-center justify-between text-[10px] font-mono text-zinc-600 uppercase tracking-widest">
                            <span>v4.1.0-RC</span>
                            <span className="text-cyan-500/50">Enterprise</span>
                        </div>
                    </div>
                </div>
            </aside>

            {/* Main Content */}
            <main className={`flex-grow transition-all duration-300 ${isSidebarOpen ? 'md:pl-72' : 'pl-0'}`}>
                <div className="max-w-4xl mx-auto px-6 py-12 md:px-12 md:py-20 animate-fade-in">
                    <Outlet />
                </div>
            </main>

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
