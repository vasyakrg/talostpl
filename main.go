package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"strconv"
	"path/filepath"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	image      string = "factory.talos.dev/metal-installer/956b9107edd250304169d2e7a765cdd4e0c31f9097036e2e113b042e6c01bb98:v1.10.4"
	k8sVersion string = "1.33.2"
	configDir  string = "config"
	version    = "1.0.0"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
)

type PatchConfig struct {
	Machine map[string]interface{} `yaml:"machine,omitempty"`
	Cluster map[string]interface{} `yaml:"cluster,omitempty"`
}

// checkRequiredTools проверяет наличие необходимых инструментов в системе
func checkRequiredTools() error {
	tools := map[string]string{
		"talosctl": "talosctl",
		"kubectl":  "kubectl",
	}

	missingTools := []string{}

	for toolName, binaryName := range tools {
		if _, err := exec.LookPath(binaryName); err != nil {
			missingTools = append(missingTools, toolName)
		}
	}

	if len(missingTools) > 0 {
		fmt.Printf("%s❌ Required tools not found:%s\n", colorRed, colorReset)
		for _, tool := range missingTools {
			fmt.Printf("   - %s\n", tool)
		}
		fmt.Println("\n📋 Installation instructions:")
		osType := runtime.GOOS
		switch osType {
		case "linux":
			fmt.Println("\n1. Install talosctl:")
			fmt.Println("   curl -sL https://talos.dev/install | sh")
			fmt.Println("   or")
			fmt.Println("   wget -O - https://talos.dev/install | sh")
			fmt.Println("\n2. Install kubectl:")
			fmt.Println("   curl -LO \"https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl\"")
			fmt.Println("   chmod +x kubectl")
			fmt.Println("   sudo mv kubectl /usr/local/bin/")
			fmt.Println("\n   For more details see:")
			fmt.Println("   https://kubernetes.io/docs/tasks/tools/install-kubectl/")
		case "darwin":
			fmt.Println("\n1. Install talosctl:")
			fmt.Println("   brew install siderolabs/tap/talosctl")
			fmt.Println("\n2. Install kubectl:")
			fmt.Println("   brew install kubectl")
			fmt.Println("\n   For more details see:")
			fmt.Println("   https://kubernetes.io/docs/tasks/tools/install-kubectl/")
		case "windows":
			fmt.Println("\nPlease refer to the official documentation for installation:")
			fmt.Println("   talosctl: https://www.talos.dev/v1.5/introduction/getting-started/#installing-talosctl")
			fmt.Println("   kubectl: https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/")
		default:
			fmt.Println("\nPlease refer to the official documentation for installation:")
			fmt.Println("   talosctl: https://www.talos.dev/v1.5/introduction/getting-started/#installing-talosctl")
			fmt.Println("   kubectl: https://kubernetes.io/docs/tasks/tools/install-kubectl/")
		}
		fmt.Println("\nAfter installation, restart the program.")
		return fmt.Errorf("required tools missing: %v", missingTools)
	}

	fmt.Printf("%s✅ All required tools found%s\n", colorGreen, colorReset)
	return nil
}

func askNumbered(prompt, def string) string {
	for {
		fmt.Print(prompt)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			if def != "" {
				return def
			} else {
				fmt.Printf("%sField is required.%s\n", colorRed, colorReset)
				continue
			}
		}
		return input
	}
}

func askYesNoNumbered(prompt, def string) bool {
	for {
		ans := strings.ToLower(askNumbered(prompt+" (y/n) ["+def+"]: ", def))
		if ans == "y" || ans == "yes" {
			return true
		} else if ans == "n" || ans == "no" {
			return false
		}
		fmt.Printf("%sInvalid input. Enter 'y'/'yes' or 'n'/'no'.%s\n", colorRed, colorReset)
	}
}

