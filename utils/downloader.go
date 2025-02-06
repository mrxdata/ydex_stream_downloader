package utils

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func DownloadFromStream(baseQuery QueryParams, headers Headers, intervalMs int, maxTimeout int, userhash, playlisthash, videohash, outputPath string) error {
	fileIndex := 0
	outputFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}

	defer func() {
		if err := outputFile.Close(); err != nil {
			log.Printf("Error closing file: %v\n", err)
		}
	}()

	client := &http.Client{
		Timeout: time.Duration(maxTimeout) * time.Second,
	}

	tryDownload := func(url string) (io.Reader, error) {
		var resp *http.Response
		for attempt := 1; attempt <= 3; attempt++ {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Printf("Attempt %d: Error occurred: %v\n", attempt, err)
				if attempt < 3 {
					log.Println("Retrying...")
					time.Sleep(5 * time.Second)
					continue
				}
				return nil, fmt.Errorf("failed to create request after %d attempts: %w", attempt, err)
			}

			for key, value := range map[string]string{
				"Accept":             headers.Accept,
				"Accept-Language":    headers.AcceptLanguage,
				"Sec-Ch-Ua":          headers.SecChUA,
				"Sec-Ch-Ua-Mobile":   headers.SecChUAMobile,
				"Sec-Ch-Ua-Platform": headers.SecChUAPlatform,
				"Sec-Fetch-Dest":     headers.SecFetchDest,
				"Sec-Fetch-Mode":     headers.SecFetchMode,
				"Sec-Fetch-Site":     headers.SecFetchSite,
				"Referer":            headers.Referer,
				"Referrer-Policy":    headers.ReferrerPolicy,
			} {
				req.Header.Set(key, value)
			}

			resp, err = client.Do(req)
			if err != nil {
				log.Printf("Attempt %d: Error occurred: %v\n", attempt, err)
				if attempt < 3 {
					log.Println("Retrying...")
					time.Sleep(time.Second)
					continue
				}
				return nil, fmt.Errorf("failed to download after %d attempts: %w", attempt, err)
			}

			if resp.StatusCode != http.StatusOK {
				log.Printf("Attempt %d: HTTP error: %s\n", attempt, resp.Status)
				if attempt < 3 {
					log.Println("Retrying...")
					err := resp.Body.Close()
					if err != nil {
						return nil, fmt.Errorf("failed to close response body: %w", err)
					}
					time.Sleep(time.Second)
					continue
				}
				return nil, fmt.Errorf("HTTP error after %d attempts: %s", attempt, resp.Status)
			}

			return resp.Body, nil
		}
		return nil, fmt.Errorf("failed to download after retries")
	}

	for {
		queryParams := baseQuery
		queryParams.T = time.Now().UnixMilli()

		url_ := BuildURL("", "", "hls", userhash, playlisthash, videohash, "720p", fileIndex, queryParams)
		log.Printf("Downloading %s\n", url_)

		body, err := tryDownload(url_)
		if err != nil {
			return fmt.Errorf("download failed: %w", err)
		}

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, body)
		if err != nil {
			return fmt.Errorf("failed to buffer response body: %w", err)
		}

		_, err = buf.WriteTo(outputFile)
		if err != nil {
			return fmt.Errorf("failed to write buffered data to file: %w", err)
		}

		log.Printf("Downloaded file %d.ts\n", fileIndex)
		fileIndex++

		progress, err := outputFile.Stat()
		if err == nil {
			log.Printf("Current output file size: %.2f MB\n", float64(progress.Size())/1024/1024)
		}

		sleepDuration := time.Duration(rand.Intn(intervalMs)) * time.Millisecond
		time.Sleep(sleepDuration)
	}

}
