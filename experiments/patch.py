import os

#!/usr/bin/env python3

with open("main.go", "r") as f:
    main_go = f.read()

# Replacements in main.go
main_go = main_go.replace("var presidioAvailable = true", "var slmAvailable = true")

main_go = main_go.replace(
    "var presidioClient = &http.Client{", "var slmClient = &http.Client{"
)

main_go = main_go.replace(
    "presidioHost := getPresidioHost()", "slmHost := getSLMHost()"
)

main_go = main_go.replace(
    "checkPresidioHealth(presidioHost)", "checkSLMHealth(slmHost)"
)

# Tier 2 Logic override
tier2_presidio = """	// TIER 2: Presidio NER Scan
	// Gatekeeper: skip the HTTP round-trip entirely for very short strings,
	// strings already fully tokenised by Tier 1/1.5, or when Presidio is offline.
	if presidioAvailable && len(refined) > 15 && !isFullyTokenised(refined) {
		piiMap := askPresidioForPII(refined)
		// Metadata-only log — never log raw entity values.
		if len(piiMap) > 0 {
			log.Printf(\"[INFO] Presidio: %d entity type(s) detected in record\", len(piiMap))
		}
		for piiType, items := range piiMap {
			for _, item := range items {
				if len(strings.TrimSpace(item)) > 2 {
					refined = applyReplacement(db, refined, item, piiType, key, actor)
				}
			}
		}
	}"""

tier2_slm = """	// TIER 2: SLM NER Scan
	// Gatekeeper: skip the HTTP round-trip entirely for very short strings,
	// strings already fully tokenised by Tier 1/1.5, or when SLM is offline.
	if slmAvailable && len(refined) > 15 && !isFullyTokenised(refined) {
		piiMap := askSLMForPII(refined)
		// Metadata-only log — never log raw entity values.
		if len(piiMap) > 0 {
			log.Printf("[INFO] SLM: %d entity type(s) detected in record", len(piiMap))
		}
		for piiType, items := range piiMap {
			for _, item := range items {
				if len(strings.TrimSpace(item)) > 2 {
					refined = applyReplacement(db, refined, item, piiType, key, actor)
				}
			}
		}
	}"""
main_go = main_go.replace(tier2_presidio, tier2_slm)


# Presidio Bridge override
bridge_presidio = """// ── Presidio NER Bridge ──────────────────────────────────────────────────────

// presidioAnalyzeRequest matches the Presidio /analyze REST API request body.
type presidioAnalyzeRequest struct {
	Text     string   `json:"text"`
	Language string   `json:"language"`
	Entities []string `json:"entities"`
}

// presidioResult is one entity recognition result from Presidio.
type presidioResult struct {
	EntityType string  `json:"entity_type"`
	Start      int     `json:"start"`
	End        int     `json:"end"`
	Score      float64 `json:"score"`
	RecogName  string  `json:"recognition_metadata,omitempty"`
}

// getPresidioHost returns the Presidio analyzer URL from env or a sensible default.
func getPresidioHost() string {
	h := os.Getenv("PRESIDIO_HOST")
	if h == "" {
		return "http://localhost:5001"
	}
	return h
}

// presidioEntityToSOV maps Presidio entity type names to OCULTAR PII categories.
var presidioEntityToSOV = map[string]string{
	"PERSON":        "PERSON",
	"LOCATION":      "LOCATION",
	"ORGANIZATION":  "COMPANY",
	"PHONE_NUMBER":  "PHONE",
	"EMAIL_ADDRESS": "EMAIL",
	"NRP":           "PERSON", // Nationalities, Religious, Political groups — often names
	"DATE_TIME":     "DATE",
	"MEDICAL":       "HEALTH",
}

// presidioEntities is the fixed list of entity types we ask Presidio to detect.
var presidioEntities = []string{
	"PERSON", "LOCATION", "ORGANIZATION",
	"PHONE_NUMBER", "EMAIL_ADDRESS", "NRP", "DATE_TIME",
}

// spanKey uniquely identifies an entity span to deduplicate across language passes.
type spanKey struct {
	EntityType string
	Start      int
	End        int
}

// askPresidioForPII calls the Presidio analyzer sidecar for both French and Spanish,
// deduplicates results by span, and returns a map of SOV PII types → extracted strings.
//
// Gatekeeper: the caller (refineString) already short-circuits for short/fully-tokenised
// strings. This function adds a secondary score filter (min 0.6) to suppress low-quality
// recognitions from noisy NLP models.
func askPresidioForPII(text string) map[string][]string {
	host := getPresidioHost()
	seen := map[spanKey]bool{}
	result := map[string][]string{}

	for _, lang := range []string{"fr", "es"} {
		reqBody := presidioAnalyzeRequest{
			Text:     text,
			Language: lang,
			Entities: presidioEntities,
		}
		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			log.Printf("[WARN] Presidio: failed to marshal request: %v", err)
			continue
		}

		resp, err := presidioClient.Post(host+"/analyze", "application/json", bytes.NewBuffer(bodyBytes))
		if err != nil {
			log.Printf("[WARN] Presidio: HTTP call failed for lang=%s: %v", lang, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("[WARN] Presidio: non-200 for lang=%s: %d — %s", lang, resp.StatusCode, strings.TrimSpace(string(body)))
			continue
		}

		var results []presidioResult
		if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
			log.Printf("[WARN] Presidio: failed to decode response for lang=%s: %v", lang, err)
			continue
		}

		for _, r := range results {
			// Score filter: skip low-confidence detections
			if r.Score < 0.6 {
				continue
			}

			key := spanKey{EntityType: r.EntityType, Start: r.Start, End: r.End}
			if seen[key] {
				continue
			}
			seen[key] = true

			// Bounds check before slicing
			if r.Start < 0 || r.End > len(text) || r.Start >= r.End {
				continue
			}
			extracted := strings.TrimSpace(text[r.Start:r.End])
			if len(extracted) < 3 {
				continue
			}

			sovType, ok := presidioEntityToSOV[r.EntityType]
			if !ok {
				sovType = r.EntityType // passthrough unknown types
			}
			result[sovType] = append(result[sovType], extracted)
		}
	}

	return result
}

// checkPresidioHealth pings the Presidio sidecar at startup.
// On failure it sets presidioAvailable=false and logs a clear warning rather
// than crashing, allowing the engine to continue in Regex-only degraded mode.
// Operators MUST monitor for the [WARN] DEGRADED log in production.
func checkPresidioHealth(host string) {
	resp, err := presidioClient.Get(host + "/health")
	if err != nil {
		log.Printf("[WARN] DEGRADED MODE: Presidio is unreachable at %s (%v). NER detection disabled — running Regex-only.", host, err)
		presidioAvailable = false
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("[WARN] DEGRADED MODE: Presidio returned HTTP %d. NER detection disabled — running Regex-only.", resp.StatusCode)
		presidioAvailable = false
		return
	}
	log.Printf("[INFO] Presidio health check passed at %s", host)
	presidioAvailable = true
}"""

