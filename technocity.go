// technocity project technocity.go
package main

import (
	"encoding/json"
	"errors"
	errs "errorshandler"
	"flag"
	"fmt"
	log "logger"
	"net/http"
	parsers "parser"
	"priceloader"
	"regexp"
	"strings"

	goquery "github.com/PuerkitoBio/goquery"
	html "golang.org/x/net/html"
)

var (
	logMode      string = "info"
	city         string = ""
	HTTP_HEADERS map[string]string
	URL          string = "https://www.technocity.ru"
)

/* Start implementation parser.InterfaceCustomParser */
type ParserActions struct {
	mainParser *parsers.ParserObject
}

//Implementation
func (pCcustomAct ParserActions) ParseCategories(html string) {
	log.Info("PARSE_CATEGORIES")
	dom, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	errs.ErrorHandle(err)

	catalog := dom.Find("ul.category-menu").First()
	pCcustomAct.readCategories(catalog, 0)
}

//Implementation
func (pCcustomAct ParserActions) CheckCategoriesTree(categories map[string]*priceloader.Category, level int) {
	for _, category := range categories {
		log.Debug("CHECK_LEVEL", level, category.Name, category.URL)
		if len(category.Categories) > 0 {
			pCcustomAct.CheckCategoriesTree(category.Categories, level+1)
		}
	}
}

//Implementation
func (pCcustomAct ParserActions) ParseItems(pCategory *priceloader.Category, htmlstr string) {
	re := regexp.MustCompile("\\s+")
	ws := " "
	clearHtml := re.ReplaceAllString(strings.Replace(htmlstr, "&nbsp;", " ", -1), ws)

	scrRegexp := regexp.MustCompile("TCSectionList\\.run\\s*\\((.*)\\);\\s*}\\)\\(\\);")
	scrStr := scrRegexp.FindStringSubmatch(clearHtml)
	if len(scrStr) < 1 {
		return
	}
	quotedStr := strings.Replace(scrStr[1], "'", "\"", -1)
	itemsRegexp := regexp.MustCompile(",\\s*items\\:\\s*([\\s\\S]*),\\s*shops")
	itemsStr := itemsRegexp.FindStringSubmatch(quotedStr)[1]
	log.CheckHtml(pCategory.URL, itemsStr, "debug")
	var info interface{}
	err := json.Unmarshal([]byte(itemsStr), &info)
	if err != nil {
		err := errors.New(strings.Join([]string{err.Error(), pCategory.URL}, " AT "))
		errs.ErrorHandle(err)
	}
	itemsStruct := info.([]interface{})
	var store string
	for num, itemInfo := range itemsStruct {
		item := itemInfo.(map[string]interface{})
		code := fmt.Sprint(item["ID"].(float64))
		name := html.UnescapeString(item["NAME"].(string))
		link := html.UnescapeString(item["NAME"].(string))
		price := int64(item["PRICE"].(map[string]interface{})["PRICE"].(float64))

		if item["ITEM_AVAILABLE"].(bool) {
			store = "Есть"
		} else {
			store = "0"
		}
		log.Info(code, name, link, price, store)
		log.Info(num)
		pItem := &priceloader.Item{Name: name, Code: code, URL: link, PriceRur: price, Store: store}
		pCategory.AddItem(pItem)
	}
	log.Info("Category:", pCategory.Name, len(pCategory.Items))
}

//Implementation
func (pCcustomAct ParserActions) ParserInit(parser *parsers.ParserObject) {
	pCcustomAct.mainParser = parser
	HTTP_HEADERS = map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language":           "ru,en-US;q=0.7,en;q=0.3",
		"Cache-Control":             "max-age=0",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
	}
	parser.Options.AddHeaders(HTTP_HEADERS)
	parser.Options.Trials = 5
	parser.Options.Interval = 3
	parser.Options.Preprocess = pCcustomAct.Preprocess
}

//Implementation
func (pCcustomAct ParserActions) CreateItemsUrl(url string) string {
	s := []string{url, "#flt-allShowed:1;sortBy:default"}
	log.Debug(strings.Join(s, ""))
	return strings.Join(s, "")
}

//Implementation
func (pCcustomAct ParserActions) ParserRun() {

}

/* End implementation parser.InterfaceCustomParser */

func (pCcustomAct ParserActions) Preprocess(pReq *http.Request) {
	if len(pReq.Cookies()) == 0 {
		log.Debug("NO_REQUEST_COOKIES")
	} else {
		pReq.Header.Del("Cookie")
	}
}

func (pCcustomAct ParserActions) readCategories(catalog *goquery.Selection, level int) {
	children := catalog.Children()
	for i := range children.Nodes {
		subCategoryNode := children.Eq(i)
		if goquery.NodeName(subCategoryNode) == "li" {
			anchor := subCategoryNode.Find("a").First()
			if len(anchor.Nodes) == 0 {
				log.Debug("NULL")
				continue
			}
			categoryName := html.UnescapeString(strings.TrimSpace(anchor.Text()))
			href, _ := anchor.Attr("href")
			log.Info("LEVEL", level, ":", categoryName, href)
			pCurrentCategory := priceloader.PriceList.SetCurrentCategory(categoryName, href, level)
			//log.Debug("FOR_LOAD", pCurrentCategory.Name, pCurrentCategory.URL)
			catalogBranch := subCategoryNode.Find("ul").First()
			if len(catalogBranch.Nodes) == 0 {
				pCurrentCategory.URL = strings.Join([]string{URL, pCurrentCategory.URL}, "")
				priceloader.PriceList.AddItemsCategory(pCurrentCategory)
			} else {
				pCcustomAct.readCategories(catalogBranch, level+1)
			}
		}
	}

}

func init() {
	flag.StringVar(&logMode, "lm", logMode, "режим логгирования")
	flag.StringVar(&city, "city", logMode, "город для которого разбирается прайс")

	//logMode = "debug"
}

func main() {
	flag.Parse()
	log.SetLogLevel(logMode)
	log.Info("LOGLEVEL", logMode)
	log.Info("START")
	custom := &parsers.ParserOptions{
		Name:           "TechnoCity",
		URL:            URL,
		Loaders:        20,
		LoaderCapacity: 5,
	}
	methods := ParserActions{}
	pParser := parsers.ParserObject{
		CustomParserOptions: custom,
		CustomParserActions: methods,
	}
	pParser.Run()
}
