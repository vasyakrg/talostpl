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
	"net/http"
	"encoding/json"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	image      string = "factory.talos.dev/metal-installer/956b9107edd250304169d2e7a765cdd4e0c31f9097036e2e113b042e6c01bb98:v1.10.4"
	k8sVersion string = "1.33.2"
	configDir  string = "config"
	version    = "v1.0.8"
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

type Answers struct {
	ClusterName    string
	K8sVersion     string
	Image          string
	Iface          string
	CPCount        int
	WorkerCount    int
	Gateway        string
	Netmask        string
	DNS1           string
	DNS2           string
	NTP1           string
	NTP2           string
	NTP3           string
	UseVIP         bool
	VIPIP          string
	UseExtBalancer bool
	ExtBalancerIP  string
	Disk           string
	UseDRBD        bool
	UseZFS         bool
	UseSPL         bool
	UseVFIOPCI     bool
	UseVFIOIOMMU   bool
	UseOVS         bool
	UseMirrors     bool
	UseMaxPods     bool
}

type FileInput struct {
	ClusterName    string   `yaml:"clusterName"`
	K8sVersion     string   `yaml:"k8sVersion"`
	Image          string   `yaml:"image"`
	Iface          string   `yaml:"iface"`
	CPCount        int      `yaml:"cpCount"`
	WorkerCount    int      `yaml:"workerCount"`
	Gateway        string   `yaml:"gateway"`
	Netmask        string   `yaml:"netmask"`
	DNS1           string   `yaml:"dns1"`
	DNS2           string   `yaml:"dns2"`
	NTP1           string   `yaml:"ntp1"`
	NTP2           string   `yaml:"ntp2"`
	NTP3           string   `yaml:"ntp3"`
	UseVIP         bool     `yaml:"useVIP"`
	VIPIP          string   `yaml:"vipIP"`
	UseExtBalancer bool     `yaml:"useExtBalancer"`
	ExtBalancerIP  string   `yaml:"extBalancerIP"`
	Disk           string   `yaml:"disk"`
	UseDRBD        bool     `yaml:"useDRBD"`
	UseZFS         bool     `yaml:"useZFS"`
	UseSPL         bool     `yaml:"useSPL"`
	UseVFIOPCI     bool     `yaml:"useVFIOPCI"`
	UseVFIOIOMMU   bool     `yaml:"useVFIOIOMMU"`
	UseOVS         bool     `yaml:"useOVS"`
	UseMirrors     bool     `yaml:"useMirrors"`
	UseMaxPods     bool     `yaml:"useMaxPods"`
	CPIPs          []string `yaml:"cpIPs"`
	WorkerIPs      []string `yaml:"workerIPs"`
}

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

func checkLatestVersion() {
	const url = "https://api.github.com/repos/vasyakrg/talostpl/releases/latest"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("%s⚠️Warning: failed to check latest version%s\n", colorYellow, colorReset)
		return
	}
	defer resp.Body.Close()
	var data struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Printf("%s⚠️Warning: failed to check latest version%s\n", colorYellow, colorReset)
		return
	}
	if data.TagName != version {
		fmt.Printf("%s⚠️ Warning: your version is %s, latest is %s. Please update!%s\n", colorYellow, version, data.TagName, colorReset)
	} else {
		fmt.Printf("%s✅ You have the latest version %s%s\n", colorGreen, version, colorReset)
	}
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

