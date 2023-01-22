// Copyright Konstantin Bakanov 2023

package datastore

import ("github.com/kostik-b/metrics-store/pkg/model")

// These return codes are used by datastore to return
// results of operations
type DatastoreReturnCode int

const (
  Success        DatastoreReturnCode = iota
  ErrorKeyExists 
  ErrorKeyNotSpecified
  ErrorValueNotSpecified
)

func (d DatastoreReturnCode) String() string {
  switch d {
    case Success:
      return "Success"
    case ErrorKeyExists:
      return "Key already exists"
    case ErrorKeyNotSpecified:
      return "Key not specified"
    case ErrorValueNotSpecified:
      return "Value not specified"
    default:
      return "Unknown return code"
  }
}

// A datastore interface to add one entry to the datastore
// and to retrieve all entries from a datastore
// if there are no entries in the datastore, an empty slice will be returned
type DatastoreInterface interface {
  GetAllEntries ()([]*model.MachineMetrics)
  AddEntry(string, *model.MachineMetrics) DatastoreReturnCode
}
