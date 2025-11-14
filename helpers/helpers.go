package helpers

import (
	"fmt"
)

func GetFileIDsPath(dataDirectory string, datasetFolder string) string {
	return fmt.Sprintf("%s/%s-fileIDs.txt", dataDirectory, datasetFolder)
}

func GetStableIDsPath(dataDirectory string, datasetFolder string) string {
	return fmt.Sprintf("%s/%s-stableIDs.txt", dataDirectory, datasetFolder)
}
