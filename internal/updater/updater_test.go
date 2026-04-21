package updater

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockServer(t *testing.T, tagName, htmlURL string, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if status == http.StatusOK {
			_ = json.NewEncoder(w).Encode(githubRelease{TagName: tagName, HTMLURL: htmlURL})
		}
	}))
}

func checkWithServer(srv *httptest.Server, currentVersion string) (UpdateInfo, error) {
	return checkURL(srv.URL, currentVersion)
}

// checkURL is the testable version of Check with an injectable URL.
func checkURL(url, currentVersion string) (UpdateInfo, error) {
	if currentVersion == "dev" || currentVersion == "" {
		return UpdateInfo{}, nil
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return UpdateInfo{}, err
	}
	req.Header.Set("User-Agent", "iVault/"+currentVersion)

	resp, err := client.Do(req)
	if err != nil {
		return UpdateInfo{}, err
	}
	defer resp.Body.Close()

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return UpdateInfo{}, err
	}

	return compareVersions(release, currentVersion)
}

func TestCheck_NewerVersionAvailable(t *testing.T) {
	srv := mockServer(t, "v1.2.0", "https://github.com/diablofong/iVault/releases/tag/v1.2.0", 200)
	defer srv.Close()

	info, err := checkWithServer(srv, "v1.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.Available {
		t.Fatal("expected Available=true, got false")
	}
	if info.Version != "v1.2.0" {
		t.Fatalf("expected Version=v1.2.0, got %q", info.Version)
	}
}

func TestCheck_SameVersion(t *testing.T) {
	srv := mockServer(t, "v1.1.0", "https://github.com/diablofong/iVault/releases/tag/v1.1.0", 200)
	defer srv.Close()

	info, err := checkWithServer(srv, "v1.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Available {
		t.Fatal("expected Available=false for same version")
	}
}

func TestCheck_OlderRelease(t *testing.T) {
	srv := mockServer(t, "v1.0.0", "https://github.com/diablofong/iVault/releases/tag/v1.0.0", 200)
	defer srv.Close()

	info, err := checkWithServer(srv, "v1.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Available {
		t.Fatal("expected Available=false when release is older than current")
	}
}

func TestCheck_DevVersionSkipped(t *testing.T) {
	info, err := Check("dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Available {
		t.Fatal("dev builds should never report update available")
	}
}
