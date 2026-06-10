package indexer

import (
	"fmt"
	"math/bits"
	"math/rand"
	"testing"
)

// phashString renders a hash value in the format goimagehash.ToString
// produces and the database stores ("p:%016x").
func phashString(v uint64) string {
	return fmt.Sprintf("p:%016x", v)
}

// TestHammingDistanceProperties checks the metric axioms of HammingDistance
// over random hash pairs, using math/bits.OnesCount64 as an independent
// oracle (differential testing: the stdlib popcount defines the right
// answer, no hand-written expected values needed).
func TestHammingDistanceProperties(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 500; i++ {
		a, b := rng.Uint64(), rng.Uint64()

		// Identity: d(a, a) == 0
		if d, err := HammingDistance(phashString(a), phashString(a)); err != nil || d != 0 {
			t.Fatalf("d(a,a) = %d, err=%v; want 0, nil (a=%016x)", d, err, a)
		}

		dAB, err := HammingDistance(phashString(a), phashString(b))
		if err != nil {
			t.Fatalf("HammingDistance failed: %v", err)
		}

		// Symmetry: d(a, b) == d(b, a)
		dBA, err := HammingDistance(phashString(b), phashString(a))
		if err != nil {
			t.Fatalf("HammingDistance failed: %v", err)
		}
		if dAB != dBA {
			t.Fatalf("symmetry violated: d(a,b)=%d d(b,a)=%d (a=%016x b=%016x)", dAB, dBA, a, b)
		}

		// Bounds: 0 <= d <= 64 for 64-bit hashes
		if dAB < 0 || dAB > 64 {
			t.Fatalf("d(a,b) = %d outside [0, 64]", dAB)
		}

		// Differential oracle: distance must equal popcount(a XOR b)
		if want := bits.OnesCount64(a ^ b); dAB != want {
			t.Fatalf("d(a,b) = %d, want popcount(a^b) = %d (a=%016x b=%016x)", dAB, want, a, b)
		}

		// AreSimilar must agree with the distance and the threshold
		for _, threshold := range []int{0, 10, 64} {
			similar, err := AreSimilar(phashString(a), phashString(b), threshold)
			if err != nil {
				t.Fatalf("AreSimilar failed: %v", err)
			}
			if want := dAB <= threshold; similar != want {
				t.Fatalf("AreSimilar(threshold=%d) = %v, want %v (d=%d)", threshold, similar, want, dAB)
			}
		}
	}
}

// TestHammingDistanceRejectsMalformedHashes covers the sad path: stored
// hashes that don't parse must produce errors, not silent zero distances.
func TestHammingDistanceRejectsMalformedHashes(t *testing.T) {
	t.Parallel()

	valid := phashString(0xdeadbeef)
	for _, bad := range []string{"", "nonsense", "p:", "p:zzzz", ":0123456789abcdef"} {
		if _, err := HammingDistance(bad, valid); err == nil {
			t.Errorf("HammingDistance(%q, valid) = nil error, want parse failure", bad)
		}
		if _, err := HammingDistance(valid, bad); err == nil {
			t.Errorf("HammingDistance(valid, %q) = nil error, want parse failure", bad)
		}
	}
}
