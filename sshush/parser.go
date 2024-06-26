package sshush

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/k0kubun/pp/v3"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

type (
	Parser struct {
		GlobalConfig      map[string]any
		DefaultConfig     map[string]any
		Extensions        map[string]ExtendsConfig
		UnprocessedConfig *orderedmap.OrderedMap[string, any]
		Verbose           bool
		Debug             bool
	}

	ExtendsConfig struct {
		Identifier string
		Config     map[string]any
		Extends    string
	}
)

// Load loads the configuration from the sources.
// It processes the global and default config blocks.
// It also resolves the Extends declarations.
// It does not process the config itself.
func (p *Parser) Load(sources *sshConfigSources) error {
	// m is initialised outside the source loop such that it's appended to.
	m := orderedmap.New[string, any]()

	for _, source := range *sources {
		contents, err := os.ReadFile(source)
		if err != nil {
			return fmt.Errorf("reading file: %w", err)
		}

		err = yaml.Unmarshal(contents, &m)
		if err != nil {
			return fmt.Errorf("unmarshalling yaml: %w", err)
		}

		// if global config exists in this source, set it and remove it.
		p.extractAndSetConfig(m, &p.GlobalConfig, "global")

		// if default config exists in this source, set it and remove it.
		p.extractAndSetConfig(m, &p.DefaultConfig, "default")
	}

	p.UnprocessedConfig = m

	// process Extends declarations.
	p.extractExtensions()

	return nil
}

// extractAndSetConfig extracts a block from the config map and sets it to the
// configProperty. It then removes the block from the config map.
func (p *Parser) extractAndSetConfig(
	m *orderedmap.OrderedMap[string, any],
	configProperty *map[string]any,
	blockName string,
) {
	if config := extractBlock(blockName, m); config != nil {
		*configProperty = config
		m.Delete(blockName)
	}
}

// extractExtensions parses all the config and resolves inherited Extends declarations.
// @see https://sshush.bencromwell.com/docs/configuration/extends/
func (p *Parser) extractExtensions() {
	extensions := make(map[string]ExtendsConfig)
	for pair := p.UnprocessedConfig.Oldest(); pair != nil; pair = pair.Next() {
		configMap, ok := pair.Value.(map[string]any)
		if !ok {
			continue
		}

		config, ok := configMap["Config"].(map[string]any)
		if !ok {
			continue
		}

		extends, ok := configMap["Extends"].(string)
		if !ok {
			extends = ""
		}

		extensions[pair.Key] = ExtendsConfig{
			Identifier: pair.Key,
			Config:     config,
			Extends:    extends,
		}
	}

	for _, extension := range extensions {
		if ext, ok := extensions[extension.Extends]; ok {
			// Update the config to the merged config.
			ext.Config = mergeMaps(extension.Config, extensions[extension.Extends].Config)
			extensions[extension.Identifier] = ext
		}
	}

	p.Extensions = extensions
}

// ProduceConfig produces the SSH configuration.
func (p *Parser) ProduceConfig() ([]string, error) {
	var output []string
	var err error

	for pair := p.UnprocessedConfig.Oldest(); pair != nil; pair = pair.Next() {
		output, err = p.processConfigGroup(pair, output)
		if err != nil {
			return nil, err
		}
	}

	// Add the global config.
	if p.GlobalConfig != nil && len(p.GlobalConfig) > 0 {
		output = append(output, "# Global config")
		output = append(output, "Host *")
		for key, value := range p.GlobalConfig {
			output = append(output, fmt.Sprintf("    %s %s", key, value))
		}
	}

	return output, err
}

// processConfigGroup processes a config group.
// @see https://sshush.bencromwell.com/docs/configuration/groups/
func (p *Parser) processConfigGroup(
	pair *orderedmap.Pair[string, any],
	output []string,
) ([]string, error) {
	identifier := pair.Key
	config := pair.Value

	if p.Debug {
		pp.Println("Identifier: ", identifier)
		pp.Println("Config: ", config)
	}

	output = append(output, fmt.Sprintf("# %s", identifier))

	if configMap, ok := config.(map[string]any); ok {
		var err error
		output, err = p.processConfigMap(configMap, output)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("config for %s is not a map", identifier)
	}

	return output, nil
}

// processConfigMap processes the internal representation of the config structure.
func (p *Parser) processConfigMap(
	configMap map[string]any,
	output []string,
) ([]string, error) {
	// This group may have a prefix declared.
	prefix, err := getPrefixFromConfigMap(configMap)
	if err != nil {
		return nil, err
	}

	groupConfig := p.getGroupConfig(configMap)

	if p.Debug {
		pp.Println("Group config: ", groupConfig)
	}

	if hosts, ok := configMap["Hosts"]; ok {
		// If it's a direct list of hosts, rearrange things.
		err = expandListToMapOfHosts(configMap, &hosts)
		if err != nil {
			return nil, err
		}

		keys := sortMapByKeys(hosts.(map[string]any))

		// Process hosts in the sorted order of their keys.
		for _, host := range keys {
			hostConfig := getHostConfig(hosts.(map[string]any)[host], groupConfig)

			if p.Debug {
				pp.Println("Host config: ", hostConfig)
			}

			output = append(output, fmt.Sprintf("Host %s%s", prefix, host))

			// Hoist HostName to the top.
			hostName := hostConfig["HostName"]
			delete(hostConfig, "HostName")
			if hostName != nil {
				output = appendLineToOutput(output, "HostName", hostName)
			}

			hostConfigKeys := sortMapByKeys(hostConfig)

			// Process the host config in the sorted order of its keys.
			for _, k := range hostConfigKeys {
				v := hostConfig[k]
				output = appendConfigToOutput(output, k, v)
			}

			output = append(output, "")
		}
	}

	return output, nil
}

