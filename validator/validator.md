### Script description
The `validator.sh` script can be used for
- validating the folder structure of a dataset
- validating all the metadata files against the xsd schemas
- checking that all the files referenced in the `images.xml` file exist in the inbox
- checking for exta or missing files in the inbox
- adding the datset id in the `dataset.xml` file and replacing the original one
- moving `PRIVATE` and `LANDING_PAGE` folders in the metadata bucket

### Running the script
Before running the script, the admin should login to the vault.
The script works for both prod and stage clusters and can be run with the following command:

```bash
./validator.sh -c <prod-or-staging> -u <username> -d <dataset-folder>
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
