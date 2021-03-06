/*
Sniperkit-Bot
- Status: analyzed
*/

package indexer

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/mgutz/logxi/v1"
)

func ReIndex(provider ProviderStore, index *Indexer) error {
	iter, _ := provider.Cursor()
	for iter.Valid() {
		key := iter.Key()
		val := iter.Value()
		var v map[string]interface{}
		if err := json.Unmarshal(val, &v); err == nil {
			k := string(key)
			u := strings.SplitN(k, "|", 2)
			ID := u[1]
			store := strings.TrimPrefix(u[0], "t$")
			// logger.Debug("reindexing", "ID", ID, "val", v)
			index.IndexDocument(ID, IndexedData{store, v})
		}
		iter.Next()
	}
	return nil
}

// IndexedData represents a stored row
type IndexedData struct {
	Bucket string      `json:"bucket"`
	Data   interface{} `json:"data"`
}

var logger = log.New("gostore-contrib.indexer")

type RequestOpt func(*bleve.SearchRequest) error

var OrderRequest = func(orderBy []string) RequestOpt {
	return func(req *bleve.SearchRequest) error {
		req.SortBy(orderBy)
		return nil
	}
}

type Indexer struct {
	index bleve.Index
}

func (i *Indexer) Index() bleve.Index {
	return i.index
}
func (i *Indexer) BatchIndex() *bleve.Batch {
	return i.index.NewBatch()
}
func (i *Indexer) Batch(b *bleve.Batch) error {
	return i.index.Batch(b)
}
func (i *Indexer) AddDocumentMapping(name string, dm *mapping.DocumentMapping) {
	// i.index.AddDocumentMapping(name, dm)
}

func (i Indexer) IndexDocument(id string, data interface{}) error {
	if i.index == nil {
		return errors.New("No index")
	}
	// logger.Debug("Indexing document", "id", id, "data", data)
	return i.index.Index(id, data)
}

func (i Indexer) UnIndexDocument(id string) error {
	if i.index == nil {
		return errors.New("No index")
	}
	// logger.Debug("UnIndexing document", "id", id)
	return i.index.Delete(id)
}

func (i Indexer) QueryMap(q map[string]interface{}, opts ...RequestOpt) (*bleve.SearchResult, error) {
	queryString := ""
	for k, v := range q {
		queryString = fmt.Sprintf("%s %s:%v", queryString, k, v)
	}
	return i.Query(queryString, opts...)
}
func (i Indexer) Query(q string, opts ...RequestOpt) (*bleve.SearchResult, error) {
	if i.index == nil {
		return nil, errors.New("No index")
	}
	// println(q)
	query := bleve.NewQueryStringQuery(q)
	searchRequest := bleve.NewSearchRequest(query)
	for _, opt := range opts {
		if err := opt(searchRequest); err != nil {
			logger.Warn("failed option passed")
		}
	}
	return i.index.Search(searchRequest)
}

func (i Indexer) QueryWithOptions(q string, size, from int, explain bool, fields []string, opts ...RequestOpt) (*bleve.SearchResult, error) {
	if i.index == nil {
		return nil, errors.New("No index")
	}
	query := bleve.NewQueryStringQuery(q)
	searchRequest := bleve.NewSearchRequestOptions(query, size, from, explain)
	if len(fields) > 0 {
		searchRequest.Fields = fields
	}
	for _, opt := range opts {
		if err := opt(searchRequest); err != nil {
			logger.Warn("failed option passed")
		}
	}
	return i.index.Search(searchRequest)
}

