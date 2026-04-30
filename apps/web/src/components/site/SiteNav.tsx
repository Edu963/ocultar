import { Github, ArrowRight } from "lucide-react";
import { useEffect, useState } from "react";

const REQUEST_ACCESS_URL =
  "mailto:sales@ocultar.dev?subject=Access%20Request&body=Hi%2C%0A%0AWork%20email%3A%20%0ACompany%20name%3A%20%0APrimary%20use%20case%3A%20";

const NAV_LINKS = [
  { label: "Platform", href: "/#platform" },
  { label: "Solutions", href: "/solutions" },
  { label: "Architecture", href: "/#architecture" },
  { label: "Compliance", href: "/#compliance" },
  { label: "Developers", href: "/#developers" },
  { label: "Docs", href: "/docs" },
];

export const SiteNav = () => {
  const [scrolled, setScrolled] = useState(false);
  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 8);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <header
      className={`fixed top-0 left-0 right-0 z-50 transition-all duration-300 ${
        scrolled
          ? "bg-background/80 backdrop-blur-xl border-b border-border"
          : "bg-transparent border-b border-transparent"
      }`}
    >
      <div className="container-page flex h-16 items-center justify-between">
        <a href="#top" className="flex items-center gap-2">
          <span
            className="inline-block h-2 w-2 rounded-sm bg-primary shadow-[0_0_12px_hsl(var(--primary)/0.7)]"
            aria-hidden
          />
          <span className="font-mono text-[15px] font-bold tracking-[0.18em] text-foreground">
            OCULTAR
          </span>
        </a>

        <nav className="hidden md:flex items-center gap-8">
          {NAV_LINKS.map((l) => (
            <a
              key={l.href}
              href={l.href}
              className="text-[13px] font-medium text-muted-foreground hover:text-foreground transition-colors"
            >
              {l.label}
            </a>
          ))}
        </nav>

        <div className="flex items-center gap-2">
          <a
            href="https://github.com/Edu963/ocultar"
            target="_blank"
            rel="noreferrer"
            className="hidden sm:inline-flex h-9 items-center gap-2 rounded-md px-3 text-[13px] font-medium text-muted-foreground hover:text-foreground hover:bg-surface transition-colors"
            aria-label="OCULTAR on GitHub"
          >
            <Github className="h-4 w-4" />
            GitHub
          </a>
          <a
            href={REQUEST_ACCESS_URL}
            className="group inline-flex h-9 items-center gap-1.5 rounded-md bg-foreground px-4 text-[13px] font-semibold text-background hover:bg-foreground/90 transition-colors"
          >
            Request access
            <ArrowRight className="h-3.5 w-3.5 transition-transform group-hover:translate-x-0.5" />
          </a>
        </div>
      </div>
    </header>
  );
};
