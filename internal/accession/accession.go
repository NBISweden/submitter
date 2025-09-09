package accession

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"math/big"

	"github.com/NBISweden/submitter/pkg/sdaclient"
)

var ErrFileAlreadyExists = errors.New("File already exists")

type File struct {
	InboxPath  string `json:"inboxPath"`
	FileStatus string `json:"fileStatus"`
}

func CreateAccessionIDs(sdaclient *sdaclient.Client, fileIDPath string, dryRun bool) error {
	file, err := createFileIDFile(fileIDPath, dryRun)
	if err != nil {
		fmt.Printf("Error occoured when trying to create file: %s\n", fileIDPath)
		return err
	}
	defer file.Close()

	response, err := sdaclient.GetUsersFiles()
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var files []File
	if err := json.Unmarshal(body, &files); err != nil {
		return err
	}

	var paths []string
	for _, f := range files {
		if f.FileStatus == "verified" &&
			strings.Contains(f.InboxPath, sdaclient.DatasetFolder) &&
			!strings.Contains(f.InboxPath, "PRIVATE") {
			paths = append(paths, f.InboxPath)
		}
}

	fmt.Printf("[Accession] Number of files to finalize: %d\n", len(paths))
	fmt.Println("[Accession] Paths :", paths)

	if dryRun {
		fmt.Println("[Dry-Run] No files will not be given accession ids")
		// Don't ask me why, but If I don't have this print the print above will not show
		fmt.Println()
		return nil
	}

	for _, filepath := range paths {
		accessionID, err := generateAccessionID()
		if err != nil {
			return err
		}

		payload, err := json.Marshal(map[string]string{
			"accession_id": accessionID,
			"filepath": filepath,
			"user": sdaclient.UserID,
		})
		if err != nil {
			return err
		}

		_, err = sdaclient.PostFileAccession(payload)
		if err != nil {
			return err
		}

		if _, err := file.WriteString(accessionID + "\n"); err != nil {
			return err
		}
	}
	return nil
}

func createFileIDFile(fileIDPath string, dryrun bool) (*os.File, error) {
	if dryrun {
		return nil, nil
	}

	if _, err := os.Stat(fileIDPath); err == nil {
		return nil, ErrFileAlreadyExists
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	file, err := os.Create(fileIDPath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func generateAccessionID() (string, error) {
	const chars = "abcdefghijklmnopqrstuvxyz23456789"
	const length = 6

	genPart := func() (string, error) {
		result := make([]byte, length)
		for i := range length {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
			if err != nil {
				return "", err
			}
			result[i] = chars[n.Int64()]
		}
		return  string(result), nil
	}
	partOne, err := genPart()
	if err != nil {
		return "", err
	}

	partTwo, err := genPart()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("aa-File-%s-%s", partOne, partTwo), nil
}
