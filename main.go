package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

const (
	defaultSshHost = "192.168.1.1"
	defaultSshUser = "admin"
)

func maskToCIDR(mask string) int {
	ip := net.ParseIP(mask)
	if ip == nil {
		return -1
	}
	maskIP := net.IPv4Mask(ip[12], ip[13], ip[14], ip[15])
	ones, _ := maskIP.Size()
	return ones
}

func parseRouteLine(line string) string {
	line = strings.TrimSpace(line)
	lower := strings.ToLower(line)
	if !strings.HasPrefix(lower, "route add") {
		return ""
	}

	parts := strings.Fields(line)
	if len(parts) < 6 {
		return ""
	}

	ip := parts[2]
	if strings.ToLower(parts[3]) != "mask" {
		return ""
	}
	mask := parts[4]

	cidr := maskToCIDR(mask)
	if cidr < 0 || cidr > 32 {
		return ""
	}

	return fmt.Sprintf("%s/%d", ip, cidr)
}

func readRoutesFromTxtFiles() ([]string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %v", err)
	}
	execDir := filepath.Dir(execPath)

	matches, err := filepath.Glob(filepath.Join(execDir, "*.txt"))
	if err != nil {
		return nil, fmt.Errorf("failed to find txt files: %v", err)
	}

	var subnets []string
	for _, match := range matches {
		file, err := os.Open(match)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			subnet := parseRouteLine(scanner.Text())
			if subnet != "" {
				subnets = append(subnets, subnet)
			}
		}
		file.Close()
	}

	return subnets, nil
}

func main() {
	host := flag.String("host", defaultSshHost, "SSH host address")
	user := flag.String("user", defaultSshUser, "SSH username")
	port := flag.Int("port", 22, "SSH port")
	flag.Parse()

	subnets := flag.Args()
	var fileSubnets []string
	var err error

	fileSubnets, err = readRoutesFromTxtFiles()
	if err != nil {
		fmt.Printf("Warning: %v\n", err)
	}
	subnets = append(subnets, fileSubnets...)

	if len(subnets) == 0 {
		fmt.Println("Usage: keenetic-route-add [-host <ip>] [-user <name>] <subnet1> [subnet2] ...")
		fmt.Println("   or: place .txt files with 'route add <ip> mask <mask> <gw>' lines next to the executable")
		os.Exit(1)
	}

	interfaces := []string{"OpenVPN0", "OpenVPN1"}

	var commands []string
	for _, iface := range interfaces {
		for _, subnet := range subnets {
			commands = append(commands, fmt.Sprintf("ip route %s %s", subnet, iface))
		}
	}

	fmt.Print("Enter SSH password: ")
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("\nFailed to read password")
		os.Exit(1)
	}
	fmt.Println()
	password := string(passwordBytes)

	config := &ssh.ClientConfig{
		User: *user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", *host, *port), config)
	if err != nil {
		fmt.Printf("Failed to connect to %s:%d: %v\n", *host, *port, err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Printf("Connected to %s:%d\n", *host, *port)

	for _, cmd := range commands {
		session, err := client.NewSession()
		if err != nil {
			fmt.Printf("Failed to create session: %v\n", err)
			continue
		}

		fmt.Printf("Executing: %s\n", cmd)
		output, err := session.CombinedOutput(cmd)
		session.Close()

		if err != nil {
			fmt.Printf("Error executing '%s': %v\n", cmd, err)
		}
		if len(output) > 0 {
			fmt.Printf("Output: %s\n", strings.TrimSpace(string(output)))
		}
	}

	fmt.Println("Done")
}
