package struct_viewer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type complexType struct {
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
}

var complexStruct = complexType{
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
}

func toJSON(t *testing.T, data interface{}) string {
	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)

	return string(jsonData)
}

func setQueryParams(req *http.Request, queryParamKey, queryParamVal string) {
	if queryParamVal != "" {
		q := req.URL.Query()
		q.Add(queryParamKey, queryParamVal)
		req.URL.RawQuery = q.Encode()
	}
}

func TestJSONHandler(t *testing.T) {
	tcs := []struct {
		testName string

		givenConfig   interface{}
		queryParamVal string

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
			testName:           "complex struct struct",
			givenConfig:        complexStruct,
			expectedStatusCode: http.StatusOK,
			expectedJSONOutput: toJSON(t, complexStruct),
		},
		{
			testName:           "valid field of complexStruct via query param",
			givenConfig:        complexStruct,
			queryParamVal:      "name",
			expectedStatusCode: http.StatusOK,
			expectedJSONOutput: toJSON(t, EnvVar{
				Env:         "TYK_NAME",
				Value:       complexStruct.Name,
				ConfigField: "name",
			}),
		},
		{
			testName:           "valid field from inner object of complexStruct via query param",
			givenConfig:        complexStruct,
			queryParamVal:      "data.object_1",
			expectedStatusCode: http.StatusOK,
			expectedJSONOutput: toJSON(t, EnvVar{
				Env:         "TYK_DATA_OBJECT1",
				Value:       strconv.Itoa(complexStruct.Data.Object1),
				ConfigField: "data.object_1",
			}),
		},
		{
			testName:           "invalid field complexStruct via query param",
			givenConfig:        complexStruct,
			queryParamVal:      "data.object_3",
			expectedStatusCode: http.StatusOK,
			expectedJSONOutput: toJSON(t, EnvVar{}),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
			// pass 'nil' as the third parameter.
			req, err := http.NewRequest("GET", "/", nil)
			assert.NoError(t, err)

			setQueryParams(req, JSONQueryKey, tc.queryParamVal)

			structViewerConfig := Config{Object: tc.givenConfig}
			helper, err := New(&structViewerConfig, "TYK_")
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
			assert.JSONEq(t, tc.expectedJSONOutput, rr.Body.String())
		})
	}
}

func TestEnvsHandler(t *testing.T) {
	tcs := []struct {
		testName string

		givenConfig        interface{}
		givenPrefix        string
		queryParamVal      string
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
			testName:           "complex struct struct",
			givenConfig:        complexStruct,
			expectedStatusCode: http.StatusOK,
			expectedJSONOutput: fmt.Sprintln(
				`["NAME:name_value",` +
					`"DATA_OBJECT1:1",` +
					`"DATA_OBJECT2:true",` +
					`"METADATA:map[key_99:{99 key99}]",` +
					`"OMITTEDVALUE:"]`,
			),
		},
		{
			testName:           "valid field of complexStruct via query param",
			givenConfig:        complexStruct,
			givenPrefix:        "TYK_",
			queryParamVal:      "TYK_NAME",
			expectedStatusCode: http.StatusOK,
			expectedJSONOutput: toJSON(t, EnvVar{
				Env:         "TYK_NAME",
				Value:       complexStruct.Name,
				ConfigField: "name",
			}),
		},
		{
			testName:           "valid field from inner object of complexStruct via query param",
			givenConfig:        complexStruct,
			givenPrefix:        "TYK_",
			queryParamVal:      "TYK_DATA_OBJECT1",
			expectedStatusCode: http.StatusOK,
			expectedJSONOutput: toJSON(t, EnvVar{
				Env:         "TYK_DATA_OBJECT1",
				Value:       strconv.Itoa(complexStruct.Data.Object1),
				ConfigField: "data.object_1",
			}),
		},
		{
			testName:           "invalid field complexStruct via query param",
			givenConfig:        complexStruct,
			givenPrefix:        "TYK_",
			queryParamVal:      "TYK_DATA_OBJECT3",
			expectedStatusCode: http.StatusOK,
			expectedJSONOutput: toJSON(t, EnvVar{}),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
			// pass 'nil' as the third parameter.
			req, err := http.NewRequest("GET", "/", nil)
			assert.NoError(t, err)

			setQueryParams(req, EnvQueryKey, tc.queryParamVal)

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
			assert.JSONEq(t, tc.expectedJSONOutput, rr.Body.String())
		})
	}
}
