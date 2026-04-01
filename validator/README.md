### Script description
The `validator.sh` script can be used for
- validating the folder structure of a dataset
- validating all the metadata files against the xsd schemas
- checking that all the files referenced in the `images.xml` file exist in the inbox
- checking for exta or missing files in the inbox
- adding the datset id in the `dataset.xml` file and replacing the original one
- moving `PRIVATE` and `LANDING_PAGE` folders in the metadata bucket
- create kubernetes job manifest and deploy the job in case of prod cluster

### Running the script
Before running the script, the admin should login to the vault.
The validation works for both production and staging clusters but the atomatic ingestion
process works **ONLY** for the production cluster.

#### Validation only
The script works for both prod and stage clusters and can be run with the following command:

```bash
./validator.sh -c <prod-or-staging> -u <username> -d <dataset-folder> --validation-only
```

where:
- `prod-or-staging` is the cluster where the dataset is located
- `username` is the username folder name in inbox
- `dataset-folder` is the dataset folder which should be in the form `DATASET_{identifier}`.

To perform a dry run, which simulates the script execution without making any changes, include the --dry-run option. The data will still be downloaded to the local machine.

```bash
./validator.sh -c <prod-or-staging> -u <username> -d <dataset-folder> --dry-run
```

In that case, the script will not modify the dataset.xml file and it will not move the `PRIVATE` and `LANDING_PAGE` folders
in the metadata bucket if the validation is **successfull**.

After a successfull run, the admin can remove all the files from the local machine (except the `dataset_id.txt` file)
by running the following command:

```bash
./validator.sh --clean
```

#### Validation and ingestion
Since in the case a kubernetes job will be deployed then the user needs to set the kubeconfig file:

```bash
export KUBECONFIG=path/to/kubeconfig/yaml/file
```

The ingestion works **ONLY** for the production cluster and can by run with the following command:

```bash
./validator.sh -c prod -u <username> -d <dataset-folder> -n <full-name> -e <email> -t <admin-token>
```

where:
- `prod` is the cluster where the dataset is located
- `username` is the username folder name in inbox
- `dataset-folder` is the dataset folder which should be in the form `DATASET_{identifier}`.
- `full-name` is the first and last name of the uploader as it is in the ticket
- `email` is the email address of the uploader
- `admin-token` is the download token which allows interaction with the admin API 

After the validation a kubernetes manifest (yaml file) will be created with the name of the dataset folder.

