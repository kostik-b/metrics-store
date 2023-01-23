package handler

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"

	ds "github.com/kostik-b/metrics-store/pkg/datastore"
	"github.com/kostik-b/metrics-store/pkg/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const defaultMaxBodySize = 1048576

var internalTemp int = 765

var dummyMachineMetrics model.MachineMetrics = model.MachineMetrics{
	ID:        "test-id",
	MachineID: 123,
	Stats: model.MetricsStats{
		CPUTemp:      456,
		FanSpeed:     789,
		HDDSpace:     987,
		InternalTemp: &internalTemp,
	},
	LastLoggedIn: "userA",
	SysTime:      "timestamp",
}

// implements Datastore interface
type datastoreMock struct {
	mock.Mock

	addEntryArgument *model.MachineMetrics
}

func (d *datastoreMock) GetAllEntries() []*model.MachineMetrics {
	args := d.Called()
	return args.Get(0).([]*model.MachineMetrics)
}

func (d *datastoreMock) AddEntry(key string, value *model.MachineMetrics) ds.DatastoreReturnCode {
	d.addEntryArgument = value

	args := d.Called(key, value)
	return args.Get(0).(ds.DatastoreReturnCode)
}

// implements http.ResponseWriter interface
type responseWriterMock struct {
	mock.Mock

	// this field holds the byte array that is passed to Write call
	// for inspection later
	writeArgument  string
	responseHeader http.Header
}

func (r *responseWriterMock) Header() http.Header {
	return r.responseHeader
}

func (r *responseWriterMock) Write(msg []byte) (int, error) {
	r.writeArgument = string(msg)

	args := r.Called(msg)
	return args.Int(0), args.Error(1)
}

func (r *responseWriterMock) WriteHeader(statusCode int) {
	r.Called(statusCode)
}

type MetricsHandlerTestSuite struct {
	suite.Suite
	dstoreMock     *datastoreMock
	respWriterMock *responseWriterMock
}

func (s *MetricsHandlerTestSuite) SetupTest() {
	s.dstoreMock = new(datastoreMock)
	s.respWriterMock = new(responseWriterMock)

	s.respWriterMock.responseHeader = make(http.Header)
}

// --------------- GET --------------------

func (s *MetricsHandlerTestSuite) Test_GET_EmptyDatastore_ReturnsEmptyJSONArray() {
	// set return values on datastore mock
	emptySlice := []*model.MachineMetrics{}
	s.dstoreMock.On("GetAllEntries").Return(emptySlice)

	metricsHandler := NewMetricsHandler(s.dstoreMock, false, false, defaultMaxBodySize)

	// set return values on response writer mock
	s.respWriterMock.On("WriteHeader", mock.AnythingOfType("int"))
	s.respWriterMock.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)

	request, err := http.NewRequest("GET", "http://localhost:4000/metrics", nil)
	assert.Nil(s.T(), err, "Problem creating request")

	metricsHandler.ServeHTTP(s.respWriterMock, request)

	// inspect call to response writer - should contain empty array
	jsonEmptyString := "[]"
	s.respWriterMock.AssertCalled(s.T(), "Write", []byte(jsonEmptyString))

	contentType := s.respWriterMock.responseHeader.Get("Content-Type")
	assert.Equal(s.T(), contentType, "application/json", "Content type is incorrect")
}

func (s *MetricsHandlerTestSuite) Test_GET_OneElementInDatastore_ReturnsOneElementJSONArray() {
	expectedJSON :=
		`[
  {
    "id": "test-id",
    "machineId": 123,
    "stats": {
      "cpuTemp": 456,
      "fanSpeed": 789,
      "HDDSpace": 987,
      "internalTemp": 765
    },
    "lastLoggedIn": "userA",
    "sysTime": "timestamp"
  }
]`

	// set return values on datastore mock
	machineMetrics := []*model.MachineMetrics{}
	machineMetrics = append(machineMetrics, &dummyMachineMetrics)

	s.dstoreMock.On("GetAllEntries").Return(machineMetrics)

	metricsHandler := NewMetricsHandler(s.dstoreMock, false, false, defaultMaxBodySize)

	// set return values on response writer mock
	s.respWriterMock.On("WriteHeader", mock.AnythingOfType("int"))
	s.respWriterMock.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)

	request, err := http.NewRequest("GET", "http://localhost:4000/metrics", nil)
	assert.Nil(s.T(), err, "Problem creating request")

	metricsHandler.ServeHTTP(s.respWriterMock, request)

	// inspect call to response writer - should contain the expected string
	s.respWriterMock.AssertCalled(s.T(), "Write", []byte(expectedJSON))

	contentType := s.respWriterMock.responseHeader.Get("Content-Type")
	assert.Equal(s.T(), contentType, "application/json", "Content type is incorrect")
}

func (s *MetricsHandlerTestSuite) Test_GET_ThreeElementsInDatastore_ReturnsThreeElementsJSONArrayOptionalFieldExcluded() {
	expectedJSON :=
		`[
  {
    "id": "test-id",
    "machineId": 123,
    "stats": {
      "cpuTemp": 456,
      "fanSpeed": 789,
      "HDDSpace": 987,
      "internalTemp": 765
    },
    "lastLoggedIn": "userA",
    "sysTime": "timestamp"
  },
  {
    "id": "test-1",
    "machineId": 123,
    "stats": {
      "cpuTemp": 456,
      "fanSpeed": 789,
      "HDDSpace": 987,
      "internalTemp": 765
    },
    "lastLoggedIn": "userA",
    "sysTime": "timestamp"
  },
  {
    "id": "test-2",
    "machineId": 123,
    "stats": {
      "cpuTemp": 456,
      "fanSpeed": 789,
      "HDDSpace": 987
    },
    "lastLoggedIn": "userA",
    "sysTime": "timestamp"
  }
]`

	// set return values on datastore mock
	duplicate1 := dummyMachineMetrics
	duplicate1.ID = "test-1"

	duplicate2 := dummyMachineMetrics
	duplicate2.ID = "test-2"
	duplicate2.Stats.InternalTemp = nil

	machineMetrics := []*model.MachineMetrics{}
	machineMetrics = append(machineMetrics, &dummyMachineMetrics)
	machineMetrics = append(machineMetrics, &duplicate1)
	machineMetrics = append(machineMetrics, &duplicate2)

	s.dstoreMock.On("GetAllEntries").Return(machineMetrics)

	metricsHandler := NewMetricsHandler(s.dstoreMock, false, false, defaultMaxBodySize)

	// set return values on response writer mock
	s.respWriterMock.On("WriteHeader", mock.AnythingOfType("int"))
	s.respWriterMock.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)

	request, err := http.NewRequest("GET", "http://localhost:4000/metrics", nil)
	assert.Nil(s.T(), err, "Problem creating request")

	metricsHandler.ServeHTTP(s.respWriterMock, request)

	// inspect call to response writer
	s.respWriterMock.AssertCalled(s.T(), "Write", []byte(expectedJSON))

	contentType := s.respWriterMock.responseHeader.Get("Content-Type")
	assert.Equal(s.T(), contentType, "application/json", "Content type is incorrect")
}

func (s *MetricsHandlerTestSuite) Test_GET_NilElementsFromDatastore_Returns500() {
	// set return values on datastore mock
	var machineMetrics []*model.MachineMetrics
	s.dstoreMock.On("GetAllEntries").Return(machineMetrics)

	metricsHandler := NewMetricsHandler(s.dstoreMock, false, false, defaultMaxBodySize)

	// set return values on response writer mock
	s.respWriterMock.On("WriteHeader", mock.AnythingOfType("int"))
	s.respWriterMock.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)

	request, err := http.NewRequest("GET", "http://localhost:4000/metrics", nil)
	assert.Nil(s.T(), err, "Problem creating request")

	metricsHandler.ServeHTTP(s.respWriterMock, request)

	// inspect call to response writer
	responseBody := "Internal Server Error\n"
	s.respWriterMock.AssertCalled(s.T(), "Write", []byte(responseBody))
	s.respWriterMock.AssertCalled(s.T(), "WriteHeader", http.StatusInternalServerError)
}

func (s *MetricsHandlerTestSuite) Test_GET_CannotWriteResponse_Returns500() {
	// set return values on datastore mock
	machineMetrics := []*model.MachineMetrics{}
	machineMetrics = append(machineMetrics, &dummyMachineMetrics)

	s.dstoreMock.On("GetAllEntries").Return(machineMetrics)

	metricsHandler := NewMetricsHandler(s.dstoreMock, false, false, defaultMaxBodySize)

	// set return values on response writer mock
	s.respWriterMock.On("WriteHeader", mock.AnythingOfType("int"))
	s.respWriterMock.On("Write", mock.AnythingOfType("[]uint8")).Return(0,
		fmt.Errorf("dummy error"))

	request, err := http.NewRequest("GET", "http://localhost:4000/metrics", nil)
	assert.Nil(s.T(), err, "Problem creating request")

	metricsHandler.ServeHTTP(s.respWriterMock, request)

	// inspect calls to response writer
	responseBody := "Error writing response\n"
	s.respWriterMock.AssertCalled(s.T(), "Write", []byte(responseBody))
	s.respWriterMock.AssertCalled(s.T(), "WriteHeader", http.StatusInternalServerError)
}

// --------------- POST --------------------

func (s *MetricsHandlerTestSuite) Test_POST_WrongContentType_Returns415() {
	// even though request body is irrelevant at the moment, we still have to set it since
	// the implementation can change
	requestBody :=
		`{
        "machineId": 12345,
        "stats": {
            "cpuTemp": 90,
            "fanSpeed": 400,
            "HDDSpace": 800
        },
        "lastLoggedIn": "admin/Paul",
        "sysTime": "2022-04-23T18:25:43.511Z"
    }`

	// set return values on datastore mock
	s.dstoreMock.On("AddEntry", mock.Anything, mock.Anything).Return(ds.Success)

	metricsHandler := NewMetricsHandler(s.dstoreMock, false, false, defaultMaxBodySize)

	// set return values on response writer mock
	s.respWriterMock.On("WriteHeader", mock.AnythingOfType("int"))
	s.respWriterMock.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)

	request, err := http.NewRequest("POST", "http://localhost:4000/metrics",
		strings.NewReader(requestBody))
	assert.Nil(s.T(), err, "Problem creating request")

	request.Header.Set("Content-Type", "text/plain; charset=utf-8")

	metricsHandler.ServeHTTP(s.respWriterMock, request)

	// inspect call to response writer
	responseBody := "Content-Type header is not application/json\n"
	s.respWriterMock.AssertCalled(s.T(), "Write", []byte(responseBody))
	s.respWriterMock.AssertCalled(s.T(), "WriteHeader", http.StatusUnsupportedMediaType)

	s.dstoreMock.AssertNumberOfCalls(s.T(), "AddEntry", 0)
}

func (s *MetricsHandlerTestSuite) Test_POST_MalformedJSON_Returns400() {
	requestBody :=
		`{
        "machineId": 12345,
        "stats": {
            "cpuTemp": 90,
            "fanSpeed": 400
            "HDDSpace": 800
        },
        "lastLoggedIn": "admin/Paul",
        "sysTime": "2022-04-23T18:25:43.511Z"
    }`

	// set return values on datastore mock
	s.dstoreMock.On("AddEntry", mock.Anything, mock.Anything).Return(ds.Success)

	metricsHandler := NewMetricsHandler(s.dstoreMock, false, false, defaultMaxBodySize)

	// set return values on response writer mock
	s.respWriterMock.On("WriteHeader", mock.AnythingOfType("int"))
	s.respWriterMock.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)

	request, err := http.NewRequest("POST", "http://localhost:4000/metrics",
		strings.NewReader(requestBody))
	assert.Nil(s.T(), err, "Problem creating request")

	request.Header.Set("Content-Type", "application/json")

	metricsHandler.ServeHTTP(s.respWriterMock, request)

	// inspect call to response writer
	responseBody := "Error parsing request body: invalid character '\"' after object key:value pair\n"
	s.respWriterMock.AssertCalled(s.T(), "Write", []byte(responseBody))
	s.respWriterMock.AssertCalled(s.T(), "WriteHeader", http.StatusBadRequest)

	s.dstoreMock.AssertNumberOfCalls(s.T(), "AddEntry", 0)
}

func (s *MetricsHandlerTestSuite) Test_POST_MoreThanOneJSONObject_Returns400() {
	requestBody :=
		`{
        "machineId": 12345,
        "stats": {
            "cpuTemp": 90,
            "fanSpeed": 400,
            "HDDSpace": 800
        },
        "lastLoggedIn": "admin/Paul",
        "sysTime": "2022-04-23T18:25:43.511Z"
    },`

	// set return values on datastore mock
	s.dstoreMock.On("AddEntry", mock.Anything, mock.Anything).Return(ds.Success)

	metricsHandler := NewMetricsHandler(s.dstoreMock, false, false, defaultMaxBodySize)

	// set return values on response writer mock
	s.respWriterMock.On("WriteHeader", mock.AnythingOfType("int"))
	s.respWriterMock.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)

	request, err := http.NewRequest("POST", "http://localhost:4000/metrics",
		strings.NewReader(requestBody))
	assert.Nil(s.T(), err, "Problem creating request")

	request.Header.Set("Content-Type", "application/json")

	metricsHandler.ServeHTTP(s.respWriterMock, request)

	// inspect call to response writer
	responseBody := "Request body can only contain one JSON object\n"
	s.respWriterMock.AssertCalled(s.T(), "Write", []byte(responseBody))
	s.respWriterMock.AssertCalled(s.T(), "WriteHeader", http.StatusBadRequest)

	s.dstoreMock.AssertNumberOfCalls(s.T(), "AddEntry", 0)
}

func (s *MetricsHandlerTestSuite) Test_POST_RequestBodyTooLarge_Returns400() {
	requestBody :=
		`{
        "machineId": 12345,
        "stats": {
            "cpuTemp": 90,
            "fanSpeed": 400,
            "HDDSpace": 800
        },
        "lastLoggedIn": "admin/Paul",
        "sysTime": "2022-04-23T18:25:43.511Z"
    }`

	// set return values on datastore mock
	s.dstoreMock.On("AddEntry", mock.Anything, mock.Anything).Return(ds.Success)

	metricsHandler := NewMetricsHandler(s.dstoreMock, false, false, 2)

	// set return values on response writer mock
	s.respWriterMock.On("WriteHeader", mock.AnythingOfType("int"))
	s.respWriterMock.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)

	request, err := http.NewRequest("POST", "http://localhost:4000/metrics",
		strings.NewReader(requestBody))
	assert.Nil(s.T(), err, "Problem creating request")

	request.Header.Set("Content-Type", "application/json")

	metricsHandler.ServeHTTP(s.respWriterMock, request)

	// inspect call to response writer
	responseBody := "Error parsing request body: http: request body too large\n"
	s.respWriterMock.AssertCalled(s.T(), "Write", []byte(responseBody))
	s.respWriterMock.AssertCalled(s.T(), "WriteHeader", http.StatusBadRequest)

	s.dstoreMock.AssertNumberOfCalls(s.T(), "AddEntry", 0)
}

func (s *MetricsHandlerTestSuite) Test_POST_CannotAddToDatastoreKeyExists_Returns500() {
	requestBody :=
		`{
        "machineId": 12345,
        "stats": {
            "cpuTemp": 90,
            "fanSpeed": 400,
            "HDDSpace": 800
        },
        "lastLoggedIn": "admin/Paul",
        "sysTime": "2022-04-23T18:25:43.511Z"
    }`

	// set return values on datastore mock
	s.dstoreMock.On("AddEntry", mock.Anything, mock.Anything).Return(ds.ErrorKeyExists)

	metricsHandler := NewMetricsHandler(s.dstoreMock, false, false, defaultMaxBodySize)

	// set return values on response writer mock
	s.respWriterMock.On("WriteHeader", mock.AnythingOfType("int"))
	s.respWriterMock.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)

	request, err := http.NewRequest("POST", "http://localhost:4000/metrics",
		strings.NewReader(requestBody))
	assert.Nil(s.T(), err, "Problem creating request")

	request.Header.Set("Content-Type", "application/json")

	metricsHandler.ServeHTTP(s.respWriterMock, request)

	// inspect call to response writer
	responseBody := "Internal Server Error\n"
	s.respWriterMock.AssertCalled(s.T(), "Write", []byte(responseBody))
	s.respWriterMock.AssertCalled(s.T(), "WriteHeader", http.StatusInternalServerError)

	s.dstoreMock.AssertNumberOfCalls(s.T(), "AddEntry", 2)
}

func (s *MetricsHandlerTestSuite) Test_POST_CannotAddToDatastoreKeyNotSpecified_Returns500() {
	requestBody :=
		`{
        "machineId": 12345,
        "stats": {
            "cpuTemp": 90,
            "fanSpeed": 400,
            "HDDSpace": 800
        },
        "lastLoggedIn": "admin/Paul",
        "sysTime": "2022-04-23T18:25:43.511Z"
    }`

	// set return values on datastore mock
	s.dstoreMock.On("AddEntry", mock.Anything, mock.Anything).Return(ds.ErrorKeyNotSpecified)

	metricsHandler := NewMetricsHandler(s.dstoreMock, false, false, defaultMaxBodySize)

	// set return values on response writer mock
	s.respWriterMock.On("WriteHeader", mock.AnythingOfType("int"))
	s.respWriterMock.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)

	request, err := http.NewRequest("POST", "http://localhost:4000/metrics",
		strings.NewReader(requestBody))
	assert.Nil(s.T(), err, "Problem creating request")

	request.Header.Set("Content-Type", "application/json")

	metricsHandler.ServeHTTP(s.respWriterMock, request)

	// inspect call to response writer
	responseBody := "Internal Server Error\n"
	s.respWriterMock.AssertCalled(s.T(), "Write", []byte(responseBody))
	s.respWriterMock.AssertCalled(s.T(), "WriteHeader", http.StatusInternalServerError)

	s.dstoreMock.AssertNumberOfCalls(s.T(), "AddEntry", 1)
}

func (s *MetricsHandlerTestSuite) Test_POST_UnknownFieldNotAllowed_Returns400() {
	requestBody :=
		`{
        "machineId": 12345,
        "stats": {
            "cpuTemp": 90,
            "fanSpeed": 400,
            "HDDSpace": 800
        },
        "lastLoggedIn": "admin/Paul",
        "extraField": "extraValue",
        "sysTime": "2022-04-23T18:25:43.511Z"
    }`

	// set return values on datastore mock
	s.dstoreMock.On("AddEntry", mock.Anything, mock.Anything).Return(ds.Success)

	metricsHandler := NewMetricsHandler(s.dstoreMock, false, false, defaultMaxBodySize)

	// set return values on response writer mock
	s.respWriterMock.On("WriteHeader", mock.AnythingOfType("int"))
	s.respWriterMock.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)

	request, err := http.NewRequest("POST", "http://localhost:4000/metrics",
		strings.NewReader(requestBody))
	assert.Nil(s.T(), err, "Problem creating request")

	request.Header.Set("Content-Type", "application/json")

	metricsHandler.ServeHTTP(s.respWriterMock, request)

	// inspect call to response writer
	responseBody := "Error parsing request body: json: unknown field \"extraField\"\n"
	s.respWriterMock.AssertCalled(s.T(), "Write", []byte(responseBody))
	s.respWriterMock.AssertCalled(s.T(), "WriteHeader", http.StatusBadRequest)

	s.dstoreMock.AssertNumberOfCalls(s.T(), "AddEntry", 0)
}

func (s *MetricsHandlerTestSuite) Test_POST_UnknownFieldAllowed_Returns201() {
	requestBody :=
		`{
        "machineId": 12345,
        "stats": {
            "cpuTemp": 90,
            "fanSpeed": 400,
            "HDDSpace": 800
        },
        "lastLoggedIn": "admin/Paul",
        "extraField": "extraValue",
        "sysTime": "2022-04-23T18:25:43.511Z"
    }`

	// set return values on datastore mock
	s.dstoreMock.On("AddEntry", mock.Anything, mock.Anything).Return(ds.Success)

	metricsHandler := NewMetricsHandler(s.dstoreMock, false, true, defaultMaxBodySize)

	// set return values on response writer mock
	s.respWriterMock.On("WriteHeader", mock.AnythingOfType("int"))
	s.respWriterMock.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)

	request, err := http.NewRequest("POST", "http://localhost:4000/metrics",
		strings.NewReader(requestBody))
	assert.Nil(s.T(), err, "Problem creating request")

	request.Header.Set("Content-Type", "application/json")

	metricsHandler.ServeHTTP(s.respWriterMock, request)

	// try to match the response against regex
	regexPattern :=
		`{
  "id": ".*",
  "message": "New entry added to the data store with id - .*"
}`
	assert.Regexp(s.T(), regexp.MustCompile(regexPattern), s.respWriterMock.writeArgument,
		"Response body should match expected regex")

	s.respWriterMock.AssertCalled(s.T(), "WriteHeader", http.StatusCreated)

	s.dstoreMock.AssertNumberOfCalls(s.T(), "AddEntry", 1)

	// verify the calls to AddEntry
	assert.Equal(s.T(), 12345, s.dstoreMock.addEntryArgument.MachineID,
		"MachineID in the stored model does not match that of JSON object")
	assert.Equal(s.T(), 90, s.dstoreMock.addEntryArgument.Stats.CPUTemp,
		"Stats.CPUTemp in the stored model does not match that of JSON object")
	assert.Equal(s.T(), 400, s.dstoreMock.addEntryArgument.Stats.FanSpeed,
		"Stats.FanSpeed in the stored model does not match that of JSON object")
	assert.Equal(s.T(), 800, s.dstoreMock.addEntryArgument.Stats.HDDSpace,
		"Stats.HDDSpace in the stored model does not match that of JSON object")
	assert.Equal(s.T(), "admin/Paul", s.dstoreMock.addEntryArgument.LastLoggedIn,
		"LastLoggedIn in the stored model does not match that of JSON object")
	assert.Equal(s.T(), "2022-04-23T18:25:43.511Z", s.dstoreMock.addEntryArgument.SysTime,
		"SysTime in the stored model does not match that of JSON object")
}

func (s *MetricsHandlerTestSuite) Test_POST_OptionalField_Returns201() {
	requestBody :=
		`{
        "machineId": 12345,
        "stats": {
            "cpuTemp": 90,
            "fanSpeed": 400,
            "HDDSpace": 800,
            "internalTemp": 23
        },
        "lastLoggedIn": "admin/Paul",
        "sysTime": "2022-04-23T18:25:43.511Z"
    }`

	// set return values on datastore mock
	s.dstoreMock.On("AddEntry", mock.Anything, mock.Anything).Return(ds.Success)

	metricsHandler := NewMetricsHandler(s.dstoreMock, false, true, defaultMaxBodySize)

	// set return values on response writer mock
	s.respWriterMock.On("WriteHeader", mock.AnythingOfType("int"))
	s.respWriterMock.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)

	request, err := http.NewRequest("POST", "http://localhost:4000/metrics",
		strings.NewReader(requestBody))
	assert.Nil(s.T(), err, "Problem creating request")

	request.Header.Set("Content-Type", "application/json")

	metricsHandler.ServeHTTP(s.respWriterMock, request)

	// try to match the response against regex
	regexPattern :=
		`{
  "id": ".*",
  "message": "New entry added to the data store with id - .*"
}`
	assert.Regexp(s.T(), regexp.MustCompile(regexPattern), s.respWriterMock.writeArgument,
		"Response body should match expected regex")

	s.respWriterMock.AssertCalled(s.T(), "WriteHeader", http.StatusCreated)

	s.dstoreMock.AssertNumberOfCalls(s.T(), "AddEntry", 1)

	// verify the calls to AddEntry
	assert.Equal(s.T(), 12345, s.dstoreMock.addEntryArgument.MachineID,
		"MachineID in the stored model does not match that of JSON object")
	assert.Equal(s.T(), 90, s.dstoreMock.addEntryArgument.Stats.CPUTemp,
		"Stats.CPUTemp in the stored model does not match that of JSON object")
	assert.Equal(s.T(), 400, s.dstoreMock.addEntryArgument.Stats.FanSpeed,
		"Stats.FanSpeed in the stored model does not match that of JSON object")
	assert.Equal(s.T(), 800, s.dstoreMock.addEntryArgument.Stats.HDDSpace,
		"Stats.HDDSpace in the stored model does not match that of JSON object")
	assert.Equal(s.T(), 23, *s.dstoreMock.addEntryArgument.Stats.InternalTemp,
		"Stats.HDDSpace in the stored model does not match that of JSON object")
	assert.Equal(s.T(), "admin/Paul", s.dstoreMock.addEntryArgument.LastLoggedIn,
		"LastLoggedIn in the stored model does not match that of JSON object")
	assert.Equal(s.T(), "2022-04-23T18:25:43.511Z", s.dstoreMock.addEntryArgument.SysTime,
		"SysTime in the stored model does not match that of JSON object")
}

func (s *MetricsHandlerTestSuite) Test_UnknownMethod_Returns405() {

	metricsHandler := NewMetricsHandler(s.dstoreMock, false, false, defaultMaxBodySize)

	request, err := http.NewRequest("PUT", "http://localhost:4000/metrics", nil)
	assert.Nil(s.T(), err, "Problem creating request")

	s.respWriterMock.On("WriteHeader", mock.AnythingOfType("int"))
	metricsHandler.ServeHTTP(s.respWriterMock, request)

	// inspect call to response writer
	s.respWriterMock.AssertCalled(s.T(), "WriteHeader", http.StatusMethodNotAllowed)

	assert.Equal(s.T(), "POST, GET", s.respWriterMock.responseHeader.Get("Allow"),
		"Allowed methods set incorrectly")
}

func TestMetricsHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(MetricsHandlerTestSuite))
}
