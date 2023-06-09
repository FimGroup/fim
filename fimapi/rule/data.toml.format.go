package rule

import (
	"strings"

	"github.com/FimGroup/fim/fimapi/pluginapi"
)

func ValidateFullPath(in string) bool {
	splits := SplitFullPath(in)
	lastLevel := false
	for _, v := range splits {
		if lastLevel {
			return false
		}
		typeOfNode, ok := checkElementKey(v, false)
		if !ok {
			return false
		}
		if typeOfNode == pluginapi.TypeUnknown {
			return false
		}
		switch typeOfNode {
		case pluginapi.TypeAttributeNode:
			fallthrough
		case pluginapi.TypeNsNode:
			lastLevel = true
		}
	}
	return true
}

func ValidateFullPathOfDefinition(in string) bool {
	splits := SplitFullPath(in)
	lastLevel := false
	for _, v := range splits {
		if lastLevel {
			return false
		}
		typeOfNode, ok := checkElementKey(v, true)
		if !ok {
			return false
		}
		if typeOfNode == pluginapi.TypeUnknown {
			return false
		}
		switch typeOfNode {
		case pluginapi.TypeAttributeNode:
			fallthrough
		case pluginapi.TypeNsNode:
			lastLevel = true
		}
	}
	return true
}

func SplitFullPath(in string) []string {
	return strings.Split(in, pluginapi.PathSeparator)
}

func IsPathArray(in string) bool {
	_, idx := ExtractArrayPath(in)
	return idx >= 0
}

func IsArrayDefinition(in string) bool {
	s, _ := ExtractArrayPath(in)
	return len(s)+2 == len(in) // length of '[]' is 2
}

func IsArrayAccess(in string) bool {
	return (!IsArrayDefinition(in)) && IsPathArray(in)
}

func ConcatFullPath(paths []string) string {
	return strings.Join(paths, pluginapi.PathSeparator)
}

func ExtractArrayPath(in string) (string, int) {
	// assume the path is valid

	bracketStartIdx := -1
	for idx := 0; idx < len(in); idx++ {
		ch := in[idx]
		if ch == '[' {
			bracketStartIdx = idx
			break
		}
	}

	if bracketStartIdx != -1 {
		cnt := 0
		bracketEndIdx := len(in) - 1
		for i := bracketStartIdx + 1; i < bracketEndIdx; i++ {
			cnt = cnt*10 + int(in[i]-'0')
		}
		return in[:bracketStartIdx], cnt
	} else {
		return in, -1
	}
}

func checkElementKey(key string, allowedEmptyArrayIndex bool) (pluginapi.TypeOfNode, bool) {
	if len(key) <= 0 {
		return pluginapi.TypeUnknown, false
	}
	first := key[0]
	var nameKey string
	var nodeType pluginapi.TypeOfNode
	switch first {
	case '#':
		nameKey = key[1:]
		nodeType = pluginapi.TypeAttributeNode
	case '@':
		nameKey = key[1:]
		nodeType = pluginapi.TypeNsNode
	default:
		nameKey = key
		nodeType = pluginapi.TypeDataNode
	}
	if len(nameKey) <= 0 {
		return pluginapi.TypeUnknown, false
	}

	return nodeType, checkElement(nameKey, allowedEmptyArrayIndex)
}

func checkElement(str string, allowedEmptyArrayIndex bool) bool {
	// compatible to available characters from xml
	// e.g.letters, digits, hyphens, underscores, and periods
	// note: colons are reserved

	// starting character
	first := str[0]
	if checkRange(first, 'A', 'Z') ||
		checkRange(first, 'a', 'z') ||
		first == '_' {
		// allowed
	} else {
		return false
	}

	// allowed character
	bracketStartIdx := -1
	for idx := 0; idx < len(str); idx++ {
		ch := str[idx]
		if checkRange(ch, 'A', 'Z') ||
			checkRange(ch, 'a', 'z') ||
			checkRange(ch, '0', '9') ||
			ch == '-' || ch == '_' || ch == '.' {
			continue
		} else if ch == '[' {
			bracketStartIdx = idx
			break
		}
		return false
	}

	// check array
	if bracketStartIdx != -1 {
		bracketEndIdx := len(str) - 1
		if str[bracketStartIdx] != '[' || str[bracketEndIdx] != ']' {
			return false
		}
		cnt := 0
		for i := bracketStartIdx + 1; i < bracketEndIdx; i++ {
			if !checkRange(str[i], '0', '9') {
				return false
			} else {
				cnt++
			}
		}
		if cnt == 0 && !allowedEmptyArrayIndex {
			return false
		}
		//FIXME check index format and overflow
	}

	return true
}

func checkRange(v uint8, lBound, hBound uint8) bool {
	return v >= lBound && v <= hBound
}
