package trips

import (
	"encoding/csv"
	"io"
	"strings"
)

func openCSVReader(r io.Reader) (map[string]int, *csv.Reader, error) {
	reader := csv.NewReader(r)
	header, err := reader.Read()
	if err != nil {
		return nil, nil, err
	}
	return indexByName(header), reader, nil
}

func openCSVReaderLegacy(file io.Reader, err error) (map[string]int, *csv.Reader) {
	if err != nil {
		return nil, nil
	}
	idx, reader, e := openCSVReader(file)
	if e != nil {
		return nil, nil
	}
	return idx, reader
}

func indexByName(header []string) map[string]int {
	idx := make(map[string]int, len(header))
	for i, name := range header {
		// Strip UTF-8 BOM if present on first column.
		if len(name) > 0 && name[0] == '\xef' {
			name = strings.TrimPrefix(name, "\ufeff")
		}
		idx[name] = i
	}
	return idx
}
