// parser
package parser

import (
	errs "errorshandler"
	"webreader"

	log "logger"
	"math/rand"

	"os"
	"os/signal"
	"priceloader"
	"time"
)

const (
	ENDMESSAGE = "PARSE_DONE"
)

type InterfaceCustomParser interface {
	ParseCategories(param string)
	CheckCategoriesTree(map[string]*priceloader.Category, int)
	CreateItemsUrl(url string) string
	ParseItems(pCategory *priceloader.Category, param string)
	ParserInit(*ParserObject)
	ParserRun()
	//HandleError()
}

type ParserOptions struct {
	Name           string
	URL            string
	Loaders        int
	LoaderCapacity int
	Trials         int
}

type ParserObject struct {
	Options             *webreader.RequestOptions
	CustomParserOptions *ParserOptions
	CustomParserActions InterfaceCustomParser
}

func (pParser *ParserObject) init() {
	log.Info("PARSER_INIT_START")
	priceloader.PriceList.PriceList(pParser.CustomParserOptions.Name)
	rand.Seed(time.Now().UnixNano())
	pParser.Options = webreader.GetOptions()
	pParser.Options.SetRandUserAgent()

	pParser.Options.Url = pParser.CustomParserOptions.URL
	pParser.CustomParserActions.ParserInit(pParser)
	log.Info("PARSER_INIT_DONE")
}

func (pParser *ParserObject) Run() {
	pParser.init()

	result, err := webreader.DoRequest(pParser.CustomParserOptions.URL, pParser.Options)
	errs.ErrorHandle(err)
	log.CheckHtml(pParser.CustomParserOptions.URL, result, "debug")
	pParser.CustomParserActions.ParseCategories(result)
	pParser.CustomParserActions.CheckCategoriesTree(priceloader.PriceList.Categories, 0)

	//Подготовим каналы и регулятор
	taskChan := make(chan priceloader.LoadTask)
	quitChan := make(chan bool)
	pController := &LoadController{Loaders: pParser.CustomParserOptions.Loaders, LoaderCapacity: pParser.CustomParserOptions.LoaderCapacity}
	pController.init(taskChan)
	log.Info(len(priceloader.PriceList.ItemsCategories))

	//Приготовимся перехватывать сигнал останова в канал keys
	keys := make(chan os.Signal, 1)
	signal.Notify(keys, os.Interrupt)

	go pController.balance(quitChan)
	go pParser.taskGenerator(taskChan)

	log.Info("MAIN_CYCLE_START")
	//Основной цикл программы:
	for {
		select {
		case <-keys: //пришла информация от нотификатора сигналов:
			log.Info("CTRL-C: Ожидаю завершения активных загрузок")
			quitChan <- true //посылаем сигнал останова балансировщику

		case <-quitChan: //пришло подтверждение о завершении от балансировщика
			log.Info("MAIN_CYCLE_END")
			log.Info(len(priceloader.PriceList.ItemsCategories))
			for _, pCategory := range priceloader.PriceList.ItemsCategories {
				log.Info(pCategory.Name, pCategory.URL)
			}
			return
		}
	}

}

func (pParser *ParserObject) taskGenerator(out chan priceloader.LoadTask) {
	pPriceList := priceloader.PriceList
	var toLoad bool
	toLoad = true
	for toLoad {
		for _, category := range pPriceList.ItemsCategories {
			task := priceloader.LoadTask{Pointer: category, Handler: pParser.LoadItems, Message: "TASK"}
			out <- task
		}
		toLoad = len(pPriceList.ItemsCategories) > 0
	}
	endTask := priceloader.LoadTask{Pointer: nil, Message: ENDMESSAGE}
	out <- endTask
}

func (pParser *ParserObject) LoadItems(pCategory *priceloader.Category) {
	url := pParser.CustomParserActions.CreateItemsUrl(pCategory.URL)
	result, err := webreader.DoRequest(url, pParser.Options)
	if err == nil {
		priceloader.PriceList.DeleteItemsCategory(pCategory)
	}
	log.CheckHtml(url, result, "debug")
	pParser.CustomParserActions.ParseItems(pCategory, result)
}

func LoadAndParse(itemLoadTask priceloader.LoadTask) {
	itemLoadTask.Handler(itemLoadTask.Pointer)
}
