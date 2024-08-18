compctl -K _tinyenv tinyenv

_tinyenv() {
  local words completions
  local lang cmd
  read -cA words

  if [[ ${#words} -eq 2 ]]; then
    completions="$(printf "go\njava\nnode\nperl\npython\nruby\nrust")"
  elif [[ ${#words} -eq 3 ]]; then
    completions="$(printf "global\ninit\nreahsh\nversion\nversions")"
  elif [[ ${#words} -eq 4 ]]; then
    lang=$words[2]
    cmd=$words[3]
    if [[ $cmd = global ]]; then
      completions="$(tinyenv $lang versions)"
    fi
  fi
  reply=("${(ps:\n:)completions}")
}
