package structviewer

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

// complexStructToMap returns a map representation of the complexStruct.
func complexStructToMap() map[string]*EnvVar {
	envs := parseEnvs(complexStruct, "TYK_", "")
	configMap := parseConfig(envs)

	return configMap
}

func setQueryParams(req *http.Request, queryParamKey, queryParamVal string) {
	if queryParamVal != "" {
		q := req.URL.Query()
		q.Add(queryParamKey, queryParamVal)
		req.URL.RawQuery = q.Encode()
	}
}

func TestConfigHandler(t *testing.T) {
	tcs := []struct {
		testName string

		givenConfig interface{}

		expectedStatusCode int
		expectedJSONOutput string

		shouldDeleteConfig bool
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
			testName:           "not initialized",
			givenConfig:        complexStruct,
			expectedStatusCode: http.StatusInternalServerError,
			expectedJSONOutput: "",
			shouldDeleteConfig: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
			// pass 'nil' as the third parameter.
			req, err := http.NewRequest("GET", "/", nil)
			assert.NoError(t, err)

			structViewerConfig := Config{Object: tc.givenConfig}
			helper, err := New(&structViewerConfig, "TYK_")
			assert.NoError(t, err, "failed to instantiate viewer")

			if tc.shouldDeleteConfig {
				helper.config = nil
			}

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(helper.ConfigHandler)

			// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
			// directly and pass in our Request and ResponseRecorder.
			handler.ServeHTTP(rr, req)

			// Check the status code is what we expect.
			assert.Equal(t, tc.expectedStatusCode, rr.Code)

			// Check the response body is what we expect.
			if tc.expectedStatusCode == http.StatusOK {
				assert.JSONEq(t, tc.expectedJSONOutput, rr.Body.String())
			}
		})
	}
}

func TestDetailedConfigHandler(t *testing.T) {
	tcs := []struct {
		testName string

		givenConfig   interface{}
		queryParamVal string

		expectedStatusCode int
		expectedJSONOutput string
		shouldDeleteConfig bool
	}{
		{
			testName: "simple struct",
			givenConfig: struct {
				Name string `json:"field_name"`
			}{
				"field_value",
			},
			expectedStatusCode: http.StatusOK,
			expectedJSONOutput: toJSON(t, map[string]interface{}{
				"Name": map[string]interface{}{
					"config_field": "field_name",
					"env":          "TYK_FIELDNAME",
					"value":        "field_value",
					"obfuscated":   false,
				},
			}),
		},
		{
			testName:           "complex struct",
			givenConfig:        complexStruct,
			expectedStatusCode: http.StatusOK,
			expectedJSONOutput: toJSON(t, complexStructToMap()),
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
				Obfuscated:  getPointerBool(false),
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
				Obfuscated:  getPointerBool(false),
			}),
		},
		{
			testName:           "invalid field complexStruct via query param",
			givenConfig:        complexStruct,
			queryParamVal:      "data.object_3",
			expectedStatusCode: http.StatusOK,
			expectedJSONOutput: toJSON(t, EnvVar{}),
		},
		{
			testName: "not initialized",
			givenConfig: &struct {
				Name string `json:"field_name"`
			}{
				"field_value",
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedJSONOutput: "",
			shouldDeleteConfig: true,
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

			if tc.shouldDeleteConfig {
				helper.configMap = nil
			}

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(helper.DetailedConfigHandler)

			// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
			// directly and pass in our Request and ResponseRecorder.
			handler.ServeHTTP(rr, req)

			// Check the status code is what we expect.
			assert.Equal(t, tc.expectedStatusCode, rr.Code)

			if tc.expectedStatusCode == http.StatusOK {
				assert.JSONEq(t, tc.expectedJSONOutput, rr.Body.String())
			}
		})
	}
}

func TestEnvsHandler(t *testing.T) {
	tcs := []struct {
		testName           string
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
			expectedJSONOutput: fmt.Sprintln(`["FIELDNAME=field_value"]`),
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
			expectedJSONOutput: fmt.Sprintln(`["TEST_FIELDNAME=field_value"]`),
		},
		{
			testName:           "complex struct",
			givenConfig:        complexStruct,
			expectedStatusCode: http.StatusOK,
			expectedJSONOutput: fmt.Sprintln(
				`["NAME=name_value",` +
					`"DATA_OBJECT1=1",` +
					`"DATA_OBJECT2=true",` +
					`"METADATA_ID=99",` +
					`"METADATA_VALUE=key99",` +
					`"OMITTEDVALUE=''"]`,
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
				Obfuscated:  getPointerBool(false),
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
				Obfuscated:  getPointerBool(false),
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
			req, err := http.NewRequest("GET", "/", nil)
			assert.NoError(t, err)

			setQueryParams(req, EnvQueryKey, tc.queryParamVal)

			structViewerConfig := Config{Object: tc.givenConfig}
			helper, err := New(&structViewerConfig, tc.givenPrefix)
			assert.NoError(t, err, "failed to instantiate viewer")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(helper.EnvsHandler)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)

			// Determine whether the expected output is an array by trying to unmarshal it into a slice
			var expectedArray, actualArray []string
			expectedArrayErr := json.Unmarshal([]byte(tc.expectedJSONOutput), &expectedArray)
			actualArrayErr := json.Unmarshal(rr.Body.Bytes(), &actualArray)

			if expectedArrayErr == nil && actualArrayErr == nil {
				// Both JSON strings are arrays; compare using unordered comparison
				assert.ElementsMatch(t, expectedArray, actualArray)
			} else {
				// Not arrays, compare as ordered JSON strings
				assert.JSONEq(t, tc.expectedJSONOutput, rr.Body.String())
			}
		})
	}
}
