package main

import (
	_ "embed"
	"net"

	"golang.org/x/crypto/ssh"
)

func reboot(ip string, signer ssh.Signer) error {
	addr := net.JoinHostPort(ip, "22")
	client, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{
		User: "isuadmin",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	if err := session.Run("sudo systemctl reboot"); err != nil {
		return err
	}

	return nil
}
