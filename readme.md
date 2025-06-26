# talostpl

## Features

- **Interactive wizard**: Step-by-step prompts for all cluster parameters.
- **Non-interactive mode**: All answers and IPs are taken from a YAML file, no prompts.
- **Automatic patch generation**: Generates all required Talos patches and config files.
- **Integration with talosctl**: Runs `talosctl` to generate secrets, configs, apply patches, and bootstrap the cluster.
- **Kubeconfig export**: Automatically exports kubeconfig to your `$HOME/.kube` directory.
- **Cluster initialization control**: You can skip cluster initialization (apply-config/bootstrap) at the final step if needed (interactive) or via `initCluster` (from-file).

## Requirements

- Installed utilities: `talosctl`, `kubectl` (must be in `$PATH`)
- Linux, macOS or Windows

## Installation

Go to the [Releases page](https://github.com/vasyakrg/talostpl/releases) and download the binary for your OS and architecture.

Or use curl/wget (replace `<version>` and `<os-arch>` with the actual values):

For MacOS arm64 (Apple Silicon) users and BREW package manager

```bash
brew tap vasyakrg/talostpl
brew install talostpl
```

or

```sh
# Example for macOS arm64
wget -O talostpl "https://github.com/vasyakrg/talostpl/releases/download/$(curl -s https://api.github.com/repos/vasyakrg/talostpl/releases/latest | grep '"tag_name":' | head -1 | cut -d '"' -f4)/talostpl-darwin-arm64"
chmod +x talostpl
sudo mv talostpl /usr/local/bin/
```

```sh
# Example for Linux amd64
curl -L -o talostpl "https://github.com/vasyakrg/talostpl/releases/download/$(curl -s https://api.github.com/repos/vasyakrg/talostpl/releases/latest | grep '"tag_name":' | head -1 | cut -d '"' -f4)/talostpl-linux-amd64"
chmod +x talostpl
sudo mv talostpl /usr/local/bin/
```

```sh
# Example for Linux arm64
curl -L -o talostpl "https://github.com/vasyakrg/talostpl/releases/download/$(curl -s https://api.github.com/repos/vasyakrg/talostpl/releases/latest | grep '"tag_name":' | head -1 | cut -d '"' -f4)/talostpl-linux-arm64"
chmod +x talostpl
sudo mv talostpl /usr/local/bin/
```

After installation, you can run `talostpl -v` from anywhere in your terminal.

## Run

### Interactive mode

```sh
./talostpl generate
```

### With flags (all flags are optional, defaults are set)

```sh
./talostpl generate [--force] \
  --image="factory.talos.dev/metal-installer/..." \
  --k8s-version="1.33.2" \
  --config-dir="config"
```

### Non-interactive mode (from file)

```sh
./talostpl generate --from-file=example-cluster.yaml [--force] [--config-dir=dir]
```

- In this mode, all parameters are taken from the YAML file, no questions are asked.
- If `--force` is specified, the config directory is always cleaned without confirmation.
- If `initCluster: true` in the file, the cluster will be initialized automatically; if `false`, only config files will be generated (no apply/bootstrap).

## Command-line flags

- `--image` — Talos installer image (default provided)
- `--k8s-version` — Kubernetes version (default provided)
- `--config-dir` — Directory for generated files (default: config)
- `--force` — Clean config directory if not empty (in interactive mode asks for confirmation, in from-file mode always cleans)
- `--from-file` — Path to YAML file with answers and IP addresses for non-interactive mode

## Notes

- All generated files will be placed in the directory specified by `--config-dir` (default: `config`).
- Makefile is only used for building the binary. All other functionality is handled by the Go application itself.
- Make sure all external dependencies (`talosctl`, `kubectl`) are installed and available in your system.

### Author

[Yegorov Vassiliy](https://egorovanet.ru)

(C)2025
