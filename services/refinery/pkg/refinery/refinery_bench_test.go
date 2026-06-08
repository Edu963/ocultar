package refinery_test

// Benchmark harness for the Refinery pipeline.
//
// Run:
//   cd services/refinery && CGO_ENABLED=1 go test ./pkg/refinery/ -bench=. -benchmem -benchtime=5s
//
// Key metrics to track across releases:
//   BenchmarkPipeline_SingleEmail     → baseline Tier 1 latency
//   BenchmarkPipeline_FullDocument    → realistic enterprise document
//   BenchmarkPipeline_Parallel        → concurrent throughput (goroutines=GOMAXPROCS)
//   BenchmarkPipeline_Tier1Only       → Tier 1 ceiling (no AI scanner overhead)

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/refinery"
	"github.com/Edu963/ocultar/vault"
)

func newBenchEngine(b *testing.B) *refinery.Refinery {
	b.Helper()
	config.InitDefaults()
	v, err := vault.New(config.Settings{VaultBackend: "duckdb"}, ":memory:")
	if err != nil {
		b.Fatalf("vault: %v", err)
	}
	b.Cleanup(func() { v.Close() })
	return refinery.NewRefinery(v, []byte("01234567890123456789012345678901"))
}

// BenchmarkPipeline_SingleEmail is the baseline — one email address, minimal payload.
func BenchmarkPipeline_SingleEmail(b *testing.B) {
	eng := newBenchEngine(b)
	input := "Please contact alice@example.com for support."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = eng.RefineString(input, "bench", nil)
	}
}

// BenchmarkPipeline_MixedPII exercises Tier 1 + Tier 1.1 + Tier 1.2 together.
func BenchmarkPipeline_MixedPII(b *testing.B) {
	eng := newBenchEngine(b)
	input := "Invoice for John Smith (john.smith@acme.fr, +33 6 12 34 56 78). " +
		"Ship to 14 Boulevard MacDonald, 75019 Paris. " +
		"Card 4539578763621486, IBAN FR7630006000011234567890189."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = eng.RefineString(input, "bench", nil)
	}
}

// BenchmarkPipeline_FullDocument simulates a realistic 1 KB HR document.
func BenchmarkPipeline_FullDocument(b *testing.B) {
	eng := newBenchEngine(b)
	input := strings.Repeat(
		"Employee Jane Doe (jane.doe@corp.com, SSN 523-45-6789, +1 415 555 0100) "+
			"joined on 2026-01-15. Manager: Bob Smith (bob.smith@corp.com). "+
			"Salary deposited to IBAN GB29NWBK60161331926819. ",
		8, // ~560 chars × 8 ≈ 4.5 KB
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = eng.RefineString(input, "bench", nil)
	}
}

// BenchmarkPipeline_NoPII measures the cost of a full pipeline pass with no hits.
func BenchmarkPipeline_NoPII(b *testing.B) {
	eng := newBenchEngine(b)
	input := "The quarterly earnings report shows a 12% increase in revenue. " +
		"Product roadmap items are on track for Q3 delivery."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = eng.RefineString(input, "bench", nil)
	}
}

// BenchmarkPipeline_Parallel measures concurrent throughput at GOMAXPROCS goroutines.
func BenchmarkPipeline_Parallel(b *testing.B) {
	eng := newBenchEngine(b)
	input := "Contact ceo@bigcorp.com or call +33 1 42 00 00 00 for approval."
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = eng.RefineString(input, "bench", nil)
		}
	})
}

// BenchmarkPipeline_ScaleByLength measures how latency grows with payload size.
func BenchmarkPipeline_ScaleByLength(b *testing.B) {
	eng := newBenchEngine(b)
	for _, n := range []int{64, 256, 1024, 4096} {
		n := n
		b.Run(fmt.Sprintf("%dB", n), func(b *testing.B) {
			chunk := "User alice@x.com said hello. "
			input := strings.Repeat(chunk, n/len(chunk)+1)[:n]
			b.SetBytes(int64(n))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = eng.RefineString(input, "bench", nil)
			}
		})
	}
}
