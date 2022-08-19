package struct_viewer

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONHandler(t *testing.T) {
	tcs := []struct {
		testName string

		givenConfig interface{}

		expectedStatusCode int
		expectedJSONOutput string
	}{
		{
			testName: "simple struct",
			givenConfig: struct {
				Name string `json:"field_name"`
			}{
				"field_value",
			},
			expectedStatusCode: http.StatusOK,
			expectedJSONOutput: fmt.Sprintln(`{"field_name":"field_value"}`),
		},
		{
			testName: "complex struct struct",
			givenConfig: struct {
				Name string `json:"name,omitempty"`
				Data struct {
					Object1 int  `json:"object_1,omitempty"`
					Object2 bool `json:"object_2,omitempty"`
				} `json:"data"`
				Metadata map[string]struct {
					ID    int    `json:"id,omitempty"`
					Value string `json:"value,omitempty"`
				} `json:"metadata,omitempty"`
				OmittedValue string `json:"omitted_value,omitempty"`
			}{
				Name: "name_value",
				Data: struct {
					Object1 int  `json:"object_1,omitempty"`
					Object2 bool `json:"object_2,omitempty"`
				}{
					Object1: 1,
					Object2: true,
				},
				Metadata: map[string]struct {
					ID    int    `json:"id,omitempty"`
					Value string `json:"value,omitempty"`
				}{
					"key_99": {ID: 99, Value: "key99"},
				},
			},
			expectedStatusCode: http.StatusOK,
			expectedJSONOutput: fmt.Sprintln(
				`{"name":"name_value",
					"data":{"object_1":1,"object_2":true},
					"metadata":{"key_99":{"id":99,"value":"key99"}}}`,
			),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
			// pass 'nil' as the third parameter.
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			structViewerConfig := Config{Object: tc.givenConfig}
			helper, err := New(&structViewerConfig, "")
			assert.NoError(t, err, "failed to instantiate viewer")

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(helper.JSONHandler)

			// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
			// directly and pass in our Request and ResponseRecorder.
			handler.ServeHTTP(rr, req)

			// Check the status code is what we expect.
			assert.Equal(t, tc.expectedStatusCode, rr.Code)

			// Check the response body is what we expect.
			assert.Equal(t, tc.expectedJSONOutput, rr.Body.String())
		})
	}
}

func TestEnvsHandler(t *testing.T) {
	tcs := []struct {
		testName string

		givenConfig        interface{}
		givenPrefix        string
		expectedStatusCode int
		expectedJSONOutput string
	}{
		{
			testName: "simple struct",
			givenConfig: struct {
				Name string `json:"field_name"`
			}{
				"field_value",
			},
			expectedStatusCode: http.StatusOK,
			expectedJSONOutput: fmt.Sprintln(`["FIELDNAME:field_value"]`),
		},
		{
			testName: "simple struct with prefix",
			givenConfig: struct {
				Name string `json:"field_name"`
			}{
				"field_value",
			},
			givenPrefix:        "TEST_",
			expectedStatusCode: http.StatusOK,
			expectedJSONOutput: fmt.Sprintln(`["TEST_FIELDNAME:field_value"]`),
		},
		{
			testName: "complex struct struct",
			givenConfig: struct {
				Name string `json:"name,omitempty"`
				Data struct {
					Object1 int  `json:"object_1,omitempty"`
					Object2 bool `json:"object_2,omitempty"`
				} `json:"data"`
				Metadata map[string]struct {
					ID    int    `json:"id,omitempty"`
					Value string `json:"value,omitempty"`
				} `json:"metadata,omitempty"`
				OmittedValue string `json:"omitted_value,omitempty"`
			}{
				Name: "name_value",
				Data: struct {
					Object1 int  `json:"object_1,omitempty"`
					Object2 bool `json:"object_2,omitempty"`
				}{
					Object1: 1,
					Object2: true,
				},
				Metadata: map[string]struct {
					ID    int    `json:"id,omitempty"`
					Value string `json:"value,omitempty"`
				}{
					"key_99": {ID: 99, Value: "key99"},
				},
			},
			expectedStatusCode: http.StatusOK,
			expectedJSONOutput: fmt.Sprintln(
				`["NAME:name_value",
					 "DATA_OBJECT1:1",
					 "DATA_OBJECT2:true",
					 "METADATA:map[key_99:{99 key99}]",
					 "OMITTEDVALUE:"]`,
			),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
			// pass 'nil' as the third parameter.
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			structViewerConfig := Config{Object: tc.givenConfig}
			helper, err := New(&structViewerConfig, tc.givenPrefix)
			assert.NoError(t, err, "failed to instantiate viewer")

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(helper.EnvsHandler)

			// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
			// directly and pass in our Request and ResponseRecorder.
			handler.ServeHTTP(rr, req)

			// Check the status code is what we expect.
			assert.Equal(t, tc.expectedStatusCode, rr.Code)

			// Check the response body is what we expect.
			assert.Equal(t, tc.expectedJSONOutput, rr.Body.String())
		})
	}
}
