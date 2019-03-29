// options
package webreader

import (
	log "logger"
	"math/rand"
	"net/http"
	"time"
)

type RequestOptions struct {
	Url         string
	Method      string
	PostFields  map[string]Dyad
	CookieFile  string
	HttpHeaders map[string]Dyad
	UserAgent   string
	Trials      int
	Interval    time.Duration
	Preprocess  func(req *http.Request)
}

var currentOptions = new(RequestOptions)

func (options *RequestOptions) AddPostField(fieldName string, fieldValue string) {
	field := Dyad{KeyName: fieldName, Value: fieldValue}
	options.PostFields[fieldName] = field
}

func (options *RequestOptions) AddHeader(headerName string, headerValue string) {
	log.Debug(headerName+":", headerValue)
	header := Dyad{KeyName: headerName, Value: headerValue}
	options.HttpHeaders[headerName] = header
}

func (options *RequestOptions) AddHeaders(headers map[string]string) {
	log.Debug("HEADERS")
	for name, value := range headers {
		options.AddHeader(name, value)
	}
}

func (options *RequestOptions) SetRandUserAgent() {
	options.UserAgent = useragents[rand.Intn(len(useragents))]
	log.Info("UA:", options.UserAgent)
}

func GetOptions() *RequestOptions {
	currentOptions.PostFields = make(map[string]Dyad)
	currentOptions.HttpHeaders = make(map[string]Dyad)
	return currentOptions
}

func processOptions() {
	if len(currentOptions.Method) == 0 {
		currentOptions.Method = "GET"
	}

	if len(currentOptions.CookieFile) != 0 {
		cookieHandler.SetCookieFileName(currentOptions.CookieFile)
	}
}
