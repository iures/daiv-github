package github

import (
	"os"
	"strings"
)

// ConfigProvider defines an interface for retrieving configuration values
type ConfigProvider interface {
	GetString(key string) (string, bool)
	GetBool(key string) (bool, bool)
	GetStringSlice(key string) ([]string, bool)
}

// MapConfigProvider implements ConfigProvider using a map
type MapConfigProvider struct {
	values map[string]any
}

// NewMapConfigProvider creates a new MapConfigProvider
func NewMapConfigProvider(values map[string]any) *MapConfigProvider {
	return &MapConfigProvider{
		values: values,
	}
}

// GetString retrieves a string value from the configuration
func (p *MapConfigProvider) GetString(key string) (string, bool) {
	if value, ok := p.values[key]; ok {
		if strValue, ok := value.(string); ok {
			return strValue, true
		}
	}
	return "", false
}

// GetBool retrieves a boolean value from the configuration
func (p *MapConfigProvider) GetBool(key string) (bool, bool) {
	if value, ok := p.values[key]; ok {
		if boolValue, ok := value.(bool); ok {
			return boolValue, true
		}
		if strValue, ok := value.(string); ok {
			return strValue == "true", true
		}
	}
	return false, false
}

// GetStringSlice retrieves a string slice from the configuration
func (p *MapConfigProvider) GetStringSlice(key string) ([]string, bool) {
	if value, ok := p.values[key]; ok {
		if strValue, ok := value.(string); ok {
			parts := strings.Split(strValue, ",")
			for i, part := range parts {
				parts[i] = strings.TrimSpace(part)
			}
			return parts, true
		}
		if sliceValue, ok := value.([]string); ok {
			return sliceValue, true
		}
	}
	return nil, false
}

// EnvConfigProvider implements ConfigProvider using environment variables
type EnvConfigProvider struct {
	prefix string
}

// NewEnvConfigProvider creates a new EnvConfigProvider
func NewEnvConfigProvider(prefix string) *EnvConfigProvider {
	return &EnvConfigProvider{
		prefix: prefix,
	}
}

// GetString retrieves a string value from environment variables
func (p *EnvConfigProvider) GetString(key string) (string, bool) {
	envKey := p.formatEnvKey(key)
	value, exists := os.LookupEnv(envKey)
	return value, exists
}

// GetBool retrieves a boolean value from environment variables
func (p *EnvConfigProvider) GetBool(key string) (bool, bool) {
	envKey := p.formatEnvKey(key)
	value, exists := os.LookupEnv(envKey)
	if !exists {
		return false, false
	}
	return value == "true" || value == "1" || value == "yes", true
}

// GetStringSlice retrieves a string slice from environment variables
func (p *EnvConfigProvider) GetStringSlice(key string) ([]string, bool) {
	envKey := p.formatEnvKey(key)
	value, exists := os.LookupEnv(envKey)
	if !exists {
		return nil, false
	}
	parts := strings.Split(value, ",")
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}
	return parts, true
}

// formatEnvKey formats a configuration key as an environment variable
func (p *EnvConfigProvider) formatEnvKey(key string) string {
	result := strings.ToUpper(key)
	result = strings.ReplaceAll(result, ".", "_")
	if p.prefix != "" {
		result = p.prefix + "_" + result
	}
	return result
}

// CompositeConfigProvider implements ConfigProvider using multiple providers
type CompositeConfigProvider struct {
	providers []ConfigProvider
}

// NewCompositeConfigProvider creates a new CompositeConfigProvider
func NewCompositeConfigProvider(providers ...ConfigProvider) *CompositeConfigProvider {
	return &CompositeConfigProvider{
		providers: providers,
	}
}

// GetString retrieves a string value from the first provider that has it
func (p *CompositeConfigProvider) GetString(key string) (string, bool) {
	for _, provider := range p.providers {
		if value, exists := provider.GetString(key); exists {
			return value, true
		}
	}
	return "", false
}

// GetBool retrieves a boolean value from the first provider that has it
func (p *CompositeConfigProvider) GetBool(key string) (bool, bool) {
	for _, provider := range p.providers {
		if value, exists := provider.GetBool(key); exists {
			return value, true
		}
	}
	return false, false
}

// GetStringSlice retrieves a string slice from the first provider that has it
func (p *CompositeConfigProvider) GetStringSlice(key string) ([]string, bool) {
	for _, provider := range p.providers {
		if value, exists := provider.GetStringSlice(key); exists {
			return value, true
		}
	}
	return nil, false
}

// ConfigLoader loads configuration from a ConfigProvider
type ConfigLoader struct {
	provider ConfigProvider
}

// NewConfigLoader creates a new ConfigLoader
func NewConfigLoader(provider ConfigProvider) *ConfigLoader {
	return &ConfigLoader{
		provider: provider,
	}
}

// LoadConfig loads a GitHubConfig from the provider
func (l *ConfigLoader) LoadConfig() (*GitHubConfig, error) {
	// Load required values
	username, ok := l.provider.GetString("github.username")
	if !ok || username == "" {
		return nil, NewValidationError("username is required", nil)
	}

	token, ok := l.provider.GetString("github.token")
	if !ok || token == "" {
		return nil, NewValidationError("token is required", nil)
	}

	organization, ok := l.provider.GetString("github.organization")
	if !ok || organization == "" {
		return nil, NewValidationError("organization is required", nil)
	}

	repositories, ok := l.provider.GetStringSlice("github.repositories")
	if !ok || len(repositories) == 0 {
		return nil, NewValidationError("repositories are required", nil)
	}

	// Load optional values with defaults
	format, ok := l.provider.GetString("github.format")
	if !ok || format == "" {
		format = "markdown"
	}

	return &GitHubConfig{
		Username:     username,
		Token:        token,
		Organization: organization,
		Repositories: repositories,
		Format:       format,
	}, nil
} 
