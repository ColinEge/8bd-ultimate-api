package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/ColinEge/8bd-ultimate-api/client"
	"github.com/ColinEge/8bd-ultimate-api/keydefs"
)

const usage = `Usage: 8bdkbd <command> [flags]

Commands:
  scan      List connected keyboards
  action    Read a key's stored mapping override
  bind      Bind a key to an action (requires --yes)
  restore   Rebind a key to its default action (requires --yes)

Flags follow the command. Use <command> --help for details.
Key IDs are decimal or hexadecimal (e.g. 4 or 0x04), not key labels.
Restore may retain a stored mapping override.
`

type clientFactory func(time.Duration) client.Client

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr, func(timeout time.Duration) client.Client {
		return client.New(client.WithTimeout(timeout))
	}); err != nil {
		fmt.Fprintf(os.Stderr, "8bdkbd: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, out, errOut io.Writer, newClient clientFactory) (err error) {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		_, err := io.WriteString(out, usage)
		return err
	}
	command := args[0]
	switch command {
	case "scan", "action", "bind", "restore":
	default:
		return fmt.Errorf("unknown command %q; use --help", command)
	}
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(errOut)
	timeout := flags.Duration("timeout", 10*time.Second, "operation timeout (must be positive)")
	var device, key, action *string
	var yes *bool
	if command != "scan" {
		device = flags.String("device", "", "device ID from scan (required)")
		key = flags.String("key", "", "physical key ID, decimal or 0x-prefixed hex (required)")
	}
	if command == "bind" {
		action = flags.String("action", "", "3-to-6 dash-separated hex bytes, e.g. 07-00-05 (required)")
	}
	if command == "bind" || command == "restore" {
		yes = flags.Bool("yes", false, "confirm writing to the device")
	}
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	if *timeout <= 0 {
		return errors.New("--timeout must be positive")
	}
	var keyID keydefs.KeyID
	var actionCode keydefs.ActionCode
	if command != "scan" {
		if strings.TrimSpace(*device) == "" {
			return errors.New("--device is required; run scan to find a device ID")
		}
		if *key == "" {
			return errors.New("--key is required")
		}
		base := 10
		if strings.HasPrefix(*key, "0x") || strings.HasPrefix(*key, "0X") {
			base = 0
		}
		n, err := strconv.ParseUint(*key, base, 8)
		if err != nil {
			return fmt.Errorf("invalid --key %q: %w", *key, err)
		}
		keyID = keydefs.KeyID(n)
	}
	if command == "bind" {
		parts := strings.Split(*action, "-")
		if len(parts) < 3 || len(parts) > 6 {
			return errors.New("--action must contain 3 to 6 dash-separated hex bytes")
		}
		actionCode, err = keydefs.ParseActionCode(*action)
		if err != nil {
			return fmt.Errorf("invalid --action: %w", err)
		}
	}
	if yes != nil && !*yes {
		return errors.New("writing requires --yes; back up the current action with the action command first")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	c := newClient(*timeout)
	defer func() {
		if closeErr := c.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close client: %w", closeErr))
		}
	}()
	if command == "scan" {
		devices, err := c.Scan(ctx)
		if err != nil {
			return err
		}
		for _, d := range devices {
			if _, err := fmt.Fprintf(out, "%s\t%s\tvendor=0x%04x product=0x%04x\n", d.ID, d.Product, d.VendorID, uint16(d.Product)); err != nil {
				return err
			}
		}
		return nil
	}
	var mapped bool
	err = c.WithKeyboard(ctx, client.DeviceID(*device), func(s client.Keyboard) error {
		switch command {
		case "action":
			var err error
			binding, readErr := s.Binding(ctx, keyID)
			actionCode, mapped, err = binding.Action, readErr == nil && binding.State != client.BindingDefault, readErr
			return err
		case "bind":
			return s.Bind(ctx, keyID, actionCode)
		case "restore":
			return s.Restore(ctx, keyID)
		default:
			return fmt.Errorf("unsupported command %q", command)
		}
	})
	if err != nil {
		return err
	}
	if command == "action" {
		if mapped {
			_, err = fmt.Fprintln(out, actionCode.String())
		} else {
			_, err = fmt.Fprintln(out, "no mapping override")
		}
	} else {
		_, err = fmt.Fprintf(out, "%s completed for device %s key %s\n", command, *device, keyID)
	}
	return err
}
