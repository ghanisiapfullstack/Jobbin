package helpers

import (
	"strings"
	"testing"
)

// TestSanitizeStringPreservesLongURL is a regression test for the bug where
// long LinkedIn/job-board apply URLs (which frequently exceed 500 characters
// due to tracking query params) failed to save. The DB column was widened to
// TEXT; this test guards the application layer so sanitization does not
// truncate or corrupt a long, valid URL.
func TestSanitizeStringPreservesLongURL(t *testing.T) {
	longURL := "https://www.linkedin.com/jobs/view/4012345678/?alternateChannel=search&refId=" +
		strings.Repeat("a", 500) + "&trackingId=" + strings.Repeat("b", 120) +
		"&trk=flagship3_search_srp_jobs"

	if len(longURL) <= 500 {
		t.Fatalf("test fixture should exceed 500 chars, got %d", len(longURL))
	}

	got := SanitizeString(longURL)

	if got != longURL {
		t.Fatalf("SanitizeString altered a long valid URL.\n want len=%d\n got  len=%d", len(longURL), len(got))
	}
}

// TestSanitizeStringStripsHTMLInURLContext ensures sanitization still removes
// embedded HTML even inside long inputs (XSS defense must not regress).
func TestSanitizeStringStripsHTMLInURLContext(t *testing.T) {
	malicious := "https://example.com/?q=" + strings.Repeat("x", 600) + "<script>alert(1)</script>"
	got := SanitizeString(malicious)

	if strings.Contains(got, "<script>") {
		t.Fatal("SanitizeString failed to strip <script> from long input")
	}
}
