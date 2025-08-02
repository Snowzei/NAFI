# NAFI: Not Another Form Interpreter

NAFI is a simple configuration file parser for Go, supporting multiple formats including `.conf`, `.ini`, `.json`, and `.yaml`. It provides a quick interface for loading and accessing configuration values from files, making it easy to manage application settings regardless of the file format.

## Installation

Add to your `go.mod`:

```bash
go get github.com/Snowzei/NAFI
```

Import in your code:

```go
import "github.com/Snowzei/NAFI"
```

## Usage

### Load a Configuration File

```go
config, err := nafi.ConfigParser("path/to/config.json", "json")
if err != nil {
    panic(err)
}
```

### Retrieve Values

For flat key-value formats (`.conf`):

```go
val, err := config.Get("keyname")
```

For sectioned formats (`.ini`):

```go
val, err := config.Get("section.keyname")
```

For nested formats (`.json`, `.yaml`):

```go
val, err := config.Get("parent.child.keyname")
```

### Supported File Types

- `conf`: Simple key-value pairs, one per line (`key = value`)
- `ini`: INI files with sections and keys
- `json`: JSON files with nested objects
- `yaml`: YAML files with nested structures

## API Reference

### ConfigParser

```go
func ConfigParser(filepath string, fileType string) (ConfigParserObj, error)
```

Reads and parses a configuration file, returning a `ConfigParserObj`.

### ConfigParserObj.Get

```go
func (c *ConfigParserObj) Get(key string) (string, error)
```

Retrieves the value for the specified key, supporting dot notation for nested/sectioned formats.

## Example

#### config.yaml

```yaml
server:
  host: "localhost"
  port: 8080
```

#### main.go

```go
config, err := nafi.ConfigParser("config.yaml", "yaml")
if err != nil {
    panic(err)
}

host, _ := config.Get("server.host") // returns "localhost"
port, _ := config.Get("server.port") // returns "8080"
```

## License

MIT

---

**NAFI**: Not Another Form Interpreter – Unified, simple config parsing for Go.