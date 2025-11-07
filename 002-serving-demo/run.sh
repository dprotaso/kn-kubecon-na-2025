
source $(git rev-parse --show-toplevel)/lib.sh

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $SCRIPT_DIR

kubectl delete -f hello-1.yaml --ignore-not-found > /dev/null 

bx bat hello-1.yaml
bx kubectl apply -f hello-1.yaml
bx kubectl get ksvc

bx curl -k https://hello.default.margarita.dev
bx curl -v -k https://hello.default.margarita.dev

bx bat hello-2.yaml --highlight-line 10:12
bx kubectl apply -f hello-2.yaml

bx kubectl get ksvc
bx kubectl get rev

bx curl -k https://hello.default.margarita.dev

bx bat hello-3.yaml --highlight-line 13:17
bx kubectl apply -f hello-3.yaml

bx kubectl get ksvc
bx kubectl get rev

bx curl -k https://hello.default.margarita.dev
bx curl -k https://hello.default.margarita.dev

bx bat hello-4.yaml --highlight-line 8
bx kubectl apply -f hello-4.yaml

bx kubectl get ksvc
bx kubectl get rev

bx hey -z 30s -c 10 https://hello.default.margarita.dev

echo "Diff hello-00001 and hello-00002"
tmp1=$(mktemp)
tmp2=$(mktemp)
tmp3=$(mktemp)

kubectl get rev hello-00001 -o jsonpath='{.spec}' | yq -oyaml -P > $tmp1
kubectl get rev hello-00002 -o jsonpath='{.spec}' | yq -oyaml -P > $tmp2
kubectl get rev hello-00003 -o jsonpath='{.spec}' | yq -oyaml -P > $tmp3

diff --color $tmp1 $tmp2
read -n1
echo "Diff hello-00002 and hello-00003"
diff --color $tmp2 $tmp3
