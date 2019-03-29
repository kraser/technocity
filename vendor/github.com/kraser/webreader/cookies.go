// cookies
package webreader

import (
	errs "errorshandler"
	"logger"
	"net/http"
	"os"
)

type Cookies struct {
	Cookies            []*http.Cookie
	CookiesFileName    string
	cookiesFileHandler *os.File
}

func (handler *Cookies) SaveCookies(resp *http.Response) {
	handler.Cookies = resp.Cookies()
	logger.Debug("COOKIEFILE", currentOptions.CookieFile)
	if len(handler.CookiesFileName) != 0 {
		cookiesFileHandler, err := os.OpenFile(currentOptions.CookieFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		errs.ErrorHandle(err)
		defer cookiesFileHandler.Close()
		cookiesFileHandler.Truncate(0)
		logger.Debug("RESPONSE_COOKIES")
		for _, value := range handler.Cookies {
			logger.Debug(value.String())
			cookiesFileHandler.WriteString(value.String() + "\n")
		}
	}
}

func (handler *Cookies) SetCookieFileName(fileName string) {
	handler.CookiesFileName = fileName
}
