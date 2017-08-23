package flag

import (
	"code.cloudfoundry.org/cli/types"
	flags "github.com/jessevdk/go-flags"
)

type Port struct {
	types.NullInt
}

func (i *Port) UnmarshalFlag(val string) error {
	err := i.ParseFlagValue(val)
	if err != nil || i.Value < 0 {
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: "invalid argument for flag '-i' (expected int > 0)",
		}
	}
	return nil
}
