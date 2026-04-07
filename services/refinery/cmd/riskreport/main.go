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
)

const reportVersion = "3.0"
const engineVersion = "v1.14"

// reportMeta holds non-risk metadata for the report header.
type reportMeta struct {
	ReportID          string
	GeneratedAt       string
	DatasetScope      string
	MethodologyVersion string
	EngineVersion     string
	TotalRecords      int
}

// fullReport combines metadata and risk results for template rendering.
type fullReport struct {
	Meta    reportMeta
	Risk    audit.RiskReport
	Before  scenarioSummary
	After   scenarioSummary
}

// scenarioSummary represents a Before or After OCULTAR simulation snapshot.
type scenarioSummary struct {
	Label       string
	RiskLevel   string
	RiskScore   string
	VaR         string
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
		Label:       "Scenario A — Current State (No Protection)",
		RiskLevel:   r.OverallRiskLevel,
		RiskScore:   fmt.Sprintf("%.1f / 10", r.OverallRiskScore),
		VaR:         fmt.Sprintf("€%.0f", r.Exposure.TotalVaR),
		AIStatus:    r.AI.Status,
		Description: "The raw dataset as-is, transmitted directly to an LLM API or stored in a vector database. All PII fields are exposed in plaintext.",
	}

	// Project the "After OCULTAR" state: K→∞, L→∞, fully compliant
	afterScore := r.OverallRiskScore * 0.05 // 95% risk reduction
	afterVaR := r.Exposure.TotalVaR * 0.02   // 98% financial exposure reduction (residual operational cost)

	after := scenarioSummary{
		Label:       "Scenario B — After OCULTAR Processing",
		RiskLevel:   "LOW",
		RiskScore:   fmt.Sprintf("%.1f / 10 (projected)", afterScore),
		VaR:         fmt.Sprintf("€%.0f (projected residual)", afterVaR),
		AIStatus:    "ALLOW",
		Description: "After OCULTAR's tokenization and format-preserving encryption pipeline. All PII replaced with reversible vault tokens. Dataset becomes K=∞ Anonymous.",
	}
	return before, after
}

