package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	urlFile    = "tools/download_icons/icon_urls.txt"
	baseURL    = "https://wow.zamimg.com/images/wow/icons"
	proxyAddr  = "http://127.0.0.1:10808"
	destBase   = "dist/wotlk/assets/icons"
	numWorkers = 20
)

type downloadTask struct {
	size     string
	iconFile string
}

func main() {
	// Read URL list
	f, err := os.Open(urlFile)
	if err != nil {
		fmt.Println("Error opening url file:", err)
		os.Exit(1)
	}
	defer f.Close()

	var tasks []downloadTask
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "\t", 2)
		if len(parts) == 2 {
			tasks = append(tasks, downloadTask{size: parts[0], iconFile: parts[1]})
		}
	}

	fmt.Printf("Total tasks: %d\n", len(tasks))

	// Create HTTP client with proxy
	proxyURL, _ := url.Parse(proxyAddr)
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	// Channel for tasks
	taskCh := make(chan downloadTask, 100)
	var wg sync.WaitGroup

	var ok, skip, fail, done int64

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for task := range taskCh {
				destDir := filepath.Join(destBase, task.size)
				os.MkdirAll(destDir, 0755)
				destPath := filepath.Join(destDir, task.iconFile)

				// Skip if already exists and non-empty
				if info, err := os.Stat(destPath); err == nil && info.Size() > 0 {
					atomic.AddInt64(&skip, 1)
				} else {
					url := fmt.Sprintf("%s/%s/%s", baseURL, task.size, task.iconFile)
					req, _ := http.NewRequest("GET", url, nil)
					req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
					req.Header.Set("Referer", "https://www.wowhead.com/")
					req.Header.Set("Accept", "image/webp,image/apng,image/*,*/*;q=0.8")

					success := false
					for attempt := 0; attempt < 3; attempt++ {
						resp, err := client.Do(req)
						if err != nil {
							if attempt < 2 {
								time.Sleep(time.Second)
								continue
							}
							break
						}

						if resp.StatusCode == 200 {
							data, _ := io.ReadAll(resp.Body)
							if len(data) > 0 {
								os.WriteFile(destPath, data, 0644)
								success = true
							}
						} else if resp.StatusCode == 404 {
							// Create empty placeholder
							os.WriteFile(destPath, []byte{}, 0644)
							break
						}
						resp.Body.Close()

						if success {
							break
						}
						if attempt < 2 {
							time.Sleep(time.Second)
						}
					}

					if success {
						atomic.AddInt64(&ok, 1)
					} else {
						atomic.AddInt64(&fail, 1)
					}
				}

				current := atomic.AddInt64(&done, 1)
				if current%500 == 0 || current == int64(len(tasks)) {
					fmt.Printf("Progress: %d/%d (ok=%d skip=%d fail=%d)\n",
						current, len(tasks),
						atomic.LoadInt64(&ok),
						atomic.LoadInt64(&skip),
						atomic.LoadInt64(&fail))
				}
			}
		}(i)
	}

	// Feed tasks
	for _, task := range tasks {
		taskCh <- task
	}
	close(taskCh)

	wg.Wait()

	fmt.Printf("\n=== Download Complete ===\n")
	fmt.Printf("  Total: %d\n", len(tasks))
	fmt.Printf("  Downloaded: %d\n", ok)
	fmt.Printf("  Skipped: %d\n", skip)
	fmt.Printf("  Failed: %d\n", fail)
}
