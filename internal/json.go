package internal

import (
	"bytes"
	"encoding/json"
)

func prettify(b []byte) []byte {
	var out bytes.Buffer
	json.Indent(&out, b, "", "  ")
	return out.Bytes()
}
