import { useState, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { Copy, Check, Info, AlertTriangle, ShieldCheck } from 'lucide-react';
import slugify from 'slugify';

const Callout = ({ type, children }) => {
    const styles = {
        NOTE: { bg: 'bg-blue-500/10', border: 'border-blue-500/20', icon: <Info className="w-4 h-4 text-blue-400" />, text: 'text-blue-400' },
        WARNING: { bg: 'bg-amber-500/10', border: 'border-amber-500/20', icon: <AlertTriangle className="w-4 h-4 text-amber-400" />, text: 'text-amber-400' },
        SECURITY: { bg: 'bg-cyan-500/10', border: 'border-cyan-500/20', icon: <ShieldCheck className="w-4 h-4 text-cyan-400" />, text: 'text-cyan-400' },
        ENTERPRISE: { bg: 'bg-purple-500/10', border: 'border-purple-500/20', icon: <ShieldCheck className="w-4 h-4 text-purple-400" />, text: 'text-purple-400' }
    };

    const s = styles[type] || styles.NOTE;

    return (
        <div className={`my-6 p-4 rounded-xl border ${s.bg} ${s.border} flex gap-4 items-start`}>
            <div className="mt-1">{s.icon}</div>
            <div className="flex-grow">
                <div className={`text-[10px] font-bold uppercase tracking-widest mb-1 ${s.text}`}>{type}</div>
                <div className="text-sm text-slate-300 leading-relaxed">{children}</div>
            </div>
        </div>
    );
};

const CodeBlock = ({ language, value }) => {
    const [copied, setCopied] = useState(false);

    const onCopy = () => {
        navigator.clipboard.writeText(value);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <div className="relative group my-8">
            <div className="absolute right-4 top-4 z-10 opacity-0 group-hover:opacity-100 transition-opacity">
                <button 
                    onClick={onCopy}
                    className="p-2 bg-white/5 hover:bg-white/10 border border-white/10 rounded-lg text-slate-400 hover:text-white transition-all"
                >
                    {copied ? <Check className="w-4 h-4 text-cyan-500" /> : <Copy className="w-4 h-4" />}
                </button>
            </div>
            <SyntaxHighlighter 
                language={language || 'text'} 
                style={vscDarkPlus}
                customStyle={{
                    padding: '1.5rem',
                    borderRadius: '0.75rem',
                    fontSize: '13px',
                    lineHeight: '1.6',
                    backgroundColor: '#09090b',
                    border: '1px solid rgba(255,255,255,0.05)'
                }}
            >
                {value}
            </SyntaxHighlighter>
        </div>
    );
};

export default function MarkdownRenderer({ content }) {
    return (
        <article className="prose prose-invert prose-cyan max-w-none">
            <ReactMarkdown
                components={{
                    h1: ({ children }) => (
                        <h1 id={slugify(String(children), { lower: true })} className="text-3xl md:text-4xl font-bold text-white mb-8 tracking-tight">
                            {children}
                        </h1>
                    ),
                    h2: ({ children }) => (
                        <h2 id={slugify(String(children), { lower: true })} className="text-xl font-bold text-white mt-12 mb-6 tracking-tight border-b border-white/5 pb-2">
                            {children}
                        </h2>
                    ),
                    h3: ({ children }) => (
                        <h3 id={slugify(String(children), { lower: true })} className="text-lg font-bold text-slate-200 mt-8 mb-4 tracking-tight">
                            {children}
                        </h3>
                    ),
                    p: ({ children }) => <p className="text-slate-400 leading-relaxed mb-6">{children}</p>,
                    code: ({ node, inline, className, children, ...props }) => {
                        const match = /language-(\w+)/.exec(className || '');
                        return !inline ? (
                            <CodeBlock 
                                language={match ? match[1] : ''} 
                                value={String(children).replace(/\n$/, '')} 
                            />
                        ) : (
                            <code className="bg-white/5 text-cyan-400 px-1.5 py-0.5 rounded text-[13px] font-mono border border-white/5" {...props}>
                                {children}
                            </code>
                        );
                    },
                    blockquote: ({ children }) => {
                        // Check for custom callouts like > [!NOTE]
                        const text = children?.[1]?.props?.children?.[0] || '';
                        const match = text.match(/^\[!(NOTE|WARNING|SECURITY|ENTERPRISE)\]/);
                        
                        if (match) {
                            const type = match[1];
                            const content = children[1].props.children.slice(1);
                            return <Callout type={type}>{content}</Callout>;
                        }

                        return (
                            <blockquote className="border-l-4 border-cyan-500/50 bg-cyan-500/5 px-6 py-4 my-8 rounded-r-xl text-slate-300 italic">
                                {children}
                            </blockquote>
                        );
                    },
                    ul: ({ children }) => <ul className="list-disc list-outside ml-6 space-y-2 mb-8 text-slate-400">{children}</ul>,
                    li: ({ children }) => <li className="pl-2">{children}</li>,
                    a: ({ href, children }) => (
                        <a href={href} className="text-cyan-400 hover:text-cyan-300 underline underline-offset-4 decoration-cyan-500/30 transition-colors">
                            {children}
                        </a>
                    )
                }}
            >
                {content}
            </ReactMarkdown>
        </article>
    );
}
