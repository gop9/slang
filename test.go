package slang

import (
	"io"
	"io/ioutil"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"reflect"
)

type TestCaseDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Data struct {
		In  []interface{} `json:"in"`
		Out []interface{} `json:"out"`
	}
}

type TestDef struct {
	OperatorFile string        `json:"operatorFile"`
	Description  string        `json:"description"`
	TestCases    []TestCaseDef `json:"testCases"`
	valid        bool
}

// TestOperator reads a file with test data and its corresponding operator and performs the tests.
// It returns the number of failed and succeeded tests and and error in case something went wrong.
// Test failures do not lead to an error. Test failures are printed to the writer.
func TestOperator(testDataFilePath string, writer io.Writer, failFast bool) (int, int, error) {
	b, err := ioutil.ReadFile(testDataFilePath)

	if err != nil {
		return 0, 0, err
	}

	test := TestDef{}
	json.Unmarshal(b, &test)

	if !test.Valid() {
		err := test.Validate()
		if err != nil {
			return 0, 0, err
		}
	}

	o, err := ReadOperator(path.Join(path.Dir(testDataFilePath), test.OperatorFile))

	if err != nil {
		return 0, 0, err
	}

	fmt.Fprintf(writer, "Parsed %s successfully\n", test.OperatorFile)

	compiled := o.Compile()
	fmt.Fprintln(writer, "Operator compiled")
	fmt.Fprintf(writer, "Operators simplified:  %3d\n", compiled)
	fmt.Fprintf(writer, "Total basic operators: %3d\n", len(o.Children()))

	o.Out().Bufferize()

	o.Start()
	defer o.Stop()
	fmt.Fprintln(writer, "Operator started")

	succs := 0
	fails := 0

	for i, tc := range test.TestCases {
		fmt.Fprintf(writer, "Test case %3d/%3d: %s\n", i+1, len(test.TestCases), tc.Name)

		success := true

		for j := range tc.Data.In {
			in := tc.Data.In[j]
			expected := tc.Data.Out[j]

			o.In().Push(in)
			actual := o.Out().Pull()

			if !reflect.DeepEqual(expected, actual) {
				fmt.Fprintf(writer, "  expected: %v\n", expected)
				fmt.Fprintf(writer, "  actual:   %v\n", actual)

				success = false

				if failFast {
					return succs, fails + 1, nil
				}
			}
		}

		if success {
			succs++
		} else {
			fails++
		}
	}

	return succs, fails, nil
}

func (t TestDef) Validate() error {
	if len(t.OperatorFile) == 0 {
		return errors.New("no operator file given")
	}

	for _, tc := range t.TestCases {
		if len(tc.Data.In) != len(tc.Data.Out) {
			return errors.New(fmt.Sprintf(`data count unequal in test case "%s"`, tc.Name))
		}
	}

	t.valid = true
	return nil
}

func (t TestDef) Valid() bool {
	return t.valid
}