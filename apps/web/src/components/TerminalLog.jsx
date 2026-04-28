import { useState, useEffect, useRef } from 'react';

const TerminalLog = () => {
    const [lines, setLines] = useState([]);
    const containerRef = useRef(null);

    const logData = [
        { type: 'input', text: '➜ /usr/local/ocultar --start-proxy', delay: 500 },
        { type: 'system', text: 'Initialising local SLM deep scan... [OK]', delay: 800 },
        { type: 'system', text: 'Listening on http://localhost:8080', delay: 400 },
        { type: 'incoming', text: '[INCOMING] POST /v1/chat/completions', delay: 1200 },
        { type: 'detect', text: '[PII_DETECTED] "John Doe" (NAME)', delay: 200 },
        { type: 'detect', text: '[PII_DETECTED] "john.doe@gmail.com" (EMAIL)', delay: 150 },
        { type: 'action', text: '[REDACTING] Tokenizing payload...', delay: 600 },
        { type: 'action', text: '[FORWARDING] Clean data sent to api.openai.com', delay: 300 },
        { type: 'system', text: '--- Waiting for next request ---', delay: 2000 },
    ];

    useEffect(() => {
        let currentLine = 0;
        let timeoutId;

        const addLine = () => {
            if (currentLine >= logData.length) {
                setTimeout(() => {
                    setLines([]);
                    currentLine = 0;
                    addLine();
                }, 3000);
                return;
            }

            setLines((prev) => [...prev, logData[currentLine]]);
            timeoutId = setTimeout(() => {
                currentLine++;
                addLine();
            }, logData[currentLine].delay);
        };

        addLine();

        return () => clearTimeout(timeoutId);
    }, []);

    useEffect(() => {
        if (containerRef.current) {
            containerRef.current.scrollTop = containerRef.current.scrollHeight;
        }
    }, [lines]);

    return (
        <div className="relative group w-full max-w-2xl mx-auto mt-12">
            <div className="absolute -inset-1 bg-gradient-to-r from-cyan-500/20 to-sky-500/20 rounded-lg blur opacity-25 group-hover:opacity-40 transition duration-1000"></div>
            <div className="relative bg-[#0d0d0f] border border-white/10 rounded-lg overflow-hidden shadow-2xl">
                <div className="flex items-center gap-2 px-4 py-3 bg-white/5 border-b border-white/10">
                    <div className="flex gap-1.5">
                        <div className="w-3 h-3 rounded-full bg-red-500/20"></div>
                        <div className="w-3 h-3 rounded-full bg-yellow-500/20"></div>
                        <div className="w-3 h-3 rounded-full bg-cyan-500/20"></div>
                    </div>
                    <div className="text-[10px] font-mono text-slate-500 uppercase tracking-widest ml-4">Local Proxy Log — v2.4</div>
                </div>
                <div ref={containerRef} className="p-6 font-mono text-sm leading-relaxed h-[300px] overflow-y-auto scrollbar-hide">
                    {lines.map((line, i) => (
                        <div key={i} className="mb-1 animate-in fade-in slide-in-from-left-2 duration-300">
                            {line.type === 'input' && <span className="text-cyan-500">{line.text}</span>}
                            {line.type === 'system' && <span className="text-slate-500">{line.text}</span>}
                            {line.type === 'incoming' && <span className="text-sky-400">{line.text}</span>}
                            {line.type === 'detect' && <span className="text-yellow-400">{line.text}</span>}
                            {line.type === 'action' && <span className="text-cyan-400">{line.text}</span>}
                            {line.type === 'rehydrate' && <span className="text-purple-400">{line.text}</span>}
                        </div>
                    ))}
                    <div className="inline-block w-2 h-5 bg-cyan-500 animate-pulse ml-1 align-middle"></div>
                </div>
            </div>
        </div>
    );
};

export default TerminalLog;
