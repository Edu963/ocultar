import { SiteNav } from "@/components/site/SiteNav";
import { Hero } from "@/components/site/Hero";
import { ProblemSection } from "@/components/site/ProblemSection";
import { Architecture } from "@/components/site/Architecture";
import { Developers } from "@/components/site/Developers";
import { Compliance } from "@/components/site/Compliance";
import { FinalCTA } from "@/components/site/FinalCTA";
import { SiteFooter } from "@/components/site/SiteFooter";

const Index = () => {
  return (
    <main className="min-h-screen bg-background text-foreground">
      <SiteNav />
      <Hero />
      <ProblemSection />
      <Architecture />
      <Developers />
      <Compliance />
      <FinalCTA />
      <SiteFooter />
    </main>
  );
};

export default Index;
