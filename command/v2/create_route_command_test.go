package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = FDescribe("Create Route Command", func() {
	var (
		cmd             CreateRouteCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeCreateRouteActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeCreateRouteActor)

		cmd = CreateRouteCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.Space = "some-space"
		cmd.RequiredArgs.Domain = "some-domain"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeActor.CloudControllerAPIVersionReturns(command.MinVersionTCPRouting)
	})

	DescribeTable("argument combinations",
		func(expectedErr error, hostname string, path string, port flag.Port, randomPort bool) {
			cmd.Port = port
			cmd.Hostname = hostname
			cmd.Path = path
			cmd.RandomPort = randomPort

			executeErr := cmd.Execute(nil)
			if expectedErr == nil {
				Expect(executeErr).To(BeNil())
			} else {
				Expect(executeErr).To(Equal(expectedErr))
			}
		},
		Entry("hostname", nil, "some-hostname", "", flag.Port{types.NullInt{IsSet: false}}, false),
		Entry("path", nil, "", "some-path", flag.Port{types.NullInt{IsSet: false}}, false),
		Entry("hostname and path", nil, "some-hostname", "some-path", flag.Port{types.NullInt{IsSet: false}}, false),
		Entry("hostname and port", translatableerror.ArgumentCombinationError2{Args: []string{"--hostname", "--port"}}, "some-hostname", "", flag.Port{types.NullInt{IsSet: true}}, false),
		Entry("path and port", translatableerror.ArgumentCombinationError2{Args: []string{"--path", "--port"}}, "", "some-path", flag.Port{types.NullInt{IsSet: true}}, false),
		Entry("hostname, path, and port", translatableerror.ArgumentCombinationError2{Args: []string{"--hostname", "--path", "--port"}}, "some-hostname", "some-path", flag.Port{types.NullInt{IsSet: true}}, false),
		Entry("hostname and random port", translatableerror.ArgumentCombinationError2{Args: []string{"--hostname", "--random-port"}}, "some-hostname", "", flag.Port{types.NullInt{IsSet: false}}, true),
		Entry("path and random port", translatableerror.ArgumentCombinationError2{Args: []string{"--path", "--random-port"}}, "", "some-path", flag.Port{types.NullInt{IsSet: false}}, true),
		Entry("hostname, path, and random port", translatableerror.ArgumentCombinationError2{Args: []string{"--hostname", "--path", "--random-port"}}, "some-hostname", "some-path", flag.Port{types.NullInt{IsSet: false}}, true),
		Entry("port", nil, "", "", flag.Port{types.NullInt{IsSet: true}}, false),
		Entry("random port", nil, "", "", flag.Port{types.NullInt{IsSet: false}}, true),
		Entry("port and random port", translatableerror.ArgumentCombinationError2{Args: []string{"--port", "--random-port"}}, "", "", flag.Port{types.NullInt{IsSet: true}}, true),
	)

	DescribeTable("minimum api version checks",
		func(expectedErr error, port flag.Port, randomPort bool, path string, apiVersion string) {
			cmd.Port = port
			cmd.RandomPort = randomPort
			cmd.Path = path
			fakeActor.CloudControllerAPIVersionReturns(apiVersion)

			executeErr := cmd.Execute(nil)
			if expectedErr == nil {
				Expect(executeErr).To(BeNil())
			} else {
				Expect(executeErr).To(Equal(expectedErr))
			}
		},
		Entry("port, CC Version 2.52.0", translatableerror.MinimumAPIVersionNotMetError{
			CurrentVersion: "2.52.0",
			MinimumVersion: command.MinVersionTCPRouting,
		}, flag.Port{types.NullInt{IsSet: true}}, false, "", "2.52.0"),
		Entry("port, CC Version 2.53.0", nil, flag.Port{types.NullInt{IsSet: true}}, false, "", command.MinVersionTCPRouting),

		Entry("random-port, CC Version 2.52.0", translatableerror.MinimumAPIVersionNotMetError{
			CurrentVersion: "2.52.0",
			MinimumVersion: command.MinVersionTCPRouting,
		}, flag.Port{}, true, "", "2.52.0"),
		Entry("random-port, CC Version 2.53.0", nil, flag.Port{}, true, "", command.MinVersionTCPRouting),

		Entry("path, CC Version 2.35.0", translatableerror.MinimumAPIVersionNotMetError{
			CurrentVersion: "2.35.0",
			MinimumVersion: command.MinVersionHTTPRoutePath,
		}, flag.Port{}, false, "some-path", "2.35.0"),
		Entry("path, CC Version 2.36.0", nil, flag.Port{}, false, "some-path", command.MinVersionHTTPRoutePath),
	)

	Context("when all the arguments check out", func() {
		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		Context("when checking target fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(sharedaction.NotLoggedInError{BinaryName: binaryName})
			})

			It("returns an error if the check fails", func() {
				Expect(executeErr).To(MatchError(translatableerror.NotLoggedInError{BinaryName: "faceman"}))

				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				_, checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkTargetedOrg).To(BeTrue())
				Expect(checkTargetedSpace).To(BeFalse())
			})
		})

		Context("when getting the current user returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("getting current user error")
				fakeConfig.CurrentUserReturns(
					configv3.User{},
					expectedErr)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		Context("when the user is logged in, and the org is targeted", func() {
			BeforeEach(func() {
				fakeConfig.HasTargetedOrganizationReturns(true)
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
				fakeConfig.CurrentUserReturns(
					configv3.User{Name: "some-user"},
					nil)
			})

		})
	})
})
