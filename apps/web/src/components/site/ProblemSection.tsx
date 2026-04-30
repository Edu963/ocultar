import { AlertTriangle, FileWarning, Clock, EyeOff } from "lucide-react";

const ITEMS = [
  {
    icon: AlertTriangle,
    label: "Exposed",
    body: "Every call to OpenAI, Gemini or Claude carries real names, IDs and health records. Their breach becomes your breach.",
  },
  {
    icon: FileWarning,
    label: "Non-compliant",
    body: "GDPR Art. 28, HIPAA BAAs and SOC 2 CC6.1 prohibit sending PII to third parties without explicit controls. Most teams already are.",
  },
  {
    icon: Clock,
    label: "Blocked",
    body: "Security review kills AI projects for 6–18 months. Engineers ship workarounds, compliance says no, the productivity win never lands.",
  },
  {
    icon: EyeOff,
    label: "Invisible",
    body: "No audit trail. No way to prove what data touched which model. No evidence when a regulator asks.",
  },
];

export const ProblemSection = () => (
  <section id="platform" className="section relative">
    <div className="container-page">
      <SectionHeader
        eyebrow="The problem"
        title="Enterprise AI today is a compliance time bomb."
        description="Four structural failure modes shared by almost every AI deployment we audit."
      />
      <div className="mt-14 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {ITEMS.map((item) => (
          <article
            key={item.label}
            className="card-surface card-surface-hover group relative p-6 flex flex-col gap-4"
          >
            <div className="flex items-center gap-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-md border border-border bg-surface-elevated text-primary">
                <item.icon className="h-4 w-4" />
              </div>
              <span className="mono-label">{item.label}</span>
            </div>
            <p className="text-[14px] leading-relaxed text-muted-foreground">
              {item.body}
            </p>
          </article>
        ))}
      </div>
    </div>
  </section>
);

export const SectionHeader = ({
  eyebrow,
  title,
  description,
}: {
  eyebrow: string;
  title: string;
  description?: string;
}) => (
  <div className="flex flex-col items-start gap-4 max-w-2xl">
    <span className="eyebrow eyebrow-accent">
      <span className="h-1.5 w-1.5 rounded-full bg-primary" />
      {eyebrow}
    </span>
    <h2 className="text-balance text-[32px] sm:text-[40px] font-semibold leading-[1.1] tracking-tight text-foreground">
      {title}
    </h2>
    {description && (
      <p className="text-[15px] leading-relaxed text-muted-foreground">
        {description}
      </p>
    )}
  </div>
);
