package querygen

import (
	"fmt"
	"strings"
)

func normalizeSpannerExternalDatasetVerificationStatus(status string) (string, error) {
	switch strings.ToLower(emptyDefault(status, "not_checked")) {
	case "not_checked":
		return "not_checked", nil
	case "verified":
		return "verified", nil
	case "mismatch":
		return "mismatch", nil
	case "failed":
		return "failed", nil
	default:
		return "", fmt.Errorf("unsupported access verification status %q; use not_checked, verified, mismatch, or failed", status)
	}
}

func normalizeSpannerExternalDatasetVerificationSource(source, status string) (string, error) {
	source = strings.ToLower(strings.TrimSpace(source))
	if source == "" {
		if status == "" || status == "not_checked" {
			return "", nil
		}
		return "external_evidence", nil
	}
	switch source {
	case "user_config":
		if status != "" && status != "not_checked" {
			return "", fmt.Errorf("access verification source user_config is only valid with not_checked status")
		}
		return source, nil
	case "external_evidence", "live_probe":
		return source, nil
	default:
		return "", fmt.Errorf("unsupported access verification source %q; use user_config, external_evidence, or live_probe", source)
	}
}

func containsName(names []string, name string) bool {
	for _, value := range names {
		if strings.EqualFold(value, name) {
			return true
		}
	}
	return false
}
