import { Lock, ShieldCheck, Cpu, Database, Network, ScrollText } from "lucide-react";
import { SectionHeader } from "./ProblemSection";

const PILLARS = [
  {
    icon: Lock,
    title: "Zero-Egress",
    body: "Sensitive data is technically incapable of leaving your VPC. Not a policy — a network and process boundary.",
  },
  {
    icon: ShieldCheck,
    title: "Fail-Closed",
    body: "Any error blocks the request. OCULTAR never falls back to passthrough. Safe by construction.",
  },
  {
    icon: Cpu,
    title: "Local SLM",
    body: "Contextual NER runs on-premise. No external inference, no third-party model calls.",
  },
  {
    icon: Database,
    title: "Deterministic Vault",
    body: "AES-256-GCM with HKDF-SHA256 derivation. Stable tokens enable safe re-hydration.",
  },
  {
    icon: Network,
    title: "Drop-in Proxy",
    body: "OpenAI-compatible endpoint. Change a base URL, ship in an afternoon. Works with any SDK.",
  },
  {
    icon: ScrollText,
    title: "Signed Audit Chain",
    body: "Ed25519-signed, SHA-256 hash-chained log. SIEM-ready, regulator-ready, tamper-evident.",
  },
];

export const Architecture = () => (
  <section id="architecture" className="section border-t border-border bg-surface/30">
    <div className="container-page">
      <SectionHeader
        eyebrow="Architecture"
        title="Six guarantees, enforced at the boundary."
        description="OCULTAR is not a wrapper SDK or an after-the-fact log scanner. It is a sovereign runtime that enforces compliance before bytes ever leave your environment."
      />
      <div className="mt-14 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-px bg-border rounded-xl overflow-hidden border border-border">
        {PILLARS.map((p) => (
          <div
            key={p.title}
            className="group bg-background p-7 flex flex-col gap-4 hover:bg-surface transition-colors"
          >
            <div className="flex items-center gap-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-md bg-primary/10 text-primary border border-primary/20">
                <p.icon className="h-4 w-4" />
              </div>
              <h3 className="text-[15px] font-semibold text-foreground">
                {p.title}
              </h3>
            </div>
            <p className="text-[14px] leading-relaxed text-muted-foreground">
              {p.body}
            </p>
          </div>
        ))}
      </div>
    </div>
  </section>
);
