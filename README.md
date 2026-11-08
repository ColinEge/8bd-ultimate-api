# 8BitDo keyboard API

A Go library and command-line tool for discovering supported 8BitDo keyboards,
reading key bindings, and changing them. The API is pre-release and may change.

## Usage

The module is `github.com/ColinEge/8bd-ultimate-api`. Create a client, scan for
a device, then open it by the ID returned from the scan:

```go
import (
    "context"
    "errors"

    "github.com/ColinEge/8bd-ultimate-api/client"
)

func rebind(ctx context.Context) (err error) {
    c := client.New()
    defer func() { err = errors.Join(err, c.Close()) }()

    devices, err := c.Scan(ctx)
    if err != nil {
        return err
    }
    if len(devices) == 0 {
        return errors.New("no keyboards found")
    }

    return c.WithKeyboard(ctx, devices[0].ID, func(k client.Keyboard) error {
        return k.Bind(ctx, client.KeyA, client.ActionB())
    })
}
```

Import `github.com/ColinEge/8bd-ultimate-api/client`. `WithKeyboard` manages
the connection for the callback. Use `Binding` to inspect a key, `Restore` to
set its factory-default action, and `ApplyBindings` to apply multiple changes.
Batch writes are not atomic and are not automatically rolled back; check the
returned error before deciding whether to retry.

Local profile capture, application, and persistence are provided by the
`profiles` package. `profiles.Capture` reads a profile from an open keyboard;
`profiles.Apply` writes a profile to one. The profile store saves profiles
locally and does not access USB.

### CLI

Run from this module directory, or install the tool with
`go install ./cmd/8bdkbd` or `go install github.com/ColinEge/8bd-ultimate-api/cmd/8bdkbd@latest`:

```sh
8bdkbd scan
8bdkbd binding 001-003 A
8bdkbd bind 001-003 A B --write
8bdkbd apply-bindings 001-003 A=B C=LeftCtrl+C --write
8bdkbd --help
```

Commands mirror the `Keyboard` API: `info`, `firmware-version`, `keys`,
`actions`, `binding`, `bindings`, `bind`, `disable`, `restore`, `apply-bindings`,
`get-profile`, `set-profile`, and `delete-profile`. Connections are opened and
closed automatically; firmware querying currently returns an unsupported error.

Use the device ID reported by `scan`; it can change after reconnecting. Keys
and actions use case-insensitive API names without the `Key`/`Action` prefix,
such as `Enter`, `Numpad1`, and `VolumeUp`. List them with `keys DEVICE` and
`actions DEVICE`. Chords take one modifier, such as `LeftCtrl+C`. Quote profile
titles containing spaces: `set-profile DEVICE "Gaming profile" --write`.

All writes require `--write`. Back up the current binding with `binding` and use
a nonessential key. `restore` sets the factory default, not a previous custom
binding; batches have no rollback. Options can precede or follow arguments;
`--timeout 10s` is the default.

## Testing

Run the full hardware-free test suite from this directory. The explicit
environment-variable resets ensure opt-in hardware tests remain disabled:

```sh
env -u TEST_8BITDO_READONLY -u TEST_8BITDO_ACCEPTANCE \
  -u TEST_8BITDO_ACCEPTANCE_WRITE -u TEST_8BITDO_PLANNED go test ./...
```

The API behavior and feature scenarios are tested with scripted devices and do
not require a physical keyboard. To run the strict supported feature scenarios
or the race-enabled tests:

```sh
go test ./test -run '^TestFeatures$' -count=1 -v
go test -race ./test ./test/fixtures/keyboard -count=1
```

Physical-device tests are separate opt-ins. Read-only tests use
`TEST_8BITDO_READONLY=1`; feature acceptance tests require an explicit device
and key with `TEST_8BITDO_ACCEPTANCE=1`. Mutating acceptance tests use the separate
`TEST_8BITDO_ACCEPTANCE_WRITE=1` opt-in. Use a spare key, back up its original
binding, and do not run write tests unless you can recover the device if
cleanup is interrupted.

## Contributing

- Keep changes focused and add or update tests for changed behavior.
- Run `gofmt` on changed Go files and the hardware-free test suite above.
- For API or device-behavior changes, update the relevant scripted tests and
  feature scenarios; do not make hardware access a default test requirement.
- Update this README when usage, testing, or contribution workflows change.
