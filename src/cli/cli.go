package cli

import (
	"errors"
	"flag"
	"github.com/google/uuid"
	"os"
)

/*
Authentication
command: "--login/-l"
	1. "Anonymous" type
		e.g. dataServer --login "anonymous"

	2. "Basic" auth type
		e.g. dataServer --login "JohnSmith" --pass "123"
*/

type CLI struct {
	authFlag *authFlag
}

type authFlag struct {
	username *string
	password *string
}

func (cli *CLI) auth() *flag.FlagSet {
	auth := flag.NewFlagSet("auth", flag.ExitOnError)

	username := auth.String("username", "", "anonymous, basic")
	password := auth.String("password", "", "password")

	cli.authFlag = &authFlag{
		username,
		password,
	}

	return auth
}

func RunCLI() error {
	cli := CLI{}

	switch os.Args[1] {

	case "auth":
		err := cli.auth().Parse(os.Args[2:])
		if err != nil {
			return err
		}

		_, err = uuid.NewRandom()
		if err != nil {
			return err
		}

	default:
		return errors.New("auth flag expected")
	}

	return nil
}
