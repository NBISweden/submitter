package landingpage

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"path"
	"strings"

	"github.com/NBISweden/sda-bpctl/cmd"
	"github.com/NBISweden/sda-bpctl/internal/config"
	storage "github.com/NBISweden/sda-bpctl/internal/storage"
	storageClient "github.com/NBISweden/sda-bpctl/internal/storage/client"
	"github.com/neicnordic/crypt4gh/keys"
	"github.com/neicnordic/crypt4gh/streaming"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/chacha20poly1305"
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

		err = Run(cfg)
		if err != nil {
			return err
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

func Run(cfg *config.Config) error {
	bucketUserID := strings.ReplaceAll(cfg.UserID, "@", "_")
	datasetFolder := cfg.DatasetFolder
	inboxBucket := cfg.S3InboxBucket
	metadataBucket := cfg.S3MetadataBucket
	sslCaCert := cfg.SslCaCert

	privateKey, err := loadKey(cfg)
	if err != nil {
		return err
	}

	inboxStorage, err := storageClient.NewMinio(cfg.S3InboxEndpoint, cfg.S3InboxAccessKey, cfg.S3InboxSecretKey, sslCaCert)
	if err != nil {
		return err
	}

	metadataStorage, err := storageClient.NewMinio(cfg.S3MetadataEndpoint, cfg.S3MetadataAccessKey, cfg.S3MetadataSecretKey, sslCaCert)
	if err != nil {
		return err
	}

	objects, err := getLandingPages(bucketUserID, datasetFolder, inboxBucket, inboxStorage)
	if err != nil {
		return err
	}

	slog.Info("found landing pages", "nr_objects", len(objects))

	for _, object := range objects {
		if dryRun {
			slog.Info("found", "object", object.Key)
			continue
		}

		encryptedReader, err := inboxStorage.GetObject(inboxBucket, object.Key)
		if err != nil {
			return err
		}
		defer encryptedReader.Close()

		decryptedReader, err := streaming.NewCrypt4GHReader(encryptedReader, privateKey, nil)
		if err != nil {
			return err
		}
		defer decryptedReader.Close()

		inboxLocation := object.Key
		prefix := fmt.Sprintf("%s/%s", bucketUserID, datasetFolder)
		remainingPath := strings.TrimPrefix(object.Key, prefix)
		metadataLocation := path.Join("datasets", cfg.DatasetID, remainingPath)
		metadataLocation = strings.TrimSuffix(metadataLocation, ".c4gh")
		err = move(inboxBucket, inboxLocation, metadataBucket, metadataLocation, encryptedReader, decryptedReader, inboxStorage, metadataStorage)
		if err != nil {
			return fmt.Errorf("could not move object %s : %v", object.Key, err)
		}
	}

	if dryRun {
		slog.Info("dry run enabled, no objects moved")
	}

	return nil
}

func loadKey(cfg *config.Config) ([chacha20poly1305.KeySize]byte, error) {
	pemReader := bytes.NewReader([]byte(cfg.C4GHSecPem))
	passphraseBytes := []byte(cfg.C4GHPassphrase)
	key, err := keys.ReadPrivateKey(pemReader, passphraseBytes)
	if err != nil {
		return [chacha20poly1305.KeySize]byte{}, err
	}
	return key, nil
}

func getLandingPages(bucketUserID, datasetFolder, bucketName string, inboxStorage storage.StorageClient) ([]storage.ObjectInfo, error) {
	prefix := fmt.Sprintf("%s/%s/%s", bucketUserID, datasetFolder, "LANDING_PAGE")
	slog.Info("listing landing pages", "source_bucket", bucketName, "prefix", prefix)
	objects, err := inboxStorage.ListObjects(bucketName, prefix)
	if err != nil {
		return nil, err
	}

	return objects, nil
}

func move(sourceBucket, sourceLocation, destinationBucket, destinationLocation string, sourceReader, destinationReader io.Reader, sourceClient, destinationClient storage.StorageClient) error {
	slog.Info("uploading", "destination_bucket", destinationBucket, "object_location", destinationLocation)
	err := destinationClient.PutObject(destinationBucket, destinationLocation, destinationReader, -1)
	if err != nil {
		return fmt.Errorf("failed to put %s to %s : %v", destinationLocation, destinationBucket, err)
	}

	slog.Info("deleting", "source_bucket", sourceBucket, "source_location", sourceLocation)
	err = sourceClient.RemoveObject(sourceBucket, sourceLocation, sourceReader)
	if err != nil {
		return fmt.Errorf("failed to delete %s from %s: %v", sourceLocation, sourceBucket, err)
	}

	return nil
}
