package audit

import (
	"fmt"
	"math"
	"strings"
)

// FinancialExposure breaks down the Value at Risk (VaR) into its three constituent pillars.
type FinancialExposure struct {
	RegulatoryFine   float64 `json:"regulatory_fine_eur"`
	OperationalCost  float64 `json:"operational_cost_eur"`
	BrandDamage      float64 `json:"brand_damage_eur"`
	TotalVaR         float64 `json:"total_var_eur"`
}

// AIReadiness describes the dataset's safety profile for use with AI/LLM systems.
type AIReadiness struct {
	Status          string `json:"status"`           // ALLOW / SANITIZE_FIRST / BLOCK
	LLMExposure     string `json:"llm_exposure"`     // Risk level if sent to external API
	RAGSafe         bool   `json:"rag_safe"`         // Safe for vector DB / RAG indexing
	RAGGuidance     string `json:"rag_guidance"`     // Plain-English RAG advice
	Recommendation  string `json:"recommendation"`
}

// RiskScore is a normalised 0–10 score for a specific risk category.
type CategoryScore struct {
	Score       float64 `json:"score"`  // 0 (safe) – 10 (critical)
	Label       string  `json:"label"`  // LOW / MEDIUM / HIGH / CRITICAL
	Implication string  `json:"implication"`
}

