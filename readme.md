# talostpl

## Requirements

- Go 1.20+ (latest stable recommended)
- Installed utilities: `talosctl`, `kubectl` (must be in `$PATH`)
- Linux or macOS

## Build

### With Make (recommended)

1. Create a `Makefile` with the following content:

    ```makefile
    build:
    	go build -o talostpl main.go
    ```

2. Run:

    ```sh
    make build
    ```

   This will produce the `talostpl` binary.

### Without Make

```sh
go build -o talostpl main.go
```

## Run

```sh
./talostpl generate
```

or with flags (all flags are optional, defaults are set):

```sh
./talostpl generate \
  --image="factory.talos.dev/metal-installer/..." \
  --k8s-version="1.33.2" \
  --config-dir="config"
```

## Features

- **Interactive wizard**: Step-by-step prompts for all cluster parameters.
- **Automatic patch generation**: Generates all required Talos patches and config files.
- **Integration with talosctl**: Runs `talosctl` to generate secrets, configs, apply patches, and bootstrap the cluster.
- **Kubeconfig export**: Automatically exports kubeconfig to your `$HOME/.kube` directory.
- **Cluster initialization control**: You can skip cluster initialization (apply-config/bootstrap) at the final step if needed.
- **All prompts and messages are in English.**

## Notes

- All generated files will be placed in the directory specified by `--config-dir` (default: `config`).
- Makefile is only used for building the binary. All other functionality is handled by the Go application itself.
- Make sure all external dependencies (`talosctl`, `kubectl`) are installed and available in your system.
