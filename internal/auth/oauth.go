package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

const (
	defaultBackendURL = "https://stage1.doxantro.com"
	credentialsDir    = ".insighta"
	credentialsFile   = "credentials.json"
)

type Credentials struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	BackendURL   string `json:"backend_url"`
	Username     string `json:"username"`
}

func getCredentialsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, credentialsDir, credentialsFile)
}

func LoadCredentials() (*Credentials, error) {
	path := getCredentialsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("not logged in")
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}

	return &creds, nil
}

func SaveCredentials(creds *Credentials) error {
	path := getCredentialsPath()
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func ClearCredentials() error {
	path := getCredentialsPath()
	return os.Remove(path)
}

// Login performs the full OAuth PKCE flow
func Login() error {
	backendURL := os.Getenv("INSIGHTA_BACKEND_URL")
	if backendURL == "" {
		backendURL = defaultBackendURL
	}

	// Generate PKCE values
	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		return fmt.Errorf("failed to generate code verifier: %w", err)
	}
	codeChallenge := generateCodeChallenge(codeVerifier)

	// Generate state
	stateBytes := make([]byte, 16)
	rand.Read(stateBytes)
	state := hex.EncodeToString(stateBytes)

	// Start local callback server
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return fmt.Errorf("failed to start local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	// Channel to receive the callback result
	resultCh := make(chan *Credentials, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		accessToken := r.URL.Query().Get("access_token")
		refreshToken := r.URL.Query().Get("refresh_token")
		username := r.URL.Query().Get("username")

		if accessToken == "" || refreshToken == "" {
			errMsg := r.URL.Query().Get("error")
			if errMsg == "" {
				errMsg = "missing tokens in callback"
			}
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<html><body><h2>❌ Login failed: %s</h2><p>You can close this window.</p></body></html>`, errMsg)
			errCh <- fmt.Errorf(errMsg)
			return
		}

		creds := &Credentials{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			BackendURL:   backendURL,
			Username:     username,
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><body><h2>✓ Login successful!</h2><p>Logged in as <strong>@%s</strong>. You can close this window.</p></body></html>`, username)
		resultCh <- creds
	})

	server := &http.Server{Handler: mux}
	go server.Serve(listener)

	// Open browser
	authURL := fmt.Sprintf(
		"%s/auth/github?client=cli&port=%d&state=%s&code_challenge=%s",
		backendURL, port, state, codeChallenge,
	)

	fmt.Println("Opening browser for GitHub authentication...")
	fmt.Printf("If the browser doesn't open, visit:\n%s\n\n", authURL)
	openBrowser(authURL)

	// Wait for callback (timeout 2 minutes)
	select {
	case creds := <-resultCh:
		server.Close()
		if err := SaveCredentials(creds); err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}
		fmt.Printf("\n✓ Logged in as @%s\n", creds.Username)
		return nil
	case err := <-errCh:
		server.Close()
		return err
	case <-time.After(2 * time.Minute):
		server.Close()
		return fmt.Errorf("login timed out — no callback received")
	}
}

// Logout clears local credentials and invalidates server-side token
func Logout() error {
	creds, err := LoadCredentials()
	if err != nil {
		// Already logged out
		ClearCredentials()
		return nil
	}

	// Try to invalidate server-side
	client := &http.Client{Timeout: 10 * time.Second}
	body := fmt.Sprintf(`{"refresh_token": "%s"}`, creds.RefreshToken)
	req, _ := http.NewRequest("POST", creds.BackendURL+"/auth/logout",
		nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+creds.AccessToken)

	// Use a proper body
	req, _ = http.NewRequest("POST", creds.BackendURL+"/auth/logout",
		jsonReader(body))
	req.Header.Set("Content-Type", "application/json")
	client.Do(req) // Best effort — don't fail if server is unreachable

	return ClearCredentials()
}

// RefreshTokens refreshes the access token using the stored refresh token
func RefreshTokens() error {
	creds, err := LoadCredentials()
	if err != nil {
		return fmt.Errorf("not logged in")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	body := fmt.Sprintf(`{"refresh_token": "%s"}`, creds.RefreshToken)
	req, _ := http.NewRequest("POST", creds.BackendURL+"/auth/refresh",
		jsonReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh failed (HTTP %d)", resp.StatusCode)
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse refresh response: %w", err)
	}

	creds.AccessToken = result.AccessToken
	creds.RefreshToken = result.RefreshToken

	return SaveCredentials(creds)
}

// PKCE helpers
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// openBrowser opens the given URL in the default browser
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}

// jsonReader creates a reader from a JSON string
func jsonReader(s string) *jsonStringReader {
	return &jsonStringReader{s: s, i: 0}
}

type jsonStringReader struct {
	s string
	i int
}

func (r *jsonStringReader) Read(p []byte) (n int, err error) {
	if r.i >= len(r.s) {
		return 0, io.EOF
	}
	n = copy(p, r.s[r.i:])
	r.i += n
	return n, nil
}
