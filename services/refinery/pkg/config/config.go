package config

import (
	"embed"
	"encoding/json"
	"log"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

//go:embed data/*
var embeddedData embed.FS

type RegexRule struct {
	Type     string         `yaml:"type" json:"type"`
	Pattern  string         `yaml:"pattern" json:"pattern"`
	Compiled *regexp.Regexp `yaml:"-" json:"-"`
}

type DictRule struct {
	Type  string   `yaml:"type" json:"type"`
	Terms []string `yaml:"terms" json:"terms"`
}

type Settings struct {
	Regexes            []RegexRule `yaml:"regexes"`
	Dictionaries       []DictRule  `yaml:"dictionaries"`
	PresidioConfidence float64     `yaml:"presidio_confidence"`

	// Phase 4: Distributed Enterprise Vaulting
	// VaultBackend selects the storage backend: "duckdb" (default) or "postgres".
	VaultBackend string `yaml:"vault_backend"`
	// PostgresDSN is the PostgreSQL connection string used when VaultBackend is "postgres".
	// Example: "host=db.corp.internal port=5432 user=ocultar password=s3cr3t dbname=ocultar_vault sslmode=require"
	PostgresDSN string `yaml:"postgres_dsn"`

	// Phase 5: SR-SLMs (Domain Snapshots)
	// DomainSnapshot selects the AI model: "standard", "clinical", or "fintech".
	DomainSnapshot string `yaml:"domain_snapshot" json:"domain_snapshot"`

	// Governance: Regulatory Policy
	RegulatoryPolicy map[string]interface{} `yaml:"-" json:"regulatory_policy"`

	// AliasMapping (Task 1) - Maps internal IDs to Google InfoTypes
	AliasMapping map[string]string `yaml:"-" json:"alias_mapping"`
}

var Global Settings

func initDefaultConfig() {
	Global = Settings{
		Regexes: []RegexRule{
			{Type: "EMAIL", Pattern: `(?i)[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}`},
			{Type: "URL", Pattern: `(?i)https?://[^\s"<>\{\}\[\]\\]+|\bwww\.[a-zA-Z0-9\-]+\.[a-zA-Z]{2,}[^\s"<>\{\}\[\]\\]*`},
			{Type: "SSN", Pattern: `\b\d{3}-\d{2}-\d{4}\b`},
			{Type: "CREDENTIAL", Pattern: `(?i)\bpassword\s*[:=]\s*[^\s,]+`},
			{Type: "SECRET", Pattern: `(?i)\b(?:secret|key|token)\s*[:=]\s*[^\s,]+`},
			{Type: "CREDIT_CARD", Pattern: `\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|6(?:011|5[0-9][0-9])[0-9]{12}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|(?:2131|1800|35\d{3})\d{11})\b`},
			{Type: "PATIENT_ID", Pattern: `\b[A-Z]{2,3}[0-9]{6,10}\b`},
			{Type: "MEDICAL_RECORD", Pattern: `\bMRN[- ]?[0-9]{7,10}\b`},
		},
		Dictionaries: []DictRule{
			{Type: "PERSON_VIP", Terms: []string{"Héctor Eduardo Trejos", "Héctor Eduardo", "Eduardo Trejos", "Héctor", "Hector", "Eduardo", "Trejos", "Project Phoenix", "Ouroboros Protocol"}},
		},
		PresidioConfidence: 0.6,
		DomainSnapshot:     "standard",
	}
	loadProtectedEntities()
	LoadRegulatoryPolicy()
	LoadAliasMapping()
	CompileRegexes()
}

// LoadAliasMapping reads the Ocultar -> Google InfoType registry from configs/mapping.json
func LoadAliasMapping() {
	data, err := os.ReadFile("configs/mapping.json")
	if err != nil {
		// Try relative path from service root if not found at absolute root
		data, err = os.ReadFile("services/refinery/configs/mapping.json")
	}

	if err != nil {
		log.Printf("[WARN] Failed to read mapping.json: %v. CanonicalType will be empty.", err)
		return
	}

	var mapping map[string]string
	if err := json.Unmarshal(data, &mapping); err != nil {
		log.Printf("[ERROR] Failed to parse mapping.json: %v", err)
		return
	}

	Global.AliasMapping = mapping
	log.Printf("[INFO] Alias Mapping loaded: %d entities mapped to Google InfoTypes.", len(mapping))
}

// LoadRegulatoryPolicy reads the centralized governance mapping from embedded security data.
func LoadRegulatoryPolicy() {
	data, err := embeddedData.ReadFile("data/regulatory_policy.json")
	if err != nil {
		log.Printf("[WARN] Failed to read embedded regulatory_policy.json: %v. Using hardcoded fallbacks.", err)
		return
	}

	var policy map[string]interface{}
	if err := json.Unmarshal(data, &policy); err != nil {
		log.Printf("[ERROR] Failed to parse embedded regulatory_policy.json: %v", err)
		return
	}

	Global.RegulatoryPolicy = policy
	log.Printf("[INFO] Embedded regulatory policy v%v loaded successfully.", policy["version"])
}

// loadProtectedEntities attempts to read local dictionary terms from embedded data.
// If found, they are injected dynamically into the Global configuration.
func loadProtectedEntities() {
	data, err := embeddedData.ReadFile("data/protected_entities.json")
	if err != nil {
		log.Fatalf("[FATAL] [VULN-004] Failed reading embedded protected_entities.json! Refinery refusing to boot (fail-closed): %v", err)
	}

	var entities []string
	if err := json.Unmarshal(data, &entities); err != nil {
		log.Fatalf("[FATAL] [VULN-004] Failed parsing embedded protected_entities.json! Refinery refusing to boot (fail-closed): %v", err)
	}

	if len(entities) > 0 {
		Global.Dictionaries = append(Global.Dictionaries, DictRule{
			Type:  "PROTECTED_ENTITY",
			Terms: entities,
		})
	} else {
		log.Fatalf("[FATAL] [VULN-004] Embedded protected_entities.json parsed successfully but contains zero entries. " +
			"This would boot the refinery with no Dictionary Shield. Refinery refusing to start (fail-closed).")
	}
}

func CompileRegexes() {
	for i := range Global.Regexes {
		Global.Regexes[i].Compiled = regexp.MustCompile(Global.Regexes[i].Pattern)
	}
}

// Load applies base rules for community edition.
func Load() {
	initDefaultConfig()
}

// InitDefaults explicitly initializes the default rules primarily for testing purposes.
func InitDefaults() {
	initDefaultConfig()
}

func AddRegexRule(rule RegexRule) {
	rule.Compiled = regexp.MustCompile(rule.Pattern)
	Global.Regexes = append(Global.Regexes, rule)
}

func RemoveRegexRule(ruleType string) {
	var n []RegexRule
	for _, r := range Global.Regexes {
		if r.Type != ruleType {
			n = append(n, r)
		}
	}
	Global.Regexes = n
}

func AddDictionaryTerm(dictType, term string) {
	for i, d := range Global.Dictionaries {
		if d.Type == dictType {
			for _, existing := range d.Terms {
				if existing == term {
					return
				}
			}
			Global.Dictionaries[i].Terms = append(Global.Dictionaries[i].Terms, term)
			return
		}
	}
	Global.Dictionaries = append(Global.Dictionaries, DictRule{
		Type:  dictType,
		Terms: []string{term},
	})
}

func Save() error {
	var protected []string
	var saveDicts []DictRule

	for _, d := range Global.Dictionaries {
		if d.Type == "PROTECTED_ENTITY" {
			protected = append(protected, d.Terms...)
		} else {
			saveDicts = append(saveDicts, d)
		}
	}

	if len(protected) > 0 {
		b, _ := json.MarshalIndent(protected, "", "  ")
		os.WriteFile("configs/protected_entities.json", b, 0644)
	}

	saveObj := Global
	saveObj.Dictionaries = saveDicts

	b, err := yaml.Marshal(saveObj)
	if err != nil {
		return err
	}
	return os.WriteFile("configs/config.yaml", b, 0644)
}
