package main

import (
	"encoding/json"
	"flag"
	"fmt"
	htmltmpl "html/template"
	"log"
	"os"
	"strings"
	texttmpl "text/template"
	"time"

	"github.com/Edu963/ocultar/pkg/audit"
	"github.com/google/uuid"
	"os/exec"
	"runtime"
)

const reportVersion = "3.1"
const engineVersion = "v1.14"

// reportMeta holds non-risk metadata for the report header.
type reportMeta struct {
	ReportID           string
	GeneratedAt        string
	DatasetScope       string
	MethodologyVersion string
	EngineVersion      string
	TotalRecords       int
}

// fullReport combines metadata and risk results for template rendering.
type fullReport struct {
	Meta   reportMeta
	Risk   audit.RiskReport
	Before scenarioSummary
	After  scenarioSummary
}

// scenarioSummary represents a Before or After OCULTAR simulation snapshot.
type scenarioSummary struct {
	Label       string
	RiskLevel   string
	RiskScore   string
	VaRRange    string // e.g. "€12,000 – €87,000 (estimated)"
	AIStatus    string
	Description string
}

func buildMeta(datasetPath string, total int) reportMeta {
	return reportMeta{
		ReportID:           strings.ToUpper(uuid.New().String()[:8]),
		GeneratedAt:        time.Now().UTC().Format("02 January 2006, 15:04 UTC"),
		DatasetScope:       datasetPath,
		MethodologyVersion: reportVersion,
		EngineVersion:      engineVersion,
		TotalRecords:       total,
	}
}

func buildScenarios(r audit.RiskReport) (scenarioSummary, scenarioSummary) {
	before := scenarioSummary{
		Label:     "Scenario A — Current State (No Protection)",
		RiskLevel: r.OverallRiskLevel,
		RiskScore: fmt.Sprintf("%.1f / 10", r.OverallRiskScore),
		VaRRange: fmt.Sprintf("€%.0f – €%.0f (estimated)",
			r.Exposure.VaRMin, r.Exposure.VaRMax),
		AIStatus:    r.AI.Status,
		Description: "The raw dataset as-is, transmitted directly to an LLM API or stored in a vector database. All PII fields are exposed in plaintext.",
	}

	// Project the "After OCULTAR" state with appropriate uncertainty language.
	// Direct identifiers are removed; residual risk is not mathematically zero.
	afterScoreMin := r.OverallRiskScore * 0.05 // ~95% estimated risk reduction
	afterScoreMax := r.OverallRiskScore * 0.15 // conservative upper bound
	afterVaRMin := r.Exposure.VaRMin * 0.02    // residual operational baseline (low)
	afterVaRMax := r.Exposure.VaRMin * 0.08    // residual operational baseline (high)

	after := scenarioSummary{
		Label:     "Scenario B — After OCULTAR Processing",
		RiskLevel: audit.ScoreToLabel(afterScoreMax),
		RiskScore: fmt.Sprintf("%.1f – %.1f / 10 (projected, subject to contextual factors)", afterScoreMin, afterScoreMax),
		VaRRange: fmt.Sprintf("€%.0f – €%.0f (projected residual)",
			afterVaRMin, afterVaRMax),
		AIStatus: "ALLOW",
		Description: "After OCULTAR tokenization and format-preserving encryption pipeline. " +
			"Direct identifiers are removed and re-identification risk is significantly reduced, though not mathematically eliminated. " +
			"Dataset achieves strong pseudonymization; residual risk reflects standard operational overhead.",
	}
	return before, after
}

