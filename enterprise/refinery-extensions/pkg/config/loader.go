package config

import (
	"log"
	"os"

	"github.com/Edu963/ocultar/pkg/config"
	"gopkg.in/yaml.v3"
)

// Load reads config.yaml and applies custom enterprise rules dynamically.
func Load() {
	data, err := os.ReadFile("configs/config.yaml")
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("[WARN] Failed to read configs/config.yaml: %v", err)
		}
		return
	}

	var userConfig config.Settings
	if err := yaml.Unmarshal(data, &userConfig); err != nil {
		log.Printf("[ERR] Failed to parse configs/config.yaml: %v", err)
		return
	}

	// Merge enterprise configs dynamically
	config.Global.Regexes = append(config.Global.Regexes, userConfig.Regexes...)
	config.Global.Dictionaries = append(config.Global.Dictionaries, userConfig.Dictionaries...)

	if userConfig.SLMConfidence > 0 {
		config.Global.SLMConfidence = userConfig.SLMConfidence
	}

	// Phase 4: Distributed Enterprise Vaulting — vault backend selection
	if userConfig.VaultBackend != "" {
		config.Global.VaultBackend = userConfig.VaultBackend
	}
	if userConfig.PostgresDSN != "" {
		config.Global.PostgresDSN = userConfig.PostgresDSN
	}

	log.Printf("[INFO] Loaded custom rules from configs/config.yaml (Enterprise mode enabled)")
	config.CompileRegexes()
}
