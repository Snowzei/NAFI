package nafi

import (
	"errors"
	"strings"
	"testing"
)

func TestConfigParser(t *testing.T) {
	// Set testing object for subtesting
	tests := []struct {
		name     string
		fileType string
		content  string
		cases    map[string]string
	}{
		{
			name:     "conf file",
			fileType: "conf",
			content: `
# comment
key1=value1

key2 = value2
key3= value3
key4 = rAR#vW@='4EV
`,
			cases: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "rAR#vW@='4EV",
				"key5": "",
			},
		},
		{
			name:     "ini file",
			fileType: "ini",
			content: `
[section1]
foo= bar
baz =bat

[section2]
qux = quux
`,
			cases: map[string]string{
				"section1.foo": "bar",
				"section1.baz": "bat",
				"section2.qux": "quux",
				"section1.missing": "",
			},
		},
		{
			name:     "json file flat",
			fileType: "json",
			content: `
{
  "key1": "value1",
  "key2": "value2",
  "intkey": 22
}
`,
			cases: map[string]string{
				"key1":   "value1",
				"key2":   "value2",
				"intkey": "22",
				"missing": "",
			},
		},
		{
			name:     "json file nested",
			fileType: "json",
			content: `
{
  "section1": {
    "foo": "bar"
  },
  "plain": "top"
}
`,
			cases: map[string]string{
				"section1.foo": "bar",
				"plain":        "top",
				"missing":      "",
			},
		},
		{
			name:     "yaml file nested",
			fileType: "yaml",
			content: `
section1:
  foo: bar
  baz: bat
plain: topvalue
`,
			cases: map[string]string{
				"section1.foo": "bar",
				"section1.baz": "bat",
				"plain":        "topvalue",
				"missing":      "",
			},
		},
	}

	// Iterate through tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := newConfigParserFromBytes(tt.fileType, []byte(tt.content))
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			for lookup, expected := range tt.cases {
				val, err := parser.Get(lookup)
				if err != nil {
					t.Errorf("Get(%q) unexpected error: %v", lookup, err)
				}
				if val != expected {
					t.Errorf("Get(%q) = %q; want %q", lookup, val, expected)
				}
			}
			// Check that a random missing key returns "" and no error
			val, err := parser.Get("does_not_exist")
			if err != nil {
				t.Errorf("Get(%q) unexpected error: %v", "does_not_exist", err)
			}
			if val != "" {
				t.Errorf("Get(%q) = %q; want empty string", "does_not_exist", val)
			}
		})
	}
}

// Test script can parse from filepath corretcly
func TestConfFileParsing(t *testing.T) {
	t.Run("successful parse", func(t *testing.T) {
		mockData := []byte(`
username = foo
password = bar
`)
		// Override file reading for testing
		readFile = func(_ string) ([]byte, error) {
			return mockData, nil
		}

		_, err := ConfigParser("dummy.conf", "conf")
		if err != nil {
			t.Fatalf("Failed to parse config: %v", err)
		}
	})

	t.Run("read fail", func(t *testing.T) {
		// Override file reading for testing to always fail
		readFile = func(_ string) ([]byte, error) {
			return nil, errors.New("mock read error")
		}

		_, err := ConfigParser("dummy.conf", "conf")
		if err == nil {
			t.Fatalf("Expected error on file read, got nil")
		}
		if err.Error() != "mock read error" {
			t.Errorf("Expected 'mock read error', got %v", err)
		}
	})
}

// Test config object generation errors
func TestNewConfigParserFromBytesErrors(t *testing.T) {
	t.Run("unsupported file type", func(t *testing.T) {
		_, err := newConfigParserFromBytes("unknown", []byte("foo"))
		if err == nil || !strings.Contains(err.Error(), "unsupported file type") {
			t.Errorf("Expected unsupported file type error, got %v", err)
		}
	})

	t.Run("ini parse error", func(t *testing.T) {
		_, err := newConfigParserFromBytes("ini", []byte("invalid ini"))
		if err == nil {
			t.Errorf("Expected ini parse error, got nil")
		}
	})

	t.Run("json parse error", func(t *testing.T) {
		_, err := newConfigParserFromBytes("json", []byte("{invalid json"))
		if err == nil {
			t.Errorf("Expected json parse error, got nil")
		}
	})

	t.Run("yaml parse error", func(t *testing.T) {
		_, err := newConfigParserFromBytes("yaml", []byte("invalid: [yaml"))
		if err == nil {
			t.Errorf("Expected yaml parse error, got nil")
		}
	})
}

// Test edge cases for retrieving nested values
func TestGetNestedValueEdgeCases(t *testing.T) {
	t.Run("not found at leaf", func(t *testing.T) {
		data := map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}}
		val, ok := getNestedValue(data, "foo.baz")
		if ok {
			t.Errorf("Expected not found, got %v", val)
		}
	})

	t.Run("not map in path", func(t *testing.T) {
		data := map[string]interface{}{"foo": "bar"}
		val, ok := getNestedValue(data, "foo.baz")
		if ok {
			t.Errorf("Expected not found, got %v", val)
		}
	})
}

// Test for errors in file format parsing
func TestConfigParserObjGetErrors(t *testing.T) {
	t.Run("conf missing key", func(t *testing.T) {
		parser, _ := newConfigParserFromBytes("conf", []byte("foo=bar"))
		val, err := parser.Get("missing")
		if err != nil {
			t.Errorf("Get(%q) unexpected error: %v", "missing", err)
		}
		if val != "" {
			t.Errorf("Get(%q) = %q; want empty string", "missing", val)
		}
	})

	t.Run("ini missing key", func(t *testing.T) {
		parser, _ := newConfigParserFromBytes("ini", []byte("[s]\nfoo=bar"))
		val, err := parser.Get("s.missing")
		if err != nil {
			t.Errorf("Get(%q) unexpected error: %v", "s.missing", err)
		}
		if val != "" {
			t.Errorf("Get(%q) = %q; want empty string", "s.missing", val)
		}
	})

	t.Run("json unsupported type", func(t *testing.T) {
		parser := &configParserObj{fileType: "unsupported"}
		_, err := parser.Get("any")
		if err == nil {
			t.Errorf("Expected error for unsupported file type")
		}
	})
}