import { Check } from "lucide-react";
import { SectionHeader } from "./ProblemSection";

const ENTRIES = [
  { time: "09:14:32", op: "VAULTED", type: "PERSON", input: '"John Doe"', output: "[PERSON_a1b2c3d4]" },
  { time: "09:14:32", op: "VAULTED", type: "CREDIT_CARD", input: '"4111111111111111"', output: "[CC_7d9e1f2a]" },
  { time: "09:14:33", op: "FORWARDED", type: "—", input: "api.openai.com", output: "status: 200" },
  { time: "09:14:33", op: "REHYDRATED", type: "PERSON", input: "[PERSON_a1b2c3d4]", output: '"John Doe"' },
];

const opColor: Record<string, string> = {
  VAULTED: "text-primary",
  FORWARDED: "text-muted-foreground",
  REHYDRATED: "text-success",
};

export const Compliance = () => (
  <section id="compliance" className="section border-t border-border bg-surface/30">
    <div className="container-page grid grid-cols-1 lg:grid-cols-12 gap-14 items-start">
      <div className="lg:col-span-5">
        <SectionHeader
          eyebrow="Compliance"
          title="Cryptographic proof for every request."
          description="Every vault, forward and re-hydration is written to a tamper-evident, SIEM-ready audit chain. Answer regulator questions in seconds, not weeks."
        />
        <ul className="mt-8 flex flex-col gap-3">
          {[
            "Ed25519-signed entries, hash-chained with SHA-256",
            "Syslog UDP forwarder, PII-scrubbed by construction",
            "GDPR Art. 5(2) accountability, out of the box",
            "Maps to ISO 27001, SOC 2 CC6/CC7, HIPAA §164.312",
          ].map((p) => (
            <li key={p} className="flex items-start gap-3 text-[14px]">
              <span className="mt-0.5 flex h-5 w-5 items-center justify-center rounded-full border border-primary/30 bg-primary/10">
                <Check className="h-3 w-3 text-primary" />
              </span>
              <span className="text-muted-foreground">{p}</span>
            </li>
          ))}
        </ul>
      </div>

      <div className="lg:col-span-7">
        <div className="rounded-xl border border-border bg-background/80 backdrop-blur-sm overflow-hidden shadow-card">
          <div className="flex items-center justify-between border-b border-border px-5 py-3">
            <div className="flex items-center gap-2">
              <span className="h-1.5 w-1.5 rounded-full bg-primary animate-pulse-dot" />
              <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-muted-foreground">
                audit.log · real-time
              </span>
            </div>
            <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-muted-foreground">
              req 8f3a · ed25519
            </span>
          </div>
          <div className="divide-y divide-border font-mono text-[11.5px]">
            {ENTRIES.map((e, i) => (
              <div
                key={i}
                className="grid grid-cols-12 gap-3 px-5 py-3 hover:bg-surface/60 transition-colors"
              >
                <span className="col-span-2 text-muted-foreground">[{e.time}]</span>
                <span className={`col-span-2 font-semibold ${opColor[e.op]}`}>{e.op}</span>
                <span className="col-span-2 text-muted-foreground">{e.type}</span>
                <span className="col-span-3 text-foreground/80 truncate">{e.input}</span>
                <span className="col-span-3 text-foreground truncate">→ {e.output}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="mt-4 grid grid-cols-3 gap-3">
          {[
            { k: "GDPR", v: "Art. 5, 25, 28, 32" },
            { k: "EU AI Act", v: "Aug 2, 2026 ready" },
            { k: "SOC 2 / ISO", v: "Evidence-ready" },
          ].map((c) => (
            <div
              key={c.k}
              className="rounded-md border border-border bg-surface/60 px-4 py-3"
            >
              <div className="font-mono text-[10px] uppercase tracking-[0.22em] text-muted-foreground">
                {c.k}
              </div>
              <div className="mt-1 text-[13px] font-medium text-foreground">{c.v}</div>
            </div>
          ))}
        </div>
      </div>
    </div>
  </section>
);
