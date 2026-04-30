import { useEffect, useState } from "react";
import { useParams, useNavigate, Link } from "react-router-dom";
import ReactMarkdown from "react-markdown";
import { SiteNav } from "@/components/site/SiteNav";
import { SiteFooter } from "@/components/site/SiteFooter";

const SECTIONS = [
  {
    label: "Guides",
    items: [
      { label: "Setup Guide", path: "guides/SETUP_GUIDE.md" },
      { label: "Developer Guide", path: "guides/DEVELOPER_GUIDE.md" },
      { label: "Enterprise Setup", path: "guides/ENTERPRISE_SETUP_GUIDE.md" },
      { label: "Connectors", path: "guides/CONNECTORS_GUIDE.md" },
      { label: "Refinery Proxy", path: "guides/refinery_proxy_setup.md" },
      { label: "Sombra Guide", path: "guides/SOMBRA_GUIDE.md" },
      { label: "Testing Guide", path: "guides/TESTING_GUIDE.md" },
      { label: "Release Guide", path: "guides/RELEASE_GUIDE.md" },
    ],
  },
  {
    label: "API",
    items: [
      { label: "API Reference", path: "api/API_REFERENCE.md" },
    ],
  },
  {
    label: "Reference",
    items: [
      { label: "Architecture", path: "reference/ARCHITECTURE.md" },
      { label: "PII Detection", path: "reference/PII_DETECTION.md" },
      { label: "FAQ", path: "reference/FAQ.md" },
      { label: "Product Context", path: "reference/PRODUCT_CONTEXT.md" },
      { label: "EU Sovereign Pack", path: "reference/EU_SOVEREIGN_PACK_V1.md" },
      { label: "Training Program", path: "reference/TRAINING_PROGRAM.md" },
    ],
  },
  {
    label: "Compliance & Security",
    items: [
      { label: "GDPR / Privacy by Design", path: "compliance/GDPR_PRIVACY_BY_DESIGN.md" },
      { label: "Security Model", path: "security/SECURITY_MODEL.md" },
    ],
  },
  {
    label: "Other",
    items: [
      { label: "Benchmarks", path: "benchmarks/privacy_filter_eval.md" },
      { label: "Secret Management", path: "SECRETS.md" },
      { label: "Blog: Zero-Egress Supply Chain", path: "blog/zero-egress-supply-chain.md" },
    ],
  },
];

const DEFAULT_DOC = "guides/SETUP_GUIDE.md";

export default function Docs() {
  const { "*": docPath } = useParams();
  const navigate = useNavigate();
  const activePath = docPath || DEFAULT_DOC;

  const [content, setContent] = useState<string | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    setContent(null);
    setError(false);
    fetch(`/docs/${activePath}`)
      .then((r) => {
        if (!r.ok) throw new Error("not found");
        return r.text();
      })
      .then(setContent)
      .catch(() => setError(true));
  }, [activePath]);

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col">
      <SiteNav />
      <div className="flex flex-1 pt-16">
        {/* Sidebar */}
        <aside className="hidden md:flex flex-col w-60 shrink-0 border-r border-border px-4 py-8 sticky top-16 h-[calc(100vh-4rem)] overflow-y-auto">
          {SECTIONS.map((section) => (
            <div key={section.label} className="mb-6">
              <p className="text-[11px] font-semibold uppercase tracking-widest text-muted-foreground mb-2">
                {section.label}
              </p>
              <ul className="space-y-0.5">
                {section.items.map((item) => (
                  <li key={item.path}>
                    <Link
                      to={`/docs/${item.path}`}
                      className={`block rounded px-2 py-1 text-[13px] transition-colors ${
                        activePath === item.path
                          ? "bg-primary/10 text-primary font-medium"
                          : "text-muted-foreground hover:text-foreground hover:bg-surface"
                      }`}
                    >
                      {item.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </aside>

        {/* Content */}
        <main className="flex-1 px-6 md:px-12 py-10 max-w-4xl">
          {/* Mobile doc picker */}
          <div className="md:hidden mb-6">
            <select
              className="w-full rounded border border-border bg-background text-foreground text-sm px-3 py-2"
              value={activePath}
              onChange={(e) => navigate(`/docs/${e.target.value}`)}
            >
              {SECTIONS.map((section) =>
                section.items.map((item) => (
                  <option key={item.path} value={item.path}>
                    {section.label} — {item.label}
                  </option>
                ))
              )}
            </select>
          </div>

          {error && (
            <p className="text-destructive">Could not load document.</p>
          )}
          {!content && !error && (
            <p className="text-muted-foreground text-sm">Loading…</p>
          )}
          {content && (
            <article className="prose prose-invert max-w-none prose-headings:font-mono prose-code:text-primary prose-pre:bg-surface prose-pre:border prose-pre:border-border">
              <ReactMarkdown>{content}</ReactMarkdown>
            </article>
          )}
        </main>
      </div>
      <SiteFooter />
    </div>
  );
}
