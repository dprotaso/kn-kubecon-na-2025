
source $(git rev-parse --show-toplevel)/lib.sh

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $SCRIPT_DIR

# bx kubectl apply -f manifest.yaml
# bx bat data.json
# bx curl -k https://svc-intake.keventmesh.margarita.dev -H 'Content-Type: application/json" -d @data.json