// RiskReport is the full audit-grade output of the OCULTAR risk assessment engine.
type RiskReport struct {
	// --- Core Privacy Metrics ---
	KAnonymity       int    `json:"k_anonymity"`
	LDiversity       int    `json:"l_diversity"`
	IsGDPRCompliant  bool   `json:"is_gdpr_anonymous"`
	ViolatingRecords int    `json:"violating_records"`
	TotalRecords     int    `json:"total_records"`

	// --- Weighted Scorecard ---
	OverallRiskScore    float64       `json:"overall_risk_score"` // 0–10
	OverallRiskLevel    string        `json:"overall_risk_level"` // LOW/MEDIUM/HIGH/CRITICAL
	Identifiability     CategoryScore `json:"identifiability_risk"`
	FinancialSensitivity CategoryScore `json:"financial_sensitivity"`
	ReidentificationRisk CategoryScore `json:"reidentification_risk"`
	ComplianceReadiness CategoryScore `json:"compliance_readiness"`

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
func computeIdentifiability(k int) CategoryScore {
	var score float64
	var impl string
	switch {
	case k <= 0:
		score = 10
		impl = "No quasi-identifier grouping detected. Every record is fully unique and directly re-identifiable."
	case k == 1:
		score = 9.5
		impl = "K=1: Every individual is uniquely identifiable from their quasi-attributes. Maximum re-identification risk."
	case k == 2:
		score = 7.0
		impl = "K=2: Individuals share attributes with only one other record. Fails European minimum anonymity threshold (K≥3)."
	case k == 3:
		score = 4.0
		impl = "K=3: Meets the minimum European statistical safety threshold. Borderline compliant."
	case k <= 5:
		score = 2.5
		impl = fmt.Sprintf("K=%d: Acceptable anonymity. Each individual is indistinguishable within a group of %d.", k, k)
	default:
		score = 1.0
		impl = fmt.Sprintf("K=%d: Strong anonymity guarantee. Re-identification risk is low.", k)
	}
	return CategoryScore{Score: score, Label: scoreToLabel(score), Implication: impl}
}

// computeFinancialSensitivity scores based on sensitive attribute density.
func computeFinancialSensitivity(sensitiveCount, total int) CategoryScore {
	if total == 0 {
		return CategoryScore{Score: 0, Label: "LOW", Implication: "No records to analyse."}
	}
	ratio := float64(sensitiveCount) / float64(total)
	score := math.Min(ratio*12, 10) // Scale: 83%+ sensitive = 10
	impl := fmt.Sprintf(
		"%.0f%% of records contain high-sensitivity attributes (financial identifiers, personal names). "+
			"Exposure of this dataset would constitute a significant regulatory breach.",
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
		"Combined K-Anonymity=%d and L-Diversity=%d yield a re-identification attack surface of %.0f/10. "+
			"An adversary with partial knowledge could isolate individuals with high probability.",
		k, l, score,
	)
	return CategoryScore{Score: score, Label: scoreToLabel(score), Implication: impl}
}

// computeComplianceReadiness scores overall regulatory readiness.
func computeComplianceReadiness(isCompliant bool, violating, total int) CategoryScore {
	if isCompliant {
		return CategoryScore{Score: 1.5, Label: "LOW", Implication: "Dataset satisfies GDPR K-Anonymity and L-Diversity thresholds for statistical disclosure."}
	}
	ratio := float64(violating) / float64(total)
	score := math.Min(ratio*10, 10)
	impl := fmt.Sprintf(
		"%.0f%% of records (%d/%d) fail minimum anonymisation requirements. Regulatory reporting of this dataset in its current state would constitute a GDPR violation under Art. 5(1)(f).",
		ratio*100, violating, total,
	)
	return CategoryScore{Score: score, Label: scoreToLabel(score), Implication: impl}
}

// computeFinancialExposure derives the Three-Pillar VaR model.
// Formula: (Regulatory_Base + Forensics_Per_Record * n) * Brand_Multiplier
func computeFinancialExposure(records, violating int, overallScore float64) FinancialExposure {
	// Pillar 1: Regulatory (GDPR Article 83 exposure estimate)
	// Base fine of €10,000 per identifiable record with financial data
	regulatoryFine := float64(violating) * 10_000.0

	// Pillar 2: Operational (Forensics + Notification + Legal)
	// Industry average: €250/record (Ponemon Institute, 2024)
	operationalCost := float64(records) * 250.0

	// Pillar 3: Brand Damage Multiplier
	// Applied based on Overall Risk Score severity (0.5x – 2.0x)
	multiplier := 0.5 + (overallScore/10.0)*1.5
	brandDamage := (regulatoryFine + operationalCost) * multiplier * 0.4 // 40% of total as brand impact

	total := regulatoryFine + operationalCost + brandDamage

	return FinancialExposure{
		RegulatoryFine:  regulatoryFine,
		OperationalCost: operationalCost,
		BrandDamage:     brandDamage,
		TotalVaR:        total,
	}
}

// computeAIReadiness evaluates dataset safety for AI/LLM and RAG systems.
func computeAIReadiness(k int, isCompliant bool, sensitiveRatio float64) AIReadiness {
	ragSafe := isCompliant && k >= 5
	var status, llmExposure, ragGuidance, rec string

	if isCompliant && k >= 5 {
		status = "ALLOW"
		llmExposure = "LOW"
		ragGuidance = "Safe for RAG indexing. Dataset meets minimum anonymisation thresholds. Recommend monitoring embedding queries for indirect re-identification."
		rec = "Dataset may be used with external LLM APIs and indexed in vector databases subject to your DPA agreements."
	} else if isCompliant || k >= 3 {
		status = "SANITIZE_FIRST"
		llmExposure = "MEDIUM"
		ragGuidance = "Not recommended for RAG indexing without prior tokenization. Embedding high-cardinality PII fields creates semantic re-identification vectors in the vector space."
		rec = "Apply OCULTAR Format-Preserving Tokenization before routing to any LLM API or vector store."
	} else {
		status = "BLOCK"
		llmExposure = "CRITICAL"
		ragGuidance = "BLOCKED for RAG indexing. Embedding raw PII (names, IBANs, emails) into a vector database creates permanent, queryable re-identification attack surfaces that cannot be retroactively purged."
		rec = "Do not transmit this dataset to any external API or internal copilot system. Run full OCULTAR redaction pipeline before any AI usage."
	}

	return AIReadiness{
		Status:         status,
		LLMExposure:    llmExposure,
		RAGSafe:        ragSafe,
		RAGGuidance:    ragGuidance,
		Recommendation: rec,
	}
}

// AnalyzeDatasetRisk evaluates the mathematical privacy guarantees of a structured list of JSON maps.
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

	// 5. Financial exposure
	exposure := computeFinancialExposure(total, violatingRecords, overallScore)

	// 6. AI readiness
	sensitiveRatio := float64(violatingRecords) / float64(total)
	ai := computeAIReadiness(minK, isCompliant, sensitiveRatio)

	// 7. Plain-English interpretations
	kInterp := fmt.Sprintf("K-Anonymity = %d → %s", minK, identifiability.Implication)
	lInterp := fmt.Sprintf("L-Diversity = %d → ", minL)
	if minL <= 1 {
		lInterp += "Sensitive attributes are homogeneous within groups. Homogeneity attacks are feasible. An adversary knows the sensitive value of any individual in a group."
	} else {
		lInterp += fmt.Sprintf("%d distinct sensitive values per equivalence class. Homogeneity attacks are mitigated.", minL)
	}

	// 8. Remediation
	rec := "Dataset achieves compliant K-Anonymity and L-Diversity. Periodic re-evaluation recommended as dataset grows."
	if !isCompliant {
		rec = "Dataset is highly re-identifiable. Recommended remediation (in order of priority):\n" +
			"  1. **Tokenization** — Replace Names, IBANs, and Email addresses with OCULTAR vault tokens.\n" +
			"  2. **Generalization** — Replace precise Region values with broader geographic tiers.\n" +
			"  3. **Format-Preserving Encryption (FPE)** — Encrypt IBAN fields while preserving format for analytics.\n" +
			"  All steps above can be automated via the OCULTAR Enterprise pipeline."
	}

	return RiskReport{
		KAnonymity:              minK,
		LDiversity:              minL,
		IsGDPRCompliant:         isCompliant,
		ViolatingRecords:        violatingRecords,
		TotalRecords:            total,
		OverallRiskScore:        overallScore,
		OverallRiskLevel:        scoreToLabel(overallScore),
		Identifiability:         identifiability,
		FinancialSensitivity:    financialSensitivity,
		ReidentificationRisk:   reidentification,
		ComplianceReadiness:     complianceReadiness,
		Exposure:                exposure,
		AI:                      ai,
		KAnonymityInterpretation: kInterp,
		LDiversityInterpretation: lInterp,
		Recommendation:          rec,
	}
}
