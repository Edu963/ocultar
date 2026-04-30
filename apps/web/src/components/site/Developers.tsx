import { useState } from "react";
import { Check, Copy } from "lucide-react";
import { SectionHeader } from "./ProblemSection";

const BEFORE = `const response = await openai.chat.completions.create({
  model: "gpt-4o",
  messages: [{ role: "user", content: userMessage }],
});`;

const AFTER = `const response = await openai.chat.completions.create({
  baseURL: "https://ocultar.your-vpc/proxy/openai", // ← only change
  model: "gpt-4o",
  messages: [{ role: "user", content: userMessage }],
});`;

export const Developers = () => (
  <section id="developers" className="section border-t border-border">
    <div className="container-page">
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-14 items-start">
        <div className="lg:col-span-5">
          <SectionHeader
            eyebrow="Developers"
            title="One line. Drop-in replacement for the OpenAI SDK."
            description="OCULTAR runs as a transparent reverse proxy. Point your existing AI calls at it. Detection, vaulting and re-hydration happen automatically."
          />
          <ul className="mt-8 flex flex-col gap-3">
            {[
              "Zero changes to application logic",
              "Works with any OpenAI-compatible SDK",
              "Deploy as sidecar, central gateway or K8s operator",
              "Streaming, tools and function calling supported",
            ].map((item) => (
              <li key={item} className="flex items-start gap-3 text-[14px]">
                <span className="mt-0.5 flex h-5 w-5 items-center justify-center rounded-full border border-primary/30 bg-primary/10">
                  <Check className="h-3 w-3 text-primary" />
                </span>
                <span className="text-muted-foreground">{item}</span>
              </li>
            ))}
          </ul>
        </div>

        <div className="lg:col-span-7 flex flex-col gap-4">
          <CodeBlock
            label="Before — PII reaches OpenAI"
            tone="warn"
            code={BEFORE}
          />
          <CodeBlock label="After — Zero-egress" tone="ok" code={AFTER} highlight />
        </div>
      </div>
    </div>
  </section>
);

const CodeBlock = ({
  label,
  tone,
  code,
  highlight,
}: {
  label: string;
  tone: "warn" | "ok";
  code: string;
  highlight?: boolean;
}) => {
  const [copied, setCopied] = useState(false);
  return (
    <div
      className={`rounded-xl overflow-hidden border ${
        tone === "ok" ? "border-primary/40 shadow-glow" : "border-border opacity-80"
      } bg-surface/80 backdrop-blur-sm`}
    >
      <div className="flex items-center justify-between border-b border-border px-4 py-2.5">
        <div className="flex items-center gap-2">
          <span
            className={`h-1.5 w-1.5 rounded-full ${
              tone === "ok" ? "bg-primary animate-pulse-dot" : "bg-destructive/70"
            }`}
          />
          <span className="font-mono text-[10px] uppercase tracking-[0.22em] text-muted-foreground">
            {label}
          </span>
        </div>
        <button
          onClick={() => {
            navigator.clipboard?.writeText(code);
            setCopied(true);
            setTimeout(() => setCopied(false), 1500);
          }}
          className="inline-flex items-center gap-1.5 rounded text-[10px] font-mono text-muted-foreground hover:text-foreground transition-colors"
        >
          {copied ? (
            <>
              <Check className="h-3 w-3 text-primary" /> Copied
            </>
          ) : (
            <>
              <Copy className="h-3 w-3" /> Copy
            </>
          )}
        </button>
      </div>
      <pre className="p-5 text-[12.5px] leading-relaxed font-mono text-muted-foreground overflow-x-auto">
        {highlight ? (
          <code>
            {`const response = await openai.chat.completions.create({\n`}
            <span className="text-primary">{`  baseURL: "https://ocultar.your-vpc/proxy/openai", // ← only change\n`}</span>
            {`  model: "gpt-4o",\n  messages: [{ role: "user", content: userMessage }],\n});`}
          </code>
        ) : (
          <code>{code}</code>
        )}
      </pre>
    </div>
  );
};
