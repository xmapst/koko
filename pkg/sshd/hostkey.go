package sshd

import (
	"golang.org/x/crypto/ssh"
)

func ParsePrivateKeyFromString(content string) (signer ssh.Signer, err error) {
	return ssh.ParsePrivateKey([]byte(content))
}
