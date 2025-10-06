package accession

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"

	"github.com/NBISweden/submitter/pkg/sdaclient"
	"github.com/schollz/progressbar/v3"
)

var ErrFileAlreadyExists = errors.New("file already exists")

type File struct {
	InboxPath  string `json:"inboxPath"`
	FileStatus string `json:"fileStatus"`
}

func CreateAccessionIDs(client *sdaclient.Client, dryRun bool) error {
	filePath := fmt.Sprintf("data/%s-fileIDs.txt", client.DatasetFolder)
	file, err := createFileIDFile(filePath, dryRun)
	if err != nil {
		fmt.Printf("error occoured when trying to create file: %s\n", filePath)
		return err
	}
	defer file.Close() //nolint:errcheck

	fmt.Println("[accession] waiting on response from sda api ...")
	response, err := client.GetUsersFiles()
	if err != nil {
		return err
	}
	defer response.Body.Close() //nolint:errcheck

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
			strings.Contains(f.InboxPath, client.DatasetFolder) &&
			!strings.Contains(f.InboxPath, "PRIVATE") {
			paths = append(paths, f.InboxPath)
		}
	}

	fmt.Printf("[accession] files found for accession id creation: %d\n", len(paths))

	if dryRun {
		fmt.Println("[dry-run] no files will not be given accession ids")
		return nil
	}

	bar := progressbar.Default(int64(len(paths)), "[accession] creating accession ids")
	for _, filepath := range paths {
		_ = bar.Add(1)
		accessionID, err := generateAccessionID()
		if err != nil {
			return err
		}

		payload, err := json.Marshal(map[string]string{
			"accession_id": accessionID,
			"filepath":     filepath,
			"user":         client.UserID,
		})
		if err != nil {
			return err
		}

		resp, err := client.PostFileAccession(payload)
		if err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				continue
			}
			return err
		}
		defer resp.Body.Close() //nolint:errcheck

		if _, err := file.WriteString(accessionID + "\n"); err != nil {
			return err
		}
	}

	fmt.Printf("[accesion] all %d files assigned accession IDs\n", len(paths))

	return nil
}

func GetVerifiedFilePaths(client *sdaclient.Client) ([]string, error) {
	fmt.Println("[accession] waiting on response from sda api ...")
	response, err := client.GetUsersFiles()
	if err != nil {
		return nil, err
	}
	defer response.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("[accession] failed to read response body %w", err)
	}

	var files []File
	if err := json.Unmarshal(body, &files); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user files: %w", err)
	}

	var paths []string
	for _, f := range files {
		if f.FileStatus == "verified" &&
			strings.Contains(f.InboxPath, client.DatasetFolder) &&
			!strings.Contains(f.InboxPath, "PRIVATE") {
			paths = append(paths, f.InboxPath)
		}
	}
	return paths, nil
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
		return string(result), nil
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
