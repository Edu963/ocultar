import { ArrowRight, ShieldCheck } from "lucide-react";

const REQUEST_ACCESS_URL =
  "mailto:sales@ocultar.dev?subject=Access%20Request&body=Hi%2C%0A%0AWork%20email%3A%20%0ACompany%20name%3A%20%0APrimary%20use%20case%3A%20";

export const Hero = () => {
  return (
    <section
      id="top"
      className="relative overflow-hidden pt-32 pb-20 md:pt-40 md:pb-28"
    >
      {/* Background layers */}
      <div className="pointer-events-none absolute inset-0 bg-grid bg-grid-fade opacity-60" />
      <div className="pointer-events-none absolute inset-x-0 top-0 h-[640px] bg-gradient-hero" />
      <div className="pointer-events-none absolute inset-x-0 bottom-0 h-32 bg-gradient-to-b from-transparent to-background" />

      <div className="container-page relative">
        <div className="flex flex-col items-center text-center">
          <div className="pill pill-accent animate-fade-up">
            <ShieldCheck className="h-3.5 w-3.5" />
            <span>EU-Sovereign · Zero-Egress · Fail-Closed</span>
          </div>

          <h1
            className="mt-8 max-w-4xl text-balance text-[44px] sm:text-[56px] md:text-[72px] font-semibold leading-[1.02] tracking-tightest text-gradient animate-fade-up"
            style={{ animationDelay: "60ms" }}
          >
            The compliance layer
            <br className="hidden sm:block" /> for{" "}
            <span className="text-foreground">enterprise AI</span>.
          </h1>

          <p
            className="mt-7 max-w-2xl text-pretty text-[17px] leading-relaxed text-muted-foreground animate-fade-up"
            style={{ animationDelay: "120ms" }}
          >
            OCULTAR sits between your applications and any LLM — detecting,
            vaulting and re-hydrating PII inside your own infrastructure. GDPR
            and EU AI Act ready. Operational before the August 2, 2026
            enforcement deadline.
          </p>

          <div
            className="mt-10 flex flex-col sm:flex-row items-center gap-3 animate-fade-up"
            style={{ animationDelay: "180ms" }}
          >
            <a
              href={REQUEST_ACCESS_URL}
              className="group inline-flex h-11 items-center gap-2 rounded-md bg-foreground px-5 text-sm font-semibold text-background hover:bg-foreground/90 transition-colors"
            >
              Request enterprise access
              <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5" />
            </a>
            <a
              href="#architecture"
              className="inline-flex h-11 items-center gap-2 rounded-md border border-border bg-surface/60 px-5 text-sm font-semibold text-foreground hover:border-border-strong hover:bg-surface transition-colors"
            >
              See the architecture
            </a>
          </div>

          <p
            className="mt-6 font-mono text-[11px] uppercase tracking-[0.2em] text-muted-foreground animate-fade-up"
            style={{ animationDelay: "240ms" }}
          >
            Apache 2.0 core · Self-hosted · No data leaves your VPC
          </p>
        </div>

        {/* Hero product surface */}
        <div
          className="mt-16 md:mt-20 relative animate-fade-up"
          style={{ animationDelay: "300ms" }}
        >
          <div className="absolute inset-x-12 -top-10 h-40 bg-primary/10 blur-3xl rounded-full" />
          <HeroFlow />
        </div>
      </div>
    </section>
  );
};

const HeroFlow = () => (
  <div className="relative mx-auto max-w-5xl rounded-2xl border border-border bg-surface/70 backdrop-blur-md shadow-elevated overflow-hidden">
    <div className="flex items-center justify-between border-b border-border px-5 py-3">
      <div className="flex items-center gap-2">
        <span className="h-2 w-2 rounded-full bg-primary animate-pulse-dot" />
        <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-muted-foreground">
          ocultar · request pipeline
        </span>
      </div>
      <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-muted-foreground">
        live
      </span>
    </div>

    <div className="grid grid-cols-1 md:grid-cols-3 divide-y md:divide-y-0 md:divide-x divide-border">
      <FlowStep
        index="01"
        title="Detect"
        body="Local SLM identifies 40+ PII entity types in-stream."
        sample={
          <code className="text-foreground">
            "Hi, my name is{" "}
            <span className="rounded bg-destructive/15 px-1.5 py-0.5 text-destructive">
              John Doe
            </span>
            , card{" "}
            <span className="rounded bg-destructive/15 px-1.5 py-0.5 text-destructive">
              4111&nbsp;1111&nbsp;1111&nbsp;1111
            </span>
            "
          </code>
        }
      />
      <FlowStep
        index="02"
        title="Vault"
        body="AES-256-GCM encrypted, replaced with deterministic tokens."
        sample={
          <code className="text-muted-foreground">
            "Hi, my name is{" "}
            <span className="rounded bg-primary/15 px-1.5 py-0.5 text-primary">
              [PERSON_a1b2]
            </span>
            , card{" "}
            <span className="rounded bg-primary/15 px-1.5 py-0.5 text-primary">
              [CC_7d9e]
            </span>
            "
          </code>
        }
      />
      <FlowStep
        index="03"
        title="Restore"
        body="Tokens re-hydrated for the authorized caller. Provider sees nothing."
        sample={
          <code className="text-foreground">
            "Hi, my name is{" "}
            <span className="rounded bg-success/15 px-1.5 py-0.5 text-success">
              John Doe
            </span>
            , card{" "}
            <span className="rounded bg-success/15 px-1.5 py-0.5 text-success">
              4111&nbsp;1111&nbsp;1111&nbsp;1111
            </span>
            "
          </code>
        }
      />
    </div>
  </div>
);

const FlowStep = ({
  index,
  title,
  body,
  sample,
}: {
  index: string;
  title: string;
  body: string;
  sample: React.ReactNode;
}) => (
  <div className="p-6 md:p-7 flex flex-col gap-4">
    <div className="flex items-center gap-3">
      <span className="font-mono text-[10px] tracking-[0.22em] text-muted-foreground">
        {index}
      </span>
      <span className="h-px flex-1 bg-border" />
      <span className="text-[13px] font-semibold text-foreground">{title}</span>
    </div>
    <p className="text-[13px] leading-relaxed text-muted-foreground">{body}</p>
    <div className="rounded-md border border-border bg-background/60 p-3 font-mono text-[11px] leading-relaxed">
      {sample}
    </div>
  </div>
);