func mustAtoi(val string) int {
	i, err := strconv.Atoi(val)
	if err != nil {
		fmt.Printf("%sError: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	return i
}

func fileWriteYAML(path string, data interface{}) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("%sError creating file %s: %v%s\n", colorRed, path, err, colorReset)
		os.Exit(1)
	}
	defer f.Close()
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(data); err != nil {
		fmt.Printf("%sError writing YAML: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func clearDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	files, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, f := range files {
		err = os.RemoveAll(filepath.Join(dir, f))
		if err != nil {
			return err
		}
	}
	return nil
}

func generateCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Interactive Talos K8s config generator",
		Run: func(cmd *cobra.Command, args []string) {
			if configDir == "" {
				configDir = "config"
			}
			os.MkdirAll(configDir, 0o755)

			// Проверка содержимого директории
			entries, err := os.ReadDir(configDir)
			if err != nil {
				fmt.Printf("%sError reading config directory: %v%s\n", colorRed, err, colorReset)
				os.Exit(1)
			}
			if len(entries) > 0 {
				if !force {
					fmt.Printf("%sConfig directory '%s' is not empty. Use --force to overwrite or clean it.%s\n", colorRed, configDir, colorReset)
					os.Exit(1)
				} else {
					if !askYesNoNumbered(fmt.Sprintf("Config directory '%s' is not empty. Do you want to delete all its contents?", configDir), "n") {
						fmt.Printf("%sAborted by user. Directory not cleaned.%s\n", colorRed, colorReset)
						os.Exit(1)
					}
					err := clearDir(configDir)
					if err != nil {
						fmt.Printf("%sFailed to clean directory: %v%s\n", colorRed, err, colorReset)
						os.Exit(1)
					}
					fmt.Printf("%sDirectory '%s' cleaned.%s\n", colorGreen, configDir, colorReset)
				}
			}

			// Вопросы задаются последовательно, без нумерации
			clusterName := askNumbered("Enter cluster name [talos-demo]: ", "talos-demo")
			k8sVer := askNumbered("Enter Kubernetes version ["+k8sVersion+"]: ", k8sVersion)
			imageVal := askNumbered("Enter Talos installer image ["+image+"]: ", image)
			iface := askNumbered("Enter network interface name [ens18]: ", "ens18")
			cpCount := mustAtoi(askNumbered("Enter number of control planes (odd, max 7) [1]: ", "1"))
			workerCount := mustAtoi(askNumbered("Enter number of worker nodes (max 15, min 0) [3]: ", "3"))
			gateway := askNumbered("Enter default gateway: ", "")
			netmask := askNumbered("Enter network mask [24]: ", "24")
			dns1 := askNumbered("Enter first DNS server [8.8.8.8]: ", "8.8.8.8")
			dns2 := askNumbered("Enter second DNS server [8.8.4.4]: ", "8.8.4.4")
			ntp1 := askNumbered("Enter first NTP server [1.ru.pool.ntp.org]: ", "1.ru.pool.ntp.org")
			ntp2 := askNumbered("Enter second NTP server [2.ru.pool.ntp.org]: ", "2.ru.pool.ntp.org")
			ntp3 := askNumbered("Enter third NTP server [3.ru.pool.ntp.org]: ", "3.ru.pool.ntp.org")
			useVIP := false
			vipIP := ""
			if cpCount > 1 {
				useVIP = askYesNoNumbered("Do you need a VIP address?", "y")
				if useVIP {
					vipIP = askNumbered("Enter VIP address: ", "")
				}
			}
			useExtBalancer := askYesNoNumbered("Do you need an external load balancer?", "n")
			extBalancerIP := ""
			if useExtBalancer {
				extBalancerIP = askNumbered("Enter external load balancer IP: ", "")
			}
			disk := askNumbered("Enter disk for base OS installation [/dev/sda]: ", "/dev/sda")
			useDRBD := askYesNoNumbered("Enable drbd support?", "y")
			useZFS := askYesNoNumbered("Enable zfs support?", "n")
			useSPL := askYesNoNumbered("Enable spl support?", "n")
			useVFIOPCI := askYesNoNumbered("Enable vfio_pci support?", "n")
			useVFIOIOMMU := askYesNoNumbered("Enable vfio_iommu_type1 support?", "n")
			useOVS := askYesNoNumbered("Enable openvswitch support?", "n")
			useMirrors := askYesNoNumbered("Use timeweb.cloud and gcr.io mirrors for docker.io?", "y")
			useMaxPods := askYesNoNumbered("Set maxPods: 512 for kubelet? (default is 110 per node)", "n")

			// Генерация patch.yaml
			patch := PatchConfig{
				Machine: map[string]interface{}{
					"network": map[string]interface{}{
						"nameservers": []string{dns1, dns2},
					},
					"install": map[string]interface{}{
						"disk":  disk,
						"image": imageVal,
					},
					"time": map[string]interface{}{
						"servers": []string{ntp1, ntp2, ntp3},
					},
				},
				Cluster: map[string]interface{}{},
			}
			if useMirrors {
				patch.Machine["registries"] = map[string]interface{}{
					"mirrors": map[string]interface{}{
						"docker.io": map[string]interface{}{
							"endpoints": []string{"https://dockerhub.timeweb.cloud", "https://mirror.gcr.io"},
						},
					},
				}
			}
			if useExtBalancer && extBalancerIP != "" {
				patch.Machine["certSANs"] = []string{extBalancerIP}
			}
			if useDRBD && workerCount == 0 {
				mods := []map[string]interface{}{
					{"name": "drbd", "parameters": []string{"usermode_helper=disabled"}},
					{"name": "drbd_transport_tcp"},
					{"name": "dm-thin-pool"},
				}
				if useZFS {
					mods = append(mods, map[string]interface{}{ "name": "zfs" })
				}
				if useSPL {
					mods = append(mods, map[string]interface{}{ "name": "spl" })
				}
				if useVFIOPCI {
					mods = append(mods, map[string]interface{}{ "name": "vfio_pci" })
				}
				if useVFIOIOMMU {
					mods = append(mods, map[string]interface{}{ "name": "vfio_iommu_type1" })
				}
				if useOVS {
					mods = append(mods, map[string]interface{}{ "name": "openvswitch" })
				}
				patch.Machine["kernel"] = map[string]interface{}{ "modules": mods }
			}
			if workerCount == 0 {
				patch.Cluster["allowSchedulingOnControlPlanes"] = true
			}
			patch.Cluster["network"] = map[string]interface{}{
				"cni": map[string]interface{}{ "name": "none" },
			}
			patch.Cluster["proxy"] = map[string]interface{}{ "disabled": true }
			fileWriteYAML(filepath.Join(configDir, "patch.yaml"), patch)
			fmt.Printf("%sCreated patch.yaml%s\n", colorGreen, colorReset)
			fmt.Println("--------------------------------")

			// --- Генерация патчей для control-plane ---
			cpIPs := make([]string, 0, cpCount)
			usedIPs := map[string]struct{}{gateway: {}}
			for i := 1; i <= cpCount; i++ {
				var cpIP string
				for {
					cpIP = askNumbered(fmt.Sprintf("Enter IP address for control plane %d: ", i), "")
					if cpIP == "" {
						fmt.Printf("%sIP address cannot be empty.%s\n", colorRed, colorReset)
						continue
					}
					if _, ok := usedIPs[cpIP]; ok {
						fmt.Printf("%sThis IP address is already used. Enter a unique address.%s\n", colorRed, colorReset)
						continue
					}
					usedIPs[cpIP] = struct{}{}
					break
				}
				cpIPs = append(cpIPs, cpIP)
				filename := filepath.Join(configDir, fmt.Sprintf("cp%d.patch", i))
				patch := map[string]interface{}{
					"machine": map[string]interface{}{
						"network": map[string]interface{}{
							"hostname": fmt.Sprintf("cp-%d", i),
							"interfaces": []map[string]interface{}{
								{
									"interface": iface,
									"dhcp": false,
									"addresses": []string{fmt.Sprintf("%s/%s", cpIP, netmask)},
									"routes": []map[string]interface{}{
										{"network": "0.0.0.0/0", "gateway": gateway},
									},
								},
							},
						},
					},
				}
				if useVIP && vipIP != "" {
					patch["machine"].(map[string]interface{})["network"].(map[string]interface{})["interfaces"].([]map[string]interface{})[0]["vip"] = map[string]interface{}{"ip": vipIP}
				}
				if useMaxPods {
					patch["machine"].(map[string]interface{})["kubelet"] = map[string]interface{}{
						"extraConfig": map[string]interface{}{"maxPods": 512},
					}
				}
				fileWriteYAML(filename, patch)
				fmt.Printf("%sCreated file: %s%s\n", colorGreen, filename, colorReset)
			}
			fmt.Println("--------------------------------")
			// --- Генерация патчей для worker-nodes ---
			workerIPs := make([]string, 0, workerCount)
			if workerCount > 0 {
				for i := 1; i <= workerCount; i++ {
					var workerIP string
					for {
						workerIP = askNumbered(fmt.Sprintf("Enter IP address for worker %d: ", i), "")
						if workerIP == "" {
							fmt.Printf("%sIP address cannot be empty.%s\n", colorRed, colorReset)
							continue
						}
						if _, ok := usedIPs[workerIP]; ok {
							fmt.Printf("%sThis IP address is already used. Enter a unique address.%s\n", colorRed, colorReset)
							continue
						}
						usedIPs[workerIP] = struct{}{}
						break
					}
					workerIPs = append(workerIPs, workerIP)
					filename := filepath.Join(configDir, fmt.Sprintf("worker%d.patch", i))
					patch := map[string]interface{}{
						"machine": map[string]interface{}{
							"network": map[string]interface{}{
								"hostname": fmt.Sprintf("worker-%d", i),
								"interfaces": []map[string]interface{}{
									{
										"deviceSelector": map[string]interface{}{"physical": true},
										"dhcp": false,
										"addresses": []string{fmt.Sprintf("%s/%s", workerIP, netmask)},
										"routes": []map[string]interface{}{
											{"network": "0.0.0.0/0", "gateway": gateway},
										},
									},
								},
							},
						},
					}
					if useDRBD {
						mods := []map[string]interface{}{
							{"name": "drbd", "parameters": []string{"usermode_helper=disabled"}},
							{"name": "drbd_transport_tcp"},
							{"name": "dm-thin-pool"},
						}
						if useZFS {
							mods = append(mods, map[string]interface{}{ "name": "zfs" })
						}
						if useSPL {
							mods = append(mods, map[string]interface{}{ "name": "spl" })
						}
						if useVFIOPCI {
							mods = append(mods, map[string]interface{}{ "name": "vfio_pci" })
						}
						if useVFIOIOMMU {
							mods = append(mods, map[string]interface{}{ "name": "vfio_iommu_type1" })
						}
						if useOVS {
							mods = append(mods, map[string]interface{}{ "name": "openvswitch" })
						}
						patch["machine"].(map[string]interface{})["kernel"] = map[string]interface{}{ "modules": mods }
					}
					fileWriteYAML(filename, patch)
					fmt.Printf("%sCreated file: %s%s\n", colorGreen, filename, colorReset)
				}
			}
			fmt.Println("--------------------------------")
			// --- Генерация секретов ---
			secretsFile := filepath.Join(configDir, "secrets.yaml")
			if err := runCmd("talosctl", "gen", "secrets", "-o", secretsFile); err != nil {
				fmt.Printf("%sError generating secrets: %v%s\n", colorRed, err, colorReset)
				os.Exit(1)
			}
			fmt.Printf("%sCreated secrets.yaml%s\n", colorGreen, colorReset)
			fmt.Println("--------------------------------")
			// --- Генерация основной конфигурации ---
			endpointIP := vipIP
			if endpointIP == "" && len(cpIPs) > 0 {
				endpointIP = cpIPs[0]
			}
			endpointIP = strings.Split(endpointIP, "/")[0]
			if err := os.Chdir(configDir); err != nil {
				fmt.Printf("%sError changing directory: %v%s\n", colorRed, err, colorReset)
				os.Exit(1)
			}
			if err := runCmd("talosctl", "gen", "config", "--kubernetes-version", k8sVer, "--with-secrets", "secrets.yaml", clusterName, fmt.Sprintf("https://%s:6443", endpointIP), "--config-patch", "@patch.yaml"); err != nil {
				fmt.Printf("%sError generating config: %v%s\n", colorRed, err, colorReset)
				os.Exit(1)
			}
			fmt.Println("--------------------------------")
			// --- Применение патчей к control-plane ---
			for i := 1; i <= cpCount; i++ {
				if err := runCmd("talosctl", "machineconfig", "patch", "controlplane.yaml", "--patch", fmt.Sprintf("@cp%d.patch", i), "--output", fmt.Sprintf("cp%d.yaml", i)); err != nil {
					fmt.Printf("%sError patching cp%d: %v%s\n", colorRed, i, err, colorReset)
					os.Exit(1)
				}
				fmt.Printf("%sCreated file: cp%d.yaml%s\n", colorGreen, i, colorReset)
			}
			fmt.Println("--------------------------------")
			// --- Применение патчей к worker-nodes ---
			if workerCount > 0 {
				for i := 1; i <= workerCount; i++ {
					if err := runCmd("talosctl", "machineconfig", "patch", "worker.yaml", "--patch", fmt.Sprintf("@worker%d.patch", i), "--output", fmt.Sprintf("worker%d.yaml", i)); err != nil {
						fmt.Printf("%sError patching worker%d: %v%s\n", colorRed, i, err, colorReset)
						os.Exit(1)
					}
					fmt.Printf("%sCreated file: worker%d.yaml%s\n", colorGreen, i, colorReset)
				}
			}
			fmt.Println("--------------------------------")
			// --- Обновление talosconfig с endpoints ---
			endpoints := append([]string{}, cpIPs...)
			if useVIP && vipIP != "" {
				endpoints = append(endpoints, vipIP)
			}
			if useExtBalancer && extBalancerIP != "" {
				endpoints = append(endpoints, extBalancerIP)
			}
			endpointsStr := strings.Join(endpoints, ",")
			talosconfig := "talosconfig"
			if _, err := os.Stat(talosconfig); err == nil {
				tmpConfig := "talosconfig.tmp"
				in, _ := os.Open(talosconfig)
				out, _ := os.Create(tmpConfig)
				scanner := bufio.NewScanner(in)
				for scanner.Scan() {
					line := scanner.Text()
					if strings.Contains(line, "endpoints: []") {
						line = strings.Replace(line, "endpoints: []", fmt.Sprintf("endpoints: [%s]", endpointsStr), 1)
					}
					out.WriteString(line + "\n")
				}
				in.Close()
				out.Close()
				os.Rename(tmpConfig, talosconfig)
				fmt.Printf("%sUpdated talosconfig with endpoints: [%s]%s\n", colorGreen, endpointsStr, colorReset)
			} else {
				fmt.Println("File talosconfig not found")
			}
			fmt.Println("--------------------------------")
			// --- Применение конфигов и bootstrap ---
			os.Chdir("..")
			firstCP := cpIPs[0]
			firstCPClean := strings.Split(firstCP, "/")[0]
			if askYesNoNumbered("Do you want to start cluster initialization?", "y") {
				if err := runCmd("talosctl", "apply-config", "--insecure", "-n", firstCPClean, "--file", filepath.Join(configDir, "cp1.yaml")); err != nil {
					fmt.Printf("%sError apply-config: %v%s\n", colorRed, err, colorReset)
					os.Exit(1)
				}
			} else {
				fmt.Println("--------------------------------")
				fmt.Println("Generate completed. Cluster initialization cancelled by user.")
				fmt.Println("--------------------------------")
				return
			}
			fmt.Println("--------------------------------")
			askNumbered("Perform bootstrap on first control plane ("+firstCPClean+")? [Enter to continue]", "")
			if err := runCmd("talosctl", "bootstrap", "--nodes", firstCPClean, "--endpoints", firstCPClean, "--talosconfig="+filepath.Join(configDir, "talosconfig")); err != nil {
				fmt.Printf("%sError bootstrap: %v%s\n", colorRed, err, colorReset)
				os.Exit(1)
			}
			fmt.Println("--------------------------------")
			if cpCount > 1 {
				for i := 2; i <= cpCount; i++ {
					cpClean := strings.Split(cpIPs[i-1], "/")[0]
					askNumbered(fmt.Sprintf("Apply config to control plane %d (%s)? [Enter to continue]", i, cpClean), "")
					if err := runCmd("talosctl", "apply-config", "--insecure", "-n", cpClean, "--file", filepath.Join(configDir, fmt.Sprintf("cp%d.yaml", i))); err != nil {
						fmt.Printf("%sError apply-config cp%d: %v%s\n", colorRed, i, err, colorReset)
						os.Exit(1)
					}
				}
			}
			if workerCount > 0 {
				for i := 1; i <= workerCount; i++ {
					askNumbered(fmt.Sprintf("Apply config to worker-%d (%s)? [Enter to continue]", i, workerIPs[i-1]), "")
					if err := runCmd("talosctl", "apply-config", "--insecure", "-n", workerIPs[i-1], "--file", filepath.Join(configDir, fmt.Sprintf("worker%d.yaml", i))); err != nil {
						fmt.Printf("%sError apply-config worker%d: %v%s\n", colorRed, i, err, colorReset)
						os.Exit(1)
					}
				}
			}
			fmt.Println("--------------------------------")
			// --- Выгрузка kubeconfig ---
			kubeconfigEndpoint := vipIP
			if kubeconfigEndpoint == "" {
				kubeconfigEndpoint = firstCPClean
			}
			kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", clusterName+".yaml")
			if err := runCmd("talosctl", "kubeconfig", kubeconfigPath, "--nodes", kubeconfigEndpoint, "--endpoints", kubeconfigEndpoint, "--talosconfig", filepath.Join(configDir, "talosconfig")); err != nil {
				fmt.Printf("%sError exporting kubeconfig: %v%s\n", colorRed, err, colorReset)
				os.Exit(1)
			}
			fmt.Println("--------------------------------")
			fmt.Println("Script completed")
			fmt.Println("--------------------------------")
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Force clean config directory if not empty")
	return cmd
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "talostpl",
		Short: "Interactive Talos K8s config generator",
		Long:  `Interactive Talos K8s config generator. Utility for generating configs and running Talos K8s bootstrap.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	rootCmd.PersistentFlags().StringVar(&image, "image", image, "Talos installer image")
	rootCmd.PersistentFlags().StringVar(&k8sVersion, "k8s-version", k8sVersion, "Kubernetes version")
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", configDir, "Directory for configs")
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("talostpl version {{.Version}}\n")
	rootCmd.AddCommand(generateCmd())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
