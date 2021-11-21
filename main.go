// winsvcwrap is an adapter utility for running arbitrary daemons as Windows services.
package main

import (
	"github.com/hlandau/dexlogconfig"
	"github.com/hlandau/xlog"
	"gopkg.in/hlandau/easyconfig.v1"
	"gopkg.in/hlandau/service.v2"
	"os"
	"os/exec"
)

var log, Log = xlog.New("winsvcwrap")

// Configuration for the daemon.
type Config struct {
	Run string   `usage:"Path to service executable to spawn" default:""`
	Arg []string `usage:"Argument to pass to service executable (specify multiple times)" default:""`
	CWD string   `usage:"Working directory to use for spawned service" default:""`
}

type ctlEventType int

const (
	ctlTerminated ctlEventType = iota
	ctlStopReq
)

type ctlEvent struct {
	Type     ctlEventType
	Error    error
	DoneChan chan error
}

// Main object for this daemon.
type Supervisor struct {
	cfg     Config
	cmd     *exec.Cmd
	ctlChan chan ctlEvent
}

func New(cfg *Config) (*Supervisor, error) {
	sup := &Supervisor{
		cfg:     *cfg,
		ctlChan: make(chan ctlEvent, 2),
	}

	log.Debugf("supervisor instantiated")
	return sup, nil
}

func (sup *Supervisor) Start() error {
	log.Debugf("starting supervisor...")

	sup.cmd = exec.Command(sup.cfg.Run, sup.cfg.Arg...)
	sup.cmd.Dir = sup.cfg.CWD
	err := sup.cmd.Start()
	if err != nil {
		log.Criticale(err, "could not start service to be supervised by winsvcwrap")
		return err
	}

	go sup.ctlLoop()
	go sup.waitTerm()

	return nil
}

func (sup *Supervisor) ctlLoop() {
	var pendingStopReq chan error
	for {
		ev := <-sup.ctlChan
		switch ev.Type {
		case ctlTerminated:
			if pendingStopReq != nil {
				pendingStopReq <- ev.Error
			} else {
				if ev.Error != nil {
					log.Criticale(ev.Error, "service supervised by winsvcwrap exited unexpectedly with error")
				} else {
					log.Critical("service supervised by winsvcwrap exited unexpectedly without error")
				}
				// This should not happen, so just exit with error so the Windows service
				// manager will restart us.
				os.Exit(3)
			}
		case ctlStopReq:
			if pendingStopReq != nil {
				panic("unreachable")
			}
			pendingStopReq = ev.DoneChan

			err := sup.cmd.Process.Kill()
			log.Errore(err, "failed to kill supervised process, continuing...")
		}
	}
}

func (sup *Supervisor) waitTerm() {
	err := sup.cmd.Wait()
	sup.ctlChan <- ctlEvent{Type: ctlTerminated, Error: err}
}

func (sup *Supervisor) Stop() error {
	log.Debugf("processing request to stop supervised process...")
	doneCh := make(chan error)
	sup.ctlChan <- ctlEvent{Type: ctlStopReq, DoneChan: doneCh}
	err := <-doneCh
	if err != nil {
		log.Noticee(err, "request to stop supervised process completed with error")
	} else {
		log.Notice("request to stop supervised process completed")
	}
	return nil
}

func main() {
	cfg := &Config{}
	config := easyconfig.Configurator{
		ProgramName: "winsvcwrap",
	}
	config.ParseFatal(cfg)
	dexlogconfig.Init()

	service.Main(&service.Info{
		Name:          "winsvcwrap",
		Description:   "Windows service hosting adapter",
		DefaultChroot: service.EmptyChrootPath,
		NewFunc: func() (service.Runnable, error) {
			return New(cfg)
		},
	})
}
