
source $(git rev-parse --show-toplevel)/lib.sh

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $SCRIPT_DIR

bx kubectl apply -f manifest.yaml
x kubectl wait --for=condition=Ready ksvc --all --timeout 5m
bx kubectl get ksvc
open -na "Google Chrome" --args --new-window https://intake.default.margarita.dev 
open -na "Google Chrome" --args --new-window https://sockeye.default.margarita.dev