bridge_slm = """// ── SLM NER Bridge ──────────────────────────────────────────────────────

const slmSystemPrompt = `You are a strict Named Entity Recognition (NER) system tailored for PII redaction.
Extract PII entities of the following types: PERSON, LOCATION, ORGANIZATION, PHONE, EMAIL, DATE.
You MUST output ONLY a valid JSON array of objects.
Each object must have exactly two keys: "entity_type" and "value".
Do NOT output any conversational text, explanations, or markdown formatting blocks. If no entities are found, output an empty array [].
Example output: [{"entity_type": "PERSON", "value": "John Doe"}, {"entity_type": "EMAIL", "value": "john@example.com"}]`

type slmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type slmChatRequest struct {
	Messages    []slmMessage `json:"messages"`
	Temperature float64      `json:"temperature"`
}

type slmChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type slmEntity struct {
	EntityType string `json:"entity_type"`
	Value      string `json:"value"`
}

func getSLMHost() string {
	h := os.Getenv("SLM_HOST")
	if h == "" {
		return "http://localhost:8080"
	}
	return h
}

// askSLMForPII calls the local SLM via a fast OpenAI-compatible API to extract entity arrays
func askSLMForPII(text string) map[string][]string {
	host := getSLMHost()
	result := map[string][]string{}

	reqBody := slmChatRequest{
		Messages: []slmMessage{
			{Role: "system", Content: slmSystemPrompt},
			{Role: "user", Content: text},
		},
		Temperature: 0.1, // Deterministic extraction
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[WARN] SLM: failed to marshal request: %v", err)
		return result
	}

	resp, err := slmClient.Post(host+"/v1/chat/completions", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		log.Printf("[WARN] SLM: HTTP call failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[WARN] SLM: non-200 HTTP %d — %s", resp.StatusCode, strings.TrimSpace(string(body)))
		return result
	}

	var chatResp slmChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		log.Printf("[WARN] SLM: failed to decode response: %v", err)
		return result
	}

	if len(chatResp.Choices) == 0 {
		return result
	}

	content := strings.TrimSpace(chatResp.Choices[0].Message.Content)

	// Robust JSON array extraction when SLM hallucinated conversational text/markdown
	if strings.Contains(content, "[") && strings.Contains(content, "]") {
		start := strings.Index(content, "[")
		end := strings.LastIndex(content, "]")
		content = content[start : end+1]
	}

	var entities []slmEntity
	if err := json.Unmarshal([]byte(content), &entities); err != nil {
		log.Printf("[WARN] SLM: failed to parse JSON array: %v. Raw content: %s", err, content)
		return result
	}

	for _, e := range entities {
		val := strings.TrimSpace(e.Value)
		if len(val) > 2 {
			result[e.EntityType] = append(result[e.EntityType], val)
		}
	}

	return result
}

// checkSLMHealth pings the llama.cpp server at startup.
func checkSLMHealth(host string) {
	resp, err := slmClient.Get(host + "/health")
	if err != nil {
		log.Printf("[WARN] DEGRADED MODE: SLM is unreachable at %s (%v). NER disabled.", host, err)
		slmAvailable = false
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("[WARN] DEGRADED MODE: SLM returned HTTP %d. NER disabled.", resp.StatusCode)
		slmAvailable = false
		return
	}
	log.Printf("[INFO] SLM health check passed at %s", host)
	slmAvailable = true
}"""
main_go = main_go.replace(bridge_presidio, bridge_slm)

