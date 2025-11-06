bx() {
  local key

  while true; do
    x "$@"
    read -n1 -s -r key
    [[ $key == r ]] && continue
    [[ $key == q ]] && exit
    break
  done
}

x() {
  echo ">> $@"
  $@
}
