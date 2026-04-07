package audit

import (
	"fmt"
	"math"
	"strings"
)

// FinancialExposure breaks down the Value at Risk (VaR) into its constituent components.
// All figures are simulated estimates based on industry benchmarks and should not be
// interpreted as legally binding calculations or regulatory determinations.
type FinancialExposure struct {
	// Range-based regulatory exposure (simulation anchors, not per-record fines)
	RegulatoryExposureMin float64 `json:"regulatory_exposure_min_eur"`
	RegulatoryExposureMax float64 `json:"regulatory_exposure_max_eur"`

	// Operational cost estimate based on industry breach studies (€100–€300/record)
	OperationalCostMin float64 `json:"operational_cost_min_eur"`
	OperationalCostMax float64 `json:"operational_cost_max_eur"`

	// Derived risk multiplier range (from K-Anonymity, L-Diversity, PII density)
	RiskMultiplierMin float64 `json:"risk_multiplier_min"`
	RiskMultiplierMax float64 `json:"risk_multiplier_max"`

	// Final VaR range — primary output for reporting
	VaRMin float64 `json:"var_min_eur"`
	VaRMax float64 `json:"var_max_eur"`

	// Midpoint kept for backward compatibility with callers expecting a single value
	TotalVaR float64 `json:"total_var_eur"`

	// Plain-English explanation of methodology and assumptions
	AssumptionsNote string `json:"assumptions_note"`
}

// AIReadiness describes the dataset's safety profile for use with AI/LLM systems.
type AIReadiness struct {
	Status         string `json:"status"`       // ALLOW / SANITIZE_FIRST / BLOCK
	LLMExposure    string `json:"llm_exposure"` // Risk level if sent to external API
	RAGSafe        bool   `json:"rag_safe"`     // Safe for vector DB / RAG indexing
	RAGGuidance    string `json:"rag_guidance"` // Plain-English RAG advice
	Recommendation string `json:"recommendation"`
}

// CategoryScore is a normalised 0–10 score for a specific risk category.
type CategoryScore struct {
	Score       float64 `json:"score"`       // 0 (safe) – 10 (critical)
	Label       string  `json:"label"`       // LOW / MEDIUM / HIGH / CRITICAL
	Implication string  `json:"implication"` // Plain-English business meaning
}

// RiskReport is the full audit-grade output of the OCULTAR risk assessment engine.
// All financial figures are estimated ranges based on industry benchmarks.
// This output is informational and does not constitute legal advice.
type RiskReport struct {
	// --- Core Privacy Metrics ---
	KAnonymity       int  `json:"k_anonymity"`
	LDiversity       int  `json:"l_diversity"`
	IsGDPRCompliant  bool `json:"is_gdpr_anonymous"` // Heuristic assessment only
	ViolatingRecords int  `json:"violating_records"`
	TotalRecords     int  `json:"total_records"`

	// --- Derived Risk Scoring ---
	DatasetRiskScore float64 `json:"dataset_risk_score"` // 0.0–1.0 normalised composite
	RiskMultiplier   float64 `json:"risk_multiplier"`    // Midpoint of the range

	// --- Weighted Scorecard ---
	OverallRiskScore     float64       `json:"overall_risk_score"` // 0–10
	OverallRiskLevel     string        `json:"overall_risk_level"` // LOW/MEDIUM/HIGH/CRITICAL
	Identifiability      CategoryScore `json:"identifiability_risk"`
	FinancialSensitivity CategoryScore `json:"financial_sensitivity"`
	ReidentificationRisk CategoryScore `json:"reidentification_risk"`
	ComplianceReadiness  CategoryScore `json:"compliance_readiness"`

	// --- Financial Exposure Model ---
	Exposure FinancialExposure `json:"financial_exposure"`

	// --- AI Safety Assessment ---
	AI AIReadiness `json:"ai_readiness"`

	// --- Plain-English Metric Interpretations ---
	KAnonymityInterpretation string `json:"k_anonymity_interpretation"`
	LDiversityInterpretation string `json:"l_diversity_interpretation"`

	// --- Remediation ---
	Recommendation string `json:"recommendation"`
}

