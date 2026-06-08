import { ArrowRight } from "lucide-react";

export const FinalCTA = () => (
  <section className="relative section border-t border-border overflow-hidden">
    <div className="pointer-events-none absolute inset-0 bg-gradient-hero opacity-70" />
    <div className="container-page relative">
      <div className="mx-auto max-w-3xl text-center flex flex-col items-center">
        <span className="pill pill-accent">
          <span className="h-1.5 w-1.5 rounded-full bg-primary animate-pulse-dot" />
          Open source · Free forever
        </span>
        <h2 className="mt-7 text-balance text-[36px] sm:text-[48px] font-semibold leading-[1.05] tracking-tight text-gradient">
          Ship AI without shipping your customers' data.
        </h2>
        <p className="mt-5 text-[15px] text-muted-foreground max-w-xl">
          Deploy in minutes with Docker. No cloud account, no license key — your
          data never leaves your infrastructure.
        </p>
        <div className="mt-9 flex flex-col sm:flex-row gap-3">
          <a
            href="https://github.com/Edu963/ocultar"
            target="_blank"
            rel="noreferrer"
            className="group inline-flex h-11 items-center gap-2 rounded-md bg-foreground px-6 text-sm font-semibold text-background hover:bg-foreground/90 transition-colors"
          >
            Get started free
            <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5" />
          </a>
          <a
            href="https://github.com/Edu963/ocultar#quick-start--docker"
            target="_blank"
            rel="noreferrer"
            className="inline-flex h-11 items-center gap-2 rounded-md border border-border bg-surface/60 px-6 text-sm font-semibold text-foreground hover:border-border-strong transition-colors"
          >
            Docker quickstart
          </a>
        </div>
      </div>
    </div>
  </section>
);
