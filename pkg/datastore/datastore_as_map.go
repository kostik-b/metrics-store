package datastore

import (
  "sync"
  "github.com/kostik-b/metrics-store/pkg/model"
)

// datastoreAsMap implementes DatastoreInterface
type datastoreAsMap struct {
  entryMap  map[string]*model.MachineMetrics
  mutex     sync.Mutex // we need this for concurrent access
}

// it's a singleton
var metricsStore datastoreAsMap

func GetInstance() DatastoreInterface {
  if metricsStore.entryMap == nil {
    metricsStore.entryMap = make(map[string]*model.MachineMetrics)
  }

  return &metricsStore
}

func (d *datastoreAsMap) GetAllEntries ()[]*model.MachineMetrics {
  d.mutex.Lock()
  defer d.mutex.Unlock()

  allEntries := []*model.MachineMetrics{}
  for _, v := range d.entryMap {
    allEntries = append(allEntries, v)
  }

  return allEntries
}

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
  _, found := d.entryMap[key]
  if found {
    return ErrorKeyExists
  }

  d.entryMap[key] = entry

  return Success
}
