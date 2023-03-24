# cdflow2-build-lambda

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
config:
  image: mergermarket/cdflow2-config-acuris
terraform:
  image: hashicorp/terraform
```

### Parameters

#### target_directory

Change directory where the lambda code resides.  
Defaults to `./target` if not defined.

#### region

Override default region coming from config container.  
This allows deploying to different regions.  
The upload bucket will be created from the base bucket, coming from the config container and adding the region.  
So if you're using `mergermarket/cdflow2-config-acuris` as the config image, the base bucket will be `acuris-lambdas`.  
Overriding the region with `ap-southeast-1` the final bucket will be `acuris-lambdas-ap-southeast-1`.  
Default to the region coming from the config container.  
In this case, the bucket won't be changed, it will be used as is coming from the config container.  
