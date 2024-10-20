package structviewer

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type InnerStructType struct {
	// DummyAddr represents an address.
	DummyAddr string `json:"dummy_addr"`
}

type StructType struct {
	// Enable represents status.
	Enable bool            `json:"enable"`
	Inner  InnerStructType `json:"inner"`
}

type testStruct struct {
	// Exported represents a sample exported field.
	Exported    string `json:"exported"`
	notExported bool

	// StrField is a struct field.
	StrField struct {
		// Test is a field of struct type.
		Test  string `json:"test"`
		Other struct {
			// OtherTest represents a field of sub-struct.
			OtherTest   bool `json:"other_test"`
			nonEmbedded string
		}
	}

	// ST is another struct type.
	ST StructType `json:"st"`

	// JSONExported includes a JSON tag.
	JSONExported int `json:"json_exported"`
}

func TestViewerNew(t *testing.T) {
	cases := []struct {
		testName     string
		configStruct *Config
		expectedErr  error
	}{
		{
			testName:     "nil config struct",
			configStruct: nil,
			expectedErr:  ErrNilConfig,
		},
		{
			testName:     "parsing nil struct",
			configStruct: &Config{},
			expectedErr:  ErrEmptyStruct,
		},
		{
			testName:     "invalid object type",
			expectedErr:  ErrInvalidObjectType,
			configStruct: &Config{Object: "string"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(t *testing.T) {
			_, err := New(tc.configStruct, "")
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestParseEnvsValues(t *testing.T) {
	tcs := []struct {
		testName     string
		testStruct   interface{}
		expectedLen  int
		expectedEnvs []string
	}{
		{
			testName: "KEY:VALUE common struct",
			testStruct: struct {
				Key string
			}{
				Key: "Value",
			},
			expectedLen:  1,
			expectedEnvs: []string{"KEY=Value"},
		},
		{
			testName: "KEY:VALUE with json tag",
			testStruct: struct {
				Key string `json:"json_name"`
			}{
				Key: "Value",
			},
			expectedLen:  1,
			expectedEnvs: []string{"JSONNAME=Value"},
		},
		{
			testName: "KEY:VALUE with json tag and omitempty",
			testStruct: struct {
				Key string `json:"json_name,omitempty"`
			}{
				Key: "Value",
			},
			expectedLen:  1,
			expectedEnvs: []string{"JSONNAME=Value"},
		},
		{
			testName: "KEY:VALUE with json '-' tag",
			testStruct: struct {
				Key string `json:"-"`
			}{
				Key: "Value",
			},
			expectedLen:  1,
			expectedEnvs: []string{"KEY=Value"},
		},
		{
			testName: "struct with nested map and struct",
			testStruct: struct {
				Key string
				Map map[string]string
				Str struct {
					Inner string
				}
			}{
				Key: "Value",
				Map: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
				Str: struct {
					Inner string
				}{
					Inner: "inner",
				},
			},
			expectedLen:  4,
			expectedEnvs: []string{"KEY=Value", "MAP_KEY1=value1", "MAP_KEY2=value2", "STR_INNER=inner"},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			structViewerConfig := Config{Object: tc.testStruct}
			helper, err := New(&structViewerConfig, "")
			assert.NoError(t, err, "failed to instantiate viewer")

			envs := helper.ParseEnvs()

			assert.Len(t, envs, tc.expectedLen)
			// Compare arrays disregarding the order
			assert.ElementsMatch(t, tc.expectedEnvs, envs)
		})
	}
}

func TestParseEnvsLen(t *testing.T) {
	testStruct := testStruct{
		Exported:    "val1",
		notExported: true,
		StrField: struct {
			Test  string `json:"test"`
			Other struct {
				OtherTest   bool `json:"other_test"`
				nonEmbedded string
			}
		}{Test: "test"},
		JSONExported: 5,
	}

	structViewerConfig := Config{Object: testStruct}
	helper, err := New(&structViewerConfig, "TYK_")
	assert.NoError(t, err, "failed to instantiate viewer")

	envs := helper.ParseEnvs()

	assert.Len(t, envs, 6)
}

func TestParseEnvsPrefix(t *testing.T) {
	testStruct := testStruct{
		Exported:    "val1",
		notExported: true,

		StrField: struct {
			Test  string `json:"test"`
			Other struct {
				OtherTest   bool `json:"other_test"`
				nonEmbedded string
			}
		}{Test: "test"},
		JSONExported: 5,
	}

	prefix := "TYK_TEST_"
	structViewerConfig := Config{Object: testStruct}
	helper, err := New(&structViewerConfig, prefix)
	assert.NoError(t, err, "failed to instantiate viewer")

	envs := helper.ParseEnvs()

	for _, env := range envs {
		assert.True(t, strings.HasPrefix(env, prefix))
	}
}

func TestParseComments(t *testing.T) {
	viewer, err := New(&Config{Object: testStruct{}, Path: "./parser_test.go"}, "TYK_")
	assert.NoError(t, err, "failed to instantiate viewer")
	err = viewer.parseComments()
	assert.NoError(t, err, "failed to parse comments")

	for _, env := range viewer.Envs() {
		assert.NotEmpty(t, env.Description, "failed to parse %v comments", env.field)
	}
}

func TestEnvNotation(t *testing.T) {
	const prefix = "TYK_"
	viewerWithComment, err := New(
		&Config{
			Object: testStruct{},
			Path:   "./parser_test.go",
		},
		prefix,
	)
	assert.NoError(t, err, "failed to instantiate viewer with comment parser")

	testCases := []struct {
		viewer          *Viewer
		jsonNotation    string
		expectedEnv     string
		expectedComment string
	}{
		{
			viewer:          viewerWithComment,
			jsonNotation:    "",
			expectedEnv:     "",
			expectedComment: "",
		},
		{
			viewer:          viewerWithComment,
			jsonNotation:    "non_existent",
			expectedEnv:     "",
			expectedComment: "",
		},
		{
			viewer:          viewerWithComment,
			jsonNotation:    "st.enable",
			expectedEnv:     fmt.Sprintf("%s%s", prefix, "ST_ENABLE"),
			expectedComment: "Enable represents status.",
		},
		{
			viewer:          viewerWithComment,
			jsonNotation:    "st.inner.dummy_addr",
			expectedEnv:     fmt.Sprintf("%s%s", prefix, "ST_INNER_DUMMYADDR"),
			expectedComment: "DummyAddr represents an address.",
		},
		{
			viewer:          viewerWithComment,
			jsonNotation:    "json_exported",
			expectedEnv:     fmt.Sprintf("%s%s", prefix, "JSONEXPORTED"),
			expectedComment: "JSONExported includes a JSON tag.",
		},
	}

	for _, tc := range testCases {
		envVar := tc.viewer.EnvNotation(tc.jsonNotation)
		assert.Equal(t, tc.expectedEnv, envVar.Env, "failed to get env notation of %s", tc.jsonNotation)
	}
}

func TestJSONNotation(t *testing.T) {
	const prefix = "TYK_"
	viewerWithComment, err := New(
		&Config{
			Object:        testStruct{},
			Path:          "./parser_test.go",
			ParseComments: true,
		},
		prefix,
	)
	assert.NoError(t, err, "failed to instantiate viewer with comment parser")

	testCases := []struct {
		viewer          *Viewer
		envNotation     string
		expectedJSON    string
		expectedComment string
	}{
		{
			viewer:          viewerWithComment,
			envNotation:     "",
			expectedJSON:    "",
			expectedComment: "",
		},
		{
			viewer:          viewerWithComment,
			envNotation:     fmt.Sprintf("%s%s", prefix, "NONEXISTENT"),
			expectedJSON:    "",
			expectedComment: "",
		},
		{
			viewer:          viewerWithComment,
			envNotation:     fmt.Sprintf("%s%s", prefix, "EXPORTED"),
			expectedJSON:    "exported",
			expectedComment: "Exported represents a sample exported field.",
		},
		{
			viewer:          viewerWithComment,
			envNotation:     fmt.Sprintf("%s%s", prefix, "ST_ENABLE"),
			expectedJSON:    "st.enable",
			expectedComment: "Enable represents status.",
		},
		{
			viewer:          viewerWithComment,
			envNotation:     fmt.Sprintf("%s%s", prefix, "ST_INNER_DUMMYADDR"),
			expectedJSON:    "st.inner.dummy_addr",
			expectedComment: "DummyAddr represents an address.",
		},
		{
			viewer:          viewerWithComment,
			envNotation:     fmt.Sprintf("%s%s", prefix, "JSONEXPORTED"),
			expectedJSON:    "json_exported",
			expectedComment: "JSONExported includes a JSON tag.",
		},
	}

	for _, tc := range testCases {
		envVar := tc.viewer.JSONNotation(tc.envNotation)
		assert.Equal(t, tc.expectedJSON, envVar.ConfigField, "failed to get JSON notation of %s", tc.envNotation)
		assert.Equal(t, tc.expectedComment, envVar.Description, "failed to parse comments of %s", tc.envNotation)
	}
}

func TestObfuscateTags(t *testing.T) {
	tests := []struct {
		name    string
		given   interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name: "obfuscate single exported field",
			given: &struct {
				Exported string `json:"exported" structviewer:"obfuscate"`
				Test     int    `json:"test"`
			}{
				Exported: "value",
				Test:     10,
			},
			want: &struct {
				Exported string `json:"exported" structviewer:"obfuscate"`
				Test     int    `json:"test"`
			}{
				Exported: "*REDACTED*",
				Test:     10,
			},
			wantErr: false,
		},
		{
			name: "obfuscate nested field",
			given: &struct {
				Exported string `json:"exported"`
				Inner    struct {
					InnerField string `json:"inner_field" structviewer:"obfuscate"`
					NotObf     string `json:"not_obf"`
				} `json:"inner"`
			}{
				Exported: "value",
				Inner: struct {
					InnerField string `json:"inner_field" structviewer:"obfuscate"`
					NotObf     string `json:"not_obf"`
				}{
					InnerField: "inner",
					NotObf:     "not obfuscated",
				},
			},
			want: &struct {
				Exported string `json:"exported"`
				Inner    struct {
					InnerField string `json:"inner_field" structviewer:"obfuscate"`
					NotObf     string `json:"not_obf"`
				} `json:"inner"`
			}{
				Exported: "value",
				Inner: struct {
					InnerField string `json:"inner_field" structviewer:"obfuscate"`
					NotObf     string `json:"not_obf"`
				}{
					InnerField: "*REDACTED*",
					NotObf:     "not obfuscated",
				},
			},
			wantErr: false,
		},
		{
			name: "obfuscate deep nested field and top level",

			given: &struct {
				Exported string `json:"exported" structviewer:"obfuscate"`
				Inner    struct {
					InnerField string `json:"inner_field" structviewer:"obfuscate"`
					NotObf     string `json:"not_obf"`
				} `json:"inner"`
			}{
				Exported: "value",
				Inner: struct {
					InnerField string `json:"inner_field" structviewer:"obfuscate"`
					NotObf     string `json:"not_obf"`
				}{
					InnerField: "inner",
					NotObf:     "not obfuscated",
				},
			},
			want: &struct {
				Exported string `json:"exported" structviewer:"obfuscate"`
				Inner    struct {
					InnerField string `json:"inner_field" structviewer:"obfuscate"`
					NotObf     string `json:"not_obf"`
				} `json:"inner"`
			}{
				Exported: "*REDACTED*",
				Inner: struct {
					InnerField string `json:"inner_field" structviewer:"obfuscate"`
					NotObf     string `json:"not_obf"`
				}{
					InnerField: "*REDACTED*",
					NotObf:     "not obfuscated",
				},
			},
			wantErr: false,
		},
		{
			name:    "Handle error when config is not pointer",
			given:   struct{}{},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := obfuscateTags(tt.given)

			if (err != nil) != tt.wantErr {
				t.Errorf("obfuscateTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("obfuscateTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
