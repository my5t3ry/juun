ROOT=$(dirname $0:A)

juunfzf() {
  local selected num
  setopt localoptions noglobsubst noposixbuiltins pipefail 2>/dev/null
  selected=$($ROOT/juun.fzf $$ $BUFFER | fzf --height ${FZF_TMUX_HEIGHT:-40%} -n2..,.. --bind "ctrl-r:reload($ROOT/juun.fzf $$ {q})" --tiebreak=index --query="$BUFFER" +m)
  local ret=$?
  if [ -n "$selected" ]; then
    BUFFER=$selected
  fi
  zle reset-prompt
  return $ret
}

zle -N juunfzf
bindkey '^R' juunfzf
