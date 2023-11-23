package main

import (
	_ "embed"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

func reboot(ip string, signer ssh.Signer) error {
	addr := net.JoinHostPort(ip, "22")
	log.Printf("addr = %s\n", addr)
	log.Printf("signer = %+v\n", signer)
	client, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{
		User: "isuadmin",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		log.Printf("error = %s\n", err.Error())
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
