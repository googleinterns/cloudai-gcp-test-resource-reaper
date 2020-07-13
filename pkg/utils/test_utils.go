package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"google.golang.org/api/option"
)

// createServer is a helper function to create a fake server
// where http requsts will be rerouted for testing
func CreateServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}

func DefaultHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(`{"success": true}`))

}

func GetTestOptions(server *httptest.Server) []option.ClientOption {
	return []option.ClientOption{
		option.WithHTTPClient(server.Client()),
		option.WithEndpoint(server.URL),
	}
}

type TestInstance struct {
	Name              string
	CreationTimestamp string
	Zone              string
}

func SendResponse(w http.ResponseWriter, response interface{}) {
	json.NewEncoder(w).Encode(response)
}
