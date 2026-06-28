package app

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const Version = "v0.1.0"

type Config struct {
	APIURL string
	Bearer string
}

type APIResponse struct {
	OK     bool   `json:"ok"`
	Action string `json:"action"`
	Status int    `json:"status"`
}

func configPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "nuke", "config.toml")
}

func loadConfig() Config {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return Config{}
	}

	cfg := Config{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		line = strings.Trim(line, `"`)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"`)

		switch key {
		case "api_url":
			cfg.APIURL = val
		case "bearer":
			cfg.Bearer = val
		}
	}
	return cfg
}

func saveConfig(apiURL, bearer string) error {
	if _, err := url.ParseRequestURI(apiURL); err != nil {
		return fmt.Errorf("invalid worker url: %s", apiURL)
	}

	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	content := fmt.Sprintf("api_url = %q\nbearer = %q\n", apiURL, bearer)
	return os.WriteFile(path, []byte(content), 0600)
}

func readInput(domain string) (string, error) {
	if domain != "" {
		return domain + "\n", nil
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	if stat.Mode()&os.ModeCharDevice != 0 {
		return "", fmt.Errorf("no input. use: nuke scan -d example.com OR cat domains.txt | nuke")
	}

	body, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func request(cfg Config, method, endpoint string, body string) ([]byte, int, error) {
	if cfg.APIURL == "" {
		return nil, 0, fmt.Errorf("api_url missing. run: nuke config -url https://worker.workers.dev -token YOUR_SECRET")
	}
	if cfg.Bearer == "" {
		return nil, 0, fmt.Errorf("bearer missing. run: nuke config -url https://worker.workers.dev -token YOUR_SECRET")
	}

	apiURL := strings.TrimRight(cfg.APIURL, "/") + endpoint

	req, err := http.NewRequest(method, apiURL, bytes.NewBufferString(body))
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Authorization", "Bearer "+cfg.Bearer)
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{Timeout: 5 * time.Minute}

	res, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	out, _ := io.ReadAll(res.Body)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return out, res.StatusCode, fmt.Errorf("server returned status %d", res.StatusCode)
	}

	return out, res.StatusCode, nil
}

func countDomains(input string) int {
	count := 0
	for _, line := range strings.Split(input, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

func scan(domain string) {
	cfg := loadConfig()

	input, err := readInput(domain)
	if err != nil {
		fail(err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		fail(fmt.Errorf("empty input"))
	}

	out, _, err := request(cfg, "POST", "/scan", input+"\n")
	if err != nil {
		fmt.Println(string(out))
		fail(err)
	}

	var apiResp APIResponse
	_ = json.Unmarshal(out, &apiResp)

	fmt.Println("✓ Scan submitted")
	fmt.Println()
	fmt.Println("Workflow started")
	fmt.Printf("Domains : %d\n", countDomains(input))
	fmt.Printf("Action  : %s\n", fallback(apiResp.Action, "scan"))
	fmt.Printf("Status  : %d\n", apiResp.Status)
	fmt.Println()
	fmt.Println("Use:")
	fmt.Println("  nuke summary")
}

func simpleGet(endpoint string) {
	cfg := loadConfig()
	out, _, err := request(cfg, "GET", endpoint, "")
	if err != nil {
		fmt.Println(string(out))
		fail(err)
	}
	fmt.Println(string(out))
}

func deletePath(path string) {
	cfg := loadConfig()
	body := fmt.Sprintf(`{"path":%q}`, path)

	out, _, err := request(cfg, "POST", "/delete", body)
	if err != nil {
		fmt.Println(string(out))
		fail(err)
	}

	fmt.Println("✓ Delete request submitted")
	fmt.Println(string(out))
}

func rerun() {
	cfg := loadConfig()
	out, _, err := request(cfg, "POST", "/rerun", "{}")
	if err != nil {
		fmt.Println(string(out))
		fail(err)
	}

	fmt.Println("✓ Rerun submitted")
	fmt.Println(string(out))
}

func status(id string) {
	if id == "" {
		fail(fmt.Errorf("missing run id. use: nuke status 123456789"))
	}
	simpleGet("/status?id=" + url.QueryEscape(id))
}

func configCmd(args []string) {
	fs := flag.NewFlagSet("config", flag.ExitOnError)
	apiURL := fs.String("url", "", "Cloudflare Worker URL")
	token := fs.String("token", "", "Bearer token")
	fs.Parse(args)

	if *apiURL == "" || *token == "" {
		fmt.Println("Usage:")
		fmt.Println("  nuke config -url https://YOUR-WORKER.workers.dev -token YOUR_SECRET")
		os.Exit(1)
	}

	if err := saveConfig(*apiURL, *token); err != nil {
		fail(err)
	}

	fmt.Println("✓ Config saved")
	fmt.Println(configPath())
}

func Run() {
	if len(os.Args) < 2 {
		fs := flag.NewFlagSet("nuke", flag.ExitOnError)
		domain := fs.String("d", "", "single domain")
		fs.Parse(os.Args[1:])
		scan(*domain)
		return
	}

	switch os.Args[1] {
	case "version":
		fmt.Println("Nuke " + Version)

	case "config":
		configCmd(os.Args[2:])

	case "scan":
		fs := flag.NewFlagSet("scan", flag.ExitOnError)
		domain := fs.String("d", "", "single domain")
		fs.Parse(os.Args[2:])
		scan(*domain)

	case "import":
		scan("")

	case "logs":
		simpleGet("/logs")

	case "results":
		simpleGet("/results")

	case "summary":
		simpleGet("/summary")

	case "status":
		id := ""
		if len(os.Args) > 2 {
			id = os.Args[2]
		}
		status(id)

	case "rerun":
		rerun()

	case "delete":
		if len(os.Args) < 3 {
			fail(fmt.Errorf("missing path. use: nuke delete nuke/log/file.log"))
		}
		deletePath(os.Args[2])

	default:
		fs := flag.NewFlagSet("nuke", flag.ExitOnError)
		domain := fs.String("d", "", "single domain")
		fs.Parse(os.Args[1:])
		scan(*domain)
	}
}

func fallback(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func fail(err error) {
	fmt.Println("error:", err)
	os.Exit(1)
}
