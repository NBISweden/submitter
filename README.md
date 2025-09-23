<div align="center">
:construction: This project is still under development and things might be buggy or work in unexpected ways :construction:
</div>

# submitter
This project can be used to to help with administrative tasks during dataset submission to the big picture project. This tool wraps a set of rules and buisseness logic around API calls to the big picture api. It uses the users privellages to do the needed operations.

### usage

The CLI have one requiered argument, called a **command** and non-requiered input arguments as flags. The rest of configuration is done through a config file. See more in the configuration section.

Commands must be one of:

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

Running ingestion:

```bash
./submitter \
  -config=/home/user/myconfig.yaml \
  -dry-run=false \
  ingest
```

### configuration

To run submitter a configuration file is needed with proper input, example: 

```yaml
UserID: testu@lifescience-ri.eu
Uploader: John Doe
UploaderEmail: johndoe@email.com
DatasetID: aa-Dataset-benchmark-1k
DatasetFolder: DATASET_BENCHMARK_1K
Email: myemail@nbis.se
Password: mypassword
APIHost: https://api.bp.nbis.se
SMTPHost: tickets.nbis.se
SMTPPort: 587
S3Config: /home/user/s3cmd.conf
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

### good to know

`question:` What's the deal with the brackets in all the prints?

`anwser:` Since the goal is to have a end-to-end flow of all stages ingestion -> accession -> dataset I feelt it was usefull during testing to understand what package logged something. So I added the package name in brackets whenever something is logged / printed. That way it was easier to understand which part of the flow executed or failed. This can be reworked and / or retought at some point.

`question:` What's the deal with the -dry-run flag? 

`anwser:` Since many iterations and testing is requiered when developing this project I never wanted to accedentally send something or change db states. To avoid this the dry-run flag is (as of this writing) default set to true. This means if you want to actually execute something you need to run `./submitter --dry-run=false <COMMAND>` for it to take full effect.

`question:` How do I contribute?

`answer:` As of right now there are no explicit rules. Feel free to reach out if you have any questions `erik.zeidlitz@nbis.se`

### testing

Unit tests using [pkg.go.dev/testing](https://pkg.go.dev/testing) 

Running all tests:
```bash
go test ./...
