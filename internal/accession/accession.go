package accession

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"os"
	"strings"

	"github.com/NBISweden/submitter/cmd"
	"github.com/NBISweden/submitter/helpers"
	"github.com/NBISweden/submitter/internal/client"
	"github.com/NBISweden/submitter/internal/config"
	"github.com/spf13/cobra"
)

var dryRun bool
var configPath string
var dataDirectory string
var datasetFolder string

var accessionCmd = &cobra.Command{
	Use:   "accession [flags]",
	Short: "Trigger accession",
	Long:  "Trigger accession",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := config.NewConfig(configPath)
		if err != nil {
			return err
		}
		api, err := client.New(configPath)
		err = CreateAccessionIDs(api, conf.DatasetFolder, conf.UserID)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	cmd.AddCommand(accessionCmd)
	accessionCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Toggles dry-run mode. Dry run will not run any state changing API calls")
	accessionCmd.Flags().StringVar(&configPath, "config", "config.yaml", "Path to configuration file")
	accessionCmd.Flags().StringVar(&dataDirectory, "data-directory", "data", "Path to directory to write / read intermediate files for stableIDs and fileIDs")
}

func CreateAccessionIDs(api client.APIClient, datasetFolder string, userID string) error {
	filePath := helpers.GetFileIDsPath(dataDirectory, datasetFolder)
	file, err := createFileIDFile(filePath, dryRun)
	if err != nil {
		slog.Error("[accession] error occoured when trying to create file", "filePath", filePath)
		return err
	}
	defer file.Close() //nolint:errcheck

	files, err := api.GetUsersFiles()

	var paths []string
	for _, f := range files {
		if f.Status == "verified" &&
			strings.Contains(f.InboxPath, datasetFolder) &&
			!strings.Contains(f.InboxPath, "PRIVATE") {
			paths = append(paths, f.InboxPath)
		}
	}
	slog.Info("[accession] files found for accession id creation", "files_found", len(paths))

	if dryRun {
		slog.Info("[accession] dry-run enabled, no files will be given accession ids")
		return nil
	}

	for _, filepath := range paths {
		accessionID, err := generateAccessionID()
		if err != nil {
			return err
		}

		payload, err := json.Marshal(map[string]string{
			"accession_id": accessionID,
			"filepath":     filepath,
			"user":         userID,
		})
		if err != nil {
			return err
		}

		resp, err := api.PostFileAccession(payload)
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

	slog.Info("[accession] accession IDs assigned", "nr_files", len(paths))

	return nil
}

func createFileIDFile(fileIDPath string, dryrun bool) (*os.File, error) {
	if dryrun {
		return nil, nil
	}

	if _, err := os.Stat(fileIDPath); err == nil {
		return nil, fmt.Errorf("file already exists")
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
