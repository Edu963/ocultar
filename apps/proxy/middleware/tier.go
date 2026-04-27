// Package middleware provides HTTP middleware for the OCULTAR proxy ingress layer.
package middleware

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Edu963/ocultar-proxy/metrics"
)

// Tier represents an Ocultar subscription tier.
type Tier string

const (
	TierFree       Tier = "free"
	TierDeveloper  Tier = "developer"
	TierTeam       Tier = "team"
	TierEnterprise Tier = "enterprise"
)

const (
	HeaderTier           = "X-Ocultar-Tier"
	HeaderCallsRemaining = "X-Ocultar-Calls-Remaining"
	unlimited            = "unlimited"
)

// monthlyLimits maps each tier to its monthly call cap. -1 = unlimited.
var monthlyLimits = map[Tier]int64{
	TierFree:       0,        // no quota — blocks when Tier 2 is active
	TierDeveloper:  50_000,
	TierTeam:       200_000,
	TierEnterprise: -1,       // overridable via OCU_ENTERPRISE_LIMIT
}

// UsageStore tracks per-API-key monthly call counts.
type UsageStore interface {
	// Increment adds one call for (apiKeyHash, month) and returns the new total.
	Increment(apiKeyHash, month string) (int64, error)
	// Count returns the current call total for (apiKeyHash, month).
	Count(apiKeyHash, month string) (int64, error)
}

// TierMiddleware enforces tier-based access control and monthly call caps.
// It reads the tier from the X-Ocultar-Tier request header and adds
// X-Ocultar-Tier and X-Ocultar-Calls-Remaining to every response.
type TierMiddleware struct {
	next            http.Handler
	tier2Available  bool    // true when refinery has an active AI scanner
	store           UsageStore
	enterpriseLimit int64   // -1 = unlimited; set from OCU_ENTERPRISE_LIMIT
}

// New constructs a TierMiddleware wrapping next.
//   - tier2Available: pass eng.AIScanner.IsAvailable() at startup.
//   - store: UsageStore for call tracking; nil disables cap enforcement.
//   - enterpriseLimit: enterprise monthly cap; -1 = unlimited.
func New(next http.Handler, tier2Available bool, store UsageStore, enterpriseLimit int64) *TierMiddleware {
	return &TierMiddleware{
		next:            next,
		tier2Available:  tier2Available,
		store:           store,
		enterpriseLimit: enterpriseLimit,
	}
}

// EnterpriseLimit reads OCU_ENTERPRISE_LIMIT from the environment.
// Returns -1 (unlimited) when the variable is absent or unparseable.
func EnterpriseLimit() int64 {
	if s := os.Getenv("OCU_ENTERPRISE_LIMIT"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return n
		}
	}
	return -1
}

// ServeHTTP implements http.Handler.
func (m *TierMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	raw := strings.ToLower(strings.TrimSpace(r.Header.Get(HeaderTier)))
	if raw == "" {
		raw = string(TierFree) // missing header defaults to free tier
	}
	tier := Tier(raw)

	switch tier {
	case TierFree, TierDeveloper, TierTeam, TierEnterprise:
		// valid
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_tier",
			"message": "X-Ocultar-Tier must be one of: free, developer, team, enterprise",
		})
		return
	}

	// Free tier: block entirely when Tier 2 AI scanner is active.
	if tier == TierFree && m.tier2Available {
		metrics.IncTierUpgradeRequired()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(HeaderTier, string(TierFree))
		w.Header().Set(HeaderCallsRemaining, "0")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error":       "tier_upgrade_required",
			"tier":        "free",
			"message":     "SLM inference requires Developer tier or above",
			"upgrade_url": "https://ocultar.dev/pricing",
		})
		return
	}

	// Resolve the effective monthly cap.
	limit := monthlyLimits[tier]
	if tier == TierEnterprise && m.enterpriseLimit >= 0 {
		limit = m.enterpriseLimit
	}

	if limit > 0 && m.store != nil {
		// Enforce monthly call budget.
		apiKeyHash := HashAPIKey(r.Header.Get("Authorization"))
		month := time.Now().UTC().Format("2006-01")

		count, err := m.store.Increment(apiKeyHash, month)
		if err != nil {
			http.Error(w, "usage tracking error", http.StatusInternalServerError)
			return
		}

		remaining := limit - count
		if remaining < 0 {
			remaining = 0
		}
		w.Header().Set(HeaderTier, string(tier))
		w.Header().Set(HeaderCallsRemaining, strconv.FormatInt(remaining, 10))

		if count > limit {
			now := time.Now().UTC()
			nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
			w.Header().Set("Retry-After", strconv.Itoa(int(time.Until(nextMonth).Seconds())))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":            "monthly_limit_exceeded",
				"tier":             string(tier),
				"calls_this_month": count,
				"monthly_limit":    limit,
			})
			return
		}
	} else {
		// Unlimited: enterprise with no cap, or free tier without Tier 2.
		w.Header().Set(HeaderTier, string(tier))
		w.Header().Set(HeaderCallsRemaining, unlimited)
	}

	metrics.IncAPICall(string(tier))
	m.next.ServeHTTP(w, r)
}

// UsageHandler returns an http.Handler for GET /v1/usage.
// It requires a valid X-Ocultar-Tier header and Authorization header.
func (m *TierMiddleware) UsageHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		raw := strings.ToLower(strings.TrimSpace(r.Header.Get(HeaderTier)))
		if raw == "" {
			raw = string(TierFree)
		}
		tier := Tier(raw)

		limit := monthlyLimits[tier]
		if tier == TierEnterprise && m.enterpriseLimit >= 0 {
			limit = m.enterpriseLimit
		}

		apiKeyHash := HashAPIKey(r.Header.Get("Authorization"))
		month := time.Now().UTC().Format("2006-01")

		var count int64
		if m.store != nil {
			count, _ = m.store.Count(apiKeyHash, month)
		}

		var callsRemaining string
		if limit < 0 {
			callsRemaining = unlimited
		} else {
			rem := limit - count
			if rem < 0 {
				rem = 0
			}
			callsRemaining = strconv.FormatInt(rem, 10)
		}

		now := time.Now().UTC()
		nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tier":             string(tier),
			"calls_this_month": count,
			"calls_remaining":  callsRemaining,
			"reset_date":       nextMonth.Format("2006-01-02"),
		})
	})
}

// HashAPIKey returns the first 8 bytes of SHA-256(key) as hex.
// Strips the "Bearer " prefix if present; uses "anonymous" for empty keys.
func HashAPIKey(auth string) string {
	key := strings.TrimPrefix(auth, "Bearer ")
	if key == "" {
		key = "anonymous"
	}
	sum := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", sum[:8])
}
