// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package cli

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/keybase/slackbot"

	"gopkg.in/alecthomas/kingpin.v2"
)

// IsParseContextValid checks if the kingpin context is valid
func IsParseContextValid(app *kingpin.Application, args []string) error {
	if pcontext, perr := app.ParseContext(args); pcontext == nil {
		return perr
	}
	return nil
}

// Parse kingpin args and return valid command, usage, and error
func Parse(app *kingpin.Application, args []string, stringBuffer *bytes.Buffer) (string, string, error) {
	log.Printf("Parsing args: %#v", args)
	// Make sure context is valid otherwise showing Usage on error will fail later.
	// This is a workaround for a kingpin bug.
	if err := IsParseContextValid(app, args); err != nil {
		return "", "", err
	}

	cmd, err := app.Parse(args)

	if err != nil && stringBuffer.Len() == 0 {
		log.Printf("Error in parsing command: %s. got %s", args, err)
		_, _ = io.WriteString(stringBuffer, fmt.Sprintf("I don't know what you mean by `%s`.\nError: `%s`\nHere's my usage:\n\n", strings.Join(args, " "), err.Error()))
		// Print out help page if there was an error parsing command
		app.Usage([]string{})
	}

	if stringBuffer.Len() > 0 {
		return "", slackbot.BlockQuote(stringBuffer.String()), nil
	}

	return cmd, "", err
}
