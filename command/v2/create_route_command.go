package v2

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/types"
)

//go:generate counterfeiter . CreateRouteActor

type CreateRouteActor interface {
	CloudControllerAPIVersion() string
	CreateRouteWithExistenceCheck(orgGUID string, spaceName string, host string, path string, port types.NullInt, generatePort bool) (v2action.Route, v2action.Warnings, error)
}

type CreateRouteCommand struct {
	RequiredArgs    flag.SpaceDomain `positional-args:"yes"`
	Hostname        string           `long:"hostname" short:"n" description:"Hostname for the HTTP route (required for shared domains)"`
	Path            string           `long:"path" description:"Path for the HTTP route"`
	Port            flag.Port        `long:"port" description:"Port for the TCP route"`
	RandomPort      bool             `long:"random-port" description:"Create a random port for the TCP route"`
	usage           interface{}      `usage:"Create an HTTP route:\n      CF_NAME create-route SPACE DOMAIN [--hostname HOSTNAME] [--path PATH]\n\n   Create a TCP route:\n      CF_NAME create-route SPACE DOMAIN (--port PORT | --random-port)\n\nEXAMPLES:\n   CF_NAME create-route my-space example.com                             # example.com\n   CF_NAME create-route my-space example.com --hostname myapp            # myapp.example.com\n   CF_NAME create-route my-space example.com --hostname myapp --path foo # myapp.example.com/foo\n   CF_NAME create-route my-space example.com --port 5000                 # example.com:5000"`
	relatedCommands interface{}      `related_commands:"check-route, domains, map-route"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       CreateRouteActor
}

func (cmd *CreateRouteCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	cmd.SharedActor = sharedaction.NewActor()

	_, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	return nil
}

func (cmd CreateRouteCommand) Execute(args []string) error {
	err := cmd.validateArguments()
	if err != nil {
		return shared.HandleError(err)
	}

	err = cmd.SharedActor.CheckTarget(cmd.Config, true, false)
	if err != nil {
		return shared.HandleError(err)
	}

	err = cmd.minimumFlagVersions()
	if err != nil {
		return shared.HandleError(err)
	}

	_, err = cmd.Config.CurrentUser()
	if err != nil {
		return shared.HandleError(err)
	}

	// actor.CreateRouteWithExistenceCheck(.....)
	// if err {
	//	if exists print ok
	//	else fail
	// }
	// print ok

	return nil
}

func (cmd CreateRouteCommand) minimumFlagVersions() error {
	if err := command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), command.MinVersionHTTPRoutePath); cmd.Path != "" && err != nil {
		return err
	}
	if err := command.MinimumAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), command.MinVersionTCPRouting); (cmd.Port.IsSet || cmd.RandomPort) && err != nil {
		return err
	}
	return nil
}

func (cmd CreateRouteCommand) validateArguments() error {
	var failedArgs []string

	if cmd.Hostname != "" {
		failedArgs = append(failedArgs, "--hostname")
	}
	if cmd.Path != "" {
		failedArgs = append(failedArgs, "--path")
	}
	if cmd.Port.IsSet {
		failedArgs = append(failedArgs, "--port")
	}
	if cmd.RandomPort {
		failedArgs = append(failedArgs, "--random-port")
	}

	switch {
	case (cmd.Hostname != "" || cmd.Path != "") && (cmd.Port.IsSet || cmd.RandomPort),
		cmd.Port.IsSet && cmd.RandomPort:
		return translatableerror.ArgumentCombinationError2{Args: failedArgs}
	}

	return nil
}
