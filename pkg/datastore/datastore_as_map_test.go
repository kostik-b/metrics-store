package datastore

import (
  "testing"

  "github.com/kostik-b/metrics-store/pkg/model"

  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/suite"
)

var internalTemp int = 765

var dummyMachineMetrics model.MachineMetrics = model.MachineMetrics{
  ID: "test-id",
  MachineID: 123,
  Stats: model.MetricsStats{
    CPUTemp: 456,
    FanSpeed: 789,
    HDDSpace: 987,
    InternalTemp: &internalTemp,
  },
  LastLoggedIn: "userA",
  SysTime: "timestamp",
}

type DatastoreTestSuite struct {
  suite.Suite
}

// this "hack" will be run before each test to clear the map
func (suite *DatastoreTestSuite) SetupTest() {
  GetInstance() // to make sure the map had been created

  for k := range metricsStore.entryMap {
      delete(metricsStore.entryMap, k)
  }
}

func (suite *DatastoreTestSuite) Test_AddEntryWithEmptyKey_ReturnsKeyNotSpecifiedError() {
  datastore := GetInstance()

  err := datastore.AddEntry("", &dummyMachineMetrics)

  assert.Equal(suite.T(), err, ErrorKeyNotSpecified, "Error should be " + ErrorKeyNotSpecified.String())
}

func (suite *DatastoreTestSuite) Test_AddEntryWithNilValue_ReturnsValueNotSpecifiedError() {
  datastore := GetInstance()

  err := datastore.AddEntry("dummyKey", nil)

  assert.Equal(suite.T(), err, ErrorValueNotSpecified, "Error should be " + ErrorValueNotSpecified.String())
}

func (suite *DatastoreTestSuite) Test_AddEntryWithExistingKey_ReturnsKeyExistsError() {
  datastore := GetInstance()

  datastore.AddEntry("dummyKey", &dummyMachineMetrics)
  err := datastore.AddEntry("dummyKey", &dummyMachineMetrics)

  assert.Equal(suite.T(), err, ErrorKeyExists, "Error should be " + ErrorKeyExists.String())
}

func (suite *DatastoreTestSuite) Test_AddEntryWithNonExistingKeyNonNilvalue_ReturnsSuccess() {
  datastore := GetInstance()

  err := datastore.AddEntry("dummyKey", &dummyMachineMetrics)

  assert.Equal(suite.T(), err, Success, "ReturnCode should be " + Success.String())
}

func (suite *DatastoreTestSuite) Test_GetAllEntries_MapEmpty_ReturnsEmptySlice() {
  datastore := GetInstance()

  allEntries := datastore.GetAllEntries()

  emptySlice := []*model.MachineMetrics {}

  assert.Equal(suite.T(), allEntries, emptySlice, "Empty slice should be returned")
}

func (suite *DatastoreTestSuite) Test_GetAllEntriesMapContainsOneEntry_ReturnsSliceWithiSameOneEntry() {
  datastore := GetInstance()

  err := datastore.AddEntry("dummyKey", &dummyMachineMetrics)
  assert.Equal(suite.T(), err, Success, "Return Code should be " + Success.String())

  allEntries := datastore.GetAllEntries()

  assert.Equal(suite.T(), len(allEntries), 1, "Slice should contain one entry")

  assert.EqualValues(suite.T(), &dummyMachineMetrics, allEntries[0], "Actual values of MachineMetries do not match the expected ones")
}

func (suite *DatastoreTestSuite) Test_GetAllEntriesMapContainsThreeEntries_ReturnsSliceWithSameThreeElements() {
  datastore := GetInstance()

  duplicate1 := dummyMachineMetrics
  duplicate1.ID = "test-1"

  duplicate2 := dummyMachineMetrics
  duplicate2.ID = "test-2"

  err := datastore.AddEntry("dummyKey", &dummyMachineMetrics)
  assert.Equal(suite.T(), err, Success, "Return Code should be " + Success.String())

  err = datastore.AddEntry("dummyKey1", &duplicate1)
  assert.Equal(suite.T(), err, Success, "Return Code should be " + Success.String())

  err = datastore.AddEntry("dummyKey2", &duplicate2)
  assert.Equal(suite.T(), err, Success, "Return Code should be " + Success.String())

  allEntries := datastore.GetAllEntries()

  assert.Equal(suite.T(), len(allEntries), 3, "Slice should contain three entries")

  var expectedSlice []*model.MachineMetrics

  expectedSlice = append(expectedSlice, &dummyMachineMetrics)
  expectedSlice = append(expectedSlice, &duplicate1)
  expectedSlice = append(expectedSlice, &duplicate2)

  assert.ElementsMatch(suite.T(), expectedSlice, allEntries, "slices do not match")
}

func (suite *DatastoreTestSuite) Test_GetInstance_EntryMapIsSet() {
  datastore := GetInstance()

  impl, ok := datastore.(*datastoreAsMap)

  assert.True(suite.T(), ok, "Could not cast datastore interface to datastoreAsMap")

  assert.NotNil(suite.T(), impl.entryMap, "datastoreAsMap.entryMap should have been initialized")
}

func (suite *DatastoreTestSuite) Test_GetInstanceTwice_HaveSameAddress() {
  datastore  := GetInstance()
  datastore2 := GetInstance()

  assert.Same(suite.T(), datastore, datastore2, "GetInstance should return the same object")
}

func TestDatastoreTestSuite(t *testing.T) {
    suite.Run(t, new(DatastoreTestSuite))
}
