source $(git rev-parse --show-toplevel)/lib.sh

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $SCRIPT_DIR

kubectl delete ksvc --all > /dev/null

bx echo "show mcp code"

bx bat render/mcp.yaml --highlight-line 7:9
bx kubectl apply -f render/mcp.yaml
bx kubectl get ksvc

bx echo "show mcp inspector"

bx bat render/add.yaml
bx kubectl apply -f render/add.yaml
bx kubectl get ksvc

bx echo "show add tool"

bx bat render/multiply.yaml
bx kubectl apply -f render/multiply.yaml
bx kubectl get ksvc

bx echo "show multiply tool"

kubectl delete ksvc --all > /dev/null
