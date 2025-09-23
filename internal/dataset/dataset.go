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

		resp, err := client.PostDatasetCreate(jsonData)
		if err != nil {
			// Se comment bellow in sendInChunks() why this might be needed
			if errors.Is(err, io.ErrUnexpectedEOF) {
			} else {
				return err
			}
		}
		defer resp.Body.Close()
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
		resp, err := client.PostDatasetCreate(jsonData)
		/*
			As of 2025-09-17 we can get theese EOF responses when sending the accession id request to the sda api.
			However the request can still have been processed on the server side, but we won't get a response back in return.
			Therefore we still want to continue and send the rest of the batches. Unsure what causes this behaviour. It is not reproducable when
			running the sda stack locally. Probably something in the network setup that we need to figure out. For now we can live with it.
		*/
		if err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				continue
			}
			return err
		}
		defer resp.Body.Close()
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

	fmt.Println("[Dataset] Waiting on response from sda api ...")
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
