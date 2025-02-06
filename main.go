package main

import (
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv"
	"io"
	"log"
	"os"
	"strings"
	"time"
	"ydxstream_downloader/utils"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var (
		fileName   = os.Getenv("MERGEDFILE_NAME")
		urlToParse = os.Getenv("STREAM_URL")
	)

	// LOGGER SETUP
	logDir := "./logs"
	err = os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		fmt.Println("Ошибка при создании папки для логов:", err)
		return
	}

	currentTime := time.Now().Format("2006-01-02_15-04-05")
	logFileName := fmt.Sprintf("%s/logfile_%s.log", logDir, currentTime)

	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Ошибка при открытии файла для записи логов:", err)
		return
	}

	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			fmt.Printf("Error closing log file: %s\n", err)
		}
	}(logFile)

	multiWriter := io.MultiWriter(os.Stdout, logFile)

	log.SetOutput(multiWriter)

	log.Println("Starting download, output file:", fileName+".ts")
	// MAIN LOGIC

	parsedUrl := utils.ParseURL(urlToParse)

	headers := utils.Headers{
		Accept:          "*/*",
		AcceptLanguage:  "ru-RU,ru;q=0.9",
		SecChUA:         `"Not A(Brand";v="8", "Chromium";v="132", "Google Chrome";v="132"`,
		SecChUAMobile:   "?0",
		SecChUAPlatform: `"Windows"`,
		SecFetchDest:    "empty",
		SecFetchMode:    "cors",
		SecFetchSite:    "cross-site",
		Referer:         "https://disk.yandex.ru/",
		ReferrerPolicy:  "strict-origin-when-cross-origin",
	}

	queryParams := utils.BuildQueryParams(
		parsedUrl.Vsid,
		"",
		"",
		parsedUrl.Vpuid,
		time.Now().Unix()-1,
	)

	err = utils.DownloadFromStream(
		queryParams,
		headers,
		200,
		20,
		parsedUrl.UserHash,
		parsedUrl.PlaylistHash,
		parsedUrl.VideoHash,
		strings.Join([]string{"output", fileName + ".ts"}, "/"),
	)
	if err != nil {
		log.Println("Error:", err)
	} else {
		log.Println("Download complete.")
	}
}
