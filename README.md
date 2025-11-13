# submitter

This project can be used as a tool to help with administrative tasks during dataset submission to the big picture project. It supports three primary functions, making data ingestion, assigning accession ids to each ingested file, and creating a dataset for all files ingested with a accession id.

It can also be run as a standalone job in kubernetes and try to complete the entire process.

### usage

The CLI have one requiered argument, called a **command** and non-requiered input arguments as flags. The rest of configuration is done through a config file. See more in the configuration section.

Commands must be one of:

- `ingest`
- `accession`
- `dataset`
- `mail`
- `job`

#### ingestion

TODO: Describe this

#### accession

TODO: Describe this

#### dataset

TODO: Describe this

#### job

TODO: Describe this

### configuration

submitter can consume configuration from either `config.yaml` or from environment variables. If both are supplied then the environment variables will take priority. If using config.yaml it is expected to be located in the root directory of the project

see the `config.yaml.example` or `job.yaml.example` for a base template with what fields to fill

### contribute

As of right now there are no explicit rules. Feel free to reach out if you have any questions `erik.zeidlitz@nbis.se`

### testing

Unit tests using [pkg.go.dev/testing](https://pkg.go.dev/testing) 

Running all tests:
```bash
go test ./...