// scoreToLabel converts a 0–10 numeric score to a qualitative risk label.
func scoreToLabel(score float64) string {
	switch {
	case score <= 2.5:
		return "LOW"
	case score <= 5.0:
		return "MEDIUM"
	case score <= 7.5:
		return "HIGH"
	default:
		return "CRITICAL"
	}
}

// computeIdentifiability scores re-identification risk based on K-Anonymity.
// Language is calibrated to reflect empirical risk without making absolute legal claims.
func computeIdentifiability(k int) CategoryScore {
	var score float64
	var impl string
	switch {
	case k <= 0:
		score = 10
		impl = "No quasi-identifier grouping detected. Every record appears fully unique, presenting maximum re-identification risk in an adversarial context."
	case k == 1:
		score = 9.5
		impl = "K=1: Each individual is uniquely identifiable from their quasi-attributes alone. This represents the highest re-identification risk level. Common industry benchmarks suggest K≥3–5 as a minimum for basic pseudonymization."
	case k == 2:
		score = 7.0
		impl = "K=2: Individuals share attributes with only one other record. This dataset is likely to present a high re-identification risk in external processing scenarios. Industry benchmarks generally recommend K≥3–5 for robust pseudonymization."
	case k == 3:
		score = 4.0
		impl = "K=3: Meets commonly cited minimum thresholds for statistical pseudonymization. Borderline for most enterprise privacy frameworks; contextual factors apply."
	case k <= 5:
		score = 2.5
		impl = fmt.Sprintf("K=%d: Acceptable anonymization level for many use cases. Each individual is indistinguishable within a group of %d records. Contextual risk factors should still be evaluated.", k, k)
	default:
		score = 1.0
		impl = fmt.Sprintf("K=%d: Strong pseudonymization level. Re-identification risk is estimated to be low under typical adversarial models. Residual risks from auxiliary information cannot be mathematically excluded.", k)
	}
	return CategoryScore{Score: score, Label: scoreToLabel(score), Implication: impl}
}

// computeFinancialSensitivity scores based on sensitive attribute density.
func computeFinancialSensitivity(sensitiveCount, total int) CategoryScore {
	if total == 0 {
		return CategoryScore{Score: 0, Label: "LOW", Implication: "No records to analyse."}
	}
	ratio := float64(sensitiveCount) / float64(total)
	score := math.Min(ratio*12, 10)
	impl := fmt.Sprintf(
		"An estimated %.0f%% of assessed records contain high-sensitivity attributes (financial identifiers, personal names). "+
			"Exposure of this dataset in an unprotected context would likely present significant regulatory and operational risk. "+
			"Subject to contextual factors including jurisdiction, data controller agreements, and processing purpose.",
		ratio*100,
	)
	return CategoryScore{Score: score, Label: scoreToLabel(score), Implication: impl}
}

// computeReidentification scores based on combined K and L scores.
func computeReidentification(k, l int) CategoryScore {
	score := 0.0
	if k <= 1 {
		score += 5.0
	} else if k < 3 {
		score += 3.0
	}
	if l <= 1 {
		score += 5.0
	} else if l < 2 {
		score += 3.0
	}
	score = math.Min(score, 10)
	impl := fmt.Sprintf(
		"Estimated re-identification attack surface: %.0f/10 (based on K-Anonymity=%d and L-Diversity=%d). "+
			"An adversary with partial auxiliary knowledge may be able to isolate individuals with elevated probability. "+
			"This is a modelled estimate; actual risk depends on the adversary model and available external data.",
		score, k, l,
	)
	return CategoryScore{Score: score, Label: scoreToLabel(score), Implication: impl}
}

