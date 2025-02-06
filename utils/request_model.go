package utils

type Headers struct {
	Accept          string `json:"accept"`
	AcceptLanguage  string `json:"accept-language"`
	SecChUA         string `json:"sec-ch-ua"`
	SecChUAMobile   string `json:"sec-ch-ua-mobile"`
	SecChUAPlatform string `json:"sec-ch-ua-platform"`
	SecFetchDest    string `json:"sec-fetch-dest"`
	SecFetchMode    string `json:"sec-fetch-mode"`
	SecFetchSite    string `json:"sec-fetch-site"`
	Referer         string `json:"Referer"`
	ReferrerPolicy  string `json:"Referrer-Policy"`
}

type QueryParams struct {
	Vsid        string `json:"vsid"`
	Vpuid       string `json:"vpuid"`
	SourceIndex int    `json:"source_index"`
	SessionData int    `json:"session_data"`
	Preview     int    `json:"preview"`
	T           int64  `json:"t"`
	Ab          int    `json:"ab"`
}
