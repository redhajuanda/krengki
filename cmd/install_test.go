package cmd

import "testing"

func TestSelectCreateNextAppTagPrefersLTS(t *testing.T) {
	tag, err := selectCreateNextAppTag(map[string]string{
		"latest": "16.2.9",
		"lts":    "15.3.9",
	})
	if err != nil {
		t.Fatalf("selectCreateNextAppTag returned error: %v", err)
	}
	if tag != "lts" {
		t.Fatalf("tag = %q, want %q", tag, "lts")
	}
}

func TestSelectCreateNextAppTagUsesHighestStablePreviousMajor(t *testing.T) {
	tag, err := selectCreateNextAppTag(map[string]string{
		"latest":    "16.2.9",
		"canary":    "16.3.0-canary.48",
		"preview":   "16.3.0-preview.3",
		"beta":      "16.0.0-beta.0",
		"next-14":   "14.2.35",
		"next-15-0": "15.1.12",
		"next-15-2": "15.2.9",
		"next-15-3": "15.3.9",
	})
	if err != nil {
		t.Fatalf("selectCreateNextAppTag returned error: %v", err)
	}
	if tag != "next-15-3" {
		t.Fatalf("tag = %q, want %q", tag, "next-15-3")
	}
}

func TestSelectCreateNextAppTagRejectsMissingPreviousStable(t *testing.T) {
	_, err := selectCreateNextAppTag(map[string]string{
		"latest": "16.2.9",
		"canary": "16.3.0-canary.48",
	})
	if err == nil {
		t.Fatal("selectCreateNextAppTag returned nil error")
	}
}
