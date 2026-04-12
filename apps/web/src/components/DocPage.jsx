import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import MarkdownRenderer from './MarkdownRenderer';

// Vite-specific dynamic importing of markdown files
const modules = import.meta.glob('../docs/content/**/*.md', { as: 'raw', eager: true });

export default function DocPage() {
    const { "*": path } = useParams();
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
            <MarkdownRenderer content={content} />
            
            <div className="mt-20 pt-10 border-t border-white/5 flex justify-between items-center text-xs text-zinc-600">
                <div>Last updated: April 2026</div>
                <div className="flex gap-4">
                    <a href="https://github.com/Edu963/ocultar" className="hover:text-cyan-500 transition-colors">Edit this page</a>
                    <a href="#" className="hover:text-cyan-500 transition-colors">Next Page &rarr;</a>
                </div>
            </div>
        </div>
    );
}
