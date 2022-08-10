package struct_viewer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Exported    string
	notExported bool

	StrField struct {
		Test  string
		Other struct {
			OtherTest  bool
			nonEmbeded string
		}
	}

	JsonExported int `json:"name"`
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
			expectedEnvs: []string{"KEY:Value"},
		},
		{
			testName: "KEY:VALUE with json tag",
			testStruct: struct {
				Key string `json:"json_name"`
			}{
				Key: "Value",
			},
			expectedLen:  1,
			expectedEnvs: []string{"JSONNAME:Value"},
		},
		{
			testName: "KEY:VALUE with json tag and omitempty",
			testStruct: struct {
				Key string `json:"json_name,omitempty"`
			}{
				Key: "Value",
			},
			expectedLen:  1,
			expectedEnvs: []string{"JSONNAME:Value"},
		},
		{
			testName: "KEY:VALUE with json '-' tag",
			testStruct: struct {
				Key string `json:"-"`
			}{
				Key: "Value",
			},
			expectedLen:  1,
			expectedEnvs: []string{"KEY:Value"},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			structViewerConfig := Config{Object: tc.testStruct}
			helper := New(&structViewerConfig, "")
			envs := helper.ParseEnvs()

			assert.Len(t, envs, tc.expectedLen)
			assert.EqualValues(t, tc.expectedEnvs, envs)
		})
	}
}

func TestParseEnvsLen(t *testing.T) {
	testStruct := testStruct{
		Exported:    "val1",
		notExported: true,

		StrField: struct {
			Test  string
			Other struct {
				OtherTest  bool
				nonEmbeded string
			}
		}{Test: "test"},
		JsonExported: 5,
	}

	structViewerConfig := Config{Object: testStruct}
	helper := New(&structViewerConfig, "TYK_")

	envs := helper.ParseEnvs()

	assert.Len(t, envs, 4)
}

func TestParseEnvsPrefix(t *testing.T) {
	testStruct := testStruct{
		Exported:    "val1",
		notExported: true,

		StrField: struct {
			Test  string
			Other struct {
				OtherTest  bool
				nonEmbeded string
			}
		}{Test: "test"},
		JsonExported: 5,
	}

	prefix := "TYK_TEST_"
	structViewerConfig := Config{Object: testStruct}
	helper := New(&structViewerConfig, prefix)

	envs := helper.ParseEnvs()

	for _, env := range envs {
		assert.True(t, strings.HasPrefix(env, prefix))
	}
}

func TestParseComments(t *testing.T) {
	prefix := "TYK_TEST_"
	structViewerConfig := Config{Object: ExampleConfig{}}
	helper := New(&structViewerConfig, prefix)
	err := helper.parseComments()
	assert.NoError(t, err, "failed to parse comments")

	for _, env := range helper.Envs() {
		assert.NotEmpty(t, env.Desc, env.Key)
	}
}
