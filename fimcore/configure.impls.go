package fimcore

import (
	"os"
	"strings"

	"github.com/FimGroup/fim/fimapi/basicapi"
)

const (
	ConfigurePrefixStatic  = "configure-static://"
	ConfigurePrefixDynamic = "configure-dynamic://"
)

type NestedConfigureManager struct {
	configureManagers []basicapi.ConfigureManager
}

func (n *NestedConfigureManager) ReplaceStaticConfigure(placeholder string) string {
	var r = placeholder
	for _, v := range n.configureManagers {
		r = v.ReplaceStaticConfigure(r)
	}
	return r
}

func (n *NestedConfigureManager) ReplaceDynamicConfigure(placeholder string) string {
	var r = placeholder
	for _, v := range n.configureManagers {
		r = v.ReplaceDynamicConfigure(r)
	}
	return r
}

func (n *NestedConfigureManager) SupportDynamicConfigure(placeholder string) bool {
	for _, v := range n.configureManagers {
		if v.SupportDynamicConfigure(placeholder) {
			return true
		}
	}
	return false
}

func (n *NestedConfigureManager) addSubConfigureManager(configureManager basicapi.ConfigureManager) {
	n.configureManagers = append(n.configureManagers, configureManager)
}

func NewNestedConfigureManager() *NestedConfigureManager {
	return &NestedConfigureManager{
		configureManagers: []basicapi.ConfigureManager{},
	}
}

type SettableConfigureManager struct {
	configures map[string]string
}

func (s *SettableConfigureManager) ReplaceStaticConfigure(placeholder string) string {
	if strings.HasPrefix(placeholder, ConfigurePrefixStatic) {
		key := placeholder[len(ConfigurePrefixStatic):]
		val, ok := s.configures[key]
		if ok {
			return val
		} else {
			return placeholder
		}
	}
	return placeholder
}

func (s *SettableConfigureManager) ReplaceDynamicConfigure(placeholder string) string {
	// do not support dynamic configure
	return placeholder
}

func (s *SettableConfigureManager) SupportDynamicConfigure(placeholder string) bool {
	// do not support dynamic configure
	return false
}

func (s *SettableConfigureManager) SetConfigure(key, value string) (string, bool) {
	val, ok := s.configures[key]
	s.configures[key] = value
	return val, ok
}

func NewSettableConfigureManager() *SettableConfigureManager {
	return &SettableConfigureManager{map[string]string{}}
}

type EnvConfigureManager struct {
}

func (e *EnvConfigureManager) ReplaceStaticConfigure(placeholder string) string {
	if strings.HasPrefix(placeholder, ConfigurePrefixStatic) {
		key := placeholder[len(ConfigurePrefixStatic):]
		val, ok := os.LookupEnv(key)
		if ok {
			return val
		} else {
			return placeholder
		}
	}
	return placeholder
}

func (e *EnvConfigureManager) ReplaceDynamicConfigure(placeholder string) string {
	return placeholder
}

func (e *EnvConfigureManager) SupportDynamicConfigure(placeholder string) bool {
	// do not support dynamic configure
	return false
}

func NewEnvConfigureManager() *EnvConfigureManager {
	return &EnvConfigureManager{}
}
