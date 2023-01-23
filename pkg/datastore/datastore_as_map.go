package datastore

import (
	"github.com/kostik-b/metrics-store/pkg/model"
	"sync"
)

// this is needed to init datastoreAsMap.entries in a thread safe way
var once sync.Once

// datastoreAsMap implementes DatastoreInterface
type datastoreAsMap struct {
	entries map[string]*model.MachineMetrics
	mutex   sync.Mutex // we need this for concurrent access
}

// datastore as map is a singleton and can only be retrieved
// using the GetInstance() method
var metricsStore datastoreAsMap

func GetInstance() DatastoreInterface {
	once.Do(func() {
		metricsStore.entries = make(map[string]*model.MachineMetrics)
	})

	return &metricsStore
}

// GetAllEntries returns all entries stored in the map
// or an empty slide otherwise
func (d *datastoreAsMap) GetAllEntries() []*model.MachineMetrics {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	allEntries := []*model.MachineMetrics{}
	for _, v := range d.entries {
		allEntries = append(allEntries, v)
	}

	return allEntries
}

// AddEntry adds entry to the map based on key
func (d *datastoreAsMap) AddEntry(key string, entry *model.MachineMetrics) DatastoreReturnCode {
	if key == "" {
		return ErrorKeyNotSpecified
	}

	if entry == nil {
		return ErrorValueNotSpecified
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	// check if such entry exists
	_, found := d.entries[key]
	if found {
		return ErrorKeyExists
	}

	d.entries[key] = entry

	return Success
}
