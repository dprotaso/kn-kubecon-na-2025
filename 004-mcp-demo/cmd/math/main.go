// nolint
package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

var op func(args []string) int

func Add(args []string) int {
	sum := 0

	for _, val := range args {
		intVal, _ := strconv.Atoi(val)
		sum += intVal
	}
	return sum
}

func Mul(args []string) int {
	sum := 1 // this is dumb but whatever

	for _, val := range args {
		intVal, _ := strconv.Atoi(val)
		sum *= intVal
	}
	return sum
}

func main() {
	opVal := os.Getenv("OPERATOR")
	switch opVal {
	case "mul":
		op = Mul
	default:
		op = Add
	}

	http.HandleFunc("/", Handle)
	http.ListenAndServe(":"+cmp.Or(os.Getenv("PORT"), "8080"), nil)
}

type CalcRequest struct {
	Args []string `json:"args,omitempty"` // numbers; sqrt will use first arg only
}

type CalcResponse struct {
	Value int `json:"value"` // numbers; sqrt will use first arg only
}

// Handle a request using your function instance.
func Handle(w http.ResponseWriter, r *http.Request) {
	var c *CalcRequest

	d := json.NewDecoder(r.Body)
	if err := d.Decode(&c); err != nil {
		fmt.Println("error decoding", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := CalcResponse{Value: op(c.Args)}
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		fmt.Println("error encoding", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
