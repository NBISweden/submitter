#!/bin/bash

# Function for echoing colored text
# At the end it resets the color to default
cecho() {
    local color="$1"
    local text="$2"

    case $color in
        red)     color="\033[31m";;
        green)   color="\033[32m";;
        yellow)  color="\033[33m";;
        blue)    color="\033[34m";;
        magenta) color="\033[35m";;
        *)       color="\033[39m";;
    esac

    # If the output is a terminal, print with colors; else just print text
    if [ -t 2 ]; then
        printf "%b%s%b\n" "$color" "$text" "$reset"
    else
        printf "%s\n" "$text"
    fi
}

# Check if the tools that are needed are installed
if [ ! "$(command -v s3cmd)" ];then
    cecho red "s3cmd command does not exist"
    exit 1
fi

if [ ! "$(command -v vault)" ];then
    cecho red "vault command does not exist"
    exit 1
fi

if [ ! "$(command -v xmllint)" ];then
    cecho red "xmllint command does not exist"
    exit 1
fi

if [ ! "$(command -v crypt4gh)" ];then
    cecho red "crypt4gh command does not exist"
    exit 1
fi

if [ ! "$(command -v kubectl)" ];then
    cecho red "kubectl command does not exist"
    exit 1
fi

# Determine whic crypt4gh version the user has (python or go)
C4GHGEN=$(crypt4gh generate 2>&1)
if [[ $C4GHGEN != *"the required flag"* ]]; then
    c4gh_decrypt() {
        local sk="$1"
        local file="$2"
        crypt4gh decrypt --sk "$sk" < "$file" > "${file%.c4gh}"
    }
    c4gh_encrypt() {
        local sk="$1"
        local pk="$2"
        local file="$3"
        crypt4gh encrypt --sk "$sk" --recipient_pk "$pk" < "$file" > "$file.c4gh"
    }
else
    c4gh_decrypt() {
        local sk="$1"
        local file="$2"
        crypt4gh decrypt -s "$sk" -f "$file"
    }
    c4gh_encrypt() {
        local sk="$1"
        local pk="$2"
        local file="$3"
        crypt4gh encrypt -s "$sk" -p "$pk" -f "$file"
    }
fi

# OS-specific configurations
if [[ "$OSTYPE" == "darwin"* ]]; then
    NUMFMT="gnumfmt"
    sed_i() { sed -i '' "$@"; }
    sed_i_bak() { sed -i '.bak' "$@"; }
else
    NUMFMT="numfmt"
    sed_i() { sed -i "$@"; }
    sed_i_bak() { sed -i.bak "$@"; }
fi

if [ ! "$(command -v $NUMFMT)" ];then
    cecho red "$NUMFMT command does not exist. On macOS, you can install it with 'brew install coreutils'."
    exit 1
fi

crypt4gh_required_version="1.9.0"
crypt4gh_current_version=$(crypt4gh -v)

if [ "$(printf '%s\n' "$crypt4gh_required_version" "$crypt4gh_current_version" | sort -V | head -n 1)" != "$crypt4gh_required_version" ]; then
    cecho red "crypt4gh version must be at least $crypt4gh_required_version (found $crypt4gh_current_version)"
    exit 1
fi  
INBOX_ACCESS_KEY=""
INBOX_SECRET_KEY=""
HOST_BUCKET=""
INBOX_BUCKET=""
METADATA_ACCESS_KEY=""
METADATA_SECRET_KEY=""
C4GH_PASSPHRASE=""
METADATA_BUCKET=""
ERROR_STATUS=0
PRIVATE_FOLDER=true
LANDING_PAGE=false
DRY_RUN=false
VALIDATION_ONLY=false
MIN_FILE_SIZE=152
ORG_NAME=""
NAME=""
EMAIL=""
ACCESS_TOKEN=""
version=""
dataset_id=""

function help {
    cat << END_USAGE
    USAGE: $0 -c <cluster> -u <user-id> -d <dataset-name> -n <user-name> -e <user-email> or $0 --clean
    parameters:
    -c, --cluster       Cluster name (prod or staging)
    -u, --user          Username folder in the inbox bucket
    -d, --dataset       Dataset folder (or path) name in the inbox bucket
    -n, --name          The actual name of the uploader
    -e, --email         The actual e-mail of the uploader (for sending email)
    -t, --token         Token for accessing the admin API
    --dry-run           Flag for running only the validation without modifying and moving metadata
    --validation-only   Flag for running only the validation
    --clean             Clean up the files that are created by the script (except the dataset_id.txt file)
END_USAGE
    exit 1
}

# Function for cleaning up the files that are created by the script
# - Removes the xsd and xml folders
# - Removes the crypt4gh private key
# - Removes the crypt4gh private key passphrase from env
# - Removes the json files
# - Removes the public key
# - Removes the error files
# - Removes the general_errors.logs file
function cleanup {
    cecho yellow "Cleaning up ..."

    rm -rf xsd-files xml-files PRIVATE LANDING_PAGE

    unset C4GH_PASSPHRASE

    rm -f bp_key.pub

    rm -f general_errors.logs

    rm -f *.error

    cecho green "Done"

    exit 0
}

function remove_private_key {
    cecho yellow "Removing private key ..."
    rm -f c4gh.sec.pem
    cecho green "Private key removed"
}

