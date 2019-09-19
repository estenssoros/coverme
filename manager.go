package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

type manager struct {
	*configuration
	Restart    chan bool
	gil        *sync.Once
	context    context.Context
	cancelFunc context.CancelFunc
}

func newManager(ctx context.Context, c *configuration) *manager {
	ctx, cancelFunc := context.WithCancel(ctx)
	return &manager{
		configuration: c,
		Restart:       make(chan bool),
		gil:           &sync.Once{},
		context:       ctx,
		cancelFunc:    cancelFunc,
	}
}

// runs build and reports on error
func (m *manager) buildTransaction(fn func() error) {
	if err := fn(); err != nil {
		logrus.Error(err)
	}
}

// runAndListen runs a command and reports the output to console
func (m *manager) runAndListen(cmd *exec.Cmd) error {
	var stderr bytes.Buffer
	mw := io.MultiWriter(&stderr, os.Stderr)
	cmd.Stderr = mw
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("%s\n%s", err, stderr.String())
	}

	logrus.Infof("running: %s (PID: %d)", strings.Join(cmd.Args, " "), cmd.Process.Pid)
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("%s\n%s", err, stderr.String())
	}
	return nil
}

// test builds a go binary to run
func (m *manager) test(event fsnotify.Event) {
	logrus.Info("testing...")
	m.gil.Do(func() {
		defer func() {
			m.gil = &sync.Once{}
		}()

		m.buildTransaction(func() error {
			time.Sleep(m.BuildDelay * time.Millisecond)
			now := time.Now()
			logrus.Infof("retesting on: %s", event.Name)

			if _, err := os.Stat("test-coverage"); err != nil {
				os.Mkdir("test-coverage", 0700)
			}

			{
				args := []string{
					"test",
					"./...",
					"-coverprofile=test-coverage/c.out",
				}
				cmd := exec.Command("go", args...)

				if err := m.runAndListen(cmd); err != nil {
					fmt.Println(err)
					return err
				}
			}
			{
				args := []string{
					"tool",
					"cover",
					"-html=test-coverage/c.out",
					"-o=test-coverage/index.html",
				}
				cmd := exec.Command("go", args...)

				if err := m.runAndListen(cmd); err != nil {
					fmt.Println(err)
					return err
				}
			}

			logrus.Infof("building completed (time: %s)", time.Since(now))
			return nil
		})

	})
}

// start creates a watcher for fs events, builds, and monitors
func (m *manager) start() error {
	w := newWatcher(m)
	w.start()

	go m.test(fsnotify.Event{Name: ":start:"})

	// watch files
	go func() {
		logrus.Info("watching files...")
		for {
			select {
			case event := <-w.Events:
				if event.Op != fsnotify.Chmod {
					go m.test(event)
				}
				w.Remove(event.Name)
				w.Add(event.Name)
			case <-m.context.Done():
				break
			}
		}
	}()

	go func() {
		for {
			select {
			case err := <-w.Errors:
				logrus.Error(err)
			case <-m.context.Done():
				break
			}
		}
	}()

	for {
		_, err := os.Stat("test-coverage/index.html")
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	exec.Command("live-server", "test-coverage").Run()
	return nil
}
