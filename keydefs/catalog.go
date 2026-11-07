package keydefs

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed testdata/bindables.yaml
var catalogData []byte

type CatalogEntry struct {
	Action   ActionCode
	Label    string
	KeyLabel string
	Category KeyCategory
}

var builtInCatalog = loadCatalog()

func loadCatalog() []CatalogEntry {
	var root yaml.Node
	if err := yaml.Unmarshal(catalogData, &root); err != nil {
		panic(fmt.Errorf("invalid embedded action catalog: %w", err))
	}
	var entries []CatalogEntry
	top := root.Content[0]
	for i := 0; i < len(top.Content); i += 2 {
		category := KeyCategory(top.Content[i].Value)
		items := top.Content[i+1].Content
		for j := 0; j < len(items); j += 2 {
			action, err := ParseActionCode(items[j].Value)
			if err != nil {
				panic(fmt.Errorf("invalid embedded action: %w", err))
			}
			var entry KeyEntry
			if err := items[j+1].Decode(&entry); err != nil {
				panic(err)
			}
			entries = append(entries, CatalogEntry{Action: action, Label: entry.ActionLabel, KeyLabel: entry.KeyLabel, Category: category})
		}
	}
	return entries
}

// Catalog returns a copy in the built-in catalog's display order.
// It describes known firmware actions, not a model-specific physical layout.
func Catalog() []CatalogEntry { return append([]CatalogEntry(nil), builtInCatalog...) }

func SupportedAction(action ActionCode) bool {
	for _, entry := range builtInCatalog {
		if entry.Action == action && entry.Category != CategoryKeyOnly {
			return true
		}
	}
	if action.IsKeyboardUsage() && action[1] >= 0xe0 && action[1] <= 0xe7 && action[2] != 0 {
		return SupportedAction(ActionCode{7, 0, action[2]})
	}
	return false
}
