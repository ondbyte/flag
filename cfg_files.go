package flag

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/go-yaml/yaml"
)

// EnvToMap parses an environment file content and returns the key-value pairs as a map.
func EnvToMap(content string) (map[string]string, error) {
	envMap := make(map[string]string)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, errors.New("invalid line format: " + line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		envMap[key] = value
	}

	return envMap, nil
}

// MapToYAML writes a map to a YAML string
func MapToYAML(data map[string]interface{}) (string, error) {
	yamlContent, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(yamlContent), nil
}

// MapToJSON writes a map to a JSON string
func MapToJSON(data map[string]interface{}) (string, error) {
	m, err := stringMap(data)
	if err != nil {
		return "", fmt.Errorf("error map to JSON: %v", err)
	}
	jsonContent, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(jsonContent), nil
}

// map to a TOML string
func MapToTOML(data map[string]interface{}) (string, error) {
	b := new(bytes.Buffer)
	w := toml.NewEncoder(b)
	config, err := stringMap(data)
	if err != nil {
		return "", fmt.Errorf("unable to map to TOML : %v", err)
	}
	err = w.Encode(config)
	if err != nil {
		return "", fmt.Errorf("unable to map to TOML : %v", err)
	}
	return b.String(), nil
}

// JSONToMap reads the contents of a JSON file from a string and returns a map[interface{}]interface{}
func JSONToMap(content string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := json.Unmarshal([]byte(content), &result)
	if err != nil {
		return nil, err
	}

	result2 := make(map[string]interface{})
	for k, v := range result {
		result2[k] = v
	}
	return result2, nil
}

// YAMLToMap reads the contents of a YAML file from a string and returns a map[interface{}]interface{}
func YAMLToMap(content string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(content), &result)
	if err != nil {
		return nil, err
	}
	m, err := stringMap(result)
	if err != nil {
		return m, fmt.Errorf("unable create map from YAML : %v", err)
	}
	return m, nil
}

func TOMLToMap(content string) (map[string]interface{}, error) {
	var config map[string]interface{}
	_, err := toml.Decode(content, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

/* // ReadProperties reads the contents of a Property file from a string and returns a map[string]string
func ReadProperties(content string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
	result2 := make(map[string]interface{})
	for k, v := range result {
		result2[k] = v
	}
		}
		parts :2= strings.Split(line, "#")
		keyValue := strings.TrimSpace(parts[0])
		if len(keyValue) == 0 {
			continue
		}
		parts = strings.SplitN(keyValue, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			result[key] = value
		}
	}

	return result, nil
} */

func jsonnify(v interface{}) (string, error) {
	if v == nil {
		return "", nil
	}
	_, ok := v.(map[string]interface{})
	if ok {
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return fmt.Sprintf("%v", v), nil
}

func getValueByNotationArray(inputMap map[string]interface{}, notation []string) (s string, err error) {
	data, err := stringMap(inputMap)
	if err != nil {
		return "", fmt.Errorf("unable to get value by dot notation : %v", err)
	}
	key := notation[0]
	notation = notation[1:]
	nextData := data[key]
	if len(notation) == 0 {
		return jsonnify(nextData)
	}
	if nextMap, ok := nextData.(map[string]interface{}); ok {
		//value is a map, go recursive
		v, err := getValueByNotationArray(nextMap, notation)
		if err != nil {
			return "", fmt.Errorf("unable to get value by dot notation : %v", err)
		}
		return v, nil
	}
	return "", fmt.Errorf("value not found for dot notation")
}

func getValueByDotNotation(inputMap map[string]interface{}, not string) (s string, err error) {
	sNotation := strings.Split(not, ".")
	return getValueByNotationArray(inputMap, sNotation)
}

func stringMap(inputMap interface{}) (map[string]interface{}, error) {
	ip, ok := inputMap.(map[string]interface{})
	ip2 := make(map[interface{}]interface{})
	outputMap := make(map[string]interface{})
	if !ok {
		ip2, ok = inputMap.(map[interface{}]interface{})
		if !ok {
			return outputMap, fmt.Errorf("unable to make a string keyed map")
		}
		for key, value := range ip2 {
			// Convert key to string
			strKey := fmt.Sprintf("%v", key)

			switch value.(type) {
			case map[interface{}]interface{}:
				// Recursively convert nested maps
				newVal, err := stringMap(value)
				if err != nil {
					return outputMap, fmt.Errorf("unable to create string key map as value for key %v beacuse error: %v", strKey, err)
				}
				outputMap[strKey] = newVal
			default:
				// For other types, directly assign the value
				outputMap[strKey] = value
			}
		}

		return outputMap, nil
	}

	for key, value := range ip {
		// Convert key to string
		strKey := fmt.Sprintf("%v", key)

		switch value.(type) {
		case map[interface{}]interface{}:
			// Recursively convert nested maps
			newVal, err := stringMap(value)
			if err != nil {
				return outputMap, fmt.Errorf("unable to create string key map as value for key %v beacuse error: %v", strKey, err)
			}
			outputMap[strKey] = newVal
		default:
			// For other types, directly assign the value
			outputMap[strKey] = value
		}
	}

	return outputMap, nil
}

func setValueByDotNotation(data map[string]interface{}, notation string, value interface{}) (map[string]interface{}, error) {
	current, err := stringMap(data)
	lastData := current
	if err != nil {
		return lastData, err
	}
	keys := strings.Split(notation, ".")

	for {
		key := keys[0]
		keys = keys[1:]

		lastKey := len(keys) == 0
		if lastKey {
			current[key] = value
			break
		} else {
			m, ok := current[key]
			if ok {
				mm, ok := m.(map[string]interface{})
				if ok {
					current = mm
				} else {
					v := map[string]interface{}{}
					current[key] = v
					current = v
				}
			} else {
				v := map[string]interface{}{}
				current[key] = v
				current = v
			}
		}
	}
	return lastData, nil
}
