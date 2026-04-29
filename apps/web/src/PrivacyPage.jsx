const SECTIONS = [
    {
        num: '01',
        title: 'What OCULTAR Does',
        body: 'OCULTAR is a local, zero-egress PII detection and redaction engine. It runs entirely within your own infrastructure. No data you submit to OCULTAR is transmitted to any external server, cloud service, or third party by OCULTAR itself.',
    },
    {
        num: '02',
        title: 'Data Processing',
        items: [
            {
                label: 'What is processed',
                text: 'Text submitted through the refine_text tool or the /api/refine endpoint is analyzed locally to detect and redact PII. The redacted output and an encrypted form of the original values are stored in a local vault on your own machine or server.',
            },
            {
                label: 'Where processing happens',
                text: 'All detection, tokenization, and vault storage occur on the machine running the OCULTAR Refinery. No text, tokens, or vault contents are transmitted off that machine by OCULTAR.',
            },
            {
                label: 'What is stored',
                text: 'A local encrypted vault (AES-256-GCM) mapping deterministic token IDs to encrypted PII ciphertext — this file remains on your infrastructure. An optional audit log (Ed25519 hash-chained) records operation metadata (actor, action type, token ID, timestamp). No plaintext PII is written to the audit log.',
            },
        ],
    },
    {
        num: '03',
        title: 'No Telemetry',
        body: 'OCULTAR collects no usage analytics, crash reports, or telemetry of any kind. No data is sent to the OCULTAR project, its author, or any analytics platform.',
    },
    {
        num: '04',
        title: 'MCP Extensions',
        body: 'The ocultar-claude-mcp and ocultar-goose-mcp extensions communicate exclusively with the locally running OCULTAR Refinery over localhost. They make no outbound network calls to any external service. If the local Refinery is unreachable, both extensions fail closed — they return an error and refuse to forward your text elsewhere.',
    },
    {
        num: '05',
        title: 'Your Role as Data Controller',
        body: 'Because all data stays within your infrastructure, you — the operator deploying OCULTAR — are the data controller under GDPR and similar regulations. OCULTAR acts as a local data processor running entirely under your control. You are responsible for configuring access controls, key management, and audit log retention in accordance with your applicable data protection obligations.',
    },
    {
        num: '06',
        title: 'Third-Party Services',
        body: 'OCULTAR does not integrate with any third-party services by default. If you configure an upstream API target (OCU_PROXY_TARGET), OCULTAR forwards only the redacted output — never raw PII — to that target. The privacy practices of that upstream service are governed by its own policy.',
    },
    {
        num: '07',
        title: 'Data Retention and Deletion',
        body: 'Vault contents and audit logs are stored on your infrastructure and subject to your own retention policies. You can delete them at any time. OCULTAR provides no mechanism to transmit this data externally and retains no copy of it.',
    },
    {
        num: '08',
        title: "Children's Data",
        body: 'OCULTAR is a developer infrastructure tool not directed at children. We do not knowingly process data submitted by or about children.',
    },
    {
        num: '09',
        title: 'Changes to This Policy',
        body: 'Material changes will be noted in the CHANGELOG and reflected in the effective date above.',
    },
    {
        num: '10',
        title: 'Contact',
        body: 'For privacy questions or data requests: edu@ocultar.dev',
    },
];

export default function PrivacyPage() {
    return (
        <div className="animate-fade-in-up">
            <section className="section-padding">
                <div className="max-container max-w-3xl">
                    <h4 className="text-orange-500 text-[10px] uppercase font-mono tracking-[0.4em] mb-4">Legal</h4>
                    <h1 className="mb-4">Privacy Policy</h1>
                    <p className="text-[11px] font-mono text-zinc-500 uppercase tracking-widest mb-16">
                        Effective date: 28 April 2026 &nbsp;·&nbsp; Product: OCULTAR PII Refinery
                    </p>

                    <div className="space-y-16">
                        {SECTIONS.map(section => (
                            <div key={section.num} className="flex gap-8">
                                <div className="text-[11px] font-mono font-bold text-orange-500/40 tracking-widest shrink-0 mt-1 w-6">
                                    {section.num}
                                </div>
                                <div className="flex-1">
                                    <h3 className="text-base font-bold text-white mb-4 uppercase tracking-wide">
                                        {section.title}
                                    </h3>
                                    {section.body && (
                                        <p className="text-sm text-zinc-400 leading-relaxed">{section.body}</p>
                                    )}
                                    {section.items && (
                                        <div className="space-y-6">
                                            {section.items.map(item => (
                                                <div key={item.label} className="border-l-2 border-orange-500/20 pl-4">
                                                    <div className="text-[10px] font-bold text-orange-500 uppercase tracking-widest mb-2">
                                                        {item.label}
                                                    </div>
                                                    <p className="text-sm text-zinc-400 leading-relaxed">{item.text}</p>
                                                </div>
                                            ))}
                                        </div>
                                    )}
                                </div>
                            </div>
                        ))}
                    </div>

                    <div className="mt-20 pt-8 border-t border-zinc-800 text-[10px] text-zinc-500 font-mono uppercase tracking-widest">
                        OCULTAR SECURITY &nbsp;·&nbsp; ZERO-EGRESS BY ARCHITECTURE
                    </div>
                </div>
            </section>
        </div>
    );
}
