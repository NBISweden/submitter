package dataset

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/NBISweden/submitter/pkg/sdaclient"
)

type Payload struct {
	FileIDs []string `json:"fileIDs"`
}

func CreateDataset(sdaClient *sdaclient.Client, fileIDsPath string, dryRun bool) error {

	var fileIDsList []string

	// We can keep this stored in memory instead of reading from file. TODO: Refactor this.
	file, err := os.Open(fileIDsPath)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Println("Reading ", fileIDsPath)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fileIDsList = append(fileIDsList, scanner.Text())
	}
	fmt.Println("[Dataset] Number of files included in dataset:", len(fileIDsList))
	if dryRun {
		fmt.Println("[Dry-Run] No datasets will be created")
		return nil
	}

	if len(fileIDsList) > 100 {
		err := sendInChunks(fileIDsList, sdaClient)
		if err != nil {
			return err
		}
	}

	if len(fileIDsList) <= 100 {
		payload := Payload{
			FileIDs: fileIDsList,
		}
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		_, err = sdaClient.PostDatasetCreate(jsonData)
		if err != nil {
			return err
		}
	}

	return nil
}

func sendInChunks(fileIDsList []string, sdaClient *sdaclient.Client) error {
	fmt.Println("[Dataset] More than 100 entries. Sending in chunks of 100")
	chunks := slices.Chunk(fileIDsList, 100)
	for chunk := range chunks {
		payload := Payload{
			FileIDs: chunk,
		}
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		_, err = sdaClient.PostDatasetCreate(jsonData)
		if err != nil {
			return err
		}
	}
	return nil
}
