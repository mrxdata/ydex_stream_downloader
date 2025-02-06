package utils

import (
	"fmt"
	"time"
)

func BuildURL(protocol, domain, tag, userhash, playlisthash, videohash, quality string, fileindex int, queryparams QueryParams) string {
	if protocol == "" {
		protocol = "https"
	}
	if domain == "" {
		domain = "streaming.disk.yandex.net"
	}
	if quality == "" {
		quality = "720p"
	}

	base := fmt.Sprintf("%s://%s/%s/%s/%s/%s/%s/%d.ts", protocol, domain, tag, userhash, playlisthash, videohash, quality, fileindex)
	return fmt.Sprintf("%s?vsid=%s&vpuid=%s&source_index=%d&session_data=%d&preview=%d&t=%d&ab=%d",
		base,
		queryparams.Vsid,
		queryparams.Vpuid,
		queryparams.SourceIndex,
		queryparams.SessionData,
		queryparams.Preview,
		queryparams.T,
		queryparams.Ab,
	)
}

func BuildQueryParams(hash, param1, param2, vpuid string, tunixStaticSeconds int64) QueryParams {
	if param1 == "" || param1 == "default" {
		param1 = "xWEB"
	}
	if param2 == "" || param2 == "default" {
		param2 = "x2402"
	}

	vsid := fmt.Sprintf("%s%s%sx%d", hash, param1, param2, tunixStaticSeconds)
	currentUnixMilliseconds := time.Now().UnixMilli()

	return QueryParams{
		Vsid:        vsid,
		Vpuid:       vpuid,
		SourceIndex: 0,
		SessionData: 1,
		Preview:     1,
		T:           currentUnixMilliseconds,
		Ab:          1,
	}
}
