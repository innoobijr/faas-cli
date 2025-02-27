package commands

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/innoobijr/faas-cli/schema"
	"github.com/innoobijr/faas-provider/types"
)

func Test_getFunctionURLs(t *testing.T) {
	cases := []struct {
		name              string
		gateway           string
		functionName      string
		functionNamespace string
		expectedURL       string
		expectedAsyncURL  string
	}{
		{"localhost", "http://127.0.0.1:8080", "figlet", "alpha", "http://127.0.0.1:8080/function/figlet.alpha", "http://127.0.0.1:8080/async-function/figlet.alpha"},
		{"secure site", "https://example.com", "nodeinfo", "beta", "https://example.com/function/nodeinfo.beta", "https://example.com/async-function/nodeinfo.beta"},
		{"no namespace", "https://example.com:31112", "nodeinfo", "", "https://example.com:31112/function/nodeinfo", "https://example.com:31112/async-function/nodeinfo"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			url, asyncURL := getFunctionURLs(tc.gateway, tc.functionName, tc.functionNamespace)

			if url != tc.expectedURL || asyncURL != tc.expectedAsyncURL {
				t.Fatalf("incorrect URL(s), want: %q and %q, got: %q and %q", tc.expectedURL, tc.expectedAsyncURL, url, asyncURL)
			}
		})
	}
}

func TestDescribeOuput(t *testing.T) {
	spaces := regexp.MustCompile(`[ ]{2,}`)
	cases := []struct {
		name           string
		function       schema.FunctionDescription
		verbose        bool
		expectedOutput string
	}{
		{
			name: "non-verbose minimal output",
			function: schema.FunctionDescription{
				FunctionStatus: types.FunctionStatus{
					Name:        "figlet",
					Image:       "openfaas/figlet:latest",
					Labels:      &map[string]string{},
					Annotations: &map[string]string{},
				},
				Status: "Ready",
			},
			verbose:        false,
			expectedOutput: "Name:\tfiglet\nStatus:\tReady\nReplicas:\t0\nAvailable Replicas: 0\nInvocations:\t0\nImage:\topenfaas/figlet:latest\nFunction Process:\t<default>\n",
		},
		{
			name: "verbose minimal output",
			function: schema.FunctionDescription{
				FunctionStatus: types.FunctionStatus{
					Name:        "figlet",
					Image:       "openfaas/figlet:latest",
					Labels:      &map[string]string{},
					Annotations: &map[string]string{},
				},
				Status: "Ready",
			},
			verbose:        true,
			expectedOutput: "Name:\tfiglet\nStatus:\tReady\nReplicas:\t0\nAvailable Replicas: 0\nInvocations:\t0\nImage:\topenfaas/figlet:latest\nFunction Process:\t<default>\nURL:\t<none>\nAsync URL:\t<none>\nLabels:\t<none>\nAnnotations:\t<none>\nConstraints:\t<none>\nEnvironment:\t<none>\nSecrets:\t<none>\nRequests:\t<none>\nLimits:\t<none>\nUsage:\t<none>\n",
		},
		{
			name: "non-verbose formats output with non-empty labels, env variables, and secrets",
			function: schema.FunctionDescription{
				FunctionStatus: types.FunctionStatus{
					Name:        "figlet",
					Image:       "openfaas/figlet:latest",
					Labels:      &map[string]string{"quadrant": "alpha"},
					Annotations: &map[string]string{},
					EnvVars:     map[string]string{"FOO": "bar"},
					Secrets:     []string{"db-password"},
				},
				Status: "Ready",
			},
			verbose:        false,
			expectedOutput: "Name:\tfiglet\nStatus:\tReady\nReplicas:\t0\nAvailable Replicas: 0\nInvocations:\t0\nImage:\topenfaas/figlet:latest\nFunction Process:\t<default>\nLabels:\n quadrant: alpha\nEnvironment:\n FOO: bar\nSecrets:\n - db-password\n",
		},
		{
			name: "verbose formats output with non-empty labels, env variables, and secrets",
			function: schema.FunctionDescription{
				FunctionStatus: types.FunctionStatus{
					Name:        "figlet",
					Image:       "openfaas/figlet:latest",
					Labels:      &map[string]string{"quadrant": "alpha"},
					Annotations: &map[string]string{},
					EnvVars:     map[string]string{"FOO": "bar"},
					Secrets:     []string{"db-password"},
				},
				Status: "Ready",
			},
			verbose:        true,
			expectedOutput: "Name:\tfiglet\nStatus:\tReady\nReplicas:\t0\nAvailable Replicas: 0\nInvocations:\t0\nImage:\topenfaas/figlet:latest\nFunction Process:\t<default>\nURL:\t<none>\nAsync URL:\t<none>\nLabels:\n quadrant: alpha\nAnnotations:\t<none>\nConstraints:\t<none>\nEnvironment:\n FOO: bar\nSecrets:\n - db-password\nRequests:\t<none>\nLimits:\t<none>\nUsage:\t<none>\n",
		},
		{
			name: "formats non-empty usage",
			function: schema.FunctionDescription{
				FunctionStatus: types.FunctionStatus{
					Name:        "figlet",
					Image:       "openfaas/figlet:latest",
					Labels:      &map[string]string{},
					Annotations: &map[string]string{},
					Usage: &types.FunctionUsage{
						TotalMemoryBytes: 1024 * 1024 * 1024,
						CPU:              1.5,
					},
				},
				Status: "Ready",
			},
			verbose:        false,
			expectedOutput: "Name:\tfiglet\nStatus:\tReady\nReplicas:\t0\nAvailable Replicas: 0\nInvocations:\t0\nImage:\topenfaas/figlet:latest\nFunction Process:\t<default>\nUsage:\n\tRAM:\t1024.00 MB\n\tCPU:\t2 Mi\n",
		},
		{
			name: "Multiple env variables",
			function: schema.FunctionDescription{
				FunctionStatus: types.FunctionStatus{
					Name:        "figlet",
					Image:       "openfaas/figlet:latest",
					Labels:      &map[string]string{"quadrant": "alpha"},
					Annotations: &map[string]string{},
					EnvVars:     map[string]string{"DDD": "ddd", "AAA": "aaa", "BBB": "bbb", "CCC": "ccc"},
					Secrets:     []string{"db-password"},
				},
				Status: "Ready",
			},
			verbose:        true,
			expectedOutput: "Name:\tfiglet\nStatus:\tReady\nReplicas:\t0\nAvailable Replicas: 0\nInvocations:\t0\nImage:\topenfaas/figlet:latest\nFunction Process:\t<default>\nURL:\t<none>\nAsync URL:\t<none>\nLabels:\n quadrant: alpha\nAnnotations:\t<none>\nConstraints:\t<none>\nEnvironment:\n AAA: aaa\n BBB: bbb\n CCC: ccc\n DDD: ddd\nSecrets:\n - db-password\nRequests:\t<none>\nLimits:\t<none>\nUsage:\t<none>\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var dst bytes.Buffer
			printFunctionDescription(&dst, tc.function, tc.verbose)
			result := spaces.ReplaceAllString(dst.String(), "\t")
			if result != tc.expectedOutput {
				t.Fatalf("incorrect output,\nwant: %q\nnorm: %q\ngot: %q", tc.expectedOutput, result, dst.String())
			}
		})
	}
}

