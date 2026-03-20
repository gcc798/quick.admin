package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	var (
		outPath = flag.String("out", "", "output openapi yaml path")
		title   = flag.String("title", "", "override info.title")
	)
	flag.Parse()
	if strings.TrimSpace(*outPath) == "" {
		fail("-out is required")
	}
	inputs := flag.Args()
	if len(inputs) == 0 {
		fail("at least one input openapi yaml is required")
	}

	merged := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":   *title,
			"version": "0.0.1",
		},
		"paths":      map[string]any{},
		"components": map[string]any{},
	}

	var version string
	var anyPaths bool
	for _, input := range inputs {
		doc, err := readYAMLMap(input)
		if err != nil {
			fail(err.Error())
		}
		if v := nestedString(doc, "info", "version"); version == "" && v != "" {
			version = v
		}
		if t := nestedString(doc, "info", "title"); strings.TrimSpace(*title) == "" && nestedString(merged, "info", "title") == "" && t != "" {
			merged["info"].(map[string]any)["title"] = t
		}
		if paths := asMap(doc["paths"]); len(paths) > 0 {
			anyPaths = true
			mergeMap(asMap(merged["paths"]), paths)
		}
		if components := asMap(doc["components"]); len(components) > 0 {
			mergeMap(asMap(merged["components"]), components)
		}
		mergeTags(merged, doc)
	}

	if !anyPaths {
		return
	}
	if version != "" {
		merged["info"].(map[string]any)["version"] = version
	}
	if strings.TrimSpace(nestedString(merged, "info", "title")) == "" {
		merged["info"].(map[string]any)["title"] = inferTitle(*outPath)
	}

	content, err := yaml.Marshal(merged)
	if err != nil {
		fail(err.Error())
	}
	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(*outPath, content, 0o644); err != nil {
		fail(err.Error())
	}
}

func readYAMLMap(path string) (map[string]any, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func asMap(value any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	if result, ok := value.(map[string]any); ok {
		return result
	}
	return map[string]any{}
}

func mergeMap(dst, src map[string]any) {
	for key, value := range src {
		existing, ok := dst[key]
		if !ok {
			dst[key] = value
			continue
		}
		existingMap, existingIsMap := existing.(map[string]any)
		valueMap, valueIsMap := value.(map[string]any)
		if existingIsMap && valueIsMap {
			mergeMap(existingMap, valueMap)
			continue
		}
		existingSlice, existingIsSlice := existing.([]any)
		valueSlice, valueIsSlice := value.([]any)
		if existingIsSlice && valueIsSlice {
			dst[key] = mergeSlice(existingSlice, valueSlice)
			continue
		}
		if reflect.DeepEqual(existing, value) {
			continue
		}
		if isZeroValue(existing) && !isZeroValue(value) {
			dst[key] = value
		}
	}
}

func mergeSlice(dst, src []any) []any {
	result := append([]any{}, dst...)
	for _, item := range src {
		found := false
		for _, existing := range result {
			if reflect.DeepEqual(existing, item) {
				found = true
				break
			}
		}
		if !found {
			result = append(result, item)
		}
	}
	return result
}

func mergeTags(dst, src map[string]any) {
	dstTags, _ := dst["tags"].([]any)
	srcTags, _ := src["tags"].([]any)
	if len(srcTags) == 0 {
		return
	}
	merged := mergeSlice(dstTags, srcTags)
	sort.SliceStable(merged, func(i, j int) bool {
		left := tagName(merged[i])
		right := tagName(merged[j])
		return left < right
	})
	dst["tags"] = merged
}

func tagName(value any) string {
	if item, ok := value.(map[string]any); ok {
		if name, ok := item["name"].(string); ok {
			return name
		}
	}
	return fmt.Sprint(value)
}

func nestedString(doc map[string]any, keys ...string) string {
	current := doc
	for idx, key := range keys {
		value, ok := current[key]
		if !ok {
			return ""
		}
		if idx == len(keys)-1 {
			if str, ok := value.(string); ok {
				return str
			}
			return ""
		}
		next, ok := value.(map[string]any)
		if !ok {
			return ""
		}
		current = next
	}
	return ""
}

func inferTitle(outPath string) string {
	base := filepath.Base(filepath.Dir(outPath))
	if strings.EqualFold(base, "v1") {
		base = filepath.Base(filepath.Dir(filepath.Dir(outPath)))
	}
	if base == "" || base == "." || base == string(filepath.Separator) {
		return "OpenAPI"
	}
	return strings.ToUpper(base[:1]) + base[1:] + " API"
}

func isZeroValue(value any) bool {
	if value == nil {
		return true
	}
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	default:
		return false
	}
}

func fail(message string) {
	_, _ = fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
