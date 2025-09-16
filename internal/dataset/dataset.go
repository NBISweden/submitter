package dataset

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/NBISweden/submitter/pkg/sdaclient"
	"github.com/schollz/progressbar/v3"
)

var ErrFileAlreadyExists = errors.New("File already exists")

type Payload struct {
	AccessionIDs []string `json:"accession_ids"`
	DatasetID    string   `json:"dataset_id"`
	User         string   `json:"user"`
}

// TODO: Unify the naming here. Stable ID and Accession ID is interchanged?
type StableID struct {
	AccessionID string `json:"accessionID"`
	InboxPath   string `json:"inboxPath"`
}

func CreateDataset(client *sdaclient.Client, dryRun bool) error {
	if !dryRun {
		err := createStableIDsFile(client)
		if err != nil {
			fmt.Println("[Dataset] failed to create file with stable ids")
		}
	}

	var fileIDsList []string
	filePath := fmt.Sprintf("data/%s-fileIDs.txt", client.DatasetFolder)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Printf("[Dataset] Reading %s\n", filePath)
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

	fmt.Println("[Dataset] creation of dataset completed!")
	return nil
}

func sendInChunks(fileIDsList []string, client *sdaclient.Client) error {
	fmt.Println("[Dataset] More than 100 entries. Sending in chunks of 100")
	chunks := slices.Chunk(fileIDsList, 100)
	allChunks := slices.Collect(chunks)
	totalChunks := len(allChunks)
	bar := progressbar.Default(int64(totalChunks), "[Dataset] Creating dataset")
	for _, chunk := range allChunks {
		bar.Add(1)
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
		if r.StatusCode != 200 {
			fmt.Printf("[Dataset] bad response from SDA API %s for %s", r.Status, chunk)
		}
	}
	return nil
}

func createStableIDsFile(client *sdaclient.Client) error {
	filePath := fmt.Sprintf("data/%s-stableIDs.txt", client.DatasetFolder)
	if _, err := os.Stat(filePath); err == nil {
		return ErrFileAlreadyExists
	} else if !os.IsNotExist(err) {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	r, err := client.GetUsersFilesWithPrefix()
	if err != nil {
		return err
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var stableIDs []StableID
	if err := json.Unmarshal(body, &stableIDs); err != nil {
		return err
	}

	for _, f := range stableIDs {
		fmt.Fprintf(file, "%s %s\n", f.AccessionID, f.InboxPath)
	}

	fmt.Printf("[Dataset] Created file with stable ids in %s\n", filePath)
	return nil
}
