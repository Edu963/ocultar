package audit

import (
	"fmt"
	"strings"
)

type RiskReport struct {
	KAnonymity       int    `json:"k_anonymity"`
	LDiversity       int    `json:"l_diversity"`
	IsGDPRCompliant  bool   `json:"is_gdpr_anonymous"`
	Recommendation   string `json:"recommendation"`
	ViolatingRecords int    `json:"violating_records"`
}

// AnalyzeDatasetRisk evaluates the mathematical privacy guarantees of a structured list of JSON maps.
func AnalyzeDatasetRisk(dataset []map[string]interface{}, quasiIdentifiers []string, sensitiveAttributes []string) RiskReport {
	if len(dataset) == 0 {
		return RiskReport{}
	}

	// 1. Group records by their quasi-identifiers
	groups := make(map[string][]map[string]interface{})

	for _, rec := range dataset {
		var qiValues []string
		for _, qi := range quasiIdentifiers {
			if val, ok := rec[qi]; ok {
				qiValues = append(qiValues, fmt.Sprintf("%v", val))
			} else {
				qiValues = append(qiValues, "*") // Standardize missing fields as generic wildcards
			}
		}
		key := strings.Join(qiValues, "|")
		groups[key] = append(groups[key], rec)
	}

	minK := -1
	minL := -1
	violatingRecords := 0

	// 2. Mathematically evaluate bucket dimensions
	for _, group := range groups {
		k := len(group)
		if minK == -1 || k < minK {
			minK = k
		}
		if k < 3 { // K=3 is the general minimum legal barrier for European statistical safety
			violatingRecords += k
		}

		// 3. Prevent 'Homogeneity Attacks' by scoring L-Diversity
		if len(sensitiveAttributes) > 0 {
			sensitiveValues := make(map[string]bool)
			for _, rec := range group {
				var saValues []string
				for _, sa := range sensitiveAttributes {
					if val, ok := rec[sa]; ok {
						saValues = append(saValues, fmt.Sprintf("%v", val))
					}
				}
				saKey := strings.Join(saValues, "|")
				sensitiveValues[saKey] = true
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
	rec := "Dataset achieves compliant K-Anonymity and L-Diversity."
	if !isCompliant {
		rec = "Dataset is highly re-identifiable. Further bucketization or Format-Preserving Encryption required before sharing."
	}

	return RiskReport{
		KAnonymity:       minK,
		LDiversity:       minL,
		IsGDPRCompliant:  isCompliant,
		Recommendation:   rec,
		ViolatingRecords: violatingRecords,
	}
}