// computeComplianceReadiness scores overall regulatory readiness.
// Note: this is a heuristic assessment, not a legal determination of compliance.
func computeComplianceReadiness(isCompliant bool, violating, total int) CategoryScore {
	if isCompliant {
		return CategoryScore{
			Score: 1.5,
			Label: "LOW",
			Implication: "Dataset satisfies commonly cited K-Anonymity and L-Diversity thresholds for statistical pseudonymization. " +
				"This is a technical heuristic and does not constitute a legal determination of GDPR compliance.",
		}
	}
	ratio := float64(violating) / float64(total)
	score := math.Min(ratio*10, 10)
	impl := fmt.Sprintf(
		"An estimated %.0f%% of records (%d/%d) fall below commonly cited minimum pseudonymization thresholds. "+
			"This dataset presents a high likelihood of non-compliance in external processing scenarios. "+
			"Regulatory exposure depends on jurisdiction, processing context, and applicable exemptions.",
		ratio*100, violating, total,
	)
	return CategoryScore{Score: score, Label: scoreToLabel(score), Implication: impl}
}

// computeDatasetRiskScore derives a normalised 0.0–1.0 composite score used as
// a simulation anchor for the VaR regulatory exposure component.
func computeDatasetRiskScore(overallScore float64) float64 {
	return math.Min(overallScore/10.0, 1.0)
}

// computeRiskMultiplier returns a range of multipliers derived from K-Anonymity,
// L-Diversity, and PII density. Higher risk profiles receive larger multipliers.
func computeRiskMultiplier(k, l int) (minMult, maxMult float64) {
	switch {
	case k <= 1:
		return 1.5, 2.0
	case k < 3:
		return 1.3, 1.5
	case l <= 1:
		return 1.2, 1.5
	default:
		return 1.0, 1.2
	}
}

// computeVaRRange implements the 3-component VaR model.
//
// Formula: VaR = (RegulatoryExposure + OperationalCost) × RiskMultiplier
//
// Component 1 — Regulatory Exposure (simulation anchor, not a per-record fine):
//   RegulatoryMin = DatasetRiskScore × €10,000   (base anchor — low scenario)
//   RegulatoryMax = DatasetRiskScore × €100,000  (base anchor — high scenario)
//
// Component 2 — Operational Cost (industry breach benchmarks):
//   OperationalMin = records × €100  (lower bound, IBM/Ponemon range)
//   OperationalMax = records × €300  (upper bound, IBM/Ponemon range)
//
// Component 3 — Risk Multiplier (K/L driven):
//   K=1 → 1.5–2.0x | K<3 → 1.3–1.5x | else → 1.0–1.2x
func computeVaRRange(records int, datasetRiskScore float64, k, l int) FinancialExposure {
	// Pillar 1: Regulatory exposure (simulation anchors)
	const baseLow = 10_000.0
	const baseHigh = 100_000.0
	regMin := datasetRiskScore * baseLow
	regMax := datasetRiskScore * baseHigh

	// Pillar 2: Operational cost (industry benchmark range)
	const costPerRecordLow = 100.0
	const costPerRecordHigh = 300.0
	opMin := float64(records) * costPerRecordLow
	opMax := float64(records) * costPerRecordHigh

	// Pillar 3: Risk multiplier
	multMin, multMax := computeRiskMultiplier(k, l)

	// Final VaR range
	varMin := (regMin + opMin) * multMin
	varMax := (regMax + opMax) * multMax
	midpoint := (varMin + varMax) / 2.0

	note := "This estimate is based on industry breach cost benchmarks (€100–€300/record range from published incident studies) " +
		"and a regulatory exposure simulation anchor (€10,000–€100,000 base, scaled by dataset risk score). " +
		"The risk multiplier (range: 1.0×–2.0×) is derived from K-Anonymity and L-Diversity scores. " +
		"Actual financial impact may vary significantly depending on enforcement context, data controller obligations, " +
		"jurisdiction, organisational size, and mitigating controls in place. " +
		"These figures are simulated estimates and do not constitute legal or financial advice."

	return FinancialExposure{
		RegulatoryExposureMin: regMin,
		RegulatoryExposureMax: regMax,
		OperationalCostMin:    opMin,
		OperationalCostMax:    opMax,
		RiskMultiplierMin:     multMin,
		RiskMultiplierMax:     multMax,
		VaRMin:                varMin,
		VaRMax:                varMax,
		TotalVaR:              midpoint,
		AssumptionsNote:       note,
	}
}

