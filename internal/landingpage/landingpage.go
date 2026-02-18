package landingpage

import (
	"fmt"
	"log/slog"

	"github.com/NBISweden/sda-bpctl/cmd"
	"github.com/NBISweden/sda-bpctl/internal/config"
	storage "github.com/NBISweden/sda-bpctl/internal/storage/client"
	"github.com/spf13/cobra"
)

var dryRun bool
var configPath string
var dataDirectory string

var storageCmd = &cobra.Command{
	Use:   "landingpage [flags]",
	Short: "Moves landingpages from inbox bucket to a dedicated metadata bucket",
	Long:  "Moves landingpages from inbox bucket to a dedicated metadata bucket",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.NewConfig(configPath)
		if err != nil {
			return err
		}

		archiveBucket := cfg.S3ArchiveBucket
		metadataBucket := cfg.S3MetadataBucket
		sslCaCert := cfg.SslCaCert

		archiveStorage, err := storage.NewMinio(cfg.S3ArchiveEndpoint, cfg.S3ArchiveAccessKey, cfg.S3ArchiveSecretKey, sslCaCert)
		if err != nil {
			return err
		}

		metadataStorage, err := storage.NewMinio(cfg.S3MetadataEndpoint, cfg.S3MetadataAccessKey, cfg.S3MetadataSecretKey, sslCaCert)
		if err != nil {
			return err
		}

		prefix := fmt.Sprintf("%s/%s", cfg.UserID, cfg.DatasetFolder)
		slog.Info("listing landing pages", "source", archiveBucket, "prefix", prefix)
		objects, err := archiveStorage.ListObjects(archiveBucket, prefix)
		if err != nil {
			return err
		}

		if len(objects) == 0 {
			slog.Info("No landing pages found")
		}

		for _, object := range objects {
			slog.Info("moving laning pages", "object", object.Key, "destination", metadataBucket)
			reader, err := archiveStorage.GetObject(archiveBucket, object.Key)
			if err != nil {
				return fmt.Errorf("failed to get %s from %s : %v", object.Key, archiveBucket, err)
			}
			err = metadataStorage.PutObject(metadataBucket, object.Key, reader, object.Size)
			reader.Close()
			if err != nil {
				return fmt.Errorf("failed to put %s to %s : %v", object.Key, metadataBucket, err)
			}
		}
		return nil
	},
}

func init() {
	cmd.AddCommand(storageCmd)
	storageCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Toggles dry-run mode. Dry run will not run any state changing API calls")
	storageCmd.Flags().StringVarP(&configPath, "config", "c", "config.yaml", "Path to configuration file")
	storageCmd.Flags().StringVarP(&dataDirectory, "dataDirectory", "d", "data", "Path to a directory to store itermidate data when needed")
}
