# submitter

This project can be used to to help with administrative tasks during dataset submission to the big picture project. This tool wraps a set of rules and buisseness logic around API calls to the big picture api. It uses the users privellages to do the needed operations.

## usage

The CLI requires a **command** and several input arguments. Commands must be one of:

- `ingest`
- `accession`
- `dataset`
- `mail`
- `all`

### arguments

| Flag              | Default                          | Required | Description                                                                 |
|-------------------|----------------------------------|----------|-----------------------------------------------------------------------------|
| `-dry-run`        | `true`                           | No       | Run without executing state changing API calls                              |
| `-config`         | `config.yaml`                    | No       | The config file with all input information and other needed metadata        |

### example

Running ingestion with specific input arguments:

```bash
./submitter \
  -config=/home/user/myconfig.yaml \
  -dry-run=false \
  ingest
```

### configuration

To run submitter a configuration file is needed with proper input, example of the config file: `config.yaml`

example: 
```yaml
Uploader: John Doe
UserID: testu@lifescience-ri.eu
DatasetID: aa-Dataset-benchmark-1k
DatasetFolder: DATASET_BENCHMARK_1K
Email: myemail@nbis.se
Password: mypassword
APIHost: api.bp.nbis.se
SMTPHost: tickets.nbis.se
SMTPPort: 587
S3Config: /home/erik/s3-secrets/s3cmd-inbox.conf
```

| Name          | Description                                          |
| ------------- | ---------------------------------------------------- |
| Uploader      | The name of the dataset uploader                     |
| UserID        | The user id of the uploader                          |
| DatasetID     | The dataset id                                       |
| DatasetFolder | The folder where the dataset resides                 |
| Email         | Your nbis email, used for sending out notifications  |
| Password      | Your password associated with your nbis email        |
| APIHost       | The hostname associated with the SDA api             |
| SMTPHost      | The smtp host, used for relaying email notifications |
| SMTPPort      | The port for the smtp host                           |
| S3Config      | The path where the s3cmd config resides              |


## testing

Unit tests using [pkg.go.dev/testing](https://pkg.go.dev/testing) 

Running all tests:
```bash
go test ./...