#parse input
while (( "$#" )); do
    case "$1" in
        -c|--cluster)
            shift
            cluster="$1"
            ;;
        -u|--user)
            shift
            user="$1"
            ;;
        -d|--dataset)
            shift
            dataset="$1"
            ;;
        -n|--name)
            shift
            NAME="$1"
            while [[ -n "$2" && "$2" != -* ]]; do
                NAME="$NAME $2"
                shift
            done
            # Trim leading and trailing whitespace
            NAME="${NAME#"${NAME%%[![:space:]]*}"}"
            NAME="${NAME%"${NAME##*[![:space:]]}"}"
            ;;
        -e|--email)
            shift
            EMAIL="$1"
            ;;
        -t|--token)
            shift
            ACCESS_TOKEN="$1"
            ;;
        --dry-run)
            DRY_RUN=true
            ;;
        --validation-only)
            VALIDATION_ONLY=true
            ;;
        -h|--help)
            help
            ;;
        --clean)
            cleanup
            ;;
    esac
    shift
done

# Check if all arguments are in place
if [ -z "$user" ];then
    cecho red "ERROR: No user given"
    help
fi

if [ -z "$cluster" ];then
    cecho red "ERROR: No cluster given"
    help
fi

if [ -z "$dataset" ];then
    cecho red "ERROR: No dataset given"
    help
fi

if [ "$DRY_RUN" == "false" ] && [ "$VALIDATION_ONLY" == "false" ]; then
    if [ -z "$NAME" ];then
        cecho red "ERROR: No user name given"
        help
    fi

    if [ -z "$EMAIL" ];then
        cecho red "ERROR: No user email given"
        help
    fi

    if [ -z "$ACCESS_TOKEN" ];then
        cecho red "ERROR: No access token given"
        help
    fi
fi

if [ "$DRY_RUN" == "true" ] && [ "$VALIDATION_ONLY" == "true" ]; then
    cecho red "ERROR: cannot use both --dry-run and --validation-only flags"
    help
fi

