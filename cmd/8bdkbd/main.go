package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/ColinEge/8bd-ultimate-api/client"
)

const usage = `Usage: 8bdkbd <command> [arguments] [options]

Commands:
  scan                                 List connected keyboards
  info DEVICE                          Show device metadata
  firmware-version DEVICE              Query firmware (currently unsupported)
  keys DEVICE                          List physical key names
  actions DEVICE                       List available action names
  binding DEVICE KEY                   Read a key's state and effective action
  bindings DEVICE                      List stored binding overrides
  bind DEVICE KEY ACTION               Bind a key to an action
  disable DEVICE KEY                   Disable a key
  reset-to-default DEVICE KEY          Bind the factory-default action
  apply-bindings DEVICE KEY=ACTION...   Apply an ordered batch of changes
  get-profile DEVICE                   Read the device profile title
  set-profile DEVICE TITLE             Set the device profile title
  delete-profile DEVICE                Clear the device profile title

Use device IDs from scan. Key and Action names are case-insensitive.
Chords use one modifier: LeftCtrl+C. Quote profile titles containing spaces.

Options may appear before or after arguments:
  --timeout DURATION   Command-exchange timeout (default 10s)
  --write              Required for every write command
  --help               Show command help

Back up bindings before writing. Reset-to-default uses the factory default, not a
previous custom binding. Batches are not atomic and have no rollback.
`

type commandSpec struct {
	arguments string
	minArgs   int
	maxArgs   int
	writes    bool
}

