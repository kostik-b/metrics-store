// Copyright Konstantin Bakanov 2023

package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/kostik-b/metrics-store/pkg/datastore"
	"github.com/kostik-b/metrics-store/pkg/model"
)

// an HTTP handler to handle incoming requests
// making it unexported as its member variables have to be set
type metricsHandler struct {
	MetricsDatastore   datastore.DatastoreInterface
	Debug              bool
	AllowUnknownFields bool
	MaxBodySize        int64
}

func NewMetricsHandler(metricsDatastore datastore.DatastoreInterface,
	debug, allowUnknownFields bool,
	maxBodySize int64) *metricsHandler {
	return &metricsHandler{
		MetricsDatastore:   metricsDatastore,
		Debug:              debug,
		AllowUnknownFields: allowUnknownFields,
		MaxBodySize:        maxBodySize,
	}
}

// implementing http.Handler interface
func (m *metricsHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {

	// differentiate between post and get
	// if unknown return 405
	if request.Method == "GET" {
		m.handleGetRequest(responseWriter, request)
	} else if request.Method == "POST" {
		m.handlePostRequest(responseWriter, request)
	} else {
		if m.Debug {
			log.Printf("Received unknown request method: %s\n", request.Method)
		}
		errorResponseMethodNotAllowed(responseWriter)
	}
}

func errorResponseMethodNotAllowed(responseWriter http.ResponseWriter) {
	responseWriter.Header().Set("Allow", "POST, GET")
	responseWriter.WriteHeader(http.StatusMethodNotAllowed)
}

func (m *metricsHandler) handleGetRequest(responseWriter http.ResponseWriter, request *http.Request) {
	if m.Debug {
		log.Println("Handling GET request")
	}

	allEntries := m.MetricsDatastore.GetAllEntries()

	if allEntries == nil {
		log.Println("ERROR: GET - could not get entries from the datastore")
		http.Error(responseWriter, "Internal Server Error", http.StatusInternalServerError)
	} else {
		allEntriesAsBytes, err := json.MarshalIndent(allEntries, "", "  ") // for easier readability
		if err != nil {
			log.Printf("ERROR: GET - could not marshal entries as a byte array: %s\n", err.Error())
			http.Error(responseWriter, "Error marshalling entries", http.StatusInternalServerError)
		} else {
			responseWriter.Header().Set("Content-Type", "application/json")

			// the 200 header will be set automatically
			_, err := responseWriter.Write(allEntriesAsBytes)

			if err != nil {
				log.Printf("ERROR: GET - could not write response: %s\n", err.Error())
				http.Error(responseWriter, "Error writing response", http.StatusInternalServerError)
			}
		}
	}
}

func (m *metricsHandler) handlePostRequest(responseWriter http.ResponseWriter, request *http.Request) {
	if m.Debug {
		log.Printf("Handling POST request")
	}

	// check whether content type is correct
	if contentType := request.Header.Get("Content-Type"); contentType != "" {
		if contentType != "application/json" {
			if m.Debug {
				log.Printf("POST - received incorrect content type %s\n", contentType)
			}
			http.Error(responseWriter, "Content-Type header is not application/json", http.StatusUnsupportedMediaType)
			return
		}
	}

	requestBodyMaxBytesReader := http.MaxBytesReader(responseWriter, request.Body, m.MaxBodySize)

	decoder := json.NewDecoder(requestBodyMaxBytesReader)

	// disallow unknown fields in the request
	if !m.AllowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	var machineMetrics *model.MachineMetrics = &model.MachineMetrics{}
	err := decoder.Decode(&machineMetrics)

	if err != nil {
		errMsg := "Error parsing request body: " + err.Error()
		log.Printf("ERROR: %s\n", errMsg)
		http.Error(responseWriter, errMsg, http.StatusBadRequest)
		return
	}

	// check that there is no additional data in the body
	err = decoder.Decode(&struct{}{})
	if err != io.EOF {
		log.Println("ERROR: POST - request body contains more than one object")
		http.Error(responseWriter, "Request body can only contain one JSON object", http.StatusBadRequest)
		return
	}

	// if we got to here, then it's all good
	if m.Debug {
		log.Printf("POST - successfully received and decoded request: %#v\n", machineMetrics)
	}

	machineMetrics.ID = uuid.New().String()

	rc := m.MetricsDatastore.AddEntry(machineMetrics.ID, machineMetrics)

	// in a super rare case when the UUID is duplicate, generate another one
	if rc == datastore.ErrorKeyExists {
		machineMetrics.ID = uuid.New().String()

		rc = m.MetricsDatastore.AddEntry(machineMetrics.ID, machineMetrics)

		// if there is still an error, just return 500 regardless of error type
		if rc != datastore.Success {
			log.Printf("ERROR: POST - could not add entry to the datastore: error - %s, key - %s, value - %#v\n", rc.String(), machineMetrics.ID, machineMetrics)
			http.Error(responseWriter, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else if rc != datastore.Success {
		// if it's any other error, just return 500
		log.Printf("ERROR: POST - could not add entry to the datastore: error - %s, key - %s, value - %#v\n", rc.String(), machineMetrics.ID, machineMetrics)
		http.Error(responseWriter, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusCreated)

	response := make(map[string]string)

	// this is for human readability
	response["message"] = "New entry added to the data store with id - " + machineMetrics.ID
	// this is for machine parsing
	response["id"] = machineMetrics.ID

	responseAsBytes, err := json.MarshalIndent(response, "", "  ") // for readability
	if err != nil {
		// log the error but still send the response
		log.Printf("ERROR: POST - could not marshal 201 JSON response with id %s.\nError is %s\n", machineMetrics.ID, err.Error())
	}
	_, err = responseWriter.Write(responseAsBytes)

	// we still need to send 201 back to the client to indicate
	// that an entry had been written to the DB
	if err != nil {
		log.Printf("ERROR: GET - could not write response: %s, but sending 201 anyways\n", err.Error())
	} else {
		if m.Debug {
			log.Printf("POST - sending response: %v\n", string(responseAsBytes))
		}
	}

}