# Function for removing the leading and trailing "/" from the user and dataset
# and replacing the "@" with "_" in the user if needed
function sanitize_user_dataset {
    user=${user#/}; user=${user%/}; user=${user//@/_}
    dataset=${dataset#/}; dataset=${dataset%/}
}

# Function for getting the xsd files from github repo
# - Gets the download urls of the xsd files from github api
# - Downloads the xsd files by using the download urls
# - Puts the files in the xsd-files directory
function get_xsd_files {
    cecho yellow "Getting xsd files ..."
    mkdir -p xsd-files

    curl -H "Accept: application/vnd.github.v3+json" \
       https://api.github.com/repos/imi-bigpicture/bigpicture-metaflex/contents/src?ref=$version.0.0 | jq -r '.[] | .download_url' |
    while IFS= read -r url; do
        xsd_name=$(basename "$url")
        curl -o "xsd-files/${xsd_name%\?*}" -J -L "$url" >/dev/null 2>&1
    done

    cecho green "Done"
}

# Helper function for making the s3cmd commands in inbox for the prod cluster
# and for all buckets in the staging cluster
function s3cmd_command {
    command s3cmd --host="$HOST_BUCKET" \
    --host-bucket="$HOST_BUCKET" \
    --access_key="$INBOX_ACCESS_KEY" \
    --secret_key="$INBOX_SECRET_KEY" \
    "$@"
}

# Helper function for making the s3cmd commands in metadata bucket for the prod cluster
function s3cmd_metadata {
    command s3cmd --host="$HOST_BUCKET" \
    --host-bucket="$HOST_BUCKET" \
    --access_key="$METADATA_ACCESS_KEY" \
    --secret_key="$METADATA_SECRET_KEY" \
    "$@"
}

function get_credentials {
    if [[ "$cluster" == "prod" ]]; then
        HOST_BUCKET=$(vault kv get -field=endpoints bp-secrets/S3_keys/STO2 | cut -d, -f2)
        INBOX_ACCESS_KEY=$(vault kv get -field=access_key bp-secrets/S3_keys/STO2/inbox)
        INBOX_SECRET_KEY=$(vault kv get -field=secret_key bp-secrets/S3_keys/STO2/inbox)
        INBOX_BUCKET=$(s3cmd_command ls | cut -d'/' -f3 | grep -v "staging")
        METADATA_ACCESS_KEY=$(vault kv get -field=access_key bp-secrets/S3_keys/STO2/private)
        METADATA_SECRET_KEY=$(vault kv get -field=secret_key bp-secrets/S3_keys/STO2/private)
        METADATA_BUCKET=$(s3cmd_metadata ls | cut -d'/' -f3)
    elif [[ "$cluster" == "staging" ]]; then
        INBOX_ACCESS_KEY=$(vault kv get -field=access_key bp-secrets/S3_keys/STO2/inbox)
        INBOX_SECRET_KEY=$(vault kv get -field=secret_key bp-secrets/S3_keys/STO2/inbox)
        HOST_BUCKET=$(vault kv get -field=endpoints bp-secrets/S3_keys/STO2 | cut -d, -f2)
        INBOX_BUCKET="staging-inbox"
        METADATA_BUCKET="bigpicture-test-metadata"
    else
        cecho red "ERROR: Cluster name is not valid"
        exit 1
    fi

    C4GH_PASSPHRASE=$(vault kv get -field=password bp-secrets/crypt4gh)
}

function validate_private_files {
    cecho yellow "Validating PRIVATE files ..."
    private_files=("$@")
    # Depending on the version the expected private files are different
    # If the number of private files is not the expected then exit
    if [[ $version == "v1" ]]; then
        expected_private_files=("DAC" "submission")
        cecho yellow "Expected private files: ${expected_private_files[*]}"
        # check the length is 2
        if [[ ${#private_files[@]} -ne 2 ]]; then
            cecho red "ERROR: The number of private files is not 2" | tee -a general_errors.logs
            ERROR_STATUS=1
        fi
    else
        if [[ ${#private_files[@]} -eq 2 ]]; then
            expected_private_files=("rems" "organisation")
            cecho yellow "Expected private files: ${expected_private_files[*]}"
        elif [[ ${#private_files[@]} -eq 3 ]]; then
            expected_private_files=("rems" "organisation" "datacite")
            cecho yellow "Expected private files: ${expected_private_files[*]}"
        else
            cecho red "ERROR: The number of private files are not 2 or 3" | tee -a general_errors.logs
            ERROR_STATUS=1
        fi
    fi

    # Check if the expected_private_files exist in the private_files
    for private_file in "${expected_private_files[@]}"; do
        if [[ ! " ${private_files[@]} " =~ " ${private_file}."* ]]; then
            cecho red "ERROR: Private file $private_file is missing" | tee -a general_errors.logs
            ERROR_STATUS=1
        fi
    done
    cecho green "Done"
}

function validate_files {
    cecho yellow "Validating METADATA folder files ..."
    expected_metadata_content=("dataset" "policy" "image" "annotation" "observation" "sample" "staining")
    inbox_metadata_files=$(s3cmd_command ls s3://"$INBOX_BUCKET"/"$user"/"$dataset"/METADATA/ | awk -F'_lifescience-ri.eu/|_elixir-europe.org/' '{print $2}' | cut -d'/' -f3 | tr '\n' ' ')
    # Convert multiple lines to array
    read -a metadata_files <<<"$inbox_metadata_files"
    inbox_private_files=$(s3cmd_command ls s3://"$INBOX_BUCKET"/"$user"/"$dataset"/PRIVATE/ | awk -F'_lifescience-ri.eu/|_elixir-europe.org/' '{print $2}' | cut -d'/' -f3 | tr '\n' ' ')
    # Convert multiple lines to array
    read -a metadata_private_files <<<"$inbox_private_files"
    # if the folder ANNOTATIONS is missing, make sure annotation.xml and observer.xml are not present
    if [[ "$1" == "false" ]]; then
        # Remove the annotation and observer from the expected_metadata_content
        unset 'expected_metadata_content[3]'
        expected_metadata_content=("${expected_metadata_content[@]}")
        # Check if "annotation" or "observer" is in the meatadata_files
        if [[ " ${metadata_files[@]} " =~ "annotation" ]]; then
            cecho red "ERROR: 'annotation.xml' should not exist because the 'ANNOTATIONS' directory is missing." | tee -a general_errors.logs
            ERROR_STATUS=1
        fi
    fi
    
    # Check if any metadata files are missing
    for expected_metadata in "${expected_metadata_content[@]}"; do
        if [[ ! " ${metadata_files[@]} " =~ " ${expected_metadata}."* ]]; then
            cecho red "ERROR: Metadata file $expected_metadata is missing" | tee -a general_errors.logs
            ERROR_STATUS=1
        fi
    done

    # Check if there are any extra metadata files
    for metadata_file in "${metadata_files[@]}"; do
        # Keep only the basename of the file without the extension
        file_basename=$(basename "$metadata_file" | cut -d'.' -f1)
        # If the metadata file is the observer then continue because it is not mandatory
        if [[ "$file_basename" == "observer" ]]; then
            continue
        fi
        if [[ ! " ${expected_metadata_content[@]} " =~ " ${file_basename}."* ]]; then
            extra_metadata_files+=("$metadata_file")
            cecho red "ERROR: Extra metadata file $metadata_file found" | tee -a general_errors.logs
            ERROR_STATUS=1
        fi
    done

    cecho green "Done"

    if [[ "$PRIVATE_FOLDER" == "true" ]]; then
        validate_private_files "${metadata_private_files[@]}"
    fi
}

function validate_structure {
    cecho yellow "Validating folder structure ..."
    annotations=true
    main_folder="$dataset"
    inbox_subfolders=$(s3cmd_command ls s3://"$INBOX_BUCKET"/"$user"/"$dataset"/ | awk -F'_lifescience-ri.eu/|_elixir-europe.org/' '{print $2}' | cut -d'/' -f2)
    # Convert multiple lines to array
    IFS=$'\n' read -r -d '' -a subfolders <<<"$inbox_subfolders"$'\n'
    expected_subfolders=("METADATA" "IMAGES" "ANNOTATIONS" "PRIVATE" "LANDING_PAGE")
    # Check if the main folder starts with "DATASET_"
    if [[ "$main_folder" != "DATASET_"* ]]; then
        cecho red "ERROR: Main folder does not start with DATASET_" | tee -a general_errors.logs
        ERROR_STATUS=1
    fi

    # Check if annotations folder exists in the subfolders
    # If it does not then remove it from the expected_subfolders
    if [[ ! " ${subfolders[@]} " =~ "ANNOTATIONS" ]]; then
        unset 'expected_subfolders[2]'
        expected_subfolders=("${expected_subfolders[@]}")
        annotations=false
    fi
    
    # Check if landing pages folder exists in the subfolders
    # If it does not then remove it from the expected_subfolders
    # If it exists, check if THUMBNAILS folder exists and contains files
    if [[ ! " ${subfolders[@]} " =~ "LANDING_PAGE" ]]; then
        if [[ "$annotations" == "false" ]]; then
            unset 'expected_subfolders[3]'
            expected_subfolders=("${expected_subfolders[@]}")
        else
            unset 'expected_subfolders[4]'
            expected_subfolders=("${expected_subfolders[@]}")
        fi
    else
        thumbnail_files=$(s3cmd_command ls s3://"$INBOX_BUCKET"/"$user"/"$dataset"/LANDING_PAGE/THUMBNAILS/ --recursive | wc -l)
        if [[ "$thumbnail_files" -eq 0 ]]; then
            cecho red "ERROR: THUMBNAILS folder is missing or empty" | tee -a general_errors.logs
            ERROR_STATUS=1
        fi
        LANDING_PAGE=true
    fi

    # Check if landing page folder is empty in case it exists
    if [[ "$LANDING_PAGE" == "true" ]]; then
        landing_page_files=$(s3cmd_command ls s3://"$INBOX_BUCKET"/"$user"/"$dataset"/LANDING_PAGE/ --recursive | wc -l)
        if [[ "$landing_page_files" -eq 0 ]]; then
            cecho red "ERROR: LANDING_PAGE folder is empty" | tee -a general_errors.logs
            ERROR_STATUS=1
        fi
    fi

    # Compare the two arrays of subfolders and expected_subfolders
    # and print if they are not the same
    local extra_subfolders=()
    local missing_subfolders=()
    for expected_subfolder in "${expected_subfolders[@]}"; do
        if [[ ! " ${subfolders[@]} " =~ " ${expected_subfolder} " ]]; then
            missing_subfolders+=("$expected_subfolder")
            # If the missing subfolder is "PRIVATE" then set the PRIVATE_FOLDER to false
            if [[ "$expected_subfolder" == "PRIVATE" ]]; then
                PRIVATE_FOLDER=false
            fi
        fi
    done

    for subfolder in "${subfolders[@]}"; do
        if [[ ! " ${expected_subfolders[@]} " =~ " ${subfolder} " ]]; then
            extra_subfolders+=("$subfolder")
        fi
    done

    if [[ ${#extra_subfolders[@]} -ne 0 ]]; then
        cecho red "ERROR: Extra subfolders found: ${extra_subfolders[*]}" | tee -a general_errors.logs
        ERROR_STATUS=1
    fi

    if [[ ${#missing_subfolders[@]} -ne 0 ]]; then
        cecho red "ERROR: Missing subfolders: ${missing_subfolders[*]}" | tee -a general_errors.logs
        ERROR_STATUS=1
    fi

    cecho green "Done"

    validate_files "$annotations"
}

# Function for downloading the xml metadata files from the inbox
function get_xml_files {
    mkdir -p xml-files
    cecho yellow "Getting xml files ..."
    metadata_path=$(s3cmd_command ls s3://"$INBOX_BUCKET"/"$user"/"$dataset"/ | grep -i METADATA | awk '{print $2}')
    private_path=$(s3cmd_command ls s3://"$INBOX_BUCKET"/"$user"/"$dataset"/ | grep -i PRIVATE | awk '{print $2}')
    s3cmd_command get "$metadata_path" --recursive  xml-files/ >/dev/null 2>&1
    s3cmd_command get "$private_path" --recursive  xml-files/ >/dev/null 2>&1
    if [[ "$LANDING_PAGE" == "true" ]]; then
        landing_page_path=s3://"$INBOX_BUCKET"/"$user"/"$dataset"/LANDING_PAGE/landing_page.xml
        s3cmd_command get "$landing_page_path" --recursive  xml-files/ >/dev/null 2>&1
        # Throw error if the landing_page.xml file does not exist
        if [[ ! -f xml-files/landing_page.xml.c4gh ]]; then
            cecho red "ERROR: landing_page.xml is missing" | tee -a general_errors.logs
            ERROR_STATUS=1
            LANDING_PAGE=false
        fi
    fi

    # Check if the folder xml-files is empty
    if [ -z "$(ls -A xml-files)" ]; then
        cecho red "ERROR: Metadata folder is empty. Check if the path of the metadata folder in inbox is where it should be"
        ERROR_STATUS=1
    fi
    cecho green "Done"
}

# Function for decrypting the xml file
# - Gets the crypt4gh private key from vault
# - Checks if the crypt4gh is the python or golang version
# - Decrypts all the files that are in the xml-files folder
function decrypt_xml_files {
    cecho yellow "Decrypting xml files ..."
    export C4GH_PASSPHRASE
    vault kv get -field=private_key bp-secrets/crypt4gh > c4gh.sec.pem

    for xml_file in xml-files/*.c4gh; do
        if ! c4gh_decrypt c4gh.sec.pem "$xml_file"; then
            cecho red "ERROR: Decryption failed for $xml_file"
            ERROR_STATUS=1
        fi
    done
    rm xml-files/*.c4gh
    cecho green "Done"
}

# Function for finding the metadata version
# - Checks if the METADATA_STANDARD tag exists in the dataset.xml file
# - If it does exist then the metadata version is v2
function find_metadata_version {
    xml_version=$(xmllint --xpath 'string(/DATASET_SET/DATASET/METADATA_STANDARD)' xml-files/dataset.xml)
    if [[ "$xml_version" == "" ]]; then
        version="v1"
        cecho green "Metadata version is v1"
    else
        cecho green "Metadata version is v2"
        version="v2"
    fi
}
# Function for validating xml files with xsd.
# - It creates a list of all xsd files and a list of all xml files.
# - Loops over all xml files (ignoring the c4gh files) and extracts the root element name.
# - For each xml file it loops over the xsd files (starting from the "BP" ones and ignoring
#   the common and schema), gets all the top level elements and if the root element of the
#   xml file matches one of the top level element of the xsd file, then it validates the
#   xml file with the xsd file.
# TODO: check if there are only the xml files that should be present in the xml-files folder.
# This can be done if a final decision is made about the names of the xml files.
function validate_with_xsd_v1 {
    cecho yellow "Validation started..."
    error_flag=0
    for xml in xml-files/*; do
        validate="false"
        rootElement=$(xmllint --xpath "name(/*)" "$xml")

        for xsd in xsd-files/BP.*.xsd; do
            if [[ "$xsd" =~ common|schema ]]; then
                continue
            fi
            topLevelElements=$(xmllint --xpath "/*[local-name()='schema']/*[local-name()='element']/@name" $xsd 2>/dev/null | tr ' ' '\n' | cut -d'"' -f2)
            if [[ "$topLevelElements" == *"$rootElement"* ]]; then
                validate="true"
                cecho yellow "Validating $xml with $xsd"
                if ! xmllint --noout --schema "$xsd" "$xml" >/dev/null 2>&1; then
                    touch "$rootElement".error
                    echo "$(xmllint --schema "$xsd" "$xml" 2>&1)" > "$rootElement".error
                    validate="failed"
                fi
            fi
        done

        if [ "$validate" = "false" ]; then
            for xsd in xsd-files/*.xsd; do
                if [[ "$(basename "$xsd")" =~ ^BP\.|common|schema ]]; then
                    continue
                fi
                topLevelElements=$(xmllint --xpath "/*[local-name()='schema']/*[local-name()='element']/@name" $xsd 2>/dev/null | tr ' ' '\n' | cut -d'"' -f2)
                if [[ "$topLevelElements" == *"$rootElement"* ]]; then
                    validate="true"
                    cecho yellow "Validating $xml with $xsd"
                    if ! xmllint --noout --schema "$xsd" "$xml" >/dev/null 2>&1; then
                        validate="failed"
                    fi
                    break
                fi
            done
        fi

        case $validate in
            false)
                cecho red "ERROR: Validation check is NOT happening for $xml (maybe error on root element)"
                error_flag=1
                ;;
            failed)
                error_flag=1
                ;;
        esac
    done

    if [[ $error_flag -eq 1 ]]; then
        cecho red "Validation failed"
        ERROR_STATUS=1
    else
        cecho green "Validation succeeded"
    fi
}

function validate_with_xsd_v2 {
    cecho yellow "Validation for xml files started..."
    error_flag=0
    for xml in xml-files/*; do
        # Remove folder and extension from the xml file
        xml_base=$(basename "$xml")
        xml_name=$(echo "$xml_base" | cut -d'.' -f1)
        for xsd in xsd-files/*.xsd; do
            # Remove folder, extension and prefix from the xsd file
            xsd_base=$(basename "$xsd")
            xsd_name=$(echo "$xsd_base" | cut -d'.' -f2)
            if [[ "$xsd_name" == "$xml_name" ]]; then
                cecho yellow "Validating $xml with $xsd"
                if ! xmllint --noout --schema "$xsd" "$xml" >/dev/null 2>&1; then
                    cecho red "ERROR: Validation failed for $xml" | tee -a general_errors.logs
                    touch "$xml_name".error
                    echo "$(xmllint --noout --schema "$xsd" "$xml" 2>&1)" > "$xml_name".error

                    error_flag=1
                fi
            fi
        done
    done
    if [[ $error_flag -eq 1 ]]; then
        ERROR_STATUS=1
    else
        cecho green "Validation succeeded for xml files"
    fi

}

# Function for finding matches between two lists of files
# - Gets the first argument and checks if the files exist in the second argument
# - If a file is missing then it prints it and exits with an error
function check_files {
    awk '
    NR==FNR {
        # Fill a list with the files given as first argument
        names[$0]
        next
    }
    {
        # Check each line in the files of the second argument for the presence of files from the first argument
        for (name in names) {
            if (index($0, name)) {
                # If found, delete the name from the array
                delete names[name]
            }
        }
    }
    END {
        # After processing, check if there are any files from the first argument left in the array
        for (name in names) {
            # If a name never matched, output it to indicate a failure
            print name
            exit_code = 1    # Set a custom exit code for signaling a failure
        }
        exit exit_code
    }
    ' <(echo "$1") <(echo "$2")
}

# Function for checking that there are no empty files in IMAGES
function check_file_sizes {
    cecho yellow "Checking file sizes ..."
    files_lines=$(s3cmd_command ls s3://"$INBOX_BUCKET"/"$user"/"$dataset"/ --recursive)
    while IFS= read -r line; do
        size=$(echo "$line" | awk '{print $3}')
        bytesize=$($NUMFMT --from=si $size)
        if [[ "$bytesize" -lt $MIN_FILE_SIZE ]]; then
            path=$(echo $line | awk '{print $4}')
            ERROR_STATUS=1
            cecho red "Empty or incomplete file: $path (size: $size)"
        fi
    done <<< "$files_lines"
}

# Function for checking the metadata and inbox files.
# - Gets a list of all the dataset filepaths in the inbox
# - Counts only the files in the IMAGES folder
# - Parses the filenames from metadata and counts them
# - Checks if the metadata list is empty and if it is, then exits
# - Checks if the number of files in the inbox and in the metadata are equal
# - If there are equal number of files, then it checks if the filenames from metadata exist in the inbox
# - If they are not equal, it prints the extra or missing files and exits
function comparing_files {
    cecho yellow "Checking files ..."
    all_inbox_files=$(s3cmd_command ls s3://"$INBOX_BUCKET"/"$user"/"$dataset"/ --recursive | awk '{print $4}')

    count_inbox_files=$(echo "$all_inbox_files" | grep -c "IMAGES")

    for file in xml-files/*.xml; do
        if [[ "$file" == *"image"* ]]; then
            metadata_files=$(xmllint --xpath '/IMAGE_SET/IMAGE/FILES/FILE/@filename' "$file" | awk -F= '{print $2}' | sed 's/"//g')
            break
        fi
    done

    count_metadata_files=$(echo "$metadata_files" | wc -l | xargs)
    # Replace backslashes in metadata filenames (in case they exist)
    new_metadata_files=$(echo "$metadata_files" | sed 's/\\/\//g') 
    
    if [[ "$metadata_files" == "" ]]; then
        cecho red "ERROR: No filenames found in metadata" | tee -a general_errors.logs
        ERROR_STATUS=1
    elif [ "$count_inbox_files" -lt "$count_metadata_files" ]; then
        cecho red "ERROR: There are more files in metadata than the ones that exist in the inbox (inbox=$count_inbox_files, metadata=$count_metadata_files)" | tee -a general_errors.logs
        echo "The missing files in the inbox are:"
        missing_inbox_files=$(check_files "$new_metadata_files" "$all_inbox_files")
        echo "$missing_inbox_files"
        ERROR_STATUS=1
    elif [ "$count_inbox_files" -gt "$count_metadata_files" ]; then
        cecho red "ERROR: There are more files in the inbox than the ones that are referenced in metadata (inbox=$count_inbox_files, metadata=$count_metadata_files)" | tee -a general_errors.logs
        # Modify all_inbox_files to contain only the parts that are referenced in metadata
        inbox_images_files=$(echo "$all_inbox_files" | grep "/IMAGES/" | sed -E 's|.*(IMAGES/IMAGE_[^/]+/[^.]+(\.[^.]+)*\.dcm).*|\1|')
        extra_inbox_files=$(check_files "$inbox_images_files" "$new_metadata_files")
        length_extra_files=$(echo "$extra_inbox_files" | wc -l)
        missing_files_diff=$((count_inbox_files - count_metadata_files))
        if [ "$length_extra_files" -eq "$missing_files_diff" ]; then
            echo "The extra files in the inbox are:"
            echo "$extra_inbox_files"
        else
            echo "There are extra files in inbox and other file(s) exist in metadata but not in inbox:"
            echo "$extra_inbox_files"
            ERROR_STATUS=1
        fi
    else
        matching_files=$(check_files "$new_metadata_files" "$all_inbox_files")
        if [ -n "$matching_files" ]; then
            cecho red "ERROR: The following files were not found in inbox:" | tee -a general_errors.logs
            echo "$matching_files" | tee -a general_errors.logs
            ERROR_STATUS=1
        else
            cecho green "All files listed in the metadata files are present in the inbox"
        fi
    fi

    # Check the thumbnails files
    if [[ "$LANDING_PAGE" == "true" ]]; then
        all_thumbnail_files=$(s3cmd_command ls s3://"$INBOX_BUCKET"/"$user"/"$dataset"/LANDING_PAGE/THUMBNAILS/ --recursive | awk '{print $4}')
        count_inbox_thumbnail_files=$(echo "$all_thumbnail_files" | wc -l)
        metadata_thumbnail_files=$(xmllint --xpath '/LANDING_PAGE_SET/LANDING_PAGE/SAMPLE_IMAGE_FILES/SAMPLE_IMAGE_FILE/@filename' xml-files/landing_page.xml | awk -F= '{print $2}' | sed 's/"//g')
        count_metadata_thumbnail_files=$(echo "$metadata_thumbnail_files" | wc -l)
        if [[ "metadata_thumbnail_files" == "" ]]; then
            cecho red "ERROR: No filenames found in metadata for thumbnails" | tee -a general_errors.logs
            ERROR_STATUS=1
        elif [ "$count_inbox_thumbnail_files" -lt "$count_metadata_thumbnail_files" ]; then
            cecho red "ERROR: There are more thumbnail files in metadata than the ones that exist in the inbox (inbox=$count_inbox_thumbnail_files, metadata=$count_metadata_thumbnail_files)" | tee -a general_errors.logs
            echo "The missing files in the inbox are:"
            missing_inbox_files=$(check_files "$metadata_thumbnail_files" "$all_thumbnail_files")
            echo "$missing_inbox_files"
            ERROR_STATUS=1
        elif [ "$count_inbox_thumbnail_files" -gt "$count_metadata_thumbnail_files" ]; then
            cecho red "ERROR: There are more thumbnail files in the inbox than the ones that are referenced in metadata (inbox=$count_inbox_thumbnail_files, metadata=$count_metadata_thumbnail_files)" | tee -a general_errors.logs
            # Modify all_thumbnail_files to contain only the parts that are referenced in metadata
            inbox_thumbnail_files=$(echo "$all_thumbnail_files" | sed "s/.*"$dataset"\///")
            extra_inbox_files=$(check_files "$inbox_thumbnail_files" "$metadata_thumbnail_files")
            length_extra_files=$(echo "$extra_inbox_files" | wc -l)
            missing_files_diff=$((count_inbox_thumbnail_files - count_metadata_thumbnail_files))
            if [ "$length_extra_files" -eq "$missing_files_diff" ]; then
                echo "The extra files in the inbox are:"
                echo "$extra_inbox_files"
            else
                echo "There are extra files in inbox and other file(s) exist in metadata but not in inbox:"
                echo "$extra_inbox_files"
                ERROR_STATUS=1
            fi
        else
            matching_files=$(check_files "$metadata_thumbnail_files" "$all_thumbnail_files")
            if [ -n "$matching_files" ]; then
                cecho red "ERROR: The following thumbnail files were not found in inbox:" | tee -a general_errors.logs
                echo "$matching_files" | tee -a general_errors.logs
                ERROR_STATUS=1
            else
                cecho green "All metadata thumbnail files are present in the inbox"
            fi
        fi
    fi
}

# Function for generating dataset ids (e.g: aa-Dataset-v5y9hk-nc2rfu)
function generate_dataset_id {
    local part_one=$(LC_ALL=C tr -dc 'abcdefghjkmnpqrstuvwxyz23456789' </dev/urandom | head -c 6)
    local part_two=$(LC_ALL=C tr -dc 'abcdefghjkmnpqrstuvwxyz23456789' </dev/urandom | head -c 6)

    echo "aa-Dataset-$part_one-$part_two"
}

# Function for moving the metadata files from the inbox to the metadata bucket
# - Gets the dataset id from the dataset_id.txt file
# - Find the path of the folder in the inbox.
# - In case of staging cluster:
#       - Move the files from the inbox to the metadata bucket (stable id is in the file path).
# - In case of prod cluster:
#      - Create the folder.
#      - Get the files from the inbox to the local folder.
#      - Put the files in the metadata bucket.
#      - Delete the files from the inbox.
# - Check if the metadata files are moved by checking the size of the metadata folder
#   and if the size is empty, then exit.
function move_private_metadata {
    cecho yellow "Moving metadata ..."

    stable_id=$(cat dataset_id.txt)
    metadata_inbox_path=$(s3cmd_command ls s3://"$INBOX_BUCKET"/"$user"/"$dataset"/ | grep "PRIVATE" | awk '{print $2}')
    metadata_bucket_path="s3://"$METADATA_BUCKET"/"$user"/"$stable_id"/"$dataset"/"PRIVATE"/"

    cecho yellow "Moving PRIVATE folder in metadata bucket"
       
    if [[ "$cluster" == "staging" ]]; then
        s3cmd_command mv "$metadata_inbox_path" "$metadata_bucket_path" --recursive
        metadata_size=$(s3cmd_command ls "$metadata_bucket_path" --recursive | awk '{print $3}')
        if [ -z "$metadata_size" ]; then
            cecho red "ERROR: Moving metadata failed"
            exit 1
        fi
    else
        mkdir -p "PRIVATE"
        s3cmd_command get "$metadata_inbox_path" --recursive "PRIVATE"/ >/dev/null 2>&1
        s3cmd_metadata put "PRIVATE"/ "$metadata_bucket_path" --recursive
        metadata_size=$(s3cmd_metadata ls "$metadata_bucket_path" --recursive | awk '{print $3}')
        if [ -z "$metadata_size" ]; then
            cecho red "ERROR: Moving metadata failed"
            exit 1
        fi
        s3cmd_command del "$metadata_inbox_path" --recursive
    fi

    cecho green "Done"
}

# Function for modifying the dataset.xml file
# - Checks if the dataset.xml file exists in the xml-files folder
# - Adds the dataset stable id in the dataset.xml file
# - Downloads the public key
# - Encrypts the dataset.xml file
# - Deletes the file from the metadata bucket
# - Uploads the modified encrypted file to the metadata bucket
function modify_dataset {
    cecho yellow "Modifying dataset.xml file ..."

    dataset_id=$(generate_dataset_id)
    echo "$dataset_id" > dataset_id.txt

    if [ ! -f "xml-files/dataset.xml" ]; then
        cecho red "ERROR: dataset.xml file does not exist in xml-files folder"
        exit 1
    else
        sed_i_bak -E "s/(<DATASET[^>]* alias=\"[^\"]*\")/\1 accession=\"$dataset_id\"/g" xml-files/dataset.xml
    fi

    curl https://raw.githubusercontent.com/NBISweden/EGA-SE-user-docs/main/crypt4gh_bp_key.pub -o bp_key.pub

    export C4GH_PASSPHRASE
    if ! c4gh_encrypt c4gh.sec.pem bp_key.pub xml-files/dataset.xml; then
        cecho red "ERROR: Encryption failed"
        exit 1
    fi

    s3cmd_command del s3://"$INBOX_BUCKET"/"$user"/"$dataset"/METADATA/dataset.xml.c4gh

    s3cmd_command put xml-files/dataset.xml.c4gh s3://"$INBOX_BUCKET"/"$user"/"$dataset"/METADATA/dataset.xml.c4gh

    cecho green "Done"
}

# This function extracts the organisation name from the "organisation.xml" file.
# - It first checks if the "organisation.xml" file exists in the "xml-files" directory.
# - If the file exists, it uses `xmllint` to extract the organisation name
# - If the extracted organisation name is empty, it logs an error message indicating
#   that no organisation name is present in the XML file.
function organisation_name {
    cecho yellow "Extracting organisation name ..."

    if [ ! -f "xml-files/organisation.xml" ]; then
        cecho red "ERROR: organisation.xml file does not exist in the xml-files folder"
        exit 1
    fi

    ORG_NAME=$(xmllint --xpath "//ORGANISATION_SET/ORGANISATION/NAME/text()" xml-files/organisation.xml)
    if [ -z "$ORG_NAME" ]; then
        cecho red "There is no organisation name in xml"
    else
        cecho green "Organisation name: $ORG_NAME"
    fi

    cecho green "Done"
}

# Function for checking Kubernetes access
function check_kubernetes_access {
    if [[ "$cluster" == "prod" ]]; then
        if ! kubectl -n sda-prod auth can-i create jobs >/dev/null 2>&1; then
            cecho red "ERROR: You do not have access to the 'sda-prod' namespace"
            cecho red "Please check if your KUBECONFIG is correctly set (or VPN)"
            exit 1
        fi
    fi
}

vault token renew >/dev/null 2>&1
if [[ "$?" != "0" ]] && [[ "$1" != "--clean" ]];
then
    cecho red "You must log in to vault.nbis.se before using this script"
    exit 1
fi

if [[ "$1" != "--clean" ]]; then
    if [[ "$DRY_RUN" == false ]] && [[ "$VALIDATION_ONLY" == false ]]; then
        check_kubernetes_access
    fi

    get_credentials

    sanitize_user_dataset
fi

# Validation steps
check_file_sizes

validate_structure

get_xml_files

trap remove_private_key EXIT HUP INT QUIT PIPE TERM

decrypt_xml_files

find_metadata_version

get_xsd_files

validate_with_xsd_$version

comparing_files

if [[ "$ERROR_STATUS" -eq 0 ]]; then
    cecho green "All checks passed"

    if [[ "$DRY_RUN" == "true" ]]; then
        cecho yellow "DRY RUN MODE"
        exit 0
    fi

    modify_dataset

    move_private_metadata

    organisation_name
else
    cecho red "Validation failed"
    exit $ERROR_STATUS
fi

cecho green "VALIDATION SUCCESSFUL !!!"
cecho blue "Dataset stable ID: $dataset_id"

if [[ "$VALIDATION_ONLY" == "true" ]]; then
    cecho yellow "VALIDATION ONLY MODE"
    exit 0
fi

if [[ "$cluster" == "staging" ]]; then
    cecho magenta "Kubernetes job NOT deployed for staging cluster"
    exit 0
fi

cat << EOF

     --------------------
    | Setting up k8s job |
     --------------------

EOF

# Decrypt rems file
if ! c4gh_decrypt c4gh.sec.pem "PRIVATE/rems.xml.c4gh"; then
    cecho red "ERROR: rems decryption failed"
    exit 1
fi

# Copy metadata in data folder
cp -f PRIVATE/rems.xml ../data/xml/rems.txt || exit 1
cp -f xml-files/policy.xml ../data/xml/policy.txt || exit 1
cp -f xml-files/dataset.xml ../data/xml/dataset.txt || exit 1

# Create the config file
cp  -f ../config.yaml.example ../config.yaml

# Update the config file
sed_i "s|USER_ID:.*|USER_ID: \"${user//_/@}\"|" ../config.yaml
sed_i "s|DATASET_ID:.*|DATASET_ID: \"$dataset_id\"|" ../config.yaml
sed_i "s|DATASET_FOLDER:.*|DATASET_FOLDER: \"$dataset\"|" ../config.yaml
sed_i "s|CLIENT_ACCESS_TOKEN:.*|CLIENT_ACCESS_TOKEN: \"$ACCESS_TOKEN\"|" ../config.yaml
sed_i "s|MAIL_UPLOADER:.*|MAIL_UPLOADER: \"$EMAIL\"|" ../config.yaml
sed_i "s|MAIL_UPLOADER_NAME:.*|MAIL_UPLOADER_NAME: \"$NAME\"|" ../config.yaml
sed_i "s|MAIL_UPLOADER_ORGANIZATION_NAME:.*|MAIL_UPLOADER_ORGANIZATION_NAME: \"$ORG_NAME\"|" ../config.yaml

pushd ..
go build -o bpctl .
./bpctl render -x
kubectl kustomize . -o "$dataset".yaml
kubectl -n sda-prod apply -f "$dataset".yaml
popd

trap - EXIT
remove_private_key
exit $ERROR_STATUS
