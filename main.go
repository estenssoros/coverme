package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	logrus.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
}

var rootCmd = &cobra.Command{
	Use:   "coverme",
	Short: "keeps ya covered",
	RunE: func(cmd *cobra.Command, args []string) error {
		defer func() {
			if r := recover(); r != nil {
				logrus.Error(r)
			}
		}()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		return errors.Wrap(startApp(ctx), "start dev server")
	},
}

func startApp(ctx context.Context) error {
	logrus.Infof("starting app")
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("ERROR: ", r)
		}
	}()

	c := &configuration{
		AppRoot:            ".",
		IgnoredFolders:     []string{""},
		IncludedExtensions: []string{".go"},
		BuildDelay:         time.Duration(200),
	}

	_, err := exec.LookPath("live-server")
	if err != nil {
		return errors.Wrap(err, "live server not found install with `npm install -g live-server`")
	}
	return newManager(ctx, c).start()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
