package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	defaultOutputFile = "/etc/telegraf/add_account_name.star" // Default output file for Starlark script
	defaultNatsMetricsURL = "http://localhost:8222" // Default NATS metrics URL
)

type AccountList struct {
	Accounts []string `json:"accounts"`
}

type AccountInfo struct {
	AccountDetail struct {
		DecodedJWT struct {
			Name string `json:"name"`
		} `json:"decoded_jwt"`
	} `json:"account_detail"`
}

func main() {
	output := flag.String("output", defaultOutputFile, "Path to output Starlark file")
    baseURL := flag.String("url", defaultNatsMetricsURL, "NATS metrics base URL (overrides NATS_METRICS_URL env var)")
    flag.Parse()
	
	accountMap, err := buildAccountMap(*baseURL)
	if err != nil {
		log.Printf("Error building account map: %v", err)
	} else {
	    if err := writeStarlarkFile(accountMap, *output); err != nil {
            log.Printf("Error writing Starlark file: %v", err)
        } else {
            log.Printf("Wrote %d account mappings to %s", len(accountMap), *output)
            fmt.Println("✅ Starlark file generated.")
        }
	}
}

func buildAccountMap(baseURL string) (map[string]string, error) {
	accountMap := make(map[string]string)
	ids, err := fetchAccountIDs(baseURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch NATS account IDs: %v", err)
	}

	for _, id := range ids {
		name, err := fetchAccountName(id, baseURL)
		if err != nil {
			log.Printf("Warning: failed to get info for %s: %v", id, err)
			continue
		}
		accountMap[id] = name
	}

	return accountMap, nil
}

func fetchAccountIDs(baseURL string) ([]string, error) {
	resp, err := http.Get(baseURL + "/accountz")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data AccountList
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data.Accounts, nil
}

func fetchAccountName(accountID string, baseURL string) (string, error) {
	url := fmt.Sprintf("%s/accountz?acc=%s", baseURL, accountID)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var info AccountInfo
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return "", err
	}

	if info.AccountDetail.DecodedJWT.Name == "" {
		return "unknown", nil
	}
	return info.AccountDetail.DecodedJWT.Name, nil
}


func writeStarlarkFile(mapping map[string]string, output string) error {
    tmpFile := output + ".tmp"
    outFile, err := os.Create(tmpFile)
    if err != nil {
        return err
    }
    defer outFile.Close()

    fmt.Fprintln(outFile, "def apply(metric):")
    fmt.Fprintln(outFile, "    mapping = {")
    for key, value := range mapping {
        fmt.Fprintf(outFile, "        \"%s\": \"%s\",\n", key, value)
    }
    fmt.Fprintln(outFile, "    }")
    fmt.Fprintln(outFile, "    if \"account\" in metric.tags:")
    fmt.Fprintln(outFile, "        account_id = metric.tags[\"account\"]")
    fmt.Fprintln(outFile, "        if account_id in mapping:")
    fmt.Fprintln(outFile, "            metric.tags[\"account_name\"] = mapping[account_id]")
    fmt.Fprintln(outFile, "    return metric")

    return os.Rename(tmpFile, output)
}
