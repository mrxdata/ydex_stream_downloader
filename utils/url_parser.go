package utils

import (
	"log"
	"net/url"
	"strings"
)

type UrlParser struct {
	UserHash     string
	PlaylistHash string
	VideoHash    string
	Vsid         string
	Vpuid        string
}

func ParseURL(url_ string) *UrlParser {
	urlParts := strings.Split(url_, "/")
	if len(urlParts) < 6 {
		log.Println("Ошибка: Неверный формат URL")
		return nil
	}

	urlParts[7], urlParts[8] = ParseQueryParams(urlParts[8])

	parser := &UrlParser{
		UserHash:     urlParts[4],
		PlaylistHash: urlParts[5],
		VideoHash:    urlParts[6],
		Vsid:         urlParts[7],
		Vpuid:        urlParts[8],
	}

	return parser
}

func ParseQueryParams(paramString string) (string, string) {
	parts := strings.Split(paramString, "?")
	if len(parts) < 2 {
		log.Println("Ошибка: Query параметры не найдены")
		return "", ""
	}

	params, err := url.ParseQuery(parts[1])
	if err != nil {
		log.Println("Ошибка при разборе query:", err)
		return "", ""
	}

	vsid := params.Get("vsid")
	index := strings.Index(vsid, "x")
	if index != -1 {
		vsid = vsid[:index]
	}
	vpuid := params.Get("vpuid")

	return vsid, vpuid
}