func (i Indexer) FacetedQuery(q string, facets *Facets, size, from int, explain bool, fields []string, opts ...RequestOpt) (*bleve.SearchResult, error) {
	if i.index == nil {
		return nil, errors.New("No index")
	}
	query := bleve.NewQueryStringQuery(q)
	searchRequest := bleve.NewSearchRequestOptions(query, size, from, explain)
	if len(fields) > 0 {
		searchRequest.Fields = fields
	}
	for _, opt := range opts {
		if err := opt(searchRequest); err != nil {
			logger.Warn("failed option passed")
		}
	}
	AddFacets(searchRequest, facets)
	return i.index.Search(searchRequest)
}
func (i Indexer) QueryWithOptionsHighlighted(q string, size, from int, explain bool, fields []string, opts ...RequestOpt) (*bleve.SearchResult, error) {

	if i.index == nil {
		return nil, errors.New("No index")
	}
	query := bleve.NewQueryStringQuery(q)
	searchRequest := bleve.NewSearchRequestOptions(query, size, from, explain)
	searchRequest.Highlight = bleve.NewHighlightWithStyle("ansi")
	for _, opt := range opts {
		if err := opt(searchRequest); err != nil {
			logger.Warn("failed option passed")
		}
	}
	return i.index.Search(searchRequest)
}

func (i Indexer) MatchQuery(q, field string, opts ...RequestOpt) (*bleve.SearchResult, error) {

	if i.index == nil {
		return nil, errors.New("No index")
	}
	query := bleve.NewMatchQuery(q)
	query.SetField(field)
	query.SetFuzziness(0)
	searchRequest := bleve.NewSearchRequest(query)
	for _, opt := range opts {
		if err := opt(searchRequest); err != nil {
			logger.Warn("failed option passed")
		}
	}
	return i.index.Search(searchRequest)
}

func (i Indexer) TermQuery(q string, opts ...RequestOpt) (*bleve.SearchResult, error) {

	if i.index == nil {
		return nil, errors.New("No index")
	}
	query := bleve.NewTermQuery(q)
	searchRequest := bleve.NewSearchRequest(query)
	for _, opt := range opts {
		if err := opt(searchRequest); err != nil {
			logger.Warn("failed option passed")
		}
	}
	return i.index.Search(searchRequest)
}

func (i Indexer) MatchPhraseQuery(q string, opts ...RequestOpt) (*bleve.SearchResult, error) {

	if i.index == nil {
		return nil, errors.New("No index")
	}
	query := bleve.NewMatchPhraseQuery(q)
	searchRequest := bleve.NewSearchRequest(query)
	for _, opt := range opts {
		if err := opt(searchRequest); err != nil {
			logger.Warn("failed option passed")
		}
	}
	return i.index.Search(searchRequest)
}

func (i Indexer) Close() {

	if i.index == nil {
		return
	}
	err := i.index.Close()
	if err != nil {
		logger.Warn("error while closing index")
	}
}

func GetIndex(indexPath string) (bleve.Index, bool) {
	index, err := bleve.Open(indexPath)
	if err == bleve.ErrorIndexPathDoesNotExist {
		logger.Debug("Index path does not exist", "path", "indexPath")
		return nil, false
	}
	return index, true
}
func NewIndexerFromIndex(index bleve.Index) *Indexer {
	return &Indexer{index: index}
}

// NewIndexer creates a new indexer
func NewDefaultIndexer(indexPath string) *Indexer {
	indexMapping := bleve.NewIndexMapping()
	return NewIndexer(indexPath, indexMapping)
}

// NewIndexer creates a new indexer
func NewIndexer(indexPath string, indexMapping mapping.IndexMapping) *Indexer {
	index, err := bleve.Open(indexPath)
	if err != nil {
		logger.Debug("Error opening indexpath", "path", indexPath, "verbose", string(err.Error()))
		if err == bleve.ErrorIndexMetaMissing || err == bleve.ErrorIndexPathDoesNotExist {
			logger.Debug(fmt.Sprintf("Creating new index at %s ...", indexPath))
			// indexMapping.DefaultAnalyzer = "keyword"
			index, err = bleve.New(indexPath, indexMapping)
			if err != nil {
				logger.Warn("Index could not be created", "path", indexPath, "err", string(err.Error()))
				if err != bleve.ErrorIndexPathExists {
					panic(err)
				}
				return nil
			}
			return &Indexer{index: index}
		}
		panic(err)
	}
	return &Indexer{index: index}
}
