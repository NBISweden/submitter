package accession

import (
	"errors"
	"fmt"
	"os"

	"github.com/NBISweden/submitter/pkg/sdaclient"
)

var ErrFileAlreadyExists = errors.New("File already exists")

func CreateAccessionIDs(sdaclient *sdaclient.Client, fileIDPath string, dryRun bool) error {
	fmt.Println("Creating accession IDs")
	err := createFileIDs(fileIDPath, dryRun)
	if err != nil {
		fmt.Printf("Error occoured when trying to create file: %s\n", fileIDPath)
		return err
	}

	response, err := sdaclient.GetUsersFiles()
	if err != nil {
		return err
	}

	fmt.Println(response.Status)

	return nil
}

func createFileIDs(fileIDPath string, dryrun bool) error {
	if dryrun {
		return nil
	}

	if _, err := os.Stat(fileIDPath); err == nil {
		return ErrFileAlreadyExists
	} else if !os.IsNotExist(err) {
		return err
	}

	file, err := os.Create(fileIDPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}
