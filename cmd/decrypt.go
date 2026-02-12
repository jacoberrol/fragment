package cmd

import (
	"bytes"
	"io"
	"log"
	"os"

	"filippo.io/age"
	"github.com/spf13/cobra"
	"oliverj.io/fragment/internal/shamir"
	"oliverj.io/fragment/internal/share"
)

func init() {
	rootCmd.AddCommand(decryptCmd)
}

var decryptCmd = &cobra.Command{
	Use:     "decrypt",
	Short:   "decrypts files stored in shares",
	Run:     decrypt,
	Example: "fragment decrypt share-1.txt share-2.txt ... share-n.txt",
}

func decrypt(cmd *cobra.Command, args []string) {

	shares := make([]share.Share, len(args))

	file, err := os.ReadFile(args[0])
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	s, err := share.Decode(file)
	if err != nil {
		log.Fatalf("Failed to decode share: %v", err)
	}
	payload := s.EncryptedBlob

	for i, arg := range args {
		data, err := os.ReadFile(arg)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
		shr, err := share.Decode(data)
		if err != nil {
			log.Fatalf("Failed to decode share: %v", err)
		}
		shares[i] = *shr
	}

	shareKeys := make([][]byte, len(shares))

	for i, shr := range shares {
		shareKeys[i] = shr.ShamirKey
	}

	if len(shares) < shares[0].ShareThreshold {
		log.Fatalf("Unable to recover secret")
	}

	if len(shares) > shares[0].ShareCount {
		log.Fatalf("Too many shares")
	}

	secret, err := shamir.Combine(shareKeys)

	identity, err := age.ParseX25519Identity(string(secret))

	dataReader := bytes.NewReader(payload)

	reader, err := age.Decrypt(dataReader, identity)

	contents, _ := io.ReadAll(reader)

	err = os.WriteFile("out.tar.gz", contents, 066)
	if err != nil {
		log.Fatalf("Failed to create output archive: %v", err)
	}
}
