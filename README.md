# submitter
This project can be used to to help with administrative tasks during dataset submission to the big picture project. This tool wraps a set of rules and buisseness logic around API calls to the big picture api. It uses the users privellages to do the needed operations.

### usage

The CLI have one requiered argument, called a **command** and non-requiered input arguments as flags. The rest of configuration is done through a config file. See more in the configuration section.

Commands must be one of:

- `ingest`
- `accession`
- `dataset`
- `mail`
- `job`

### arguments

| Flag              | Default                          | Required | Description                                                                 |
|-------------------|----------------------------------|----------|-----------------------------------------------------------------------------|
| `-dry-run`        | `true`                           | No       | Run without executing state changing API calls                              |
| `-config`         | `config.yaml`                    | No       | The config file with all input information and other needed metadata        |

### example

Running ingestion:

```bash
./submitter \
  -dry-run=false \
  ingest
```

### configuration

submitter can consume configuration from either `config.yaml` or from environment variables. If both are supplied then the environment variables will take priority. If using config.yaml it is expected to be located in the root directory of the project

| Name          | Description                                           |
| ------------- | ----------------------------------------------------- |
| UserID        | The user id of the uploader                           |
| Uploader      | The name of the dataset uploader                      |
| UploaderEmail | Email address of the uploader                         |
| DatasetID     | The dataset id                                        |
| DatasetFolder | The folder where the dataset resides                  |
| FileIdFolder  | The folder to store files with dataset ids in         |
| Email         | Your nbis email, used for sending out notifications   |
| Password      | Your password associated with your nbis email         |
| APIHost       | The hostname associated with the SDA api              |
| SMTPHost      | The smtp host, used for relaying email notifications  |
| SMTPPort      | The port for the smtp host                            |
| AccessToken   | Access Token for the SDA API                          |
| UseTLS        | If set, will try to setup TLS connection              |
| SSLCACert     | The location of the ca cert to use for TLS connection |

### contribute

As of right now there are no explicit rules. Feel free to reach out if you have any questions `erik.zeidlitz@nbis.se`

### testing

Unit tests using [pkg.go.dev/testing](https://pkg.go.dev/testing) 

Running all tests:
```bash
go test ./...
