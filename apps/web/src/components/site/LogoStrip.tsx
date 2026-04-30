const LOGOS = [
  "Helvetia",
  "Banque Sovrein",
  "Lumen Health",
  "Atlas Legal",
  "Northwind Bank",
  "Civitas.gov",
  "Polaris AI",
  "Meridian Group",
];

export const LogoStrip = () => (
  <section className="border-y border-border bg-surface/40">
    <div className="container-page py-10">
      <p className="text-center mono-label">
        Trusted by security and compliance teams in regulated industries
      </p>
      <div className="mt-6 marquee-mask overflow-hidden">
        <div className="flex w-max animate-scroll-x gap-14 pr-14">
          {[...LOGOS, ...LOGOS].map((name, i) => (
            <span
              key={i}
              className="whitespace-nowrap font-mono text-sm font-semibold uppercase tracking-[0.16em] text-muted-foreground/70"
            >
              {name}
            </span>
          ))}
        </div>
      </div>
    </div>
  </section>
);
