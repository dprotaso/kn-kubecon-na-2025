#!/usr/bin/env bash

source $(git rev-parse --show-toplevel)/lib.sh

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $SCRIPT_DIR

# Cleanup old demo
rm -rf square

bx func create -l go square
bx tree -C square
bx bat square/handle.go

cat <<EOF | patch square/handle.go
--- square/handle.go	2025-11-05 23:35:13
+++ square/handle.new	2025-11-05 23:37:40
@@ -1,26 +1,32 @@
 package function
 
 import (
-	"fmt"
+	"encoding/json"
 	"net/http"
-	"net/http/httputil"
+	"strconv"
 )
 
+type Input struct {
+	Value string \`json:"value"\`
+}
+
+type Output struct {
+	Value string \`json:"value"\`
+}
+
 // Handle an HTTP Request.
+// nolint
 func Handle(w http.ResponseWriter, r *http.Request) {
-	/*
-	 * YOUR CODE HERE
-	 *
-	 * Try running \`go test\`.  Add more test as you code in \`handle_test.go\`.
-	 */
+	var (
+		in  *Input
+		out Output
 
-	dump, err := httputil.DumpRequest(r, true)
-	if err != nil {
-		http.Error(w, err.Error(), http.StatusInternalServerError)
-		return
-	}
+		dec = json.NewDecoder(r.Body)
+		enc = json.NewEncoder(w)
+	)
 
-	fmt.Println("Received request")
-	fmt.Printf("%q\n", dump)
-	fmt.Fprintf(w, "%q", dump)
+	dec.Decode(&in)
+	val, _ := strconv.Atoi(in.Value)
+	out.Value = strconv.Itoa(val*val)
+	enc.Encode(out)
 }
+
EOF

bx bat square/handle.go

pushd square >/dev/null
  bx func build --registry dprotaso
  x func run &
  read
  bx func invoke --data '{"value":"9"}'
popd >/dev/null

kill %1
wait
