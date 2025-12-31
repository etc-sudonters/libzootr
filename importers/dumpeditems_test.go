package importers

import (
	"strings"
	"testing"
)

const sampleDumpedItems string = `
[
  {
    "name": "Bombs (5)",
    "advancement": false,
    "priority": false,
    "type": "Item"
  },
  {
    "name": "Deku Nuts (5)",
    "advancement": false,
    "priority": false,
    "type": "Item"
  },
  {
    "name": "Bombchus (10)",
    "advancement": true,
    "priority": false,
    "type": "Item"
  }
]
`

func TestImportsDumpedItems(t *testing.T) {
	for item, err := range DumpItems.ImportFrom(
		t.Context(),
		strings.NewReader(sampleDumpedItems),
	) {
		if err != nil {
			t.Fatalf("Failed while importing %#v", item)
		}
	}

}