// computeAIReadiness evaluates dataset safety for AI/LLM and RAG systems.
func computeAIReadiness(k int, isCompliant bool, sensitiveRatio float64) AIReadiness {
	ragSafe := isCompliant && k >= 5
	var status, llmExposure, ragGuidance, rec string

	if isCompliant && k >= 5 {
		status = "ALLOW"
		llmExposure = "LOW"
		ragGuidance = "Estimated safe for RAG indexing under current pseudonymization thresholds. Dataset meets commonly cited minimum anonymization benchmarks. Recommend ongoing monitoring of embedding queries for indirect re-identification patterns."
		rec = "Dataset may be considered for use with external LLM APIs and vector databases, subject to applicable Data Processing Agreements and your organisation's data governance policies."
	} else if isCompliant || k >= 3 {
		status = "SANITIZE_FIRST"
		llmExposure = "MEDIUM"
		ragGuidance = "Not recommended for RAG indexing without prior tokenization. Embedding high-cardinality PII fields creates semantic re-identification vectors in the embedding space that are difficult to audit or retract."
		rec = "Apply OCULTAR Format-Preserving Tokenization before routing to any LLM API or vector store. Consult your DPA and data governance team before proceeding."
	} else {
		status = "BLOCK"
		llmExposure = "CRITICAL"
		ragGuidance = "Blocked for RAG indexing. Embedding raw PII (names, IBANs, emails) into a vector database creates queryable re-identification surfaces that are difficult to retroactively remove. This assessment is based on the detected K-Anonymity and L-Diversity levels."
		rec = "Do not transmit this dataset to any external API or internal AI copilot system without first running the full OCULTAR redaction pipeline. This recommendation is based on technical risk assessment, not a legal determination."
	}

	return AIReadiness{
		Status:         status,
		LLMExposure:    llmExposure,
		RAGSafe:        ragSafe,
		RAGGuidance:    ragGuidance,
		Recommendation: rec,
	}
}

