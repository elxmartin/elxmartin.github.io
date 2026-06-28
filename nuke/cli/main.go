package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	APIURL string
	Bearer string
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "nuke", "config.txt")
}

func loadConfig() Config {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return Config{}
	}

	cfg := Config{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

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
	path := configPath()

	err := os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		return err
	}

	content := fmt.Sprintf("api_url=%s\nbearer=%s\n", apiURL, bearer)
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
		return "", fmt.Errorf("no input. use: nuke -d example.com OR cat domains.txt | nuke")
	}

	body, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func sendScan(cfg Config, domains string) error {
	if cfg.APIURL == "" {
		return fmt.Errorf("api_url missing. run: nuke config -url https://worker.workers.dev -token YOUR_SECRET")
	}
	if cfg.Bearer == "" {
		return fmt.Errorf("bearer missing. run: nuke config -url https://worker.workers.dev -token YOUR_SECRET")
	}

	apiURL := strings.TrimRight(cfg.APIURL, "/") + "/scan"

	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(domains))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+cfg.Bearer)
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	out, _ := io.ReadAll(res.Body)

	fmt.Println(string(out))

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("server returned status %d", res.StatusCode)
	}

	return nil
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "config" {
		fs := flag.NewFlagSet("config", flag.ExitOnError)
		url := fs.String("url", "", "Cloudflare Worker URL")
		token := fs.String("token", "", "Bearer token")
		fs.Parse(os.Args[2:])

		if *url == "" || *token == "" {
			fmt.Println("Usage:")
			fmt.Println("  nuke config -url https://YOUR-WORKER.workers.dev -token YOUR_SECRET")
			os.Exit(1)
		}

		if err := saveConfig(*url, *token); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}

		fmt.Println("[+] saved:", configPath())
		return
	}

	domain := flag.String("d", "", "single domain")
	flag.Parse()

	cfg := loadConfig()

	input, err := readInput(*domain)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Println("error: empty input")
		os.Exit(1)
	}

	if err := sendScan(cfg, input+"\n"); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}