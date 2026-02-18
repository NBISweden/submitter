# sda-bpctl

A tool that can be used to deal with administrative workflows for the big picture project. It supports three primary functions, making data ingestion, assigning accession ids to each ingested file, and creating a dataset for all files ingested with a accession id.

It can be used either locally as a cli tool or be packaged and run as a job in kubernetes. This is on one hand powerful and on the other hand, sometimes confusing and unintuitive, in an attempt to clarify the difference in logic based on how the tool is used the terms used will be **job** and **cli** in **bold** when describing the different logics.

The core functionallity of this tool is wrapping logic around the sensitive data archive (SDA) api, usually just referenced as 'the API' in this project. To fully understand how this is expected to work you should be familiar with the SDA and its api.

### installation

Build from source:

```bash
git clone git@github.com:NBISweden/sda-bpctl.git
cd sda-bpctl
go build -o bpctl .
./bpctl -h
```

### usage

The CLI have one required argument, called a **command** and non-required input arguments as flags. The rest of configuration is done through a config file. See more in the configuration section.

Commands must be one of:

- `ingest`
- `accession`
- `dataset`
- `mail`
- `job`
- `render`

#### examples

Some examples to demonstrate how the tool can be used

Running ingest 
```bash
./bpctl ingest
```

Running accession with a specific config.yaml
```bash
./bpctl accession --config /home/config.yaml
```

Running mail notification with the dry-run flag
```bash
./bpctl mail --dry-run
```

#### kubernetes job

The `job` command is meant to try and run all the steps of the dataset submission process in order; ingest -> accession -> dataset. It will need environment variables to configure and will need to be adjusted for the environment to run in. The `job` command also needs a input argument that represents the amount of files that is expected to be included in the finalized dataset. If at some point during the process this number does not match the job will fail and the user have to take over the process from that point.

specify your manifest, for example in a `job.yaml` or you can render a templated manifest based on your `config.yaml` using the `render` command

```bash
./bpctl render -o job.yaml
```

and apply it using `kubectl`:

```bash
kubectl apply -f job.yaml
```

will render a job.yaml manifest for you based on the configuration values you have supplied

### configuration

bpctl can consume configuration from either `config.yaml` or from environment variables. If both are supplied then the environment variables will take priority. If using config.yaml it is expected to be located in the root directory of the project. It can also be supplied by using the `--config` flag if located elsewhere.

see the `config.yaml.example` for a base template with what fields to fill
| Name | Example | Description | used by |
| --------------- | --------------- | --------------- | --------------- |
| USER_ID | "user-1234" | The user ID for the uploader, acts as identifier for the uploaded data | `ingest`, `accession`, `dataset`, `job`, `render` |
| DATASET_ID | "aa-Dataset-abc" | The ID that will be set for the finalized dataset, will be used during the `dataset` command | `ingest`, `accession`, `dataset`, `job`, `mail`, `render` |
| DATASET_FOLDER | "DATASET_ABC" | The folder where the uploaded data resides in s3inbox | `ingest`, `accession`, `dataset`, `job`, `mail`, `render` |
| JOB_TIMEOUT | 3 | A integer value, representing the number of minutes before the job times out when waiting for `accession` | `job`, `render` |
| JOB_POLL_RATE | 2 | A integer value, representing the number of minutes between each polling interval when waiting for `accession`, needs to be less than  the `JOB_TIMEOUT` value | `job`, `render` |
| JOB_EXPECTED_NR_FILES | 0 | The expected number of files to be part of the finalized dataset, set this when using `render` to include it in the rendered job.yaml | `job`, `render` |
| CLIENT_API_HOST | "https://api.example.com" | The hostname for the SDA API to communicate with | `ingest`, `accession`, `dataset`, `job` |
| CLIENT_ACCESS_TOKEN | "youraccesstoken" | The access token to authenticate towards the client api host | Yes | `ingest`, `accession`, `dataset`, `job` |
| MAIL_ADDRESS | "myemail@example.com" | Used for the `mail` command, this will be the email address the outgoing emails will be sent from | `mail` |
| MAIL_PASSWORD | "mypasswordemail" | Password associated with mail address | `mail` |
| MAIL_UPLOADER | "jane@example.com" | Mail address to the uploader, this is the address the outgoing email will be sent to | `mail` |
| MAIL_UPLOADER_NAME | "Jane Doe" | Name of the uploader | `mail` |
| MAIL_SMTP_HOST | "smtp.example.com" | Hostname to a mail server to relay mails through | `mail` |
| MAIL_SMTP_PORT | 587 | Port for the mail server | `mail` |
| CERT_SECRET_NAME | "cert-secret" | The name of the kubernetes secret that holds a tls certificate to use | `job`, `render` |

