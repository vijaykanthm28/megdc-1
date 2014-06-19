package app

import (
    "encoding/json"
	"bytes"
	"errors"
	"fmt"
	"github.com/tsuru/config"
	"github.com/megamsys/cloudinabox/action"
	"github.com/megamsys/cloudinabox/exec"
	"log"
	"text/template"
	"path"
	"bufio"
	"strings"
	"regexp"
	"os"
	"bitbucket.org/kardianos/osext"
)

const (
	keyremote_repo = "remote_repo="
	keylocal_repo  = "local_repo="
	keyproject     = "project="
	ganetipreinstall = "bash ../conf/ganeti/mganeti_preinstall.sh"
	ganetiverify = "bash ../conf/ganeti/mganeti_verify.sh"
	ganetipostinstall = "bash ../conf/ganeti/mganeti_postinstall.sh"
	ganetiinstall = "bash ../conf/ganeti/mganeti_install.sh"
	opennebulapreinstall = "bash ../conf/opennebula/one_preinstall.sh"
	opennebulaverify = "bash ../conf/opennebula/one_verify.sh"
	opennebulapostinstall = "bash ../conf/opennebula/one_postinstall.sh"
	opennebulainstall = "bash ../conf/opennebula/one_install.sh"
	rootPath  = "/tmp"
	defaultEnvPath = "conf/env.sh"	
	drbd_mnt = "/drbd_mnt"
)

var ErrAppAlreadyExists = errors.New("there is already an app with this name.")

func CommandExecutor(app *App) (action.Result, error) {
	var e exec.OsExecutor
	var b bytes.Buffer
	var commandWords []string
	commandWords = strings.Fields(app.Command)
	
	fmt.Println(commandWords, len(commandWords))

	if len(commandWords) > 0 {
		if err := e.Execute(commandWords[0], commandWords[1:], nil, &b, &b); err != nil {
			log.Printf("stderr:%s", b)		
			return nil, err
		}
	}
   
	log.Printf("%s", b)
	return &app, nil
}

//Remove the installed packages..
var remove = action.Action{
	Name: "remove",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		var app App
		switch ctx.Params[0].(type) {
		case App:
			app = ctx.Params[0].(App)
		case *App:
			app = *ctx.Params[0].(*App)
		default:
			return nil, errors.New("First parameter must be App or *App.")
		}
		
		return CommandExecutor(&app)
	},
	Backward: func(ctx action.BWContext) {
		app := ctx.FWResult.(*App)	
		log.Printf("[%s] Nothing to recover for %s", app.ClusterName)
	},
	MinParams: 1,
}

//
// Install Ganeti or Opennebula packages 
//
 var install = action.Action{
	Name: "install",
	Forward: func(ctx action.FWContext) (action.Result, error) {
		var app App
		switch ctx.Params[0].(type) {
		case App:
			app = ctx.Params[0].(App)
		case *App:
			app = *ctx.Params[0].(*App)
		default:
			return nil, errors.New("First parameter must be App or *App.")
		}
		switch app.InstallPackage {
		case "Ganeti":
		         app.command = ganetiverify
		         cmd, err := CommandExecutor(&app)
		         if err != nil {
			         log.Printf("Ganeti install pre verification failed %s", err)
			         return nil, errors.New("Ganeti install pre verification failed")
		          }
		         app.command = ganetipreinstall
		         cmd, err := CommandExecutor(&app)
		         if err != nil {
			         log.Printf("Ganeti pre installation failed %s", err)
			         return nil, errors.New("Ganeti pre installation failed")
		         }
		         app.command = ganetiinstall
		         cmd, err := CommandExecutor(&app)
		         if err != nil {
			         log.Printf("Ganeti installation failed %s", err)
			         return nil, errors.New("Ganeti installation failed")
		         }
		         app.command = ganetipostinstall
		         cmd, err := CommandExecutor(&app)
		         if err != nil {
			         log.Printf("Ganeti post installation failed %s", err)
			         return nil, errors.New("Ganeti post installation failed")
		         }
		      return cmd, err
		 case "Opennebula":
		        app.command = opennebulaverify
		        cmd, err := CommandExecutor(&app)
		        if err != nil {
			        log.Printf("Opennebula install pre verification failed %s", err)
			        return nil, errors.New("Opennebula install pre verification failed")
		        }
		        app.command = opennebulapreinstall
		        cmd, err := CommandExecutor(&app)
		        if err != nil {
			        log.Printf("Opennebula pre installation failed %s", err)
			        return nil, errors.New("Opennebula pre installation failed")
		        }
		        app.command = opennebulainstall
		        cmd, err := CommandExecutor(&app)
		        if err != nil {
			        log.Printf("Opennebula installation failed %s", err)
			        return nil, errors.New("Opennebula installation failed")
		        }
		        app.command = opennebulapostinstall
		        cmd, err := CommandExecutor(&app)
		        if err != nil {
			        log.Printf("Opennebula post installation failed %s", err)
			        return nil, errors.New("Opennebula post installation failed")
		        }
		     return cmd, err
		 default:
			return nil, errors.New("Wrong package name.")
		}
	},
	Backward: func(ctx action.BWContext) {
	app := ctx.FWResult.(*App)
		log.Printf("[%s] Nothing to recover for %s", app.ClusterName)
	},
	MinParams: 1,
	}

