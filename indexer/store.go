/*
Sniperkit-Bot
- Status: analyzed
*/

package indexer

import (
	"github.com/sniperkit/snk.fork.gostore-contrib/common"
)

// ProviderStore a store which provides data to an index store
type ProviderStore interface {
	Cursor() (common.Iterator, error)
}
