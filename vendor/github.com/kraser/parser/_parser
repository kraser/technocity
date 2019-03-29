// parser project parser.go
package parser

import (
	"math/rand"
	"time"
	"webreader"
)

const (
	ENDMESSAGE string = "LOAD_DONE"
)

type Parser struct {
	Options *webreader.RequestOptions
}

var ParserObj = new(Parser)
var isInited bool = false

func (pParser *Parser) init() {
	rand.Seed(time.Now().UnixNano())
	pParser.Options = webreader.GetOptions()
	pParser.SetUserAgent(useragents[rand.Intn(len(useragents))])

	isInited = true
}

func (pParser *Parser) SetUserAgent(ua string) {
	pParser.Options.UserAgent = ua
}

func GetParser() *Parser {
	if !isInited {
		ParserObj.init()
	}
	return ParserObj
}
