/*
Sniperkit-Bot
- Status: analyzed
*/

package indexer

import (
	"fmt"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/index/store/moss"
	"github.com/blevesearch/bleve/index/upsidedown"
	"github.com/blevesearch/bleve/mapping"
)

// NewIndexer creates a new indexer
func NewMossIndexer(indexPath string) (*Indexer, bool) {
	indexMapping := bleve.NewIndexMapping()
	return NewMossIndexerWithMapping(indexPath, indexMapping)
}

// NewIndexer creates a new indexer
func NewMossIndexerWithMapping(indexPath string, indexMapping mapping.IndexMapping) (*Indexer, bool) {
	// os.RemoveAll(indexPath)
	index, err := bleve.Open(indexPath)
	if err != nil {
		logger.Debug("Error opening MOSS indexpath", "path", indexPath, "verbose", string(err.Error()))
		if err == bleve.ErrorIndexMetaMissing || err == bleve.ErrorIndexPathDoesNotExist {
			logger.Debug(fmt.Sprintf("Creating new MOSS index at %s ...", indexPath))
			// indexMapping.DefaultAnalyzer = "keyword"
			kvconfig := map[string]interface{}{
				"mossLowerLevelStoreName": "mossStore",
			}

			index, err = bleve.NewUsing(indexPath, indexMapping, upsidedown.Name, moss.Name, kvconfig)

			if err != nil {
				logger.Warn("MOSS Index could not be created", "path", indexPath, "err", string(err.Error()))
				if err != bleve.ErrorIndexPathExists {
					panic(err)
				}
				return nil, false
			}
			time.Sleep(30 * time.Second)

		} else {
			panic(err)
		}
		return &Indexer{index: index}, true
	}
	logger.Debug("opening existing MOSS index", "stats", index.Stats())
	return &Indexer{index: index}, false
}
