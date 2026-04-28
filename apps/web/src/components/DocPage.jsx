import { useState, useEffect } from 'react';
import { useParams, useOutletContext } from 'react-router-dom';
import MarkdownRenderer from './MarkdownRenderer';
import { ThumbsUp, ThumbsDown, Clock, User, ArrowRight } from 'lucide-react';

// Vite-specific dynamic importing of markdown files
const modules = import.meta.glob('../docs/content/**/*.md', { as: 'raw', eager: true });

export default function DocPage() {
    const { "*": path } = useParams();
    const { onHeadingsExtracted } = useOutletContext();
    const [content, setContent] = useState('');
    const [error, setError] = useState(false);

    useEffect(() => {
        // Construct the expected file path
        const filePath = `../docs/content/${path}.md`;
        
        if (modules[filePath]) {
            setContent(modules[filePath]);
            setError(false);
            window.scrollTo(0, 0);
        } else {
            setError(true);
        }
    }, [path]);

    if (error) {
        return (
            <div className="flex flex-col items-center justify-center py-20 text-center">
                <div className="text-6xl mb-6">🔍</div>
                <h1 className="text-2xl font-bold text-white mb-2">Page Not Found</h1>
                <p className="text-slate-500">The document you are looking for does not exist or has been moved.</p>
            </div>
        );
    }

    return (
        <div className="animate-fade-in-up">
            <div className="flex items-center gap-6 mb-8 text-[10px] font-bold uppercase tracking-widest text-zinc-600 border-b border-white/5 pb-4">
                <div className="flex items-center gap-2">
                    <Clock className="w-3 h-3" /> 3 min read
                </div>
                <div className="flex items-center gap-2">
                    <User className="w-3 h-3" /> Core Engineering
                </div>
            </div>

            <MarkdownRenderer content={content} onHeadingsExtracted={onHeadingsExtracted} />
            
            <div className="mt-20 pt-10 border-t border-white/5">
                <div className="bg-zinc-950 border border-white/5 rounded-2xl p-8 flex flex-col md:flex-row justify-between items-center gap-6">
                    <div className="space-y-1">
                        <h4 className="text-white font-bold text-sm">Was this page helpful?</h4>
                        <p className="text-xs text-zinc-600">Your feedback helps us build a rock-solid product.</p>
                    </div>
                    <div className="flex gap-3">
                        <button className="flex items-center gap-2 px-4 py-2 bg-white/5 hover:bg-white/10 border border-white/10 rounded-lg text-xs font-bold transition-all">
                            <ThumbsUp className="w-3.5 h-3.5" /> Yes
                        </button>
                        <button className="flex items-center gap-2 px-4 py-2 bg-white/5 hover:bg-white/10 border border-white/10 rounded-lg text-xs font-bold transition-all">
                            <ThumbsDown className="w-3.5 h-3.5" /> No
                        </button>
                    </div>
                </div>

                <div className="mt-12 flex justify-between items-center text-[10px] font-bold uppercase tracking-widest text-zinc-600">
                    <div>Last updated: April 2026</div>
                    <div className="flex gap-6">
                        <a href="https://github.com/Edu963/ocultar" className="hover:text-cyan-500 transition-colors">Edit on GitHub</a>
                        <a href="#" className="hover:text-cyan-500 transition-colors">Submit Issue</a>
                    </div>
                </div>
            </div>
        </div>
    );
}