func TestGenerateMapOrder(t *testing.T) {
	var generateMapOrderTestcases = []struct {
		Name       string
		Input      map[string]string
		Output     []string
		expectFail bool
	}{
		{
			Name: "One item",
			Input: map[string]string{
				"AAA": "aaa",
			},
			Output:     []string{"AAA"},
			expectFail: false,
		},
		{
			Name: "Multiple items",
			Input: map[string]string{
				"AAA": "aaa",
				"BBB": "bbb",
				"CCC": "ccc",
				"DDD": "ddd",
			},
			Output:     []string{"AAA", "BBB", "CCC", "DDD"},
			expectFail: false,
		},
		{
			Name: "Multiple items but use a value",
			Input: map[string]string{
				"AAA": "aaa",
				"BBB": "bbb",
				"CCC": "ccc",
				"DDD": "ddd",
			},
			Output:     []string{"AAA", "BBB", "CCC", "ddd"},
			expectFail: true,
		},
	}
	for _, testcase := range generateMapOrderTestcases {
		orderedSlice := generateMapOrder(testcase.Input)
		if len(orderedSlice) != len(testcase.Output) {
			t.Errorf("Slice sizes do not match: %s", testcase.Name)
			t.Fail()
		}
		for i, v := range testcase.Output {
			if v != orderedSlice[i] && !testcase.expectFail {
				t.Errorf("Exected %s got %s: %s", v, orderedSlice[i], testcase.Name)
				t.Fail()
			}
		}
	}
}
