import { ArrowRight } from "lucide-react";

const REQUEST_ACCESS_URL =
  "mailto:sales@ocultar.dev?subject=Access%20Request&body=Hi%2C%0A%0AWork%20email%3A%20%0ACompany%20name%3A%20%0APrimary%20use%20case%3A%20";

export const FinalCTA = () => (
  <section className="relative section border-t border-border overflow-hidden">
    <div className="pointer-events-none absolute inset-0 bg-gradient-hero opacity-70" />
    <div className="container-page relative">
      <div className="mx-auto max-w-3xl text-center flex flex-col items-center">
        <span className="pill pill-accent">
          <span className="h-1.5 w-1.5 rounded-full bg-primary animate-pulse-dot" />
          Limited enterprise access · 2026 cohort
        </span>
        <h2 className="mt-7 text-balance text-[36px] sm:text-[48px] font-semibold leading-[1.05] tracking-tight text-gradient">
          Ship AI without shipping your customers' data.
        </h2>
        <p className="mt-5 text-[15px] text-muted-foreground max-w-xl">
          Talk to our team about deploying OCULTAR inside your VPC. Most pilots
          are operational within two weeks.
        </p>
        <div className="mt-9 flex flex-col sm:flex-row gap-3">
          <a
            href={REQUEST_ACCESS_URL}
            className="group inline-flex h-11 items-center gap-2 rounded-md bg-foreground px-6 text-sm font-semibold text-background hover:bg-foreground/90 transition-colors"
          >
            Request enterprise access
            <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5" />
          </a>
          <a
            href="https://github.com/Edu963/ocultar"
            target="_blank"
            rel="noreferrer"
            className="inline-flex h-11 items-center gap-2 rounded-md border border-border bg-surface/60 px-6 text-sm font-semibold text-foreground hover:border-border-strong transition-colors"
          >
            Read the source
          </a>
        </div>
      </div>
    </div>
  </section>
);