### ingest

```bash
./bpctl ingest [flags]
```

The ingest command will lookup all the files for the `USER_ID` that resides in `DATASET_FOLDER`, filter out all files that are not in either a directory `LANDING_PAGE` or `PRIVATE` and any file that does not have the event `uploaded`. 

**job:** the number of files retrieved will be compared to the `JOB_EXPECTED_NR_FILES` value, if they match the files will be sent for ingestion. If not the job will fail. 

**cli:** the files found will be sent for ingestion without evaluation. In that case the responsibility is on the user to ensure with a `--dry-run` before that the number of files are the desired amount.

Files sent to ingestion are done so trough the sda api `POST /ingest` endpoint.

### accession

```bash
./bpctl accession [flags]
```

The accession command will get a list of files for the `USER_ID` that resides in `DATASET_FOLDER` and have the event `verified`.

**job:** will poll the api according to `JOB_POLL_RATE` and wait until it finds the amount of files that matches `JOB_EXPECTED_NR_FILES` or until it times out according to `JOB_TIMEOUT`. When the expected number of files are found it will send a request to the sda api `POST /accession` endpoint with the files.

**cli:** will try create a file called `<DATASET_FOLDER>-fileIDs.txt` in the `--data-directory` directory. It will retrieve the list of files and after successful call to `POST /accession` it will write the accessionIDs to the file `<DATASET_FOLDER>-fileIDs.txt`. This is legacy logic owned from the `ingestor.sh` script and makes it so that you can store a intermediate state and keep track of the accession ids retrieved between runs of `accession` and `dataset`.

### dataset

```bash
./bpctl dataset [flags]
```

The dataset command will retrieve a list of accessionIDs and send a request to the sda api `POST /dataset`

**job:** will consume the list of accessionIDs by a in memory variable produced by the previous step in `accession` and send a list of files to be mapped to a dataset to `POST /dataset/create`

**cli:** will try to read from `<DATASET_FOLDER>-fileIDs.txt` to identify the files to be included in a dataset. If the file cannot be found it will make a call to `GET /user/files?path_prefix=<DATASET_FOLDER>` to find them and send a request to `POST /dataset/create` with the files.

### mail

```bash
./bpctl mail [flags]
```

Will send email notifications about dataset finalization to needed parties with information and attachments specifically for each.

`bigpicture submission`: recieves a mail about dataset creation and attachments with `dataset.xml` and `policy.xml`

`bigpicture PO`: recieves a mail about dataset creation and attachments with `rems.txt`, `dataset.txt` and `policy.xml`

`uploader`: recieves a mail confirming the creation of the dataset is completed with attachments `<datasetFolder>-stableIDs.txt`

**job:** will store `<datasetFolder>-stableIDs.txt` in memory during creation and relay it as attachments. Will consume the others from `data-directory/xml`

**cli:** will search for `<datasetFolder>-stableIDs.txt` in `data-directory/` and xml files in `data-directory/xml`

### job

```bash
./bpctl job <expectedFiles> [flags]
```

Will run all the above steps in order of ingest -> accession -> dataset and ensure the expectedFiles are keept between each stage where expectedFiles represents a integer value of the ammount of files that is expected to be present in the finalized dataset. If a dataset completes without error the job command will finish with sending out email notifications and move eventual landing pages from the inbox bucket to the public metadata bucket.

### testing

Unit tests using [pkg.go.dev/testing](https://pkg.go.dev/testing) 

Running all tests:
```bash
go test ./...
```