func runGeneration(ans Answers, usedIPs map[string]struct{}, cpIPs, workerIPs []string, isFromFile bool) {
	configDir := configDir

	patch := PatchConfig{
		Machine: map[string]interface{}{
			"network": map[string]interface{}{
				"nameservers": []string{ans.DNS1, ans.DNS2},
			},
			"install": map[string]interface{}{
				"disk":  ans.Disk,
				"image": ans.Image,
			},
			"time": map[string]interface{}{
				"servers": []string{ans.NTP1, ans.NTP2, ans.NTP3},
			},
		},
		Cluster: map[string]interface{}{},
	}
	if ans.UseMirrors {
		patch.Machine["registries"] = map[string]interface{}{
			"mirrors": map[string]interface{}{
				"docker.io": map[string]interface{}{
					"endpoints": []string{"https://dockerhub.timeweb.cloud", "https://mirror.gcr.io"},
				},
			},
		}
	}
	if ans.UseExtBalancer && ans.ExtBalancerIP != "" {
		ips := strings.Split(ans.ExtBalancerIP, ",")
		for i := range ips {
			ips[i] = strings.TrimSpace(ips[i])
		}
		patch.Machine["certSANs"] = ips
	}
	if ans.UseDRBD && ans.WorkerCount == 0 {
		mods := []map[string]interface{}{
			{"name": "drbd", "parameters": []string{"usermode_helper=disabled"}},
			{"name": "drbd_transport_tcp"},
			{"name": "dm-thin-pool"},
		}
		if ans.UseZFS {
			mods = append(mods, map[string]interface{}{ "name": "zfs" })
		}
		if ans.UseSPL {
			mods = append(mods, map[string]interface{}{ "name": "spl" })
		}
		if ans.UseVFIOPCI {
			mods = append(mods, map[string]interface{}{ "name": "vfio_pci" })
		}
		if ans.UseVFIOIOMMU {
			mods = append(mods, map[string]interface{}{ "name": "vfio_iommu_type1" })
		}
		if ans.UseOVS {
			mods = append(mods, map[string]interface{}{ "name": "openvswitch" })
		}
		patch.Machine["kernel"] = map[string]interface{}{ "modules": mods }
	}
	if ans.WorkerCount == 0 {
		patch.Cluster["allowSchedulingOnControlPlanes"] = true
	}
	patch.Cluster["network"] = map[string]interface{}{
		"cni": map[string]interface{}{ "name": "none" },
	}
	patch.Cluster["proxy"] = map[string]interface{}{ "disabled": true }

	// Формируем SAN'ы для cluster.apiServer.certSANs
	var certSANs []string
	if ans.UseExtBalancer && ans.ExtBalancerIP != "" {
		for _, ip := range strings.Split(ans.ExtBalancerIP, ",") {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				certSANs = append(certSANs, ip)
			}
		}
	}
	if len(certSANs) > 0 {
		if patch.Cluster["apiServer"] == nil {
			patch.Cluster["apiServer"] = map[string]interface{}{}
		}
		patch.Cluster["apiServer"].(map[string]interface{})["certSANs"] = certSANs
	}

	fileWriteYAML(filepath.Join(configDir, "patch.yaml"), patch)
	fmt.Printf("%sCreated patch.yaml%s\n", colorGreen, colorReset)
	fmt.Println("--------------------------------")

	if len(cpIPs) == 0 {
		for i := 1; i <= ans.CPCount; i++ {
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
		}
	}
	for i, cpIP := range cpIPs {
		filename := filepath.Join(configDir, fmt.Sprintf("cp%d.patch", i+1))
		patch := map[string]interface{}{
			"machine": map[string]interface{}{
				"network": map[string]interface{}{
					"hostname": fmt.Sprintf("cp-%d", i+1),
					"interfaces": []map[string]interface{}{
						{
							"interface": ans.Iface,
							"dhcp": false,
							"addresses": []string{fmt.Sprintf("%s/%s", cpIP, ans.Netmask)},
							"routes": []map[string]interface{}{
								{"network": "0.0.0.0/0", "gateway": ans.Gateway},
							},
						},
					},
				},
			},
		}
		if ans.UseVIP && ans.VIPIP != "" {
			patch["machine"].(map[string]interface{})["network"].(map[string]interface{})["interfaces"].([]map[string]interface{})[0]["vip"] = map[string]interface{}{ "ip": ans.VIPIP }
		}
		if ans.UseMaxPods {
			patch["machine"].(map[string]interface{})["kubelet"] = map[string]interface{}{
				"extraConfig": map[string]interface{}{ "maxPods": 512 },
			}
		}
		fileWriteYAML(filename, patch)
		fmt.Printf("%sCreated file: %s%s\n", colorGreen, filename, colorReset)
	}
	fmt.Println("--------------------------------")

	if len(workerIPs) == 0 && ans.WorkerCount > 0 {
		for i := 1; i <= ans.WorkerCount; i++ {
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
		}
	}
	for i, workerIP := range workerIPs {
		filename := filepath.Join(configDir, fmt.Sprintf("worker%d.patch", i+1))
		patch := map[string]interface{}{
			"machine": map[string]interface{}{
				"network": map[string]interface{}{
					"hostname": fmt.Sprintf("worker-%d", i+1),
					"interfaces": []map[string]interface{}{
						{
							"deviceSelector": map[string]interface{}{ "physical": true },
							"dhcp": false,
							"addresses": []string{fmt.Sprintf("%s/%s", workerIP, ans.Netmask)},
							"routes": []map[string]interface{}{
								{"network": "0.0.0.0/0", "gateway": ans.Gateway},
							},
						},
					},
				},
			},
		}
		if ans.UseDRBD {
			mods := []map[string]interface{}{
				{"name": "drbd", "parameters": []string{"usermode_helper=disabled"}},
				{"name": "drbd_transport_tcp"},
				{"name": "dm-thin-pool"},
			}
			if ans.UseZFS {
				mods = append(mods, map[string]interface{}{ "name": "zfs" })
			}
			if ans.UseSPL {
				mods = append(mods, map[string]interface{}{ "name": "spl" })
			}
			if ans.UseVFIOPCI {
				mods = append(mods, map[string]interface{}{ "name": "vfio_pci" })
			}
			if ans.UseVFIOIOMMU {
				mods = append(mods, map[string]interface{}{ "name": "vfio_iommu_type1" })
			}
			if ans.UseOVS {
				mods = append(mods, map[string]interface{}{ "name": "openvswitch" })
			}
			patch["machine"].(map[string]interface{})["kernel"] = map[string]interface{}{ "modules": mods }
		}
		fileWriteYAML(filename, patch)
		fmt.Printf("%sCreated file: %s%s\n", colorGreen, filename, colorReset)
	}
	fmt.Println("--------------------------------")

	secretsFile := filepath.Join(configDir, "secrets.yaml")
	if err := runCmd("talosctl", "gen", "secrets", "-o", secretsFile); err != nil {
		fmt.Printf("%sError generating secrets: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	fmt.Printf("%sCreated secrets.yaml%s\n", colorGreen, colorReset)
	fmt.Println("--------------------------------")

	endpointIP := ans.VIPIP
	if endpointIP == "" && len(cpIPs) > 0 {
		endpointIP = cpIPs[0]
	}
	endpointIP = strings.Split(endpointIP, "/")[0]
	if err := os.Chdir(configDir); err != nil {
		fmt.Printf("%sError changing directory: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	if err := runCmd("talosctl", "gen", "config", "--kubernetes-version", ans.K8sVersion, "--with-secrets", "secrets.yaml", ans.ClusterName, fmt.Sprintf("https://%s:6443", endpointIP), "--config-patch", "@patch.yaml"); err != nil {
		fmt.Printf("%sError generating config: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	fmt.Println("--------------------------------")

	for i := 1; i <= ans.CPCount; i++ {
		if err := runCmd("talosctl", "machineconfig", "patch", "controlplane.yaml", "--patch", fmt.Sprintf("@cp%d.patch", i), "--output", fmt.Sprintf("cp%d.yaml", i)); err != nil {
			fmt.Printf("%sError patching cp%d: %v%s\n", colorRed, i, err, colorReset)
			os.Exit(1)
		}
		fmt.Printf("%sCreated file: cp%d.yaml%s\n", colorGreen, i, colorReset)
	}
	fmt.Println("--------------------------------")

	if ans.WorkerCount > 0 {
		for i := 1; i <= ans.WorkerCount; i++ {
			if err := runCmd("talosctl", "machineconfig", "patch", "worker.yaml", "--patch", fmt.Sprintf("@worker%d.patch", i), "--output", fmt.Sprintf("worker%d.yaml", i)); err != nil {
				fmt.Printf("%sError patching worker%d: %v%s\n", colorRed, i, err, colorReset)
				os.Exit(1)
			}
			fmt.Printf("%sCreated file: worker%d.yaml%s\n", colorGreen, i, colorReset)
		}
	}
	fmt.Println("--------------------------------")

	endpoints := append([]string{}, cpIPs...)
	if ans.UseVIP && ans.VIPIP != "" {
		endpoints = append(endpoints, ans.VIPIP)
	}
	if ans.UseExtBalancer && ans.ExtBalancerIP != "" {
		for _, ip := range strings.Split(ans.ExtBalancerIP, ",") {
			endpoints = append(endpoints, strings.TrimSpace(ip))
		}
	}

	endpointsStr := strings.Join(endpoints, ", ")
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
	os.Chdir("..")
	firstCP := cpIPs[0]
	firstCPClean := strings.Split(firstCP, "/")[0]
	if isFromFile {
		fmt.Println("Cluster initialization skipped (non interactive mode)")
		return
	}
		input := FileInput{
			ClusterName: ans.ClusterName,
			K8sVersion: ans.K8sVersion,
			Image: ans.Image,
			Iface: ans.Iface,
			CPCount: ans.CPCount,
			WorkerCount: ans.WorkerCount,
			Gateway: ans.Gateway,
			Netmask: ans.Netmask,
			DNS1: ans.DNS1,
			DNS2: ans.DNS2,
			NTP1: ans.NTP1,
			NTP2: ans.NTP2,
			NTP3: ans.NTP3,
			UseVIP: ans.UseVIP,
			VIPIP: ans.VIPIP,
			UseExtBalancer: ans.UseExtBalancer,
			ExtBalancerIP: ans.ExtBalancerIP,
			Disk: ans.Disk,
			UseDRBD: ans.UseDRBD,
			UseZFS: ans.UseZFS,
			UseSPL: ans.UseSPL,
			UseVFIOPCI: ans.UseVFIOPCI,
			UseVFIOIOMMU: ans.UseVFIOIOMMU,
			UseOVS: ans.UseOVS,
			UseMirrors: ans.UseMirrors,
			UseMaxPods: ans.UseMaxPods,
			CPIPs: cpIPs,
			WorkerIPs: workerIPs,
	}

	if !askYesNoNumbered("Do you want to start cluster initialization?", "y") {
		fmt.Println("--------------------------------")
		fmt.Println("Cluster initialization cancelled by user.")
		fmt.Println("--------------------------------")
		printManualInitHelp(input, ans)
		return
	}

	if err := runCmd("talosctl", "apply-config", "--insecure", "-n", firstCPClean, "--file", filepath.Join(configDir, "cp1.yaml")); err != nil {
		fmt.Printf("%sError apply-config: %v%s\n", colorRed, err, colorReset)
		printManualInitHelp(input, ans)
		os.Exit(1)
	}
	fmt.Println(colorRed + "Please, wait init and reboot first control plane, before continue" + colorReset)
	if !askYesNoNumbered("Continue?", "y") {
		fmt.Println("--------------------------------")
		fmt.Println("Cluster initialization cancelled by user.")
		fmt.Println("--------------------------------")
		return
	}
	if err := runCmd("talosctl", "bootstrap", "--nodes", firstCPClean, "--endpoints", firstCPClean, "--talosconfig="+filepath.Join(configDir, "talosconfig")); err != nil {
		fmt.Printf("%sError bootstrap: %v%s\n", colorRed, err, colorReset)
		printManualInitHelp(input, ans)
		os.Exit(1)
	}

	fmt.Println(colorRed + "Please, wait bootstrap first control plane, before continue" + colorReset)
	if !askYesNoNumbered("Continue?", "y") {
		fmt.Println("--------------------------------")
		fmt.Println("Cluster initialization cancelled by user.")
		fmt.Println("--------------------------------")
		return
	}

	fmt.Println("Applying config to control planes and workers ..")
	if ans.CPCount > 1 {
		for i := 2; i <= ans.CPCount; i++ {
			cpClean := strings.Split(cpIPs[i-1], "/")[0]
			if err := runCmd("talosctl", "apply-config", "--insecure", "-n", cpClean, "--file", filepath.Join(configDir, fmt.Sprintf("cp%d.yaml", i))); err != nil {
				fmt.Printf("%sError apply-config cp%d: %v%s\n", colorRed, i, err, colorReset)
				printManualInitHelp(input, ans)
				os.Exit(1)
			}
		}
	}
	if ans.WorkerCount > 0 {
		for i := 1; i <= ans.WorkerCount; i++ {
			if err := runCmd("talosctl", "apply-config", "--insecure", "-n", workerIPs[i-1], "--file", filepath.Join(configDir, fmt.Sprintf("worker%d.yaml", i))); err != nil {
				fmt.Printf("%sError apply-config worker%d: %v%s\n", colorRed, i, err, colorReset)
				printManualInitHelp(input, ans)
				os.Exit(1)
			}
		}
	}
	fmt.Println("Done")
	fmt.Println("--------------------------------")
	fmt.Println("Generating kubeconfig ..")

	kubeconfigEndpoint := ans.VIPIP
	if kubeconfigEndpoint == "" {
		kubeconfigEndpoint = firstCPClean
	}
	kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", ans.ClusterName+".yaml")
	if err := runCmd("talosctl", "kubeconfig", kubeconfigPath, "--nodes", kubeconfigEndpoint, "--endpoints", kubeconfigEndpoint, "--talosconfig", filepath.Join(configDir, "talosconfig")); err != nil {
		fmt.Printf("%sError exporting kubeconfig: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	fmt.Println("--------------------------------")
	fmt.Println("Script completed")
	fmt.Println("--------------------------------")
	fmt.Println("Next, you need to install the network plugin Cilium")
	fmt.Println("Documentation: https://docs.cilium.io/en/stable/gettingstarted/k8s-install-default/")
	fmt.Println("-----------done-----------------")
}

func printManualInitHelp(input FileInput, ans Answers) {
	endpoint := input.CPIPs[0]
	if input.UseVIP && input.VIPIP != "" {
		endpoint = input.VIPIP
	}
	fmt.Println("\n-----------------------------")
	fmt.Println("Manual cluster initialization required. Run the following commands:")
	fmt.Println()
	fmt.Printf("talosctl apply-config --insecure -n %s --file cp1.yaml\n", input.CPIPs[0])
	fmt.Println("---------------")
	fmt.Println(colorRed + "Please, wait init and reboot first control plane, before run next commands" + colorReset)
	fmt.Println("---------------")
	fmt.Printf("talosctl bootstrap --nodes %s --endpoints %s --talosconfig=talosconfig\n", input.CPIPs[0], input.CPIPs[0])
	fmt.Println("---------------")
	fmt.Println(colorRed + "Please, wait bootstrap first control plane, before run next commands" + colorReset)
	fmt.Println("---------------")
	for i := 2; i <= input.CPCount; i++ {
		fmt.Printf("talosctl apply-config --insecure -n %s --file cp%d.yaml\n", input.CPIPs[i-1], i)
	}
	for i := 1; i <= input.WorkerCount; i++ {
		fmt.Printf("talosctl apply-config --insecure -n %s --file worker%d.yaml\n", input.WorkerIPs[i-1], i)
	}
	fmt.Printf("talosctl kubeconfig ~/.kube/%s.yaml --nodes %s --endpoints %s --talosconfig talosconfig\n", ans.ClusterName, endpoint, endpoint)
	fmt.Println("-----------------------------\n")
}

func generateCmd() *cobra.Command {
	var force bool
	var fromFile string
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Interactive Talos K8s config generator",
		Run: func(cmd *cobra.Command, args []string) {
			if err := checkRequiredTools(); err != nil {
				os.Exit(1)
			}
			if configDir == "" {
				configDir = "config"
			}
			os.MkdirAll(configDir, 0o755)

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
					if fromFile != "" {
						err := clearDir(configDir)
						if err != nil {
							fmt.Printf("%sFailed to clean directory: %v%s\n", colorRed, err, colorReset)
							os.Exit(1)
						}
						fmt.Printf("%sDirectory '%s' cleaned.%s\n", colorGreen, configDir, colorReset)
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
			}

			if fromFile != "" {
				var input FileInput
				f, err := os.Open(fromFile)
				if err != nil {
					fmt.Printf("%sFailed to open file: %v%s\n", colorRed, err, colorReset)
					os.Exit(1)
				}
				defer f.Close()
				dec := yaml.NewDecoder(f)
				if err := dec.Decode(&input); err != nil {
					fmt.Printf("%sFailed to parse YAML: %v%s\n", colorRed, err, colorReset)
					os.Exit(1)
				}
				ans := Answers{
					ClusterName:    input.ClusterName,
					K8sVersion:     input.K8sVersion,
					Image:          input.Image,
					Iface:          input.Iface,
					CPCount:        input.CPCount,
					WorkerCount:    input.WorkerCount,
					Gateway:        input.Gateway,
					Netmask:        input.Netmask,
					DNS1:           input.DNS1,
					DNS2:           input.DNS2,
					NTP1:           input.NTP1,
					NTP2:           input.NTP2,
					NTP3:           input.NTP3,
					UseVIP:         input.UseVIP,
					VIPIP:          input.VIPIP,
					UseExtBalancer: input.UseExtBalancer,
					ExtBalancerIP:  input.ExtBalancerIP,
					Disk:           input.Disk,
					UseDRBD:        input.UseDRBD,
					UseZFS:         input.UseZFS,
					UseSPL:         input.UseSPL,
					UseVFIOPCI:     input.UseVFIOPCI,
					UseVFIOIOMMU:   input.UseVFIOIOMMU,
					UseOVS:         input.UseOVS,
					UseMirrors:     input.UseMirrors,
					UseMaxPods:     input.UseMaxPods,
				}
				usedIPs := map[string]struct{}{input.Gateway: {}}
				for _, ip := range input.CPIPs {
					usedIPs[ip] = struct{}{}
				}
				for _, ip := range input.WorkerIPs {
					usedIPs[ip] = struct{}{}
				}
				runGeneration(ans, usedIPs, input.CPIPs, input.WorkerIPs, true)
				printManualInitHelp(input, ans)
				return
			}

			ans := Answers{}
			ans.ClusterName = askNumbered("Enter cluster name [talos-demo]: ", "talos-demo")
			ans.K8sVersion = askNumbered("Enter Kubernetes version ["+k8sVersion+"]: ", k8sVersion)
			ans.Image = askNumbered("Enter Talos installer image ["+image+"]: ", image)
			ans.Iface = askNumbered("Enter network interface name ens18 or eth0 may be? [ens18]: ", "ens18")
			ans.CPCount = mustAtoi(askNumbered("Enter number of control planes (odd, max 7) [1]: ", "1"))
			ans.WorkerCount = mustAtoi(askNumbered("Enter number of worker nodes (max 15, min 0) [3]: ", "3"))
			ans.Gateway = askNumbered("Enter default gateway: ", "")
			ans.Netmask = askNumbered("Enter network mask [24]: ", "24")
			ans.DNS1 = askNumbered("Enter first DNS server [8.8.8.8]: ", "8.8.8.8")
			ans.DNS2 = askNumbered("Enter second DNS server [8.8.4.4]: ", "8.8.4.4")
			ans.NTP1 = askNumbered("Enter first NTP server [1.ru.pool.ntp.org]: ", "1.ru.pool.ntp.org")
			ans.NTP2 = askNumbered("Enter second NTP server [2.ru.pool.ntp.org]: ", "2.ru.pool.ntp.org")
			ans.NTP3 = askNumbered("Enter third NTP server [3.ru.pool.ntp.org]: ", "3.ru.pool.ntp.org")
			ans.UseVIP = false
			ans.VIPIP = ""
			if ans.CPCount > 1 {
				ans.UseVIP = askYesNoNumbered("Do you need a VIP address?", "y")
				if ans.UseVIP {
					ans.VIPIP = askNumbered("Enter VIP address: ", "")
				}
			}
			ans.UseExtBalancer = askYesNoNumbered("Do you need an external load balancers (SAN's proxy)?", "n")
			ans.ExtBalancerIP = ""
			if ans.UseExtBalancer {
				ans.ExtBalancerIP = askNumbered("Enter external load balancer IPs (proxy server IPs or list of SAN's) (input comma separated, if more than one): ", "")
			}
			ans.Disk = askNumbered("Enter disk for base OS installation [/dev/sda]: ", "/dev/sda")
			ans.UseDRBD = askYesNoNumbered("Enable drbd support?", "y")
			ans.UseZFS = askYesNoNumbered("Enable zfs support?", "n")
			ans.UseSPL = askYesNoNumbered("Enable spl support?", "n")
			ans.UseVFIOPCI = askYesNoNumbered("Enable vfio_pci support?", "n")
			ans.UseVFIOIOMMU = askYesNoNumbered("Enable vfio_iommu_type1 support?", "n")
			ans.UseOVS = askYesNoNumbered("Enable openvswitch support?", "n")
			ans.UseMirrors = askYesNoNumbered("Use timeweb.cloud and gcr.io mirrors for docker.io?", "y")
			ans.UseMaxPods = askYesNoNumbered("Set maxPods: 512 for kubelet? (default is 110 per node)", "n")
			usedIPs := map[string]struct{}{ans.Gateway: {}}
			var cpIPs, workerIPs []string
			for i := 1; i <= ans.CPCount; i++ {
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
			}
			if ans.WorkerCount > 0 {
				for i := 1; i <= ans.WorkerCount; i++ {
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
				}
			}
			runGeneration(ans, usedIPs, cpIPs, workerIPs, false)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Force clean config directory if not empty")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "YAML file with all answers for non-interactive mode (see --help for example)")
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
	checkLatestVersion()
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
