package config

import (
	"encoding/json"
	"log"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

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
		},
		Dictionaries: []DictRule{
			{Type: "PERSON_VIP", Terms: []string{"Héctor Eduardo Trejos", "Héctor Eduardo", "Eduardo Trejos", "Héctor", "Hector", "Eduardo", "Trejos", "Project Phoenix", "Ouroboros Protocol"}},
		},
		PresidioConfidence: 0.6,
		DomainSnapshot:     "standard",
	}
	loadProtectedEntities()
	LoadRegulatoryPolicy()
	CompileRegexes()
}

// LoadRegulatoryPolicy reads the centralized governance mapping from security/regulatory_policy.json.
func LoadRegulatoryPolicy() {
	// Look for security/regulatory_policy.json in the current dir or parent
	path := "security/regulatory_policy.json"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = "../security/regulatory_policy.json" // Try parent (monorepo structure)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("[WARN] Failed to load regulatory_policy.json: %v. Using hardcoded fallbacks.", err)
		return
	}

	var policy map[string]interface{}
	if err := json.Unmarshal(data, &policy); err != nil {
		log.Printf("[ERROR] Failed to parse regulatory_policy.json: %v", err)
		return
	}

	Global.RegulatoryPolicy = policy
	log.Printf("[INFO] Regulatory policy v%v loaded successfully.", policy["version"])
}

// loadProtectedEntities attempts to read local dictionary terms from configs/protected_entities.json.
// If found, they are injected dynamically into the Global configuration.
func loadProtectedEntities() {
	data, err := os.ReadFile("configs/protected_entities.json")
	if err != nil {
		log.Fatalf("[FATAL] Failed reading protected_entities.json! Engine refusing to boot (fail-closed): %v", err)
	}

	var entities []string
	if err := json.Unmarshal(data, &entities); err != nil {
		log.Fatalf("[FATAL] Failed parsing protected_entities.json! Engine refusing to boot (fail-closed): %v", err)
	}

	if len(entities) > 0 {
		Global.Dictionaries = append(Global.Dictionaries, DictRule{
			Type:  "PROTECTED_ENTITY",
			Terms: entities,
		})
	} else {
		log.Fatalf("[FATAL] protected_entities.json parsed successfully but contains zero entries. " +
			"This would boot the engine with no Dictionary Shield. Engine refusing to start (fail-closed). " +
			"Add at least one protected entity to configs/protected_entities.json.")
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