var commands = map[string]commandSpec{
	"scan":             {"", 0, 0, false},
	"info":             {"DEVICE", 1, 1, false},
	"firmware-version": {"DEVICE", 1, 1, false},
	"keys":             {"DEVICE", 1, 1, false},
	"actions":          {"DEVICE", 1, 1, false},
	"binding":          {"DEVICE KEY", 2, 2, false},
	"bindings":         {"DEVICE", 1, 1, false},
	"bind":             {"DEVICE KEY ACTION", 3, 3, true},
	"disable":          {"DEVICE KEY", 2, 2, true},
	"reset-to-default": {"DEVICE KEY", 2, 2, true},
	"apply-bindings":   {"DEVICE KEY=ACTION...", 2, -1, true},
	"get-profile":      {"DEVICE", 1, 1, false},
	"set-profile":      {"DEVICE TITLE", 2, 2, true},
	"delete-profile":   {"DEVICE", 1, 1, true},
}

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
	spec, ok := commands[command]
	if !ok {
		return fmt.Errorf("unknown command %q; use --help", command)
	}
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(errOut)
	timeout := flags.Duration("timeout", 10*time.Second, "command-exchange timeout (must be positive)")
	var write *bool
	if spec.writes {
		write = flags.Bool("write", false, "confirm writing to the device")
	}
	flags.Usage = func() {
		fmt.Fprintf(errOut, "Usage: 8bdkbd %s %s [options]\n", command, spec.arguments)
		fmt.Fprintln(errOut, "Use device IDs from scan and names from keys/actions DEVICE.")
		if command == "bind" || command == "apply-bindings" {
			fmt.Fprintln(errOut, "Actions use API names (B, VolumeUp) or one-modifier chords (LeftCtrl+C).")
		}
		flags.PrintDefaults()
	}
	positional, err := parseFlags(flags, args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if len(positional) < spec.minArgs || (spec.maxArgs >= 0 && len(positional) > spec.maxArgs) {
		return fmt.Errorf("usage: 8bdkbd %s %s; use %s --help", command, spec.arguments, command)
	}
	if *timeout <= 0 {
		return errors.New("--timeout must be positive")
	}
	var device client.DeviceID
	if command != "scan" {
		if strings.TrimSpace(positional[0]) == "" {
			return errors.New("DEVICE is required; run scan to find a device ID")
		}
		device = client.DeviceID(positional[0])
	}
	var key client.KeyID
	var action client.Action
	var changes []client.BindingChange
	switch command {
	case "binding", "bind", "disable", "reset-to-default":
		key, err = parseKey(positional[1])
		if err != nil {
			return err
		}
		if command == "bind" {
			action, err = parseAction(positional[2])
			if err != nil {
				return err
			}
		}
	case "apply-bindings":
		for _, text := range positional[1:] {
			keyText, actionText, ok := strings.Cut(text, "=")
			if !ok {
				return fmt.Errorf("invalid binding %q: expected KEY=ACTION", text)
			}
			key, err := parseKey(keyText)
			if err != nil {
				return err
			}
			action, err := parseAction(actionText)
			if err != nil {
				return err
			}
			changes = append(changes, client.BindingChange{Key: key, Action: action})
		}
		if err := client.ValidateBindings(changes); err != nil {
			return err
		}
	case "set-profile":
		if err := client.ValidateProfileName(positional[1]); err != nil {
			return err
		}
	}
	if write != nil && !*write {
		return errors.New("writing requires --write; back up the current binding with the binding command first")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	c := newClient(*timeout)
	closed := false
	defer func() {
		if !closed {
			if closeErr := c.Close(); closeErr != nil {
				err = errors.Join(err, fmt.Errorf("close client: %w", closeErr))
			}
		}
	}()
	var result strings.Builder
	if command == "scan" {
		devices, err := c.Scan(ctx)
		if err != nil {
			return err
		}
		for _, d := range devices {
			writeDevice(&result, d)
		}
	} else {
		err = c.WithKeyboard(ctx, device, func(k client.Keyboard) error {
			switch command {
			case "info":
				info, err := k.Info(ctx)
				if err != nil {
					return err
				}
				writeDevice(&result, info.Device)
			case "firmware-version":
				version, err := k.FirmwareVersion(ctx)
				if err != nil {
					return err
				}
				fmt.Fprintln(&result, version)
			case "keys":
				for _, key := range k.Keys() {
					fmt.Fprintf(&result, "%s\t%s\n", actionName(key.Default), key.Label)
				}
			case "actions":
				for _, action := range k.Actions() {
					fmt.Fprintf(&result, "%s\t%s\t%s\n", actionName(action.Action), action.Label, action.Category)
				}
			case "binding":
				binding, err := k.Binding(ctx, key)
				if err != nil {
					return err
				}
				writeBinding(&result, binding)
			case "bindings":
				bindings, err := k.Bindings(ctx)
				if err != nil {
					return err
				}
				for _, binding := range bindings {
					writeBinding(&result, binding)
				}
			case "bind":
				return k.Bind(ctx, key, action)
			case "disable":
				return k.Disable(ctx, key)
			case "reset-to-default":
				return k.ResetToDefault(ctx, key)
			case "apply-bindings":
				return k.ApplyBindings(ctx, changes)
			case "get-profile":
				title, err := k.GetProfile(ctx)
				if err != nil {
					return err
				}
				fmt.Fprintln(&result, title)
			case "set-profile":
				return k.SetProfile(ctx, positional[1])
			case "delete-profile":
				return k.DeleteProfile(ctx)
			default:
				return fmt.Errorf("unsupported command %q", command)
			}
			return nil
		})
		if err != nil {
			return err
		}
		if spec.writes {
			fmt.Fprintf(&result, "%s completed for device %s\n", command, device)
		}
	}
	closed = true
	if err := c.Close(); err != nil {
		return fmt.Errorf("close client: %w", err)
	}
	_, err = io.WriteString(out, result.String())
	return err
}

// The standard flag parser stops at the first positional argument.
func parseFlags(flags *flag.FlagSet, args []string) ([]string, error) {
	var positional, options []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if len(arg) < 2 || arg[0] != '-' {
			positional = append(positional, arg)
			continue
		}
		options = append(options, arg)
		name, _, hasValue := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		if f := flags.Lookup(name); f != nil && !hasValue {
			boolean, isBool := f.Value.(interface{ IsBoolFlag() bool })
			if !isBool || !boolean.IsBoolFlag() {
				if i+1 == len(args) {
					return nil, fmt.Errorf("flag needs an argument: %s", arg)
				}
				i++
				options = append(options, args[i])
			}
		}
	}
	return positional, flags.Parse(options)
}

func writeDevice(out *strings.Builder, d client.DeviceInfo) {
	fmt.Fprintf(out, "%s\t%s\tvendor=0x%04x product=0x%04x\n", d.ID, d.Product, d.VendorID, uint16(d.Product))
}

func writeBinding(out *strings.Builder, binding client.Binding) {
	fmt.Fprintf(out, "%s\t%s\t%s\n", keyName(binding.Key), binding.State, actionName(binding.Effective))
}
