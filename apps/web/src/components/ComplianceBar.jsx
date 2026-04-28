import { Shield } from 'lucide-react';

const ComplianceBar = () => {
    const badges = [
        'GDPR COMPLIANT',
        'EU AI ACT READY',
        'HIPAA SECURE',
        'NIS2 CERTIFIED',
        'BSI C5 ALIGNED',
        'SOC2 TYPE II',
        'ISO 27001'
    ];

    return (
        <div className="w-full border-y border-white/5 bg-white/[0.02] py-8 overflow-hidden relative">
            <div className="absolute inset-y-0 left-0 w-32 bg-gradient-to-r from-[#050505] to-transparent z-10"></div>
            <div className="absolute inset-y-0 right-0 w-32 bg-gradient-to-l from-[#050505] to-transparent z-10"></div>
            
            <div className="flex animate-marquee whitespace-nowrap">
                {[...badges, ...badges].map((text, i) => (
                    <div key={i} className="flex items-center gap-3 px-12 opacity-40 hover:opacity-100 transition-opacity grayscale hover:grayscale-0">
                        <Shield className="w-5 h-5 text-cyan-500" />
                        <span className="font-mono text-xs font-bold tracking-widest text-white">{text}</span>
                    </div>
                ))}
            </div>

            <style dangerouslySetInnerHTML={{ __html: `
                @keyframes marquee {
                    0% { transform: translateX(0); }
                    100% { transform: translateX(-50%); }
                }
                .animate-marquee {
                    display: flex;
                    width: max-content;
                    animation: marquee 30s linear infinite;
                }
            `}} />
        </div>
    );
};

export default ComplianceBar;
