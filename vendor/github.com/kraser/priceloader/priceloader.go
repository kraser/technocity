// priceloader project priceloader.go
package priceloader

import (
	//"fmt"
	"crypto/md5"
	"encoding/hex"
	log "logger"
)

type Item struct {
	Id          int32
	SupplierId  int32
	ReferenceId int32
	Name        string
	Code        string
	Model       string
	Brand       string
	Price       int64
	PriceRur    int64
	PriceUsd    int64
	Store       string
	StoreNsk    string
	StoreMsk    string
	URL         string
}

type Category struct {
	Id         int32
	Name       string
	URL        string
	Categories map[string]*Category
	Items      map[int]*Item
}

func (pCat *Category) AddItem(pItem *Item) {
	pCat.Items[len(pCat.Items)+1] = pItem
}

type Price struct {
	SupplierId      int32
	supplierCode    string
	Categories      map[string]*Category
	CategoryStack   []*Category
	ItemsCategories map[string]*Category
	curLevel        int
}

type LoadTask struct {
	Pointer *Category
	Handler func(*Category)
	Message string
}

var PriceList = new(Price)
var pCurrentCategory *Category

func (price *Price) PriceList(supplierCode string) {
	price.supplierCode = supplierCode
	price.Categories = make(map[string]*Category)
	price.CategoryStack = make([]*Category, 8)
	price.ItemsCategories = make(map[string]*Category)

}

func (price *Price) SetCurrentCategory(name string, url string, level int) *Category {
	if level < 0 {
		level = 0
	}

	if level == 0 {
		pCurrentCategory = price.createAndAddCategory(name, url)
	} else {
		var pCurCategory *Category
		if level > price.curLevel {
			pCurCategory = price.CategoryStack[price.curLevel]
		} else {
			pCurCategory = price.CategoryStack[level-1]
		}

		category := Category{Id: 0, Name: name, URL: url}
		category.Categories = make(map[string]*Category)
		category.Items = make(map[int]*Item)
		pCurrentCategory = &category
		pCurCategory.Categories[name] = pCurrentCategory
		PriceList.CategoryStack[level] = pCurrentCategory

	}
	price.curLevel = level
	log.Debug("CREATED:", pCurrentCategory.Name, pCurrentCategory.URL)
	return pCurrentCategory
}

func (price *Price) createAndAddCategory(name string, url string) *Category {
	category := Category{Id: 0, Name: name, URL: url}
	category.Categories = make(map[string]*Category)
	category.Items = make(map[int]*Item)
	pCategory := &category
	price.CategoryStack[0] = pCategory
	price.Categories[name] = pCategory
	return pCategory
}

func (price *Price) AddItem(pCategory *Category, pItem *Item) {
	pCategory.AddItem(pItem)
}

func (price *Price) AddItemsCategory(pCategory *Category) {
	hash := getMd5Hash(pCategory.Name)
	PriceList.ItemsCategories[hash] = pCategory
}

func (price *Price) DeleteItemsCategory(pCategory *Category) {
	hash := getMd5Hash(pCategory.Name)
	delete(PriceList.ItemsCategories, hash)
}

func getMd5Hash(encode string) string {
	hasher := md5.New()
	hasher.Write([]byte(encode))
	return hex.EncodeToString(hasher.Sum(nil))
}
