import { SiteNav } from "@/components/site/SiteNav";
import { SiteFooter } from "@/components/site/SiteFooter";
import { useEffect } from "react";

const SECTIONS = [
  {
    id: "what-ocultar-does",
    title: "What OCULTAR Does",
    body: "OCULTAR is a local, zero-egress PII detection and redaction engine. It runs entirely within your own infrastructure. No data you submit to OCULTAR is transmitted to any external server, cloud service, or third party by OCULTAR itself.",
  },
  {
    id: "data-processing",
    title: "Data Processing",
    items: [
      {
        label: "What is processed",
        text: "Text submitted through the refine_text tool or the /api/refine endpoint is analysed locally to detect and redact PII. The redacted output and an encrypted form of the original values are stored in a local vault on your own machine or server.",
      },
      {
        label: "Where processing happens",
        text: "All detection, tokenisation, and vault storage occur on the machine running the OCULTAR Refinery. No text, tokens, or vault contents are transmitted off that machine by OCULTAR.",
      },
      {
        label: "What is stored",
        text: "A local encrypted vault (AES-256-GCM) mapping deterministic token IDs to encrypted PII ciphertext — this file remains on your infrastructure. An optional audit log (Ed25519 hash-chained) records operation metadata (actor, action type, token ID, timestamp). No plaintext PII is written to the audit log.",
      },
    ],
  },
  {
    id: "no-telemetry",
    title: "No Telemetry",
    body: "OCULTAR collects no usage analytics, crash reports, or telemetry of any kind. No data is sent to the OCULTAR project, its author, or any analytics platform.",
  },
  {
    id: "mcp-extensions",
    title: "MCP Extensions",
    body: "The ocultar-claude-mcp, ocultar-goose-mcp, and ocultar-mistral-mcp extensions communicate exclusively with the locally running OCULTAR Refinery over localhost. They make no outbound network calls to any external service. If the local Refinery is unreachable, all extensions fail closed — they return an error and refuse to forward your text elsewhere.",
  },
  {
    id: "data-controller",
    title: "Your Role as Data Controller",
    body: "Because all data stays within your infrastructure, you — the operator deploying OCULTAR — are the data controller under GDPR and similar regulations. OCULTAR acts as a local data processor running entirely under your control. You are responsible for configuring access controls, key management, and audit log retention in accordance with your applicable data protection obligations.",
  },
  {
    id: "third-party",
    title: "Third-Party Services",
    body: "OCULTAR does not integrate with any third-party services by default. If you configure an upstream API target (OCU_PROXY_TARGET), OCULTAR forwards only the redacted output — never raw PII — to that target. The privacy practices of that upstream service are governed by its own policy.",
  },
  {
    id: "retention",
    title: "Data Retention and Deletion",
    body: "Vault contents and audit logs are stored on your infrastructure and subject to your own retention policies. You can delete them at any time. OCULTAR provides no mechanism to transmit this data externally and retains no copy of it.",
  },
  {
    id: "children",
    title: "Children's Data",
    body: "OCULTAR is a developer infrastructure tool not directed at children. We do not knowingly process data submitted by or about children.",
  },
  {
    id: "changes",
    title: "Changes to This Policy",
    body: "Material changes will be noted in the CHANGELOG and reflected in the effective date above.",
  },
];

export default function Privacy() {
  useEffect(() => {
    document.title = "Privacy Policy — OCULTAR";
  }, []);

  return (
    <div className="min-h-screen bg-background text-foreground">
      <SiteNav />

      <main className="container-page max-w-3xl py-24 md:py-32">

        {/* Header */}
        <div className="mb-16 border-b border-border pb-10">
          <p className="font-mono text-[11px] uppercase tracking-[0.2em] text-primary mb-4">
            Legal
          </p>
          <h1 className="text-4xl font-semibold tracking-tight text-foreground mb-4">
            Privacy Policy
          </h1>
          <p className="text-sm text-muted-foreground leading-relaxed">
            Effective date: 28 April 2026 &nbsp;·&nbsp; Product: OCULTAR PII Refinery &nbsp;·&nbsp; Apache 2.0 open-source
          </p>
        </div>

        {/* Developer note */}
        <div className="mb-14 rounded-lg border border-primary/20 bg-primary/5 px-6 py-5">
          <p className="font-mono text-[11px] uppercase tracking-[0.2em] text-primary mb-2">
            Note for developers
          </p>
          <p className="text-sm text-muted-foreground leading-relaxed">
            OCULTAR is self-hosted software. It has no cloud component, no accounts, and no
            server that receives your data. This policy describes how the software itself
            handles data when you run it — not how we as a service handle it, because we are
            not a service.
          </p>
        </div>

        {/* Sections */}
        <div className="flex flex-col divide-y divide-border">
          {SECTIONS.map((section) => (
            <div key={section.id} id={section.id} className="py-10">
              <h2 className="text-base font-semibold text-foreground mb-4">
                {section.title}
              </h2>
              {section.body && (
                <p className="text-[14px] leading-relaxed text-muted-foreground">
                  {section.body}
                </p>
              )}
              {section.items && (
                <div className="flex flex-col gap-6 mt-2">
                  {section.items.map((item) => (
                    <div key={item.label}>
                      <p className="font-mono text-[11px] uppercase tracking-[0.16em] text-primary/70 mb-2">
                        {item.label}
                      </p>
                      <p className="text-[14px] leading-relaxed text-muted-foreground">
                        {item.text}
                      </p>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}

          {/* Contact */}
          <div id="contact" className="py-10">
            <h2 className="text-base font-semibold text-foreground mb-4">
              Contact
            </h2>
            <p className="text-[14px] leading-relaxed text-muted-foreground">
              For privacy questions or data requests:{" "}
              <a
                href="mailto:edu@ocultar.dev"
                className="text-primary hover:underline underline-offset-4"
              >
                edu@ocultar.dev
              </a>
            </p>
          </div>
        </div>

      </main>

      <SiteFooter />
    </div>
  );
}