const mdTemplate = `# OCULTAR Data Risk Assessment Report

> **CONFIDENTIAL — For Authorised Recipients Only**
> This report constitutes a technical risk and privacy assessment based on automated analysis. It is informational in nature and does not constitute legal advice or a regulatory compliance determination. Distribution is restricted to named stakeholders.

---

## Report Metadata

| Field | Value |
| :--- | :--- |
| **Report ID** | OCU-{{.Meta.ReportID}} |
| **Generated** | {{.Meta.GeneratedAt}} |
| **Dataset Scope** | ` + "`" + `{{.Meta.DatasetScope}}` + "`" + ` |
| **Records Analysed** | {{.Meta.TotalRecords}} |
| **Methodology Version** | v{{.Meta.MethodologyVersion}} |
| **Engine** | OCULTAR Enterprise {{.Meta.EngineVersion}} |

---

## Executive Risk Summary

{{if eq .Risk.OverallRiskLevel "CRITICAL"}}> [!CAUTION]
> **Overall Risk Level: {{.Risk.OverallRiskLevel}} ({{printf "%.1f" .Risk.OverallRiskScore}}/10)**
> **Compliance Likelihood: {{if .Risk.IsGDPRPseudonymized}}✅ Meets Common Pseudonymization Thresholds{{else}}⚠️ High Non-Compliance Likelihood (External Processing Scenarios){{end}}**{{end}}
{{if eq .Risk.OverallRiskLevel "HIGH"}}> [!WARNING]
> **Overall Risk Level: {{.Risk.OverallRiskLevel}} ({{printf "%.1f" .Risk.OverallRiskScore}}/10)**
> **Compliance Likelihood: {{if .Risk.IsGDPRPseudonymized}}✅ Meets Common Pseudonymization Thresholds{{else}}⚠️ High Non-Compliance Likelihood (External Processing Scenarios){{end}}**{{end}}
{{if eq .Risk.OverallRiskLevel "MEDIUM"}}> [!IMPORTANT]
> **Overall Risk Level: {{.Risk.OverallRiskLevel}} ({{printf "%.1f" .Risk.OverallRiskScore}}/10)**
> **Compliance Likelihood: {{if .Risk.IsGDPRPseudonymized}}✅ Meets Common Pseudonymization Thresholds{{else}}⚠️ Elevated Risk — Review Recommended{{end}}**{{end}}
{{if eq .Risk.OverallRiskLevel "LOW"}}> [!NOTE]
> **Overall Risk Level: {{.Risk.OverallRiskLevel}} ({{printf "%.1f" .Risk.OverallRiskScore}}/10)**
> **Compliance Likelihood: ✅ Meets Common Pseudonymization Thresholds**{{end}}

The dataset identified in this report contains an estimated **{{.Risk.ViolatingRecords}} records** that fall below commonly cited EU pseudonymization thresholds. In its current state, this data **{{if .Risk.IsGDPRPseudonymized}}satisfies commonly cited thresholds for use{{else}}presents elevated risk for use{{end}} with external AI systems and LLM APIs** without prior sanitisation.

The estimated financial exposure associated with unauthorised disclosure of this dataset is in the range of **€{{printf "%.0f" .Risk.Exposure.VaRMin}} – €{{printf "%.0f" .Risk.Exposure.VaRMax}}** (simulated estimate based on industry breach benchmarks). This range encompasses regulatory exposure modelling, operational incident response costs, and a risk multiplier derived from the dataset's anonymization profile. Actual impact may vary significantly based on enforcement context and organisational factors.

---

## Risk Scorecard

| Category | Score | Level | Business Implication |
| :--- | :---: | :---: | :--- |
| **Identifiability Risk** | {{printf "%.1f" .Risk.Identifiability.Score}}/10 | {{.Risk.Identifiability.Label}} | {{.Risk.Identifiability.Implication}} |
| **Financial Sensitivity** | {{printf "%.1f" .Risk.FinancialSensitivity.Score}}/10 | {{.Risk.FinancialSensitivity.Label}} | {{.Risk.FinancialSensitivity.Implication}} |
| **Re-identification Risk** | {{printf "%.1f" .Risk.ReidentificationRisk.Score}}/10 | {{.Risk.ReidentificationRisk.Label}} | {{.Risk.ReidentificationRisk.Implication}} |
| **Compliance Readiness** | {{printf "%.1f" .Risk.ComplianceReadiness.Score}}/10 | {{.Risk.ComplianceReadiness.Label}} | {{.Risk.ComplianceReadiness.Implication}} |
| **Overall** | **{{printf "%.1f" .Risk.OverallRiskScore}}/10** | **{{.Risk.OverallRiskLevel}}** | Weighted composite score (Identifiability 35%, Financial 25%, Re-id 25%, Compliance 15%) |

---

## Technical Metrics — Interpreted

### K-Anonymity
**Raw Score:** {{.Risk.KAnonymity}}

{{.Risk.KAnonymityInterpretation}}

> **Industry Benchmark:** Common industry frameworks suggest a minimum K-score of 3–5 for basic pseudonymization. This is a technical benchmark, not a mandatory legal threshold—contextual factors, processing purpose, and applicable exemptions determine actual compliance obligations.

### L-Diversity
**Raw Score:** {{.Risk.LDiversity}}

{{.Risk.LDiversityInterpretation}}

> **Industry Benchmark:** An L-Diversity score of ≥2 is commonly recommended to mitigate homogeneity attacks, as referenced in ISO/IEC 29101 (Privacy Architecture Framework). This is an industry guideline; applicable legal thresholds depend on jurisdictional context.

---

## Financial Exposure Model

The **Value at Risk (VaR)** range below is computed using a three-component methodology anchored to industry breach cost benchmarks. All figures are **simulated estimates** and should not be interpreted as predicted fine amounts or contractual commitments.

### VaR Components

| Component | Methodology | Min Estimate | Max Estimate |
| :--- | :--- | ---: | ---: |
| **Regulatory Exposure** | Dataset Risk Score ({{printf "%.2f" .Risk.DatasetRiskScore}}) × anchor range (€10,000–€100,000) | **€{{printf "%.0f" .Risk.Exposure.RegulatoryExposureMin}}** | **€{{printf "%.0f" .Risk.Exposure.RegulatoryExposureMax}}** |
| **Operational Cost** | €100–€300 × {{.Risk.TotalRecords}} records (industry benchmark range) | **€{{printf "%.0f" .Risk.Exposure.OperationalCostMin}}** | **€{{printf "%.0f" .Risk.Exposure.OperationalCostMax}}** |
| **Risk Multiplier** | Derived from K={{.Risk.KAnonymity}}, L={{.Risk.LDiversity}} profile | **{{printf "%.1f" .Risk.Exposure.RiskMultiplierMin}}×** | **{{printf "%.1f" .Risk.Exposure.RiskMultiplierMax}}×** |
| | | | |
| **Total Value at Risk (Estimated)** | | **€{{printf "%.0f" .Risk.Exposure.VaRMin}}** | **€{{printf "%.0f" .Risk.Exposure.VaRMax}}** |

> **Assumptions & Methodology Note:**
> {{.Risk.Exposure.AssumptionsNote}}

---

## AI & LLM Exposure Assessment

### Decision: {{.Risk.AI.Status}}

| Parameter | Assessment |
| :--- | :--- |
| **External LLM API Safety** | {{.Risk.AI.LLMExposure}} risk |
| **Internal Copilot Safety** | {{if eq .Risk.AI.Status "ALLOW"}}✅ Permitted with monitoring{{else if eq .Risk.AI.Status "SANITIZE_FIRST"}}⚠️ Permitted after OCULTAR processing{{else}}🚫 Not recommended without sanitisation{{end}} |
| **Vector DB / RAG Indexing** | {{if .Risk.AI.RAGSafe}}✅ Estimated safe for indexing{{else}}🚫 Not recommended without prior processing{{end}} |

**RAG & Vector Database Guidance:**
{{.Risk.AI.RAGGuidance}}

**Recommended Action:**
{{.Risk.AI.Recommendation}}

---

## Before / After Simulation

This section demonstrates the modelled impact of the OCULTAR Enterprise pipeline on your dataset's risk profile. Figures are projected estimates based on typical processing outcomes.

| Metric | {{.Before.Label}} | {{.After.Label}} |
| :--- | :--- | :--- |
| **Risk Level** | 🔴 {{.Before.RiskLevel}} | 🟢 {{.After.RiskLevel}} |
| **Risk Score** | {{.Before.RiskScore}} | {{.After.RiskScore}} |
| **Financial Exposure (VaR)** | {{.Before.VaRRange}} | {{.After.VaRRange}} |
| **AI / LLM Status** | {{.Before.AIStatus}} | {{.After.AIStatus}} |

**What changes:**
- **Before:** {{.Before.Description}}
- **After:** {{.After.Description}}

> The projected risk and VaR reductions above are simulated estimates based on OCULTAR's tokenization model. Direct identifiers are removed and re-identification risk is significantly reduced, though not mathematically eliminated. Residual exposure reflects standard operational overhead associated with any data processing activity.

---

## Assumptions

The following assumptions underpin all quantitative estimates in this report:

| Assumption | Value / Range | Basis |
| :--- | :--- | :--- |
| **Regulatory anchor (low)** | €10,000 | Simulation baseline; not a per-record fine |
| **Regulatory anchor (high)** | €100,000 | Simulation ceiling; not a legal determination |
| **Operational cost per record** | €100–€300 | Industry breach cost studies (range) |
| **Risk multiplier at K=1** | 1.5×–2.0× | Maximum risk profile |
| **Risk multiplier at K<3** | 1.3×–1.5× | Elevated risk profile |
| **Risk multiplier at K≥3** | 1.0×–1.2× | Baseline risk profile |
| **Dataset Risk Score** | {{printf "%.2f" .Risk.DatasetRiskScore}} (0.0–1.0) | Normalised from overall weighted score |
| **Pseudonymization threshold** | K≥3, L≥2 | Common industry benchmark |

All estimates are subject to contextual factors including jurisdiction, data controller agreements, processing purpose, security controls in place, and enforcement environment.

---

## Remediation Plan

{{.Risk.Recommendation}}

---

## Appendix: Methodology & Standards

This report applies the following analytical frameworks:

- **K-Anonymity** (Sweeney, 2002): Minimum group size for quasi-identifier equivalence classes.
- **L-Diversity** (Machanavajjhala et al., 2006): Sensitive attribute diversity within equivalence classes.
- **GDPR Article 5(1)(f)**: Integrity and confidentiality principle (reference context only).
- **GDPR Article 83**: Administrative fine schedule — cited as regulatory context, not a determination of liability.
- **Industry Breach Cost Benchmarks**: Operational cost ranges sourced from published annual breach cost studies (range: €100–€300/record).
- **ISO/IEC 29101**: Privacy architecture framework for ICT systems.

> This report was generated automatically by OCULTAR Enterprise {{.Meta.EngineVersion}}. Findings are based on the dataset provided at time of analysis and constitute a technical assessment only. Engage qualified legal counsel for regulatory compliance determinations.

---

*OCULTAR Enterprise {{.Meta.EngineVersion}} | Methodology v{{.Meta.MethodologyVersion}} | Report ID: OCU-{{.Meta.ReportID}}*
*Generated: {{.Meta.GeneratedAt}}*
`

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>OCULTAR Risk Report — OCU-{{.Meta.ReportID}}</title>
<style>
  @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
  :root {
    --critical: #dc2626; --high: #ea580c; --medium: #d97706; --low: #16a34a;
    --bg: #f8fafc; --surface: #ffffff; --border: #e2e8f0;
    --text: #0f172a; --muted: #64748b; --accent: #1e40af;
  }
  body { font-family: 'Inter', sans-serif; background: var(--bg); color: var(--text); font-size: 14px; line-height: 1.6; }
  .container { max-width: 960px; margin: 0 auto; padding: 40px 24px; }
  
  /* Header */
  .report-header { background: var(--text); color: white; padding: 40px; border-radius: 12px; margin-bottom: 32px; position: relative; overflow: hidden; }
  .report-header::before { content: ''; position: absolute; top: -60px; right: -60px; width: 240px; height: 240px; background: rgba(255,255,255,0.04); border-radius: 50%; }
  .report-header h1 { font-size: 22px; font-weight: 700; letter-spacing: -0.5px; margin-bottom: 4px; }
  .report-header .subtitle { font-size: 13px; opacity: 0.6; margin-bottom: 24px; }
  .meta-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; }
  .meta-item label { display: block; font-size: 11px; text-transform: uppercase; letter-spacing: 1px; opacity: 0.5; margin-bottom: 2px; }
  .meta-item span { font-size: 13px; font-weight: 500; }
  
  /* Risk Banner */
  .risk-banner { border-radius: 10px; padding: 24px 28px; margin-bottom: 28px; display: flex; align-items: center; gap: 20px; }
  .risk-banner.CRITICAL { background: #fef2f2; border: 1px solid #fecaca; }
  .risk-banner.HIGH { background: #fff7ed; border: 1px solid #fed7aa; }
  .risk-banner.MEDIUM { background: #fffbeb; border: 1px solid #fde68a; }
  .risk-banner.LOW { background: #f0fdf4; border: 1px solid #bbf7d0; }
  .risk-dial { width: 72px; height: 72px; border-radius: 50%; display: flex; align-items: center; justify-content: center; flex-shrink: 0; font-size: 18px; font-weight: 700; }
  .CRITICAL .risk-dial { background: var(--critical); color: white; }
  .HIGH .risk-dial { background: var(--high); color: white; }
  .MEDIUM .risk-dial { background: var(--medium); color: white; }
  .LOW .risk-dial { background: var(--low); color: white; }
  .risk-banner-text h2 { font-size: 18px; font-weight: 700; margin-bottom: 4px; }
  .risk-banner-text p { font-size: 13px; color: var(--muted); }
  
  /* Sections */
  .section { background: var(--surface); border: 1px solid var(--border); border-radius: 10px; padding: 28px; margin-bottom: 20px; }
  .section h2 { font-size: 15px; font-weight: 700; margin-bottom: 16px; padding-bottom: 10px; border-bottom: 1px solid var(--border); color: var(--accent); text-transform: uppercase; letter-spacing: 0.5px; }
  .section h3 { font-size: 13px; font-weight: 600; margin: 16px 0 8px; }
  
  /* Tables */
  table { width: 100%; border-collapse: collapse; font-size: 13px; }
  th { text-align: left; padding: 10px 12px; background: var(--bg); font-weight: 600; font-size: 11px; text-transform: uppercase; letter-spacing: 0.5px; color: var(--muted); border-bottom: 1px solid var(--border); }
  td { padding: 10px 12px; border-bottom: 1px solid #f1f5f9; vertical-align: top; }
  tr:last-child td { border-bottom: none; }
  
  /* Badges */
  .badge { display: inline-block; padding: 2px 9px; border-radius: 100px; font-size: 11px; font-weight: 600; letter-spacing: 0.5px; }
  .badge-critical { background: #fef2f2; color: var(--critical); }
  .badge-high { background: #fff7ed; color: var(--high); }
  .badge-medium { background: #fffbeb; color: var(--medium); }
  .badge-low { background: #f0fdf4; color: var(--low); }
  .badge-block { background: #fef2f2; color: var(--critical); }
  .badge-sanitize_first { background: #fffbeb; color: var(--medium); }
  .badge-allow { background: #f0fdf4; color: var(--low); }
  
  /* Score bars */
  .score-bar-wrap { display: flex; align-items: center; gap: 10px; }
  .score-bar { height: 6px; border-radius: 3px; background: #e2e8f0; flex: 1; }
  .score-bar-fill { height: 100%; border-radius: 3px; }
  .fill-critical { background: var(--critical); }
  .fill-high { background: var(--high); }
  .fill-medium { background: var(--medium); }
  .fill-low { background: var(--low); }
  
  /* VaR range display */
  .var-range { display: flex; gap: 12px; align-items: center; }
  .var-pillar { flex: 1; background: var(--bg); border: 1px solid var(--border); border-radius: 8px; padding: 12px 16px; }
  .var-pillar label { display: block; font-size: 10px; text-transform: uppercase; letter-spacing: 0.5px; color: var(--muted); margin-bottom: 4px; font-weight: 600; }
  .var-pillar .amount { font-size: 20px; font-weight: 700; color: var(--text); }
  .var-pillar.min .amount { color: var(--medium); }
  .var-pillar.max .amount { color: var(--critical); }
  .var-arrow { font-size: 20px; color: var(--muted); }
  .var-total-row td { font-weight: 700; background: #f8fafc; }
  .assumptions-box { background: #f8fafc; border-left: 3px solid var(--accent); padding: 14px 16px; border-radius: 0 6px 6px 0; margin-top: 16px; font-size: 12px; color: var(--muted); line-height: 1.6; }
  
  /* Scenario comparison */
  .scenario-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
  .scenario-card { border-radius: 8px; padding: 20px; }
  .scenario-before { background: #fef2f2; border: 1px solid #fecaca; }
  .scenario-after { background: #f0fdf4; border: 1px solid #bbf7d0; }
  .scenario-card h3 { font-size: 12px; text-transform: uppercase; letter-spacing: 0.5px; font-weight: 600; margin-bottom: 14px; }
  .scenario-stat { margin-bottom: 10px; }
  .scenario-stat label { display: block; font-size: 11px; color: var(--muted); margin-bottom: 2px; }
  .scenario-stat span { font-size: 14px; font-weight: 600; }
  .scenario-before h3 { color: var(--critical); }
  .scenario-after h3 { color: var(--low); }
  
  /* Remediation steps */
  .remediation-list { counter-reset: step-counter; list-style: none; }
  .step { display: flex; gap: 12px; margin-bottom: 14px; position: relative; }
  .step-num { width: 28px; height: 28px; border-radius: 50%; background: var(--accent); color: white; font-size: 12px; font-weight: 700; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
  .numbered-step { counter-increment: step-counter; }
  .numbered-step .step-num::before { content: counter(step-counter); }
  .step-body strong { display: block; font-size: 13px; font-weight: 600; margin-bottom: 2px; }
  .step-body p { font-size: 12px; color: var(--muted); }

  /* Assumptions table */
  .assumptions-table { font-size: 12px; }
  .assumptions-table td:first-child { font-weight: 600; white-space: nowrap; }
  .assumptions-table td:nth-child(2) { font-family: monospace; }
  
  /* Legal disclaimer banner */
  .disclaimer { background: #fffbeb; border: 1px solid #fde68a; border-radius: 6px; padding: 10px 14px; font-size: 11px; color: #92400e; margin-bottom: 20px; }
  
  /* Print / PDF */
  .print-btn { position: fixed; bottom: 28px; right: 28px; background: var(--accent); color: white; border: none; border-radius: 8px; padding: 12px 20px; font-size: 13px; font-weight: 600; cursor: pointer; box-shadow: 0 4px 12px rgba(30,64,175,0.3); display: flex; align-items: center; gap: 8px; }
  .print-btn:hover { background: #1d4ed8; }
  @media print {
    .print-btn { display: none; }
    body { background: white; }
    .container { padding: 20px; }
  }
  
  blockquote { background: #f8fafc; border-left: 3px solid var(--accent); padding: 10px 16px; border-radius: 0 6px 6px 0; margin: 12px 0; font-size: 12px; color: var(--muted); }
  p { margin-bottom: 10px; font-size: 13px; }
  code { background: #f1f5f9; padding: 2px 6px; border-radius: 4px; font-size: 12px; }
  .confidential { background: #fef9c3; border: 1px solid #fde68a; border-radius: 6px; padding: 8px 14px; font-size: 12px; color: #92400e; margin-bottom: 16px; text-align: center; font-weight: 500; }
  .footer { text-align: center; margin-top: 40px; font-size: 11px; color: var(--muted); }
</style>
</head>
<body>
<div class="container">

  <div class="confidential">⚠️ CONFIDENTIAL — For Authorised Recipients Only. Report ID: OCU-{{.Meta.ReportID}}</div>
  
  <div class="disclaimer">
    <strong>Important Notice:</strong> This report presents a technical risk assessment based on automated analysis. 
    All financial figures are simulated estimates based on industry benchmarks and do not constitute legal advice, 
    regulatory determinations, or predictions of actual fines or penalties. Engage qualified legal counsel for compliance determinations.
  </div>

  <!-- Header -->
  <div class="report-header">
    <h1>OCULTAR Data Risk Assessment</h1>
    <div class="subtitle">Enterprise Compliance &amp; AI Exposure Report — Technical Assessment</div>
    <div class="meta-grid">
      <div class="meta-item"><label>Report ID</label><span>OCU-{{.Meta.ReportID}}</span></div>
      <div class="meta-item"><label>Generated</label><span>{{.Meta.GeneratedAt}}</span></div>
      <div class="meta-item"><label>Methodology</label><span>v{{.Meta.MethodologyVersion}}</span></div>
      <div class="meta-item"><label>Dataset</label><span>{{.Meta.DatasetScope}}</span></div>
      <div class="meta-item"><label>Records Analysed</label><span>{{.Meta.TotalRecords}}</span></div>
      <div class="meta-item"><label>Engine</label><span>{{.Meta.EngineVersion}}</span></div>
    </div>
  </div>

  <!-- Executive Risk Banner -->
  <div class="risk-banner {{.Risk.OverallRiskLevel}}">
    <div class="risk-dial">{{printf "%.1f" .Risk.OverallRiskScore}}</div>
    <div class="risk-banner-text">
      {{if eq .Risk.ViolatingRecords 0}}
      <h2>High Compliance Readiness — ✅ Meets Common Thresholds</h2>
      <p>The analyzed dataset satisfies commonly cited pseudonymization benchmarks with zero identified violations.</p>
      {{else}}
      <h2>{{.Risk.OverallRiskLevel}} Risk — ⚠️ Elevated Non-Compliance Likelihood</h2>
      <p>An estimated {{.Risk.ViolatingRecords}} of {{.Risk.TotalRecords}} records fall below commonly cited pseudonymization benchmarks.</p>
      {{end}}
      <p style="margin-top:6px">Estimated financial exposure range: <strong>€{{printf "%.0f" .Risk.Exposure.VaRMin}} – €{{printf "%.0f" .Risk.Exposure.VaRMax}}</strong> (simulated — see Assumptions).</p>
      <p style="margin-top:6px; font-size:11px; opacity:0.7">This dataset <strong>{{if eq .Risk.ViolatingRecords 0}}satisfies common pseudonymization thresholds for use{{else}}presents elevated technical risk for use{{end}}</strong> with external AI APIs without prior sanitisation.</p>
    </div>
  </div>

  <!-- Risk Scorecard -->
  <div class="section">
    <h2>Risk Scorecard</h2>
    <table>
      <thead><tr><th>Category</th><th>Score</th><th>Level</th><th>Business Implication</th></tr></thead>
      <tbody>
        <tr>
          <td><strong>Identifiability Risk</strong></td>
          <td>
            <div class="score-bar-wrap">
              <span>{{printf "%.1f" .Risk.Identifiability.Score}}</span>
              <div class="score-bar"><div class="score-bar-fill fill-{{lower .Risk.Identifiability.Label}}" style="width:{{pct .Risk.Identifiability.Score}}%"></div></div>
            </div>
          </td>
          <td><span class="badge badge-{{lower .Risk.Identifiability.Label}}">{{.Risk.Identifiability.Label}}</span></td>
          <td>{{.Risk.Identifiability.Implication}}</td>
        </tr>
        <tr>
          <td><strong>Financial Sensitivity</strong></td>
          <td>
            <div class="score-bar-wrap">
              <span>{{printf "%.1f" .Risk.FinancialSensitivity.Score}}</span>
              <div class="score-bar"><div class="score-bar-fill fill-{{lower .Risk.FinancialSensitivity.Label}}" style="width:{{pct .Risk.FinancialSensitivity.Score}}%"></div></div>
            </div>
          </td>
          <td><span class="badge badge-{{lower .Risk.FinancialSensitivity.Label}}">{{.Risk.FinancialSensitivity.Label}}</span></td>
          <td>{{.Risk.FinancialSensitivity.Implication}}</td>
        </tr>
        <tr>
          <td><strong>Re-identification Risk</strong></td>
          <td>
            <div class="score-bar-wrap">
              <span>{{printf "%.1f" .Risk.ReidentificationRisk.Score}}</span>
              <div class="score-bar"><div class="score-bar-fill fill-{{lower .Risk.ReidentificationRisk.Label}}" style="width:{{pct .Risk.ReidentificationRisk.Score}}%"></div></div>
            </div>
          </td>
          <td><span class="badge badge-{{lower .Risk.ReidentificationRisk.Label}}">{{.Risk.ReidentificationRisk.Label}}</span></td>
          <td>{{.Risk.ReidentificationRisk.Implication}}</td>
        </tr>
        <tr>
          <td><strong>Compliance Readiness</strong></td>
          <td>
            <div class="score-bar-wrap">
              <span>{{printf "%.1f" .Risk.ComplianceReadiness.Score}}</span>
              <div class="score-bar"><div class="score-bar-fill fill-{{lower .Risk.ComplianceReadiness.Label}}" style="width:{{pct .Risk.ComplianceReadiness.Score}}%"></div></div>
            </div>
          </td>
          <td><span class="badge badge-{{lower .Risk.ComplianceReadiness.Label}}">{{.Risk.ComplianceReadiness.Label}}</span></td>
          <td>{{.Risk.ComplianceReadiness.Implication}}</td>
        </tr>
      </tbody>
    </table>
  </div>

  <!-- Technical Interpretation -->
  <div class="section">
    <h2>Technical Metrics — Interpreted</h2>
    <h3>K-Anonymity (Score: {{.Risk.KAnonymity}})</h3>
    <p>{{.Risk.KAnonymityInterpretation}}</p>
    <blockquote>Industry Benchmark: K≥3–5 is commonly recommended for basic pseudonymization. This is a technical guideline, not a mandatory legal threshold. Actual compliance obligations depend on processing context and applicable law.</blockquote>
    <h3>L-Diversity (Score: {{.Risk.LDiversity}})</h3>
    <p>{{.Risk.LDiversityInterpretation}}</p>
    <blockquote>Industry Benchmark: L≥2 is commonly recommended to mitigate homogeneity attacks, per ISO/IEC 29101. Applicable legal thresholds depend on jurisdictional context.</blockquote>
  </div>

  <!-- Financial Exposure -->
  <div class="section">
    <h2>Financial Exposure — Simulated VaR Range</h2>
    <p style="font-size:12px;color:var(--muted);margin-bottom:16px">All figures below are <strong>simulated estimates</strong> based on industry breach benchmarks and a risk-scored regulatory exposure model. They do not represent predicted fines, legal obligations, or accounting provisions.</p>
    
    <table>
      <thead><tr><th>Component</th><th>Methodology</th><th style="text-align:right">Min (€)</th><th style="text-align:right">Max (€)</th></tr></thead>
      <tbody>
        <tr><td><strong>Regulatory Exposure</strong></td><td>Risk score × anchor range (€10k–€100k)</td><td style="text-align:right">€{{printf "%.0f" .Risk.Exposure.RegulatoryExposureMin}}</td><td style="text-align:right">€{{printf "%.0f" .Risk.Exposure.RegulatoryExposureMax}}</td></tr>
        <tr><td><strong>Operational Cost</strong></td><td>€100–€300 × {{.Risk.TotalRecords}} records (benchmark range)</td><td style="text-align:right">€{{printf "%.0f" .Risk.Exposure.OperationalCostMin}}</td><td style="text-align:right">€{{printf "%.0f" .Risk.Exposure.OperationalCostMax}}</td></tr>
        <tr><td><strong>Risk Multiplier</strong></td><td>Derived from K={{.Risk.KAnonymity}}, L={{.Risk.LDiversity}} profile</td><td style="text-align:right">{{printf "%.1f" .Risk.Exposure.RiskMultiplierMin}}×</td><td style="text-align:right">{{printf "%.1f" .Risk.Exposure.RiskMultiplierMax}}×</td></tr>
        <tr class="var-total-row"><td colspan="2"><strong>Total Value at Risk (Estimated Range)</strong></td><td style="text-align:right"><strong>€{{printf "%.0f" .Risk.Exposure.VaRMin}}</strong></td><td style="text-align:right"><strong>€{{printf "%.0f" .Risk.Exposure.VaRMax}}</strong></td></tr>
      </tbody>
    </table>

    <div class="assumptions-box">
      <strong>Assumptions &amp; Methodology Note:</strong><br>{{.Risk.Exposure.AssumptionsNote}}
    </div>
  </div>

  <!-- AI Exposure -->
  <div class="section">
    <h2>AI &amp; LLM Exposure Assessment</h2>
    <table>
      <thead><tr><th>Parameter</th><th>Assessment</th></tr></thead>
      <tbody>
        <tr><td><strong>Decision</strong></td><td><span class="badge badge-{{lower .Risk.AI.Status}}">{{.Risk.AI.Status}}</span></td></tr>
        <tr><td><strong>External LLM API Risk</strong></td><td><span class="badge badge-{{lower .Risk.AI.LLMExposure}}">{{.Risk.AI.LLMExposure}}</span></td></tr>
        <tr><td><strong>Vector DB / RAG Indexing</strong></td><td>{{if .Risk.AI.RAGSafe}}✅ Estimated safe for indexing{{else}}🚫 Not recommended without prior processing{{end}}</td></tr>
        <tr><td><strong>RAG Guidance</strong></td><td>{{.Risk.AI.RAGGuidance}}</td></tr>
        <tr><td><strong>Recommended Action</strong></td><td><strong>{{.Risk.AI.Recommendation}}</strong></td></tr>
      </tbody>
    </table>
  </div>

  <!-- Before / After -->
  <div class="section">
    <h2>Before / After — OCULTAR Impact Simulation</h2>
    <p style="font-size:12px;color:var(--muted);margin-bottom:16px">Projected figures based on modelled processing outcomes. Actual results may vary.</p>
    <div class="scenario-grid">
      <div class="scenario-card scenario-before">
        <h3>{{.Before.Label}}</h3>
        <div class="scenario-stat"><label>Risk Level</label><span>🔴 {{.Before.RiskLevel}}</span></div>
        <div class="scenario-stat"><label>Risk Score</label><span>{{.Before.RiskScore}}</span></div>
        <div class="scenario-stat"><label>Financial Exposure</label><span>{{.Before.VaRRange}}</span></div>
        <div class="scenario-stat"><label>AI Status</label><span>{{.Before.AIStatus}}</span></div>
        <p style="margin-top:12px;font-size:12px;color:#6b7280">{{.Before.Description}}</p>
      </div>
      <div class="scenario-card scenario-after">
        <h3>{{.After.Label}}</h3>
        <div class="scenario-stat"><label>Risk Level</label><span>🟢 {{.After.RiskLevel}}</span></div>
        <div class="scenario-stat"><label>Risk Score</label><span>{{.After.RiskScore}}</span></div>
        <div class="scenario-stat"><label>Financial Exposure</label><span>{{.After.VaRRange}}</span></div>
        <div class="scenario-stat"><label>AI Status</label><span>{{.After.AIStatus}}</span></div>
        <p style="margin-top:12px;font-size:12px;color:#3d6b3d">{{.After.Description}}</p>
      </div>
    </div>
  </div>

  <!-- Remediation -->
  <div class="section">
    <h2>Remediation Plan</h2>
    <div class="remediation-list">
      {{if eq .Risk.LDiversity 1}}
      <div class="step" style="border: 1px solid var(--high); padding: 12px; border-radius: 8px; background: #fff7ed;">
        <div class="step-num" style="background: var(--high);">!</div>
        <div class="step-body">
          <strong style="color: var(--high);">Homogeneity Attack Mitigation (L=1)</strong>
          <p>The dataset exhibits critical attribute homogeneity. An adversary can infer sensitive values with 100% probability for certain groups. 
          <strong>Action:</strong> Apply <em>t-closeness</em> suppression or value generalization to sensitive fields before vector indexing.</p>
        </div>
      </div>
      {{end}}
      
      <div class="step numbered-step"><div class="step-num"></div><div class="step-body"><strong>Tokenization</strong><p>Replace all Name, IBAN, and Email fields with OCULTAR reversible vault tokens. Zero-knowledge re-hydration available on authorised request.</p></div></div>
      
      {{if lt .Risk.KAnonymity 3}}
      <div class="step numbered-step"><div class="step-num"></div><div class="step-body"><strong>Generalization (Target K≥3)</strong><p>Replace precise Region sub-categories with broader geographic tiers to increase K-Anonymity group size above the commonly recommended threshold of K≥3.</p></div></div>
      {{end}}
      
      <div class="step numbered-step"><div class="step-num"></div><div class="step-body"><strong>Format-Preserving Encryption (FPE)</strong><p>Apply FPE to IBAN and financial fields to maintain data utility for analytics while preventing plaintext exposure.</p></div></div>
      <div class="step numbered-step"><div class="step-num"></div><div class="step-body"><strong>Automate via OCULTAR Pipeline</strong><p>All steps above can be automated via the OCULTAR Enterprise proxy. Route your LLM API calls through the proxy and all PII is intercepted and redacted in real-time, with zero changes to your application code.</p></div></div>
    </div>
  </div>

  <!-- Footer -->
  <div class="footer">
    OCULTAR Enterprise {{.Meta.EngineVersion}} | Methodology v{{.Meta.MethodologyVersion}} | Report OCU-{{.Meta.ReportID}}<br>
    This report was generated automatically. All findings are based on the dataset provided at time of analysis and constitute a technical assessment only.<br>
    Standards referenced: GDPR (context) · ISO/IEC 29101 · Industry Breach Cost Benchmarks
  </div>

</div>
<button class="print-btn" onclick="window.print()">
  <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M6 9V2h12v7M6 18H4a2 2 0 01-2-2v-5a2 2 0 012-2h16a2 2 0 012 2v5a2 2 0 01-2 2h-2M6 14h12v8H6z"/></svg>
  Download PDF
</button>
</body>
</html>
`

func main() {
	datasetPath := flag.String("dataset", "", "Path to the JSON dataset file")
	outputPath := flag.String("output", "risk_report.md", "Output path for the Markdown report")
	htmlPath := flag.String("html", "", "If set, also generate an HTML report at this path")
	flag.Parse()

	if *datasetPath == "" {
		log.Fatal("Usage: riskreport -dataset <path> [-output <path>] [-html <path>]")
	}

	data, err := os.ReadFile(*datasetPath)
	if err != nil {
		log.Fatalf("Failed to read dataset: %v", err)
	}

	var dataset []map[string]interface{}
	if err := json.Unmarshal(data, &dataset); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	qi := []string{"region", "dept"}
	sa := []string{"name", "iban", "email"}

	risk := audit.AnalyzeDatasetRisk(dataset, qi, sa)
	meta := buildMeta(*datasetPath, len(dataset))
	before, after := buildScenarios(risk)

	report := fullReport{Meta: meta, Risk: risk, Before: before, After: after}

	// --- Markdown output (text/template — no HTML escaping) ---
	mdTmpl := texttmpl.Must(texttmpl.New("md").Parse(mdTemplate))
	mdFile, err := os.Create(*outputPath)
	if err != nil {
		log.Fatalf("Failed to create markdown output: %v", err)
	}
	defer mdFile.Close()
	if err := mdTmpl.Execute(mdFile, report); err != nil {
		log.Fatalf("Failed to render markdown: %v\nNote: ensure all template fields match the updated RiskReport struct.", err)
	}
	fmt.Printf("✅  Markdown report: %s\n", *outputPath)

	// --- HTML output (html/template — XSS-safe) ---
	if *htmlPath != "" {
		funcMap := htmltmpl.FuncMap{
			"lower": strings.ToLower,
			"pct":   func(score float64) int { return int(score * 10) },
		}
		htmlTmpl := htmltmpl.Must(htmltmpl.New("html").Funcs(funcMap).Parse(htmlTemplate))
		htmlFile, err := os.Create(*htmlPath)
		if err != nil {
			log.Fatalf("Failed to create HTML output: %v", err)
		}
		defer htmlFile.Close()
		if err := htmlTmpl.Execute(htmlFile, report); err != nil {
			log.Fatalf("Failed to render HTML: %v", err)
		}
		fmt.Printf("✅  HTML report: %s\n", *htmlPath)
		
		// Attempt to auto-open the dashboard
		openBrowser(*htmlPath)
	}
}

// openBrowser opens the specified URL in the default user browser.
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Printf("⚠️  Could not auto-open browser: %v\n", err)
	}
}
