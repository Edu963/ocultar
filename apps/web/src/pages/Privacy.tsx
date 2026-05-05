import { Link } from 'react-router-dom';
import { ArrowLeft } from 'lucide-react';
import { useEffect } from 'react';

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
      { label: 'What is processed', text: 'Text submitted through the refine_text tool or the /api/refine endpoint is analyzed locally to detect and redact PII. The redacted output and an encrypted form of the original values are stored in a local vault on your own machine or server.' },
      { label: 'Where processing happens', text: 'All detection, tokenization, and vault storage occur on the machine running the OCULTAR Refinery. No text, tokens, or vault contents are transmitted off that machine by OCULTAR.' },
      { label: 'What is stored', text: 'A local encrypted vault (AES-256-GCM) mapping deterministic token IDs to encrypted PII ciphertext — this file remains on your infrastructure. An optional audit log (Ed25519 hash-chained) records operation metadata (actor, action type, token ID, timestamp). No plaintext PII is written to the audit log.' },
    ],
  },
  { num: '03', title: 'No Telemetry', body: 'OCULTAR collects no usage analytics, crash reports, or telemetry of any kind. No data is sent to the OCULTAR project, its author, or any analytics platform.' },
  { num: '04', title: 'MCP Extensions', body: 'The ocultar-claude-mcp, ocultar-goose-mcp, and ocultar-mistral-mcp extensions communicate exclusively with the locally running OCULTAR Refinery over localhost. They make no outbound network calls to any external service. If the local Refinery is unreachable, all extensions fail closed — they return an error and refuse to forward your text elsewhere.' },
  { num: '05', title: 'Your Role as Data Controller', body: 'Because all data stays within your infrastructure, you — the operator deploying OCULTAR — are the data controller under GDPR and similar regulations. OCULTAR acts as a local data processor running entirely under your control. You are responsible for configuring access controls, key management, and audit log retention in accordance with your applicable data protection obligations.' },
  { num: '06', title: 'Third-Party Services', body: 'OCULTAR does not integrate with any third-party services by default. If you configure an upstream API target (OCU_PROXY_TARGET), OCULTAR forwards only the redacted output — never raw PII — to that target. The privacy practices of that upstream service are governed by its own policy.' },
  { num: '07', title: 'Data Retention and Deletion', body: 'Vault contents and audit logs are stored on your infrastructure and subject to your own retention policies. You can delete them at any time. OCULTAR provides no mechanism to transmit this data externally and retains no copy of it.' },
  { num: '08', title: "Children's Data", body: 'OCULTAR is a developer infrastructure tool not directed at children. We do not knowingly process data submitted by or about children.' },
  { num: '09', title: 'Changes to This Policy', body: 'Material changes will be noted in the CHANGELOG and reflected in the effective date above.' },
  { 
    num: '10', 
    title: 'Contact', 
    body: (
      <span>
        For privacy questions or data requests: <a href="mailto:edu@ocultar.com" className="text-emerald-500 hover:underline font-bold">edu@ocultar.com</a>
      </span>
    )
  },
];

export default function Privacy() {
  useEffect(() => {
    document.title = "Privacy Policy — OCULTAR";
  }, []);

  return (
    <div className="min-h-screen bg-[#050505]">
      <div className="max-container max-w-2xl py-12 md:py-16">

        <Link
          to="/"
          className="inline-flex items-center gap-2 text-xs font-mono font-bold uppercase tracking-widest text-slate-600 hover:text-slate-300 transition-colors mb-12"
        >
          <ArrowLeft className="w-3 h-3" /> Platform
        </Link>

        <div className="mb-16">
          <p className="text-xs font-mono font-semibold uppercase tracking-widest text-emerald-500 mb-4">Legal</p>
          <h1 className="text-3xl font-bold text-white tracking-tight mb-4">Privacy Policy</h1>
          <p className="text-xs font-mono text-slate-600 uppercase tracking-widest leading-loose">
            Effective date: 28 April 2026 &nbsp;·&nbsp; Product: OCULTAR PII Refinery &nbsp;·&nbsp; Contact: <a href="mailto:edu@ocultar.com" className="hover:text-slate-300 underline underline-offset-2">edu@ocultar.com</a>
          </p>
        </div>

        <div className="flex flex-col gap-12">
          {SECTIONS.map(section => (
            <div key={section.num} className="flex gap-8">
              <div className="text-xs font-mono font-bold text-emerald-500/30 tracking-widest shrink-0 mt-0.5 w-6">
                {section.num}
              </div>
              <div className="flex-1 flex flex-col gap-4">
                <h3 className="text-sm font-bold text-white uppercase tracking-wide">{section.title}</h3>
                {section.body && (
                  <div className="text-sm text-slate-400 leading-relaxed">
                    {section.body}
                  </div>
                )}
                {section.items && (
                  <div className="flex flex-col gap-6">
                    {section.items.map(item => (
                      <div key={item.label} className="border-l border-white/10 pl-4 flex flex-col gap-2">
                        <div className="text-xs font-mono font-semibold text-emerald-500/70 uppercase tracking-widest">{item.label}</div>
                        <p className="text-sm text-slate-400 leading-relaxed">{item.text}</p>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>

        <div className="mt-20 pt-8 border-t border-white/5 text-xs font-mono text-slate-700 uppercase tracking-widest">
          OCULTAR Security &nbsp;·&nbsp; Zero-Egress by Architecture
        </div>
      </div>
    </div>
  );
}
