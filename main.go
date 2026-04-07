package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/playwright-community/playwright-go"
)

type CrawlRequest struct {
	URLs []string `json:"urls"`
}

type CrawlResponse struct {
	URL     string `json:"url"`
	FileURL string `json:"file_url,omitempty"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

func main() {
	if err := playwright.Install(); err != nil {
		slog.Error("Error installing Playwright", "error", err)
		os.Exit(1)
	}

	pw, err := playwright.Run()
	if err != nil {
		slog.Error("Error Playwright", "error", err)
		os.Exit(1)
	}

	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		slog.Error("Error launching browser", "error", err)
		os.Exit(1)
	}
	defer browser.Close()

	app := fiber.New()
	app.Get("/outputs/*", static.New("./outputs"))

	app.Post("/api/crawl", func(c fiber.Ctx) error {
		startTime := time.Now()

		var req CrawlRequest
		if err := c.Bind().Body(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    fiber.StatusBadRequest,
				"status":  "Bad Request",
				"message": "Invalid JSON payload",
				"data":    nil,
			})
		}

		if len(req.URLs) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    fiber.StatusBadRequest,
				"status":  "Bad Request",
				"message": "URLs list cannot be empty",
				"data":    nil,
			})
		}

		responses := make([]CrawlResponse, len(req.URLs))
		var wg sync.WaitGroup

		for i, url := range req.URLs {
			wg.Add(1)
			go func(index int, targetURL string) {
				defer wg.Done()
				filename, err := crawlURL(browser, targetURL)
				if err != nil {
					responses[index] = CrawlResponse{
						URL:    targetURL,
						Status: "error",
						Error:  err.Error(),
					}
				} else {
					responses[index] = CrawlResponse{
						URL:     targetURL,
						FileURL: fmt.Sprintf("%s/outputs/%s", c.BaseURL(), filename),
						Status:  "success",
					}
				}
			}(i, url)
		}

		wg.Wait()

		successCount := 0
		failedCount := 0
		for _, res := range responses {
			if res.Status == "success" {
				successCount++
			} else {
				failedCount++
			}
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"code":    fiber.StatusOK,
			"status":  "OK",
			"message": "Crawling process completed",
			"data":    responses,
			"meta": fiber.Map{
				"total_urls":     len(req.URLs),
				"success_count":  successCount,
				"failed_count":   failedCount,
				"execution_time": time.Since(startTime).String(),
			},
		})
	})

	if err := app.Listen(":5000"); err != nil {
		slog.Error("Error starting server", "error", err)
	}

}

func crawlURL(browser playwright.Browser, url string) (string, error) {
	page, err := browser.NewPage()
	if err != nil {
		return "", err
	}
	defer page.Close()

	page.SetDefaultNavigationTimeout(45000)
	page.SetDefaultTimeout(45000)

	if _, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		return "", err
	}

	htmlContent, err := page.Content()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll("outputs", os.ModePerm); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s_%d.html", sanitizeURLForFile(url), time.Now().Unix())
	savedPath := "outputs/" + filename

	if err := os.WriteFile(savedPath, []byte(htmlContent), 0644); err != nil {
		return "", err
	}

	return filename, nil

}

func sanitizeURLForFile(rawURL string) string {
	name := strings.TrimPrefix(rawURL, "https://")
	name = strings.TrimPrefix(name, "http://")
	return name
}
