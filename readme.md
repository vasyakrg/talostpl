# talostpl

## Features

- **Interactive wizard**: Step-by-step prompts for all cluster parameters.
- **Non-interactive mode**: All answers and IPs are taken from a YAML file, no prompts.
- **Automatic patch generation**: Generates all required Talos patches and config files.
- **Integration with talosctl**: Runs `talosctl` to generate secrets, configs, apply patches, and bootstrap the cluster.
- **Kubeconfig export**: Automatically exports kubeconfig to your `$HOME/.kube` directory.
- **Cluster initialization control**: You can skip cluster initialization (apply-config/bootstrap) at the final step if needed (interactive).
- **Node addition**: Add new control plane or worker nodes to existing cluster configuration.

## Requirements

- Installed utilities: `talosctl`, `kubectl` (must be in `$PATH`)
- Linux или macOS (darwin)

## Installation

Go to the [Releases page](https://github.com/vasyakrg/talostpl/releases) and download the binary for your OS and architecture.

For MacOS (Intel/Apple Silicon) users and BREW package manager

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
# Example for macOS amd64
wget -O talostpl "https://github.com/vasyakrg/talostpl/releases/download/$(curl -s https://api.github.com/repos/vasyakrg/talostpl/releases/latest | grep '"tag_name":' | head -1 | cut -d '"' -f4)/talostpl-darwin-amd64"
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

### Add new nodes to existing cluster

Add new control plane node:

```sh
./talostpl add --cp=2 --address=192.168.1.22
```

Add new worker node:

```sh
./talostpl add --worker=4 --address=192.168.1.24
```

Add new node with automatic configuration application:

```sh
./talostpl add --cp=3 --address=192.168.1.23 --auto-apply
```

Add new node using custom configuration directory:

```sh
./talostpl add --worker=2 --address=192.168.1.25 --config=/path/to/custom/config
```

**Requirements for adding nodes:**

- Existing configuration directory with `controlplane.yaml` or `worker.yaml`
- Existing `talosconfig` file
- Existing `cp1.patch` or `worker1.patch` as base template
- Target node number must not already exist (e.g., `cp2.patch` should not exist)

## Command-line flags

### Global flags (for all commands)

- `--image` — Talos installer image (default provided)
- `--k8s-version` — Kubernetes version (default provided)
- `--config-dir` — Directory for generated files (default: config)

### Generate command flags

- `--force` — Clean config directory if not empty (in interactive mode asks for confirmation, in from-file mode always cleans)
- `--from-file` — Path to YAML file with answers and IP addresses for non-interactive mode

### Add command flags

- `--cp` — Control plane node number (e.g., `--cp=2` for cp2.patch/cp2.yaml)
- `--worker` — Worker node number (e.g., `--worker=4` for worker4.patch/worker4.yaml)
- `--address` — IP address for the new node (required)
- `--auto-apply` — Automatically apply configuration to the node after generation (optional)
- `--config` — Path to configuration directory (overrides global --config-dir flag)

## Notes

- All generated files will be placed in the directory specified by `--config-dir` (default: `config`).
- In interactive mode, a file `cluster.yaml` with all cluster parameters and answers is automatically generated in the config directory. This file is useful for documentation or for future non-interactive runs.
- The `cluster.yaml` file is not created in non-interactive (`--from-file`) mode.
- When adding nodes, the tool uses existing base patch files (`cp1.patch` or `worker1.patch`) as templates and only changes the IP address and hostname.
- The `add` command automatically detects network mask from the base patch file and applies it to the new node.
- The `--config` flag allows specifying a custom configuration directory path, overriding the global `--config-dir` setting.
- If `--auto-apply` is used and user declines or if application fails, the tool displays the manual command to apply the configuration.
- Makefile is only used for building the binary. All other functionality is handled by the Go application itself.
- Make sure all external dependencies (`talosctl`, `kubectl`) are installed and available in your system.

### Author

[Yegorov Vassiliy](https://egorovanet.ru)

(C) 2025