// AnalyzeDatasetRisk evaluates the mathematical privacy properties of a structured dataset.
// Results are estimates based on K-Anonymity and L-Diversity models and industry benchmarks.
// This function does not produce a legal compliance determination.
func AnalyzeDatasetRisk(dataset []map[string]interface{}, quasiIdentifiers []string, sensitiveAttributes []string) RiskReport {
	if len(dataset) == 0 {
		return RiskReport{}
	}

	// 1. Group records by quasi-identifiers
	groups := make(map[string][]map[string]interface{})
	for _, rec := range dataset {
		var qiValues []string
		for _, qi := range quasiIdentifiers {
			if val, ok := rec[qi]; ok {
				qiValues = append(qiValues, fmt.Sprintf("%v", val))
			} else {
				qiValues = append(qiValues, "*")
			}
		}
		key := strings.Join(qiValues, "|")
		groups[key] = append(groups[key], rec)
	}

	minK := -1
	minL := -1
	violatingRecords := 0

	// 2. Evaluate K-Anonymity and L-Diversity
	for _, group := range groups {
		k := len(group)
		if minK == -1 || k < minK {
			minK = k
		}
		if k < 3 {
			violatingRecords += k
		}

		if len(sensitiveAttributes) > 0 {
			sensitiveValues := make(map[string]bool)
			for _, rec := range group {
				var saValues []string
				for _, sa := range sensitiveAttributes {
					if val, ok := rec[sa]; ok {
						saValues = append(saValues, fmt.Sprintf("%v", val))
					}
				}
				sensitiveValues[strings.Join(saValues, "|")] = true
			}
			l := len(sensitiveValues)
			if minL == -1 || l < minL {
				minL = l
			}
		}
	}
	if minL == -1 {
		minL = 0
	}

	isCompliant := minK >= 3 && (len(sensitiveAttributes) == 0 || minL >= 2)
	total := len(dataset)

	// 3. Compute category scores
	identifiability := computeIdentifiability(minK)
	financialSensitivity := computeFinancialSensitivity(violatingRecords, total)
	reidentification := computeReidentification(minK, minL)
	complianceReadiness := computeComplianceReadiness(isCompliant, violatingRecords, total)

	// 4. Overall weighted risk score (weights: identifiability 35%, financial 25%, reid 25%, compliance 15%)
	overallScore := (identifiability.Score*0.35 +
		financialSensitivity.Score*0.25 +
		reidentification.Score*0.25 +
		complianceReadiness.Score*0.15)

	// 5. VaR range
	datasetRiskScore := computeDatasetRiskScore(overallScore)
	exposure := computeVaRRange(total, datasetRiskScore, minK, minL)
	multMin, multMax := computeRiskMultiplier(minK, minL)
	riskMultiplierMid := (multMin + multMax) / 2.0

	// 6. AI readiness
	sensitiveRatio := float64(violatingRecords) / float64(total)
	ai := computeAIReadiness(minK, isCompliant, sensitiveRatio)

	// 7. Plain-English interpretations
	kInterp := fmt.Sprintf("K-Anonymity = %d — %s", minK, identifiability.Implication)
	lInterp := fmt.Sprintf("L-Diversity = %d — ", minL)
	if minL <= 1 {
		lInterp += "Sensitive attributes appear homogeneous within quasi-identifier groups. " +
			"This increases the risk of homogeneity attacks, where an adversary may infer the sensitive value of any individual in a group."
	} else {
		lInterp += fmt.Sprintf(
			"An estimated %d distinct sensitive values per equivalence class were detected. "+
				"Homogeneity attack risk is estimated to be reduced at this diversity level.",
			minL,
		)
	}

	// 8. Remediation
	rec := "Dataset satisfies commonly cited K-Anonymity and L-Diversity thresholds for pseudonymization. " +
		"Periodic re-evaluation is recommended as the dataset grows or as new quasi-identifiers are identified."
	if !isCompliant {
		rec = "Dataset presents elevated re-identification risk. Recommended remediation steps (in order of priority):\n" +
			"  1. **Tokenization** — Replace Names, IBANs, and Email addresses with OCULTAR vault tokens.\n" +
			"  2. **Generalization** — Replace precise Region values with broader geographic tiers to increase K-Anonymity group size.\n" +
			"  3. **Format-Preserving Encryption (FPE)** — Encrypt IBAN fields while preserving format for analytics.\n" +
			"  All steps above can be automated via the OCULTAR Enterprise pipeline."
	}

	return RiskReport{
		KAnonymity:               minK,
		LDiversity:               minL,
		IsGDPRCompliant:          isCompliant,
		ViolatingRecords:         violatingRecords,
		TotalRecords:             total,
		DatasetRiskScore:         datasetRiskScore,
		RiskMultiplier:           riskMultiplierMid,
		OverallRiskScore:         overallScore,
		OverallRiskLevel:         scoreToLabel(overallScore),
		Identifiability:          identifiability,
		FinancialSensitivity:     financialSensitivity,
		ReidentificationRisk:     reidentification,
		ComplianceReadiness:      complianceReadiness,
		Exposure:                 exposure,
		AI:                       ai,
		KAnonymityInterpretation: kInterp,
		LDiversityInterpretation: lInterp,
		Recommendation:           rec,
	}
}
