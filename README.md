# submitter

This project can be used to to help with administrative tasks during dataset submission

## Usage

The CLI requires a **command** and several input arguments. Commands must be one of:

- `ingest`
- `accession`
- `dataset`

### Arguments

| Flag              | Default                          | Required | Description                                                                 |
|-------------------|----------------------------------|----------|-----------------------------------------------------------------------------|
| `-dry-run`        | `true`                           | No       | Run without executing API calls (simulation mode).                          |
| `-api-host`       | `https://api.bp.nbis.se`         | No       | The Big Picture API URL.                                                    |
| `-config`         | `s3cmd.conf`                     | No       | The `s3cmd` config file.                                                    |
| `-user-id`        | *(none)*                         | Yes      | The User ID of the uploader/submitter.                                      |
| `-dataset-id`     | *(none)*                         | Yes      | The ID of the uploaded dataset.                                             |
| `-dataset-folder` | *(none)*                         | Yes      | The folder in `s3inbox` where uploaded files reside.                        |

### Example

Running ingestion with specific input arguments:

```bash
./submitter \
  -user-id=user123 \
  -dataset-id=ds001 \
  -dataset-folder=/path/to/folder \
  -api-host=https://api.bp.nbis.se \
  -config=custom_s3cmd.conf \
  -dry-run=false \
  ingest
```

## testing

Unit tests using [pkg.go.dev/testing](https://pkg.go.dev/testing) 

Running all tests:
```bash
go test ./...
