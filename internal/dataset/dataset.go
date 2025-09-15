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
	AccessionIDs []string `json:"accession_ids"`
	DatasetID    string   `json:"dataset_id"`
	User         string   `json:"user"`
}

func CreateDataset(client *sdaclient.Client, dryRun bool) error {

	var fileIDsList []string

	filePath := fmt.Sprintf("data/%s-fileIDs.txt", client.DatasetFolder)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Println("[Dataset] Reading ", filePath)
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
		err := sendInChunks(fileIDsList, client)
		if err != nil {
			return err
		}
	}

	if len(fileIDsList) <= 100 {
		payload := Payload{
			AccessionIDs: fileIDsList,
			DatasetID:    client.DatasetID,
			User:         client.UserID,
		}
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		fmt.Printf("[Dataset] Sending payload:\n%s\n", string(jsonData))
		r, err := client.PostDatasetCreate(jsonData)
		if err != nil {
			return err
		}
		fmt.Printf("[Dataset] Response from SDA API: %s\n", r.Status)
	}

	return nil
}

func sendInChunks(fileIDsList []string, client *sdaclient.Client) error {
	fmt.Println("[Dataset] More than 100 entries. Sending in chunks of 100")
	chunks := slices.Chunk(fileIDsList, 100)
	allChunks := slices.Collect(chunks)
	totalChunks := len(allChunks)
	for i, chunk := range allChunks {
		payload := Payload{
			AccessionIDs: chunk,
			DatasetID:    client.DatasetID,
			User:         client.UserID,
		}
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		r, err := client.PostDatasetCreate(jsonData)
		if err != nil {
			return err
		}
		fmt.Printf("[Dataset] Response from SDA API: %s (%d/%d)\n", r.Status, i, totalChunks)
	}
	return nil
}
