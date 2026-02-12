package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"filippo.io/age"
	"github.com/spf13/cobra"
	"oliverj.io/fragment/internal/archive"
	"oliverj.io/fragment/internal/shamir"
	"oliverj.io/fragment/internal/share"
)

var shareCount int
var threshold int

func init() {
	encryptCmd.Flags().IntVar(&shareCount, "shares", 2, "Number of shares")
	encryptCmd.Flags().IntVar(&threshold, "threshold", 2, "Threshold")

	rootCmd.AddCommand(encryptCmd)
}

var encryptCmd = &cobra.Command{
	Use:     "encrypt",
	Short:   "encrypt file into shares",
	Args:    cobra.MinimumNArgs(1),
	Run:     encrypt,
	Example: "fragment encrypt --shares 4 --threshold 2 files_dir",
}

func encrypt(cmd *cobra.Command, args []string) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}

	out := &bytes.Buffer{}

	w, err := age.Encrypt(out, identity.Recipient())
	if err != nil {
		log.Fatalf("Failed to create encrypted file: %v", err)
	}

	arch := &bytes.Buffer{}
	stat, err := os.Stat(args[0])
	if err != nil {
		log.Fatalf("Failed to read file info: %v", err)
	}
	if stat.IsDir() {
		err = archive.CreateDirArchive(arch, args[0])
		if err != nil {
			log.Fatalf("Failed to create archive: %v", err)
		}
	} else {
		data, err := os.ReadFile(args[0])
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
		err = archive.CreateArchiveEnt(arch, []archive.FileEntry{{args[0], data}})
		if err != nil {
			log.Fatalf("Failed to create archilve: %v", err)
		}
	}

	if _, err := w.Write(arch.Bytes()); err != nil {
		log.Fatalf("Failed to write to encrypted file: %v", err)
	}

	if err := w.Close(); err != nil {
		log.Fatalf("Failed to close encrypted file: %v", err)
	}

	shares, err := shamir.Split(shareCount, threshold, []byte(identity.String()))

	for i, s := range shares {
		shr := share.Share{
			ShamirKey:      s,
			EncryptedBlob:  out.Bytes(),
			ShareCount:     shareCount,
			ShareThreshold: threshold,
		}

		shareData, err := shr.Encode()
		if err != nil {
			log.Fatalf("Failed to encode share data: %v", err)
		}

		err = os.WriteFile(fmt.Sprintf("share-%d.fragment", i+1), shareData, 066)
		if err != nil {
			log.Fatalf("Failed to create fragment: share-%d.fragment: %v", i+1, err)
		}
	}
}
