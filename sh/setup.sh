juun_restart() {
  juun_stop

  sleep 1

  juun_start

  sleep 2

  tail -10 ~/.juun.log
}

juun_stop() {
  if [ -f ~/.juun.pid ]; then
    kill $(cat ~/.juun.pid) 2>/dev/null
  fi

  if [ -f ~/.juun.vw.pid ]; then
    kill $(cat ~/.juun.vw.pid) 2>/dev/null
  fi
}

juun_start() {
  $ROOT/juun.service || echo 'unable to start juun.service'
}

juun_work() {
  $ROOT/juun.updown $1 $$ "$2"
}

juun_cleanup() {
  juun_work delete $$
}

if [[ -n "$BASH" ]]; then
  _realpath() {
    [[ $1 == /* ]] && echo "$1" || echo "$PWD/${1#./}"
  }
  ROOT=$(_realpath $(dirname $BASH_SOURCE))

  if [ ${BASH_VERSINFO[0]} -lt 4 ]; then
    echo "Sorry, you need at least bash-4.0 to use juun."
  else
    source $ROOT/preexec.sh

    juun_preexec() {
      juun_work add "$1"
    }

    juun_precmd() {
      juun_work end end
    }

    preexec_functions+=(juun_preexec)
    precmd_functions+=(juun_precmd)

    trap 'juun_cleanup' EXIT

    juun_search_start() {
      $ROOT/juun.search $$ 2>/tmp/juun.search.$$
      rc=$?
      res=$(cat /tmp/juun.search.$$)
      rm /tmp/juun.search.$$
      if [ $rc -eq 0 ]; then
        echo $res
        # FIXME: add it to the normal history?
        READLINE_LINE=""
        READLINE_POINT=""

        eval "$res"
        juun_work "add" "$res"
      else
        READLINE_LINE="$res"
        READLINE_POINT=${#READLINE_LINE}
      fi
    }

    juun_down() {
      res=$(juun_work down "$READLINE_LINE")
      READLINE_LINE="$res"
      READLINE_POINT="${#READLINE_LINE}"
    }

    juun_up() {
      res=$(juun_work up "$READLINE_LINE")
      READLINE_LINE="$res"
      READLINE_POINT="${#READLINE_LINE}"
    }

    if [ "x$JUUN_DONT_BIND_BASH" = "x" ]; then
      if [ "x$BASH_UPDOWN_BROKEN" = "x" ]; then
        bind -x '"\e[A": juun_up'
        bind -x '"\e[B": juun_down'
      fi
      bind -x '"\C-p": juun_up'
      bind -x '"\C-n": juun_down'
      bind -x '"\C-r": "juun_search_start"'
    fi

    juun_start
  fi
elif [[ -n "$ZSH_VERSION" ]]; then
  ROOT=$(dirname $0:A)

  trap 'juun_cleanup' EXIT

  zshaddhistory() {
    juun_work add "$1"
    juun_work reindex reindex
  }

  juun_precmd() {
    juun_work end end
  }

  juun_search_start() {
    zle -I
    $ROOT/juun.search </dev/tty $$ 2>/tmp/juun.search.$$
    rc=$?
    res=$(cat /tmp/juun.search.$$)
    rm /tmp/juun.search.$$

    if [ $rc -eq 0 ]; then
      BUFFER="$res"
      CURSOR=${#BUFFER}
      juun_work "add" "$res"
      zle get-line
    else
      BUFFER="$res"
      CURSOR=${#BUFFER}
    fi
  }

  juun_down() {
    BUFFER=$(juun_work down down)
    CURSOR=${#BUFFER}
  }

  juun_up() {
    BUFFER=$(juun_work up $BUFFER)
    CURSOR=${#BUFFER}
  }

  function _juun_completions() {
    local -a mywords
    mywords=$($ROOT/juun.fzf $$ $BUFFER | tr '\n' ' ')
    compadd -a mywords
  }

  compdef _juun_completions -first-

  zle -N juun_up
  zle -N juun_down
  zle -N juun_search_start

  bindkey -r "^[[A"
  bindkey -r "^[OA"
  bindkey "^[[A" juun_up
  bindkey "^[OA" juun_up
  bindkey -r "^[OB"
  bindkey -r "^[[B"
  bindkey "^[OB" juun_down
  bindkey "^[[B" juun_down
  bindkey "^p" juun_up
  bindkey "^n" juun_down

  if [ -x "$(command -v fzf)" ]; then
    source $ROOT/juun.fzf.sh
  else
    bindkey "^R" juun_search_start
  fi

  juun_start
else
  echo "Sorry, you need bash-4.+ or zsh to use juun."
fi
