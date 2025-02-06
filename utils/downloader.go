package utils

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func DownloadFromStream(baseQuery QueryParams, headers Headers, intervalMs int, maxTimeout int, userhash, playlisthash, videohash, outputPath string) error {
	fileIndex := 209
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}

	defer func() {
		if err := outputFile.Close(); err != nil {
			log.Printf("Error closing file: %v\n", err)
		}
	}()

	client := &http.Client{
		Timeout: time.Duration(maxTimeout) * time.Second,
	}

	for {
		queryParams := baseQuery
		queryParams.T = time.Now().UnixMilli()

		url_ := BuildURL("", "", "hls", userhash, playlisthash, videohash, "720p", fileIndex, queryParams)
		log.Printf("Downloading %s\n", url_)
		req, err := http.NewRequest("GET", url_, nil)
		if err != nil {
			return fmt.Errorf("failed to create utils: %w", err)
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

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Timeout or error occurred after %d files: %v\n", fileIndex, err)
			break
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("HTTP error after %d files: %s\n", fileIndex, resp.Status)
			break
		}

		_, err = io.Copy(outputFile, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to write to output file: %w", err)
		}

		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing resp body: %v\n", err)
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

	return nil
}
