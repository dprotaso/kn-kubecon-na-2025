package function

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Input struct {
	Value string `json:"value"`
}

type Output struct {
	Value string `json:"value"`
}

// Handle an HTTP Request.
// nolint
func Handle(w http.ResponseWriter, r *http.Request) {
	var (
		in  *Input
		out Output

		dec = json.NewDecoder(r.Body)
		enc = json.NewEncoder(w)
	)

	dec.Decode(&in)
	val, _ := strconv.Atoi(in.Value)
	out.Value = strconv.Itoa(val*val)
	enc.Encode(out)
}