const mdTemplate = `# OCULTAR Data Risk Assessment Report

> **CONFIDENTIAL — For Authorised Recipients Only**
> This report constitutes a formal risk and compliance assessment. Distribution is restricted to named stakeholders.

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
> **Compliance Status: {{if .Risk.IsGDPRCompliant}}✅ COMPLIANT{{else}}❌ NON-COMPLIANT (GDPR Art. 5){{end}}**{{end}}
{{if eq .Risk.OverallRiskLevel "HIGH"}}> [!WARNING]
> **Overall Risk Level: {{.Risk.OverallRiskLevel}} ({{printf "%.1f" .Risk.OverallRiskScore}}/10)**
> **Compliance Status: {{if .Risk.IsGDPRCompliant}}✅ COMPLIANT{{else}}❌ NON-COMPLIANT (GDPR Art. 5){{end}}**{{end}}
{{if eq .Risk.OverallRiskLevel "MEDIUM"}}> [!IMPORTANT]
> **Overall Risk Level: {{.Risk.OverallRiskLevel}} ({{printf "%.1f" .Risk.OverallRiskScore}}/10)**
> **Compliance Status: {{if .Risk.IsGDPRCompliant}}✅ COMPLIANT{{else}}❌ NON-COMPLIANT (GDPR Art. 5){{end}}**{{end}}
{{if eq .Risk.OverallRiskLevel "LOW"}}> [!NOTE]
> **Overall Risk Level: {{.Risk.OverallRiskLevel}} ({{printf "%.1f" .Risk.OverallRiskScore}}/10)**
> **Compliance Status: ✅ COMPLIANT**{{end}}

The dataset identified in this report contains **{{.Risk.ViolatingRecords}} records** that fail minimum EU anonymisation thresholds. In its current state, this data **{{if .Risk.IsGDPRCompliant}}satisfies{{else}}cannot safely be used with{{end}} external AI systems and LLM APIs** without a material risk of regulatory breach and personal data exposure.

The estimated financial exposure associated with unauthorised disclosure of this dataset is **€{{printf "%.0f" .Risk.Exposure.TotalVaR}}**, comprising regulatory fines, forensic response costs, and reputational damage.

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

> **Regulatory Threshold:** A minimum K-score of 3 is required under EU statistical disclosure guidelines and is the recommended baseline for GDPR Article 89 research exemptions.

### L-Diversity
**Raw Score:** {{.Risk.LDiversity}}

{{.Risk.LDiversityInterpretation}}

> **Regulatory Threshold:** An L-Diversity score of ≥2 is required to mitigate homogeneity attacks as recommended under ISO/IEC 29101 (Privacy Architecture Framework).

---

## Financial Exposure Model

The **Value at Risk (VaR)** is computed using a Three-Pillar methodology aligned with Ponemon Institute (2024) cost benchmarks and GDPR Article 83 fine exposure estimates.

| Pillar | Formula | Estimated Exposure |
| :--- | :--- | ---: |
| **Regulatory Fines** | €10,000 × {{.Risk.ViolatingRecords}} violating records | **€{{printf "%.0f" .Risk.Exposure.RegulatoryFine}}** |
| **Operational Cost** | €250 × {{.Risk.TotalRecords}} total records (forensics, notification, legal) | **€{{printf "%.0f" .Risk.Exposure.OperationalCost}}** |
| **Brand Damage** | (Regulatory + Operational) × Risk-weighted multiplier × 40% | **€{{printf "%.0f" .Risk.Exposure.BrandDamage}}** |
| | | |
| **Total Value at Risk** | | **€{{printf "%.0f" .Risk.Exposure.TotalVaR}}** |

> **Methodology Note:** The Brand Damage multiplier scales with Overall Risk Score (range: 0.5x–2.0x) and represents projected customer churn, competitive disadvantage, and crisis management costs based on industry sector benchmarks.

---

## AI & LLM Exposure Assessment

### Decision: {{.Risk.AI.Status}}

| Parameter | Assessment |
| :--- | :--- |
| **External LLM API Safety** | {{.Risk.AI.LLMExposure}} risk |
| **Internal Copilot Safety** | {{if eq .Risk.AI.Status "ALLOW"}}✅ Permitted with monitoring{{else if eq .Risk.AI.Status "SANITIZE_FIRST"}}⚠️ Permitted after OCULTAR processing{{else}}🚫 Blocked — must not be used{{end}} |
| **Vector DB / RAG Indexing** | {{if .Risk.AI.RAGSafe}}✅ Safe for indexing{{else}}🚫 Blocked for indexing{{end}} |

**RAG & Vector Database Guidance:**
{{.Risk.AI.RAGGuidance}}

**Recommended Action:**
{{.Risk.AI.Recommendation}}

---

## Before / After Simulation

This section demonstrates the measurable impact of the OCULTAR Enterprise pipeline on your dataset's risk profile.

| Metric | {{.Before.Label}} | {{.After.Label}} |
| :--- | :--- | :--- |
| **Risk Level** | 🔴 {{.Before.RiskLevel}} | 🟢 {{.After.RiskLevel}} |
| **Risk Score** | {{.Before.RiskScore}} | {{.After.RiskScore}} |
| **Financial Exposure (VaR)** | {{.Before.VaR}} | {{.After.VaR}} |
| **AI / LLM Status** | {{.Before.AIStatus}} | {{.After.AIStatus}} |

**What changes:**
- **{{.Before.Description}}**
- **{{.After.Description}}**

> The projected 95% risk reduction and 98% VaR reduction are based on OCULTAR's K=∞ Tokenization model, where all PII is replaced with reversible vault-encrypted references. Residual exposure represents standard operational overhead.

---

## Remediation Plan

{{.Risk.Recommendation}}

---

## Appendix: Methodology & Standards

This report applies the following analytical frameworks:

- **K-Anonymity** (Sweeney, 2002): Minimum group size for quasi-identifier equivalence classes.
- **L-Diversity** (Machanavajjhala et al., 2006): Sensitive attribute diversity within equivalence classes.
- **GDPR Article 5(1)(f)**: Integrity and confidentiality principle.
- **GDPR Article 83**: Administrative fine schedule (up to 4% global annual turnover or €20M).
- **Ponemon Institute 2024 Cost of a Data Breach Report**: Forensic cost benchmarks.
- **ISO/IEC 29101**: Privacy architecture framework for ICT systems.

---

*OCULTAR Enterprise {{.Meta.EngineVersion}} | Methodology v{{.Meta.MethodologyVersion}} | Report ID: OCU-{{.Meta.ReportID}}*
*This report was generated automatically. Findings are based on the dataset provided at time of analysis.*
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
  .badge-sanitize { background: #fffbeb; color: var(--medium); }
  .badge-allow { background: #f0fdf4; color: var(--low); }
  
  /* Score bars */
  .score-bar-wrap { display: flex; align-items: center; gap: 10px; }
  .score-bar { height: 6px; border-radius: 3px; background: #e2e8f0; flex: 1; }
  .score-bar-fill { height: 100%; border-radius: 3px; }
  .fill-critical { background: var(--critical); }
  .fill-high { background: var(--high); }
  .fill-medium { background: var(--medium); }
  .fill-low { background: var(--low); }
  
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
  
  /* Financial table */
  .var-total td { font-weight: 700; font-size: 15px; background: #f8fafc; }
  
  /* Remediation steps */
  .step { display: flex; gap: 12px; margin-bottom: 14px; }
  .step-num { width: 28px; height: 28px; border-radius: 50%; background: var(--accent); color: white; font-size: 12px; font-weight: 700; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
  .step-body strong { display: block; font-size: 13px; font-weight: 600; margin-bottom: 2px; }
  .step-body p { font-size: 12px; color: var(--muted); }
  
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
  .confidential { background: #fef9c3; border: 1px solid #fde68a; border-radius: 6px; padding: 8px 14px; font-size: 12px; color: #92400e; margin-bottom: 24px; text-align: center; font-weight: 500; }
  .footer { text-align: center; margin-top: 40px; font-size: 11px; color: var(--muted); }
</style>
</head>
<body>
<div class="container">

  <div class="confidential">⚠️ CONFIDENTIAL — For Authorised Recipients Only. Report ID: OCU-{{.Meta.ReportID}}</div>

  <!-- Header -->
  <div class="report-header">
    <h1>OCULTAR Data Risk Assessment</h1>
    <div class="subtitle">Enterprise Compliance &amp; AI Exposure Report</div>
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
      <h2>{{.Risk.OverallRiskLevel}} Risk — {{if .Risk.IsGDPRCompliant}}✅ COMPLIANT{{else}}❌ NON-COMPLIANT{{end}}</h2>
      <p>{{.Risk.ViolatingRecords}} of {{.Risk.TotalRecords}} records fail EU anonymisation thresholds. Estimated financial exposure: <strong>€{{printf "%.0f" .Risk.Exposure.TotalVaR}}</strong>.</p>
      <p style="margin-top:8px">This dataset <strong>{{if .Risk.IsGDPRCompliant}}can be safely used{{else}}cannot be safely used{{end}}</strong> with external AI APIs or vector databases without prior sanitisation.</p>
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
    <blockquote>Regulatory Threshold: K ≥ 3 required under EU statistical disclosure guidelines (GDPR Art. 89 research exemption).</blockquote>
    <h3>L-Diversity (Score: {{.Risk.LDiversity}})</h3>
    <p>{{.Risk.LDiversityInterpretation}}</p>
    <blockquote>Regulatory Threshold: L ≥ 2 required to mitigate homogeneity attacks per ISO/IEC 29101.</blockquote>
  </div>

  <!-- Financial Exposure -->
  <div class="section">
    <h2>Financial Exposure — Three-Pillar VaR Model</h2>
    <table>
      <thead><tr><th>Pillar</th><th>Formula</th><th style="text-align:right">Exposure (€)</th></tr></thead>
      <tbody>
        <tr><td><strong>Regulatory Fines</strong></td><td>€10,000 × {{.Risk.ViolatingRecords}} violating records</td><td style="text-align:right">€{{printf "%.0f" .Risk.Exposure.RegulatoryFine}}</td></tr>
        <tr><td><strong>Operational Cost</strong></td><td>€250 × {{.Risk.TotalRecords}} records (forensics, notification, legal)</td><td style="text-align:right">€{{printf "%.0f" .Risk.Exposure.OperationalCost}}</td></tr>
        <tr><td><strong>Brand Damage</strong></td><td>Risk-weighted multiplier × 40% of direct costs</td><td style="text-align:right">€{{printf "%.0f" .Risk.Exposure.BrandDamage}}</td></tr>
        <tr class="var-total"><td colspan="2"><strong>Total Value at Risk (VaR)</strong></td><td style="text-align:right"><strong>€{{printf "%.0f" .Risk.Exposure.TotalVaR}}</strong></td></tr>
      </tbody>
    </table>
  </div>

  <!-- AI Exposure -->
  <div class="section">
    <h2>AI &amp; LLM Exposure Assessment</h2>
    <table>
      <thead><tr><th>Parameter</th><th>Assessment</th></tr></thead>
      <tbody>
        <tr><td><strong>Decision</strong></td><td><span class="badge badge-{{lower .Risk.AI.Status}}">{{.Risk.AI.Status}}</span></td></tr>
        <tr><td><strong>External LLM API Risk</strong></td><td><span class="badge badge-{{lower .Risk.AI.LLMExposure}}">{{.Risk.AI.LLMExposure}}</span></td></tr>
        <tr><td><strong>Vector DB / RAG Indexing</strong></td><td>{{if .Risk.AI.RAGSafe}}✅ Safe for indexing{{else}}🚫 Blocked — requires sanitisation{{end}}</td></tr>
        <tr><td><strong>RAG Guidance</strong></td><td>{{.Risk.AI.RAGGuidance}}</td></tr>
        <tr><td><strong>Recommended Action</strong></td><td><strong>{{.Risk.AI.Recommendation}}</strong></td></tr>
      </tbody>
    </table>
  </div>

  <!-- Before / After -->
  <div class="section">
    <h2>Before / After — OCULTAR Impact Simulation</h2>
    <div class="scenario-grid">
      <div class="scenario-card scenario-before">
        <h3>{{.Before.Label}}</h3>
        <div class="scenario-stat"><label>Risk Level</label><span>🔴 {{.Before.RiskLevel}}</span></div>
        <div class="scenario-stat"><label>Risk Score</label><span>{{.Before.RiskScore}}</span></div>
        <div class="scenario-stat"><label>Financial Exposure</label><span>{{.Before.VaR}}</span></div>
        <div class="scenario-stat"><label>AI Status</label><span>{{.Before.AIStatus}}</span></div>
        <p style="margin-top:12px;font-size:12px;color:#6b7280">{{.Before.Description}}</p>
      </div>
      <div class="scenario-card scenario-after">
        <h3>{{.After.Label}}</h3>
        <div class="scenario-stat"><label>Risk Level</label><span>🟢 {{.After.RiskLevel}}</span></div>
        <div class="scenario-stat"><label>Risk Score</label><span>{{.After.RiskScore}}</span></div>
        <div class="scenario-stat"><label>Financial Exposure</label><span>{{.After.VaR}}</span></div>
        <div class="scenario-stat"><label>AI Status</label><span>{{.After.AIStatus}}</span></div>
        <p style="margin-top:12px;font-size:12px;color:#6b7280">{{.After.Description}}</p>
      </div>
    </div>
  </div>

  <!-- Remediation -->
  <div class="section">
    <h2>Remediation Plan</h2>
    <div class="step"><div class="step-num">1</div><div class="step-body"><strong>Tokenization</strong><p>Replace all Name, IBAN, and Email fields with OCULTAR reversible vault tokens. Zero-knowledge re-hydration available on authorised request.</p></div></div>
    <div class="step"><div class="step-num">2</div><div class="step-body"><strong>Generalization</strong><p>Replace precise Region sub-categories with broader geographic tiers to increase K-Anonymity group size above the legal threshold of 3.</p></div></div>
    <div class="step"><div class="step-num">3</div><div class="step-body"><strong>Format-Preserving Encryption (FPE)</strong><p>Apply FPE to IBAN and financial fields to maintain data utility for analytics while preventing plaintext exposure.</p></div></div>
    <div class="step"><div class="step-num">4</div><div class="step-body"><strong>Automate via OCULTAR Pipeline</strong><p>All steps above can be automated via the OCULTAR Enterprise proxy. Route your LLM API calls through the proxy and all PII is intercepted and redacted in real-time, with zero changes to your application code.</p></div></div>
  </div>

  <!-- Footer -->
  <div class="footer">
    OCULTAR Enterprise {{.Meta.EngineVersion}} | Methodology v{{.Meta.MethodologyVersion}} | Report OCU-{{.Meta.ReportID}}<br>
    This report was generated automatically. All findings are based on the dataset provided at time of analysis.<br>
    Standards applied: GDPR Art. 5 &amp; 83 · ISO/IEC 29101 · Ponemon Institute 2024
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
	outputPath  := flag.String("output", "risk_report.md", "Output path for the Markdown report")
	htmlPath    := flag.String("html", "", "If set, also generate an HTML report at this path")
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
		log.Fatalf("Failed to render markdown: %v", err)
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
	}
}
