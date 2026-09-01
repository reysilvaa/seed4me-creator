package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

type VPNNode struct {
	Country string
	Code    string
	Host    string
	Port    int
}

var DefaultNodes = []VPNNode{
	// Asia & Pacific
	{Country: "Singapore", Code: "sg", Host: "sg.seed4.me", Port: 1194},
	{Country: "Indonesia", Code: "id", Host: "id.seed4.me", Port: 1194},
	{Country: "Malaysia", Code: "my", Host: "my.seed4.me", Port: 1194},
	{Country: "Thailand", Code: "th", Host: "th.seed4.me", Port: 1194},
	{Country: "Vietnam", Code: "vn", Host: "vn.seed4.me", Port: 1194},
	{Country: "Hong Kong", Code: "hk", Host: "hk.seed4.me", Port: 1194},
	{Country: "Taiwan", Code: "tw", Host: "tw.seed4.me", Port: 1194},
	{Country: "Japan", Code: "jp", Host: "jp.seed4.me", Port: 1194},
	{Country: "South Korea", Code: "kr", Host: "kr.seed4.me", Port: 1194},
	{Country: "India", Code: "in", Host: "in.seed4.me", Port: 1194},
	{Country: "Australia", Code: "au", Host: "au.seed4.me", Port: 1194},
	{Country: "New Zealand", Code: "nz", Host: "nz.seed4.me", Port: 1194},
	{Country: "Philippines", Code: "ph", Host: "ph.seed4.me", Port: 1194},

	// Europe
	{Country: "United Kingdom", Code: "uk", Host: "uk.seed4.me", Port: 1194},
	{Country: "Germany", Code: "de", Host: "de.seed4.me", Port: 1194},
	{Country: "Netherlands", Code: "nl", Host: "nl.seed4.me", Port: 1194},
	{Country: "France", Code: "fr", Host: "fr.seed4.me", Port: 1194},
	{Country: "Switzerland", Code: "ch", Host: "ch.seed4.me", Port: 1194},
	{Country: "Sweden", Code: "se", Host: "se.seed4.me", Port: 1194},
	{Country: "Norway", Code: "no", Host: "no.seed4.me", Port: 1194},
	{Country: "Finland", Code: "fi", Host: "fi.seed4.me", Port: 1194},
	{Country: "Denmark", Code: "dk", Host: "dk.seed4.me", Port: 1194},
	{Country: "Italy", Code: "it", Host: "it.seed4.me", Port: 1194},
	{Country: "Spain", Code: "es", Host: "es.seed4.me", Port: 1194},
	{Country: "Portugal", Code: "pt", Host: "pt.seed4.me", Port: 1194},
	{Country: "Belgium", Code: "be", Host: "be.seed4.me", Port: 1194},
	{Country: "Austria", Code: "at", Host: "at.seed4.me", Port: 1194},
	{Country: "Poland", Code: "pl", Host: "pl.seed4.me", Port: 1194},
	{Country: "Czech Republic", Code: "cz", Host: "cz.seed4.me", Port: 1194},
	{Country: "Romania", Code: "ro", Host: "ro.seed4.me", Port: 1194},
	{Country: "Bulgaria", Code: "bg", Host: "bg.seed4.me", Port: 1194},
	{Country: "Greece", Code: "gr", Host: "gr.seed4.me", Port: 1194},
	{Country: "Ukraine", Code: "ua", Host: "ua.seed4.me", Port: 1194},
	{Country: "Turkey", Code: "tr", Host: "tr.seed4.me", Port: 1194},
	{Country: "Cyprus", Code: "cy", Host: "cy.seed4.me", Port: 1194},
	{Country: "Moldova", Code: "md", Host: "md.seed4.me", Port: 1194},
	{Country: "Latvia", Code: "lv", Host: "lv.seed4.me", Port: 1194},
	{Country: "Lithuania", Code: "lt", Host: "lt.seed4.me", Port: 1194},
	{Country: "Luxembourg", Code: "lu", Host: "lu.seed4.me", Port: 1194},

	// Americas
	{Country: "United States", Code: "us", Host: "us.seed4.me", Port: 1194},
	{Country: "Canada", Code: "ca", Host: "ca.seed4.me", Port: 1194},
	{Country: "Brazil", Code: "br", Host: "br.seed4.me", Port: 1194},
	{Country: "Argentina", Code: "ar", Host: "ar.seed4.me", Port: 1194},
	{Country: "Mexico", Code: "mx", Host: "mx.seed4.me", Port: 1194},
	{Country: "Chile", Code: "cl", Host: "cl.seed4.me", Port: 1194},
	{Country: "Colombia", Code: "co", Host: "co.seed4.me", Port: 1194},

	// Middle East & Africa
	{Country: "United Arab Emirates", Code: "ae", Host: "ae.seed4.me", Port: 1194},
	{Country: "Israel", Code: "il", Host: "il.seed4.me", Port: 1194},
	{Country: "South Africa", Code: "za", Host: "za.seed4.me", Port: 1194},
	{Country: "Saudi Arabia", Code: "sa", Host: "sa.seed4.me", Port: 1194},
}

const ovpnTemplate = `client
dev tun
proto udp
remote %s %d
resolv-retry infinite
nobind
persist-key
persist-tun
auth-user-pass auth.txt
data-ciphers AES-128-CBC:AES-256-CBC:AES-128-GCM:AES-256-GCM:CHACHA20-POLY1305
data-ciphers-fallback AES-128-CBC
auth SHA256
redirect-gateway def1
dhcp-option DNS 1.1.1.1
dhcp-option DNS 8.8.8.8
verb 3
`

// SaveOVPNProfiles generates .ovpn files for Linux/Windows and the auth.txt credentials
func SaveOVPNProfiles(email, password, outDir string) error {
	if outDir == "" {
		outDir = "ovpn"
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	// 1. auth.txt
	authContent := fmt.Sprintf("%s\n%s\n", email, password)
	authPath := filepath.Join(outDir, "auth.txt")
	if err := os.WriteFile(authPath, []byte(authContent), 0600); err != nil {
		return err
	}

	// 2. .ovpn profiles
	for _, n := range DefaultNodes {
		content := fmt.Sprintf(ovpnTemplate, n.Host, n.Port)
		fn := filepath.Join(outDir, fmt.Sprintf("seed4me-%s.ovpn", n.Code))
		if err := os.WriteFile(fn, []byte(content), 0644); err != nil {
			return err
		}
	}

	// 3. connect.sh for Linux/Debian
	connectSh := `#!/usr/bin/env bash
# Quick connect script for Debian/Ubuntu/Linux
NODE=${1:-sg}
CONFIG="seed4me-${NODE}.ovpn"
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ ! -f "$DIR/$CONFIG" ]; then
    echo "[!] Config $CONFIG tidak ditemukan. Contoh pakai: ./connect.sh sg (atau us, jp, uk, id)"
    exit 1
fi

echo "[*] Menghubungkan ke Seed4.Me ($CONFIG)..."
sudo openvpn --config "$DIR/$CONFIG" --auth-user-pass "$DIR/auth.txt"
`
	_ = os.WriteFile(filepath.Join(outDir, "connect.sh"), []byte(connectSh), 0755)

	return nil
}
