package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"llm-eval/internal/cache"
	"llm-eval/internal/provider"
	"llm-eval/internal/report"
	"llm-eval/internal/runner"
	"llm-eval/internal/suite"
)

func main() {
	var (
		suitePath = flag.String("suite", "examples/math.yaml", "Path to evaluation suite YAML or JSON file")
		model     = flag.String("model", "deepseek", "Model provider to use: deepseek or qwen")
		outDir    = flag.String("out", "reports", "Directory for generated reports")
		noCache   = flag.Bool("no-cache", false, "Disable local result cache")
		timeout   = flag.Duration("timeout", 60*time.Second, "Timeout for the whole evaluation run")
	)
	flag.Parse()

	if err := loadDotEnv(".env"); err != nil {
		log.Fatalf("load .env: %v", err)
	}

	evalSuite, err := suite.Load(*suitePath)
	if err != nil {
		log.Fatalf("load suite: %v", err)
	}

	var cacheStore cache.Store = cache.NewFileStore(".eval-cache")
	if *noCache {
		cacheStore = cache.NoopStore{}
	}

	modelProvider, err := provider.New(*model)
	if err != nil {
		log.Fatalf("create provider: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	result, err := runner.New(modelProvider, cacheStore).Run(ctx, evalSuite)
	if err != nil {
		log.Fatalf("run evaluation: %v", err)
	}

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		log.Fatalf("create report dir: %v", err)
	}
	if err := report.WriteJSON(result, *outDir+"/latest.json"); err != nil {
		log.Fatalf("write json report: %v", err)
	}
	if err := report.WriteMarkdown(result, *outDir+"/latest.md"); err != nil {
		log.Fatalf("write markdown report: %v", err)
	}

	fmt.Printf("Suite: %s\n", result.SuiteName)
	fmt.Printf("Cases: %d, Passed: %d, Failed: %d, Score: %.2f%%\n", result.Total, result.Passed, result.Failed, result.Score*100)
	fmt.Printf("Reports: %s/latest.json, %s/latest.md\n", *outDir, *outDir)
}

func loadDotEnv(path string) error {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key != "" && os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return nil
}
