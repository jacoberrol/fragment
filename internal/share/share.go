package share

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	nacl "github.com/nathants/go-libsodium"
	"oliverj.io/fragment/internal/archive"
)

type Share struct {
	ShamirKey      []byte
	EncryptedBlob  []byte
	ShareCount     int
	ShareThreshold int
}

type Meta struct {
	ShareCount     int `json:"shareCount"`
	ShareThreshold int `json:"shareThreshold"`
}

func (s *Share) Encode() ([]byte, error) {
	meta, err := json.Marshal(Meta{s.ShareCount, s.ShareThreshold})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to encode metadata to JSON: %v", err))
	}

	files := []archive.FileEntry{
		{"MANIFEST.age", s.EncryptedBlob},
		{"SHARE.txt", s.ShamirKey},
		{"metadata.json", meta},
	}

	var out = &bytes.Buffer{}
	err = archive.CreateArchiveEnt(out, files)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to create archive: %v", err))
	}

	nacl.Init()
	key, err := nacl.StreamKeygen()
	if err != nil {
		panic(err)
	}

	var cipher bytes.Buffer
	err = nacl.StreamEncrypt(key, out, &cipher)

	var output bytes.Buffer
	if _, err := io.Copy(&output, bytes.NewReader(key)); err != nil {
		panic(err)
	}
	if _, err := io.Copy(&output, &cipher); err != nil {
		panic(err)
	}

	return output.Bytes(), nil
}

func DecodeExternKey(data []byte, key []byte) (*Share, error) {
	nacl.Init()

	var arc bytes.Buffer
	err := nacl.StreamDecrypt(key, bytes.NewReader(data), &arc)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to decrypt fragment! %v", err))
	}

	files := archive.ReadArchiveEnt(&arc)

	var share = &Share{}

	for _, file := range files {
		switch file.Name {
		case "SHARE.txt":
			share.ShamirKey = file.Data
		case "MANIFEST.age":
			share.EncryptedBlob = file.Data
		case "metadata.json":
			var meta = &Meta{}
			err := json.Unmarshal(file.Data, meta)
			if err != nil {
				return nil, err
			}
			share.ShareCount = meta.ShareCount
			share.ShareThreshold = meta.ShareThreshold
		}
	}

	return share, nil
}

func Decode(data []byte) (*Share, error) {
	nacl.Init()

	key := data[:32]
	encryptedArc := data[32:]

	var arc bytes.Buffer
	err := nacl.StreamDecrypt(key, bytes.NewReader(encryptedArc), &arc)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to decrypt fragment! %v", err))
	}

	files := archive.ReadArchiveEnt(&arc)

	var share = &Share{}

	for _, file := range files {
		switch file.Name {
		case "SHARE.txt":
			share.ShamirKey = file.Data
		case "MANIFEST.age":
			share.EncryptedBlob = file.Data
		case "metadata.json":
			var meta = &Meta{}
			err := json.Unmarshal(file.Data, meta)
			if err != nil {
				return nil, err
			}
			share.ShareCount = meta.ShareCount
			share.ShareThreshold = meta.ShareThreshold
		}
	}

	return share, nil
}
