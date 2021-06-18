package haproxy

import (
	"bufio"
	"io"
	"os"
	"strings"
)

const (
	commentChar     = "#"
	globalKeyword   = "global"
	defaultsKeyword = "defaults"
	frontendKeyword = "frontend"
	backendKeyword  = "backend"
)

// ConfigParser is a primitive HAProxy config parser.
// It detects 4 main sections: global, defaults, frontends and backends
type ConfigParser struct {
	// ConfigPath is the full path to HAProxy config file
	ConfigPath string
	// GlobalSection holds the contents of the global section
	GlobalSection []string
	// DefaultsSection holds the contents of the defaults section
	DefaultsSection []string
	// FrontendSection holds the content of the frontend section
	FrontendSection map[string][]string
	// BackendSection holds the contents of the backend section
	BackendSection map[string][]string
}

// NewConfigParser returns an instance of ConfigParser
func NewConfigParser(configPath string) *ConfigParser {
	return &ConfigParser{
		ConfigPath:      configPath,
		GlobalSection:   []string{},
		DefaultsSection: []string{},
		FrontendSection: map[string][]string{},
		BackendSection:  map[string][]string{},
	}
}

// Parse parses the HAProxy config file detecting the 4 main sections
func (p *ConfigParser) Parse() error {
	f, err := os.Open(p.ConfigPath)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	currSection, currSubsection := "", ""

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, commentChar) || line == "" {
			continue
		}

		if strings.HasPrefix(line, globalKeyword) {
			currSection = globalKeyword
			continue
		}
		if strings.HasPrefix(line, defaultsKeyword) {
			currSection = defaultsKeyword
			continue
		}
		if strings.HasPrefix(line, frontendKeyword) {
			currSection = frontendKeyword
			words := strings.Fields(line)
			if len(words) > 1 {
				currSubsection = words[1]
				if _, exists := p.FrontendSection[currSubsection]; !exists {
					p.FrontendSection[currSubsection] = []string{}
				}
			}
			continue
		}
		if strings.HasPrefix(line, backendKeyword) {
			currSection = backendKeyword
			words := strings.Fields(line)
			if len(words) > 1 {
				currSubsection = words[1]
				if _, exists := p.BackendSection[currSubsection]; !exists {
					p.BackendSection[currSubsection] = []string{}
				}
			}
			continue
		}

		switch currSection {
		case globalKeyword:
			p.GlobalSection = append(p.GlobalSection, line)
		case defaultsKeyword:
			p.DefaultsSection = append(p.DefaultsSection, line)
		case frontendKeyword:
			p.FrontendSection[currSubsection] = append(p.FrontendSection[currSubsection], line)
		case backendKeyword:
			p.BackendSection[currSubsection] = append(p.BackendSection[currSubsection], line)
		}
	}
	return nil
}

// Frontends returns the contents of all the frontends whose names match the given name substring
func (p *ConfigParser) Frontends(nameSubstr string) map[string][]string {
	return p.findInMap(p.FrontendSection, nameSubstr, false)
}

// Frontend returns the contents of the frontend with the given name, false is returned as second value if the frontend doesn't exist
func (p *ConfigParser) Frontend(name string) ([]string, bool) {
	contents, exist := p.findInMap(p.FrontendSection, name, true)[name]
	return contents, exist
}

// Backends returns the contents of all the backends whose names match the given name substring
func (p *ConfigParser) Backends(nameSubstr string) map[string][]string {
	return p.findInMap(p.BackendSection, nameSubstr, false)
}

// returns the contents of the backend with the given name, false is returned as second value if the backend doesn't exist
func (p *ConfigParser) Backend(name string) ([]string, bool) {
	contents, exist := p.findInMap(p.BackendSection, name, true)[name]
	return contents, exist
}

func (p *ConfigParser) findInMap(m map[string][]string, s string, strict bool) map[string][]string {
	res := map[string][]string{}
	for name, contents := range m {
		if strict {
			if name == s {
				res[name] = contents
				break
			}
		} else {
			if strings.Contains(name, s) {
				res[name] = contents
			}
		}
	}
	return res
}
