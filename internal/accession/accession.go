package accession

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/NBISweden/submitter/cmd"
	"github.com/NBISweden/submitter/internal/client"
	"github.com/NBISweden/submitter/internal/config"
	"github.com/spf13/cobra"
)

var dryRun bool

var accessionCmd = &cobra.Command{
	Use:   "accession [flags]",
	Short: "Trigger accession",
	Long:  "Trigger accession",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := config.NewConfig()
		if err != nil {
			return err
		}
		sdaclient := client.NewClient(*conf)
		err = CreateAccessionIDs(sdaclient, dryRun)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	cmd.AddCommand(accessionCmd)
	accessionCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Toggles dry-run mode. Dry run will not run any state changing API calls")
}

var ErrFileAlreadyExists = errors.New("file already exists")

type File struct {
	InboxPath  string `json:"inboxPath"`
	FileStatus string `json:"fileStatus"`
}

func CreateAccessionIDs(sdaclient *client.Client, dryRun bool) error {
	filePath := fmt.Sprintf("/data/%s-fileIDs.txt", sdaclient.DatasetFolder)
	file, err := createFileIDFile(filePath, dryRun)
	if err != nil {
		slog.Error("[accession] error occoured when trying to create file", "filePath", filePath)
		return err
	}
	defer file.Close() //nolint:errcheck

	response, err := sdaclient.GetUsersFiles()
	if err != nil {
		slog.Error("[accession] error when getting user files from sdaclient", "err", err)
		return err
	}
	if response.StatusCode != http.StatusOK {
		slog.Error("[accession] got non-ok response from sda api", "status_code", response.StatusCode)
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
			strings.Contains(f.InboxPath, sdaclient.DatasetFolder) &&
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
			"user":         sdaclient.UserID,
		})
		if err != nil {
			return err
		}

		resp, err := sdaclient.PostFileAccession(payload)
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

func GetVerifiedFilePaths(client *client.Client) ([]string, error) {
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