// getGroupConfig returns the config to apply to the entire group.
func (p *Parser) getGroupConfig(configMap map[string]any) map[string]any {
	// Start empty.
	groupConfig := make(map[string]any)

	// Set the defaults.
	groupConfig = mergeMaps(groupConfig, p.DefaultConfig)

	// If we are extending another config, add that in.
	groupConfig = mergeMaps(groupConfig, p.getExtendedConfig(configMap))

	// If we have config for this specific group, add that in.
	if config, ok := configMap["Config"]; ok {
		groupConfig = mergeMaps(groupConfig, config.(map[string]any))
	}

	return groupConfig
}

// getExtendedConfig returns the fully resolved config to extend from.
// A user can define "Extends" in their config to inherit from another config.
// @see https://sshush.bencromwell.com/docs/configuration/extends/
func (p *Parser) getExtendedConfig(configMap map[string]any) map[string]any {
	if extends, extendsExists := configMap["Extends"]; extendsExists {
		extendedConfig, hasTargetConfig := p.Extensions[extends.(string)]
		if hasTargetConfig {
			if p.Debug {
				pp.Printf("Extended config %s\n", extends.(string))
				pp.Println(extendedConfig)
			}

			return extendedConfig.Config
		}
	}

	return make(map[string]any)
}

// The config to apply to a host consists of any group level config combined
// with any specific config for this single host. The specific config for this
// host takes precedence.
func getHostConfig(hostConfig any, groupConfig map[string]any) map[string]any {
	var configForThisHost map[string]any

	// If the host config is a string, it's just a HostName.
	// In which case, the config to apply is that of the group.
	if hostConfigString, ok := hostConfig.(string); ok {
		configForThisHost = groupConfig
		// If the string contains * it's a wildcard so has no specific HostName.
		if !strings.Contains(hostConfigString, "*") {
			configForThisHost["HostName"] = hostConfigString
		}
	} else {
		// If we had a map, we need to merge the group config with the host config.
		configForThisHost = mergeMaps(groupConfig, hostConfig.(map[string]any))
	}

	return configForThisHost
}

// MergeMaps merges any number of maps and returns the result.
// If a key is present, it's overriden by the last map that contains it.
func mergeMaps(maps ...map[string]any) map[string]any {
	merged := make(map[string]any)

	// Iterate over each map.
	for _, m := range maps {
		// Copy values from the current map to the merged map.
		for key, value := range m {
			merged[key] = value
		}
	}

	return merged
}

// appendLineToOutput appends a key-value pair to the output.
// It formats the value as a string if it wasn't one.
// This essentially covers Port numbers, which we get through as ints.
func appendLineToOutput(output []string, key string, value any) []string {
	// convert value to string if it's not already.
	if _, ok := value.(string); !ok {
		value = fmt.Sprintf("%v", value)
	}

	output = append(output, fmt.Sprintf("    %s %s", key, value))

	return output
}

// extractBlock extracts a block from the config map, if it existed.
// It's allowed to not exist, so returns nil rather than an error if it doesn't.
func extractBlock(
	block string,
	m *orderedmap.OrderedMap[string, any],
) map[string]any {
	if blockValue, blockExists := m.Get(block); blockExists {
		if blockMap, blockIsMap := blockValue.(map[string]any); blockIsMap {
			return blockMap
		}
	}

	return nil
}

// sortMapByKeys extracts the keys from a map and returns them sorted.
func sortMapByKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

// expandListToMapOfHosts converts a host list to a host map.
// Users may specify:
// Hosts:
//   - host1.example.com
//   - host2.example.com
//
// or:
//
// Hosts:
//
//	  host1:
//			HostName: host1.example.com
//			User: bob
//	  host2:
//			HostName: host2.example.com
//			User: alice
//
// This function converts the first form to the second form.
func expandListToMapOfHosts(
	configMap map[string]any,
	hosts *any,
) error {
	// If it's a direct list of hosts, rearrange things.
	if listOfHosts, ok := configMap["Hosts"].([]any); ok {
		tmpHosts := make(map[string]any)

		for _, host := range listOfHosts {
			host, hostIsString := host.(string)
			if !hostIsString {
				return errors.New("\"Hosts\" must be a list of strings")
			}

			tmpHosts[host] = host
		}

		// Now we can continue as though it were a map.
		*hosts = tmpHosts
	}

	return nil
}

// getPrefixFromConfigMap returns the prefix from the config map.
// If there is no prefix, it returns an empty string.
// Prefix is optional.
// @see https://sshush.bencromwell.com/docs/configuration/prefix/
func getPrefixFromConfigMap(configMap map[string]any) (string, error) {
	var prefix string
	var err error

	// Set the prefix if we have one.
	if tP, ok := configMap["Prefix"]; ok {
		prefix, ok = tP.(string)
		if !ok {
			err = errors.New("\"Prefix\" must be a string")
		}
	}

	return prefix, err
}

// appendConfigToOutput appends a key-value pair to the output.
// If the value is a map, it will recursively append the key-value pairs.
func appendConfigToOutput(
	output []string,
	key string,
	value any,
) []string {
	switch v := value.(type) {
	case []any:
		for _, nestedValue := range v {
			output = appendConfigToOutput(output, key, nestedValue)
		}
	default:
		output = appendLineToOutput(output, key, v)
	}

	return output
}
