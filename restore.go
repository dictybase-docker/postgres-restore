package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "pg-restore"
	app.Usage = "restore postgresql database from archive file(primarilly in kubernetes cluster)"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "move-from, from",
			Usage: "folder from where the archive will be moved",
		},
		cli.StringFlag{
			Name:  "move-to, to",
			Usage: "where the archive will be moved. It is also the folder from where archive will be loaded",
		},
		cli.StringFlag{
			Name:  "archive-name",
			Usage: "name of the archive file",
		},
		cli.StringFlag{
			Name:   "chado-user",
			Usage:  "name of chado database user",
			EnvVar: "CHADO_USER",
		},
		cli.StringFlag{
			Name:   "chado-pass",
			Usage:  "chado database password",
			EnvVar: "CHADO_PASS",
		},
		cli.StringFlag{
			Name:   "chado-database",
			Usage:  "name of chado database",
			EnvVar: "CHADO_DB",
		},
		cli.StringFlag{
			Name:  "service-name",
			Usage: "kubernetes service name for chado database",
		},
	}
	app.Action = restoreAction
	app.Run(os.Args)
}

func validateArgs(c *cli.Context) error {
	for _, flag := range c.GlobalFlagNames() {
		if len(c.String(flag)) == 0 {
			return fmt.Errorf("parameter for flag %s is not provided\n", flag)
		}
	}
	return nil
}

func restoreAction(c *cli.Context) error {
	if err := validateArgs(c); err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	from := filepath.Join(c.String("move-from"), c.String("archive-name"))
	to := filepath.Join(c.String("move-to"), c.String("archive-name"))
	if _, err := os.Stat(to); os.IsNotExist(err) {
		err = os.MkdirAll(to, os.ModeDir)
		if err != nil {
			return cli.NewExitError(err.Error(), 2)
		}
	}

	// move the file
	err := os.Rename(from, to)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}

	// now run the restore
	pg, err := exec.LookPath("pg-restore")
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	srv := fmt.Sprintf("$%s_%s", strings.ToUpper(c.String("service-name")), "SERVICE_HOST")
	rcmd := []string{
		"-j",
		"4",
		"-Fc",
		"-O",
		"-x",
		"-w",
		"-U",
		c.String("chado-user"),
		"-h",
		os.Getenv(srv),
		"-d",
		c.String("chado-database"),
		filepath.Join(c.String("copy-to"), c.String("archive-name")),
	}
	out, err := exec.Command(pg, rcmd...).CombinedOutput()
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	log.Println(out)
	return nil
}
