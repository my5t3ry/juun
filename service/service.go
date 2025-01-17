package main

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	. "../common"
	. "../config"
	. "github.com/jackdoe/juun/vw"
	"github.com/sevlyar/go-daemon"
	log "github.com/sirupsen/logrus"
)

func oneLine(history *History, c net.Conn) {
	hdr := make([]byte, 4)
	_, err := io.ReadFull(c, hdr)
	if err != nil {
		c.Close()
		return
	}

	dataLen := binary.LittleEndian.Uint32(hdr)
	data := make([]byte, dataLen)
	_, err = io.ReadFull(c, data)
	if err != nil {
		log.Warnf("err: %s", err.Error())
		c.Close()
		return
	}

	ctrl := &Control{}
	err = json.Unmarshal(data, ctrl)
	if err != nil {
		log.Warnf("err: %s", err.Error())
		c.Close()
		return
	}

	out := ""
	log.Infof("datalen: %d %#v", dataLen, ctrl)
	switch ctrl.Command {
	case "add":
		if len(ctrl.Payload) > 0 {
			line := strings.Trim(ctrl.Payload, "\n")
			if len(line) > 0 {
				history.add(line, ctrl.Env)
			}
			history.gotoend()
		}
	case "end":
		log.Infof("end command is deprecated")
	case "reindex":
		history.SelfReindex()
	case "save":
		history.Save()
	case "delete":
		log.Infof("delete command is deprecated")
	case "search":
		line := strings.Replace(ctrl.Payload, "\n", "", -1)
		if len(line) > 0 {
			lines := history.search(line, ctrl.Env)
			j, err := json.Marshal(lines)
			if err != nil {
				log.WithError(err).Printf("failed to encode")
			}
			out = string(j)
		}
	case "list":
		cfg := GetConfig()
		lines := history.getLastLines()[:cfg.SearchResults]
		if lines != nil {
			j, err := json.Marshal(lines)
			if err != nil {
				log.WithError(err).Printf("failed to encode")
			}
			out = string(j)
		} else {
			out = ""
		}
	case "up":
		out = history.up()
	case "down":
		out = history.down()
	}

	_, _ = c.Write([]byte(out))
	c.Close()
}

func listen(history *History, ln net.Listener) {
	for {
		fd, err := ln.Accept()
		if err != nil {
			log.Warnf("accept error: %s", err.Error())
			break
		}

		go oneLine(history, fd)
	}
}

func isRunning(pidFile string) bool {
	if piddata, err := ioutil.ReadFile(pidFile); err == nil {
		if pid, err := strconv.Atoi(string(piddata)); err == nil {
			if process, err := os.FindProcess(pid); err == nil {
				if err := process.Signal(syscall.Signal(0)); err == nil {
					return true
				}
			}
		}
	}
	return false
}

func main() {
	home := GetHome()
	socketPath := path.Join(home, ".juun.sock")
	pidFile := path.Join(home, ".juun.pid")
	modelFile := path.Join(home, ".juun.vw")
	if isRunning(pidFile) {
		os.Exit(0)
	}

	config := GetConfig()
	history := NewHistory()
	history.Load()

	ctx := &daemon.Context{
		PidFileName: pidFile,
		PidFilePerm: 0600,
		LogFileName: path.Join(home, ".juun.log"),
		LogFilePerm: 0600,
		WorkDir:     home,
		Umask:       027,
	}

	d, err := ctx.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}

	log.Infof("---------------------")
	log.Infof("listening to: %s, model: %s", socketPath, modelFile)

	if config.AutoSaveIntervalSeconds < 30 {
		log.Warnf("autosave interval is too short, limiting it to 30 seconds")
		config.AutoSaveIntervalSeconds = 30
	}
	level, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		log.Warnf("failed to parse level %s: %s", config.LogLevel, err)
	} else {
		log.SetLevel(level)
	}
	log.SetReportCaller(true)

	var vw *Bandit
	if config.EnableVowpalWabbit {
		vw = NewBandit(modelFile) // XXX: can be nil if vw is not found
	}

	history.vw = vw
	_ = syscall.Unlink(socketPath)
	sock, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal("Listen error: ", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	save := func() {
		history.Save()
	}

	cleanup := func() {
		log.Infof("juun teardown")

		save()
		sock.Close()

		_ = os.Chmod(modelFile, 0600)
		if vw != nil {
			vw.Shutdown()
		}
		_ = ctx.Release()
		os.Exit(0)
	}

	go func() {
		<-sigs
		cleanup()
	}()

	if config.AutoSaveIntervalSeconds > 0 {
		go func() {
			for {
				save()
				time.Sleep(time.Duration(config.AutoSaveIntervalSeconds) * time.Second)
			}
		}()
	}

	listen(history, sock)
	cleanup()
}
