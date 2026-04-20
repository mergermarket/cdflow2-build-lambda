# cdflow2-build-lambda

A [cdflow2](https://github.com/mergermarket/cdflow2) build container that zips Lambda code and uploads it to S3.

## Usage

```yaml
version: 2
builds:
  lambda:
    image: mergermarket/cdflow2-build-lambda
  lambda2:
    image: mergermarket/cdflow2-build-lambda
    params:
      target_directory: target2
      region: eu-west-2
  lambda3:
    image: mergermarket/cdflow2-build-lambda
    params:
      regions:
        - eu-west-1
        - us-east-1
config:
  image: mergermarket/cdflow2-config-acuris
terraform:
  image: hashicorp/terraform
```

### Parameters

#### target_directory

Directory (or file) where the Lambda code resides. The contents will be zipped and uploaded to S3.  
Defaults to `./target` if not defined.

#### region

Override the default region coming from the config container. This uploads the zip to a single additional region.  
The upload bucket name is derived from the base bucket (provided by the config container) with the region appended.  
For example, with `mergermarket/cdflow2-config-acuris` the base bucket is `acuris-lambdas`. Setting `region: ap-southeast-1` uploads to `acuris-lambdas-ap-southeast-1`.  
Defaults to `eu-west-1` (uses the base bucket name as-is).

#### regions

Deploy to multiple regions in a single build. Provide a list of AWS regions.  
Each non-default region (`eu-west-1`) gets its own bucket with the region appended (e.g. `acuris-lambdas-us-east-1`).

```yaml
params:
  regions:
    - eu-west-1
    - us-east-1
```

### Output Metadata

The build produces metadata consumed by subsequent steps:

| Key | Description |
|---|---|
| `bucket` | S3 bucket for the default region (`eu-west-1`) |
| `bucket_<region>` | S3 bucket for each additional region (e.g. `bucket_us-east-1`) |
| `key` | S3 object key for the uploaded zip |

## Building

```sh
docker build -t mergermarket/cdflow2-build-lambda .
```

## Testing

```sh
go test -v ./internal/app
```
