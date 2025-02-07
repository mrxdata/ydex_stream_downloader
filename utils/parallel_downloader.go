package utils

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

// DownloadResult хранит результат загрузки одного фрагмента.
type DownloadResult struct {
	Index int
	Data  []byte
	Err   error
}

// ParallelDownloadFromStream скачивает фрагменты параллельно, каждая горутина (worker)
// отправляет свои результаты в свой канал, а общий коллектор записывает данные в итоговый файл
// строго по порядку (0.ts, 1.ts, 2.ts, ...).
func ParallelDownloadFromStream(
	baseQuery QueryParams,
	headers Headers,
	intervalMs, maxTimeout, numWorkers int,
	userhash, playlisthash, videohash, outputPath string,
) error {
	start := time.Now()

	outputFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer func() {
		if err := outputFile.Close(); err != nil {
			log.Printf("failed to close output file: %v", err)
		}
	}()

	client := &http.Client{
		Timeout: time.Duration(maxTimeout) * time.Second,
	}

	workerChans := make([]chan DownloadResult, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workerChans[i] = make(chan DownloadResult, 20)
	}

	var wg sync.WaitGroup

	downloadChunk := func(url string) ([]byte, error) {

		var resp *http.Response

		for attempt := 1; attempt <= 3; attempt++ {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Printf("Attempt %d: failed to create request: %v\n", attempt, err)
				time.Sleep(3 * time.Second)
				continue
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
				log.Printf("Attempt %d: error downloading: %v\n", attempt, err)
				time.Sleep(time.Second)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				log.Printf("Attempt %d: HTTP error: %s\n", attempt, resp.Status)
				err := resp.Body.Close()
				if err != nil {
					return nil, fmt.Errorf("failed to close response body: %w", err)
				}
				time.Sleep(time.Second)
				continue
			}

			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, resp.Body)
			err = resp.Body.Close()
			if err != nil {
				log.Printf("Attempt %d: failed to close body: %v\n", attempt, err)
				time.Sleep(time.Second)
				continue
			}

			return buf.Bytes(), nil
		}
		return nil, fmt.Errorf("failed to download after retries")
	}

	// Worker: каждый воркер начинает с индекса startIndex и скачивает фрагменты с шагом = numWorkers.
	worker := func(startIndex int, ch chan DownloadResult) {
		defer wg.Done()
		chunkIndex := startIndex
		for {
			queryParams := baseQuery
			queryParams.T = time.Now().UnixMilli()
			url := BuildURL("", "", "hls", userhash, playlisthash, videohash, "720p", chunkIndex, queryParams)
			log.Printf("Worker %d: downloading %s\n", startIndex, url)

			data, err := downloadChunk(url)
			if err != nil {
				ch <- DownloadResult{Index: chunkIndex, Data: nil, Err: err}
				log.Printf("Worker %d: stopping at index %d due to error: %v\n", startIndex, chunkIndex, err)
				close(ch)
				return
			}

			ch <- DownloadResult{Index: chunkIndex, Data: data, Err: nil}
			log.Printf("Worker %d: downloaded %d.ts\n", startIndex, chunkIndex)
			chunkIndex += numWorkers
			time.Sleep(time.Duration(rand.Intn(intervalMs)) * time.Millisecond)
		}
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(i, workerChans[i])
	}

	go func() {
		wg.Wait()
		for i := 0; i < numWorkers; i++ {
			close(workerChans[i])
		}
	}()

	expectedIndex := 0
	for {
		workerIdx := expectedIndex % numWorkers
		res, ok := <-workerChans[workerIdx]
		if !ok {
			log.Printf("Worker %d channel closed, stopping collector at expected index %d\n", workerIdx, expectedIndex)
			break
		}
		if res.Index != expectedIndex {
			log.Printf("Out of order: expected %d, got %d\n", expectedIndex, res.Index)
			return fmt.Errorf("out of order result: expected %d, got %d", expectedIndex, res.Index)
		}
		if res.Err != nil {
			log.Printf("Error at index %d: %v. Ending file write.\n", expectedIndex, res.Err)
			break
		}
		_, err = outputFile.Write(res.Data)
		if err != nil {
			return fmt.Errorf("failed to write data for index %d: %w", expectedIndex, err)
		}
		log.Printf("Wrote chunk %d to output file.\n", expectedIndex)
		expectedIndex++
	}
	end := time.Now()
	log.Printf("Time elapsed: %v\n", end.Sub(start))
	return nil
}
