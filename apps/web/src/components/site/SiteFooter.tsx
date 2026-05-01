import { Github } from "lucide-react";
import { toast } from "sonner";

const COLS = [
  {
    title: "Platform",
    links: [
      { label: "Architecture", href: "#architecture" },
      { label: "Compliance", href: "#compliance" },
      { label: "Developers", href: "#developers" },
    ],
  },
  {
    title: "Resources",
    links: [
      { label: "GitHub", href: "https://github.com/Edu963/ocultar" },
      { label: "Documentation", href: "#" },
      { label: "Security", href: "#" },
    ],
  },
  {
    title: "Company",
    links: [
      { label: "Contact", href: "mailto:sales@ocultar.dev" },
      { label: "Engineering", href: "mailto:engineering@ocultar.dev" },
      { label: "Privacy", href: "#" },
    ],
  },
];

export const SiteFooter = () => (
  <footer className="border-t border-border bg-background">
    <div className="container-page py-16">
      <div className="grid grid-cols-2 md:grid-cols-5 gap-10">
        <div className="col-span-2 flex flex-col gap-4">
          <div className="flex items-center gap-2">
            <span className="inline-block h-2 w-2 rounded-sm bg-primary" />
            <span className="font-mono text-sm font-bold tracking-[0.18em] text-foreground">
              OCULTAR
            </span>
          </div>
          <p className="max-w-xs text-[13px] leading-relaxed text-muted-foreground">
            The EU-sovereign, zero-egress compliance layer for enterprise AI.
            PII never leaves your infrastructure.
          </p>
          <a
            href="https://github.com/Edu963/ocultar"
            className="inline-flex w-fit items-center gap-2 text-[13px] text-muted-foreground hover:text-foreground transition-colors"
          >
            <Github className="h-4 w-4" /> github.com/Edu963/ocultar
          </a>
        </div>
        {COLS.map((c) => (
          <div key={c.title} className="flex flex-col gap-3">
            <h4 className="mono-label">{c.title}</h4>
            <ul className="flex flex-col gap-2">
              {c.links.map((l) => (
                <li key={l.label}>
                  {l.href.startsWith("mailto:") ? (
                    <button
                      onClick={(e) => {
                        e.preventDefault();
                        const email = l.href.replace("mailto:", "");
                        navigator.clipboard.writeText(email);
                        toast.success("Email copied to clipboard", {
                          description: `Reach out to ${email} to get in touch.`,
                        });
                      }}
                      className="text-[13px] text-muted-foreground hover:text-foreground transition-colors cursor-pointer text-left"
                    >
                      {l.label}
                    </button>
                  ) : (
                    <a
                      href={l.href}
                      className="text-[13px] text-muted-foreground hover:text-foreground transition-colors"
                    >
                      {l.label}
                    </a>
                  )}
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>

      <div className="mt-14 pt-6 border-t border-border flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <p className="font-mono text-[10px] uppercase tracking-[0.22em] text-muted-foreground">
          © {new Date().getFullYear()} OCULTAR · All rights reserved
        </p>
        <p className="font-mono text-[10px] uppercase tracking-[0.22em] text-primary/70">
          Zero-egress · Sovereign · Fail-closed
        </p>
      </div>
    </div>
  </footer>
);
