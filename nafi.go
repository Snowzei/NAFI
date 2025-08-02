package nafi

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"
)

// type for file reading function so it can be mocked
type FileReaderFunc func(path string) ([]byte, error)

var readFile FileReaderFunc = os.ReadFile

// Config parser object
type ConfigParserObj struct {
	data     map[string]interface{}
	raw      map[string]string
	fileType string
	iniFile  *ini.File
}

// NewConfigParserFromBytes parses config data from a byte slice, based on the provided file type.
func newConfigParserFromBytes(fileType string, content []byte) (*ConfigParserObj, error) {
	parser := &ConfigParserObj{
		data:     make(map[string]interface{}),
		raw:      make(map[string]string),
		fileType: fileType,
	}

	// Perform parsing based on filetype
	switch fileType {
	case "conf":
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				parser.raw[key] = val
			}
		}
	case "ini":
		iniFile, err := ini.Load(content)
		if err != nil {
			return nil, err
		}
		parser.iniFile = iniFile
	case "json":
		var jsonData map[string]interface{}
		if err := json.Unmarshal(content, &jsonData); err != nil {
			return nil, err
		}
		parser.data = jsonData
	case "yaml":
		var yamlData map[string]interface{}
		if err := yaml.Unmarshal(content, &yamlData); err != nil {
			return nil, err
		}
		parser.data = yamlData
	default:
		return nil, errors.New("unsupported file type " + fileType)
	}
	return parser, nil
}

// retieve nested value from data
func getNestedValue(data map[string]interface{}, key string) (interface{}, bool) {
	parts := strings.Split(key, ".")
	var current interface{} = data
	for _, part := range parts {
		switch curr := current.(type) {
		case map[string]interface{}:
			next, ok := curr[part]
			if !ok {
				return nil, false
			}
			current = next
		default:
			return nil, false
		}
	}
	return current, true
}

// Reads a filepath on the disk and parses it, returning a ConfigParserObj object.
//
// Supported file types:
//
// "conf", "ini", "json", "yaml"
func ConfigParser(filepath string, fileType string) (ConfigParserObj, error) {
	content, err := readFile(filepath)
	if err != nil {
		return ConfigParserObj{}, err
	}

	parser, err := newConfigParserFromBytes(fileType, content)
	if err != nil {
		return ConfigParserObj{}, err
	}
	return *parser, nil
}

// Get returns the value for a key, using dot notation for sectioned/nested formats
// 
// Example 1 - val, err := configParser.Get("foo")
//
// Example 2 - val, err := configParser.Get("foo.bar")
func (c *ConfigParserObj) Get(key string) (string, error) {
	// Check filetype of parser
	switch c.fileType {
	// Perform action for type conf
	case "conf":
		val, ok := c.raw[key]
		if !ok {
			return "", fmt.Errorf("key %q not found", key)
		}
		return val, nil
	// Perform action for type conf
	case "ini":
		if !strings.Contains(key, ".") {
			val := c.iniFile.Section("").Key(key).String()
			if val == "" {
				return "", fmt.Errorf("key %q not found", key)
			}
			return val, nil
		}
		parts := strings.SplitN(key, ".", 2)
		section, k := parts[0], parts[1]
		val := c.iniFile.Section(section).Key(k).String()
		if val == "" {
			return "", fmt.Errorf("key %q not found in section %q", k, section)
		}
		return val, nil
	// Perform action for type json or yaml
	case "json", "yaml":
		val, found := getNestedValue(c.data, key)
		if !found {
			return "", fmt.Errorf("key %q not found", key)
		}
		return fmt.Sprintf("%v", val), nil
	default:
		return "", errors.New("unsupported file type " + c.fileType)
	}
}
