package main

import (
	"encoding/json"
	"errors"

	"golang.org/x/crypto/blake2b"
)

//AttachmentHash is a blake2b 256 (32byte) length byte array
type AttachmentHash []byte

//Attachment on replies or comments, images, pdfs, any
//type of email attachment
type Attachment struct {
	Hash     []byte `storm:"unique"`
	Data     []byte `json:"-"`
	Filename string
}

//WSCreateAttachment websocket handler for the creation of attachments
func WSCreateAttachment(userID int64, reqJSON []byte) ([]byte, error) {
	var attachment Attachment
	err := json.Unmarshal(reqJSON, &attachment)
	if err != nil {
		return nil, err
	}

	if attachment.Data == nil || len(attachment.Data) < 1 {
		return nil, errors.New("Attachment data can't be empty")
	}

	hasher, err := blake2b.New256(nil)
	if err != nil {
		return nil, err
	}

	_, err = hasher.Write(attachment.Data)
	if err != nil {
		return nil, err
	}

	attachment.Hash = hasher.Sum(nil)

	err = attachmentDB.Save(&attachment)
	if err != nil {
		return nil, err
	}

	res, err := json.Marshal(attachment)
	if err != nil {
		return nil, err
	}

	return res, nil
}
