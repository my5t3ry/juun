# juun(-fzf) - make your machine learn to be better tool, inshallah. 
### the best known approach known to $(whoami). gracefully & ctx aware build with some $BUFFER > bash > go > vowpal > go > bash > $BUFFER foo by [jackdoe](https://github.com/jackdoe/).
#### + smacked on fzf integration and compdef support


![logo](https://raw.githubusercontent.com/my5t3ry/juun-fzf/master/logo-small_ls.png)

## learn damnit

![video](https://raw.githubusercontent.com/my5t3ry/juun-fzf/master/learn.gif)

[comment]: <> (in this example the search learns that by 'd' I mean `git diff`, not `dmesg`)

## [Here be dragons](https://en.wikipedia.org/wiki/Here_be_dragons)

attempt to fix my bash history rage

* keeps my bash/zsh history across terminals
* local service that i can personalize and run some ML
* stores it in ~/.juun.json [0600]
* per terminal plus global history
+ has per terminal local state and when that is exhausted it uses the global history to go back

this annoys me so much in the current history(1) it is unbelivable

## install/run

### with curl

supported:
* macos amd64 [tested by me and some friends]
* linux amd64 [tested on travisci]
* freebsd amd64 [compiled, but not tested]

```
curl -L https://raw.githubusercontent.com/my5t3ry/juun-fzf/master/download-and-install.sh | bash
```

[comment]: <> (### with homebrew:)

[comment]: <> (```)

[comment]: <> (brew tap jackdoe/tap)

[comment]: <> (brew install juun)

[comment]: <> (```)

[comment]: <> (&#40;then you need to follow the instructions&#41;)


### from source
requires golang

```
go get github.com/sirupsen/logrus
go get github.com/chzyer/readline
go get github.com/sevlyar/go-daemon
go get github.com/rekki/go-query
go get github.com/rekki/go-query-index
go get github.com/jackdoe/juun/vw

git clone https://github.com/my5t3ry/juun-fzf
cd juun-fzf && make
```

```
make install # this will add 'source juun/dist/setup.sh' to .bash_profile and .zshrc
```

this will hook up, down and ctrl+r to use juun for fzf history widget. press ctrl+r to update query results for changed $BUFFER
if compdef is available "compdef _juun_completions -first-" is added to support ctx aware tab completion.
it also hooks to preexec()

setup.sh will always try to start `juun.service` which listens on $HOME/.juun.sock
logs are in $HOME/.juun.log and pid is $HOME/.juun.pid

## import

if you want to import your current history run:

```
$ HISTTIMEFORMAT= history | ~/.juun.dist/juun.import
```

this will add each of your history lines to juun

## scoring

running search for `m` from one terminal gives the following score
(edge gram `e_m`)

```

2018/12/23 14:34:07 tfidf: 0.086013 timeScore: -4.162952 terminalScore:0.000000 countScore:4.543295, age: 14552s - make
2018/12/23 14:34:07 tfidf: 0.086013 timeScore: -4.904991 terminalScore:0.000000 countScore:1.386294, age: 80350s - make -n
2018/12/23 14:34:07 tfidf: 0.086013 timeScore: -4.820044 terminalScore:0.000000 countScore:0.693147, age: 66075s - make clean
2018/12/23 14:34:07 tfidf: 0.086013 timeScore: -4.757343 terminalScore:0.000000 countScore:1.945910, age: 57192s - make install
2018/12/23 14:34:07 tfidf: 0.086013 timeScore: -4.754386 terminalScore:0.000000 countScore:0.693147, age: 56804s - mkdir brew
2018/12/23 14:34:07 tfidf: 0.086013 timeScore: -4.114077 terminalScore:0.000000 countScore:0.693147, age: 13003s - git commit -a -m 'make it feel more natural; fix issues with newline'

```

* [tfidf](https://en.wikipedia.org/wiki/Tf%E2%80%93idf) `(occurances of [m] in the line) * log(1-totalNumberDocuments/documentsHaving[m])`
* terminalScore `100 command was ran on this terminal session, 0 otherwise`
* countScore  `log(number of times this command line was executed)`
* timeScore `log10(seconds between now and the command)`
* score `tfidf + terminalScore + countScore + timeScore`

## learning

If you have [vowpal wabbit](https://github.com/VowpalWabbit/vowpal_wabbit) installed (`brew install vowpal-wabbit`), juun will use it to re-sort the last 5 items from the search
when you click (use) one of the recommended items it learns positive signal, if you use something else rather than the shown results, it will learn negative signal

Vowpal is started with quadratic interractions between `i` and `c` namespaces, the features are split into item and user features context, and the user context is `query` and the `time`.
For example: `git diff` is featurized as
```
|i_id id_2
|i_text git diff
|i_count count:4.454347
|i_time year_2018 day_25 month_12 hour_16
|i_score tfidf:1.870803 timeScore:-0.903090 countScore:4.454347 terminalScore_100
```
and the user is featurized as:

```
|c_user_time year_2018 day_25 month_12 hour_16
|c_query git
|c_cwd juun _Users_jack_work_juun
```

i_time is the last time this command was used, the idea is to learn patterns like: in the morning i prefer those commands, and in the evenening i prefer those
As you can see one of the features of the items is the search engine's score.


example log line in `~/.juun.log`

```
2018/12/25 16:54:26 sending 1 |i_id id_2  |i_text git diff  |i_count count:4.454347  |i_time year_2018 day_25 month_12 hour_16  |c_user_time year_2018 day_25 month_12 hour_16  |c_query git  |i_score tfidf:1.870803 timeScore:-0.903090 countScore:4.454347 terminalScore_100 |c_cwd juun _Users_jack_work_juun
2018/12/25 16:54:26 received 0.624512 0.584649 0.664374
```

## bash4
* there is *some* support for bash4, and it kind of works, but there are some issues, so use on your own risk
* bash's [preexec](https://github.com/rcaloras/bash-preexec) is super hacky, i strongly suggest to use juun with zsh; making juun is actually what made me switch from bash to zsh
* in some bash versions up/down gives `bash_execute_unix_command: cannot find keymap for command`, in the same time \C-p and \C-p work, to not hook to up/down use `BASH_UPDOWN_BROKEN=1 source setup.sh` in your bashrc
* sometimes the terminal gets broken and you have to 'reset'

## credit

logo: Icons made by <a href="https://www.freepik.com/" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" 			    title="Flaticon">www.flaticon.com</a> is licensed by <a href="http://creativecommons.org/licenses/by/3.0/" 			    title="Creative Commons BY 3.0" target="_blank">CC 3.0 BY</a>

