package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/semver"
)

const apiURL = "https://api.github.com/repos/diablofong/iVault/releases/latest"

// UpdateInfo is returned to the frontend via the update:available event.
type UpdateInfo struct {
	Available bool   `json:"available"`
	Version   string `json:"version"`
	URL       string `json:"url"`
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// Check fetches the latest GitHub release and compares it against currentVersion.
// Returns UpdateInfo{Available: false} silently on any network or parse error.
// Skips the check entirely when currentVersion == "dev".
func Check(currentVersion string) (UpdateInfo, error) {
	if currentVersion == "dev" || currentVersion == "" {
		return UpdateInfo{}, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return UpdateInfo{}, err
	}
	req.Header.Set("User-Agent", "iVault/"+currentVersion)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return UpdateInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UpdateInfo{}, fmt.Errorf("github api: status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return UpdateInfo{}, err
	}

	return compareVersions(release, currentVersion)
}

func compareVersions(release githubRelease, currentVersion string) (UpdateInfo, error) {
	latestTag := strings.TrimPrefix(release.TagName, "v")
	currentTag := strings.TrimPrefix(currentVersion, "v")

	latest, err := semver.NewVersion(latestTag)
	if err != nil {
		return UpdateInfo{}, fmt.Errorf("invalid latest version %q: %w", latestTag, err)
	}
	current, err := semver.NewVersion(currentTag)
	if err != nil {
		return UpdateInfo{}, fmt.Errorf("invalid current version %q: %w", currentTag, err)
	}

	return UpdateInfo{
		Available: latest.GreaterThan(current),
		Version:   release.TagName,
		URL:       release.HTMLURL,
	}, nil
}