with open("main.go", "w") as f:
    f.write(main_go)


# ----------------------------------------------------
# docker-compose.yml modification
# ----------------------------------------------------
with open("docker-compose.yml", "r") as f:
    dc = f.read()

dc_presidio_service = """  # ── Tier 2: Presidio Analyzer Sidecar ─────────────────────────────────────
  # Provides multilingual NER (French + Spanish) via spaCy transformer models.
  # Presidio replaces Ollama/Llama3 as the primary PII detection engine.
  presidio-analyzer:
    image: mcr.microsoft.com/presidio-analyzer:latest
    ports:
      - "5001:5001"
    environment:
      # Default Presidio port
      - PORT=5001
    # Pre-load French and Spanish spaCy models at container startup.
    # These are large models (~500MB each) — first pull will take time.
    # After the first run they are cached in the image layer.
    command: >
      sh -c "python -m spacy download fr_core_news_lg &&
             python -m spacy download es_core_news_lg &&
             python app.py"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:5001/health"]
      interval: 15s
      timeout: 10s
      retries: 5
      start_period: 60s   # Allow time for model downloads on first launch"""

dc_slm_service = """  # ── Tier 2: Specialized SLM NER Sidecar ─────────────────────────────────────
  # Replaces generic Ollama/Presidio to provide fast, locally quantized API
  # specifically prompted to extract PII natively.
  slm-ner:
    image: ghcr.io/ggerganov/llama.cpp:server
    ports:
      - "8080:8080"
    volumes:
      - ./slm_model:/models  # Directory where download_quantize_slm.py saved the GGUF model
    command: >
      -m /models/model-q4_k_m.gguf -c 4096 --host 0.0.0.0 --port 8080
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 15s
      timeout: 10s
      retries: 5"""

dc = dc.replace(dc_presidio_service, dc_slm_service)
dc = dc.replace(
    "PRESIDIO_HOST=http://presidio-analyzer:5001", "SLM_HOST=http://slm-ner:8080"
)
dc = dc.replace("presidio-analyzer:", "slm-ner:")

with open("docker-compose.yml", "w") as f:
    f.write(dc)

# We should also update dist/enterprise/docker-compose.yml and dist/community/docker-compose.yml if they exist
for dist in ["dist/enterprise", "dist/community"]:
    if os.path.exists(dist):
        try:
            with open(os.path.join(dist, "docker-compose.yml"), "r") as f:
                dist_dc = f.read()
            dist_dc = dist_dc.replace(dc_presidio_service, dc_slm_service)
            dist_dc = dist_dc.replace(
                "PRESIDIO_HOST=http://presidio-analyzer:5001",
                "SLM_HOST=http://slm-ner:8080",
            )
            dist_dc = dist_dc.replace("presidio-analyzer:", "slm-ner:")
            with open(os.path.join(dist, "docker-compose.yml"), "w") as f:
                f.write(dist_dc)
        except Exception:
            pass

    if os.path.exists(os.path.join(dist, "main.go")):
        try:
            with open(os.path.join(dist, "main.go"), "r") as f:
                dist_main = f.read()

            # Same replacements as main main.go
            dist_main = dist_main.replace(
                "var presidioAvailable = true", "var slmAvailable = true"
            )
            dist_main = dist_main.replace(
                "var presidioClient = &http.Client{", "var slmClient = &http.Client{"
            )
            dist_main = dist_main.replace(
                "presidioHost := getPresidioHost()", "slmHost := getSLMHost()"
            )
            dist_main = dist_main.replace(
                "checkPresidioHealth(presidioHost)", "checkSLMHealth(slmHost)"
            )
            dist_main = dist_main.replace(tier2_presidio, tier2_slm)
            dist_main = dist_main.replace(bridge_presidio, bridge_slm)

            with open(os.path.join(dist, "main.go"), "w") as f:
                f.write(dist_main)
        except Exception:
            pass

print("Patching complete.")
