# cdflow2-build-lambda

Still a work in progress, but planned usage in cdflow.yaml something like:

```yaml
version: 2
build:
  lambda:
    image: mergermarket/cdflow2-build-lambda
  lambda2:
    image: mergermarket/cdflow2-build-lambda
    params:
      target_directory: target2
config:
  image: mergermarket/cdflow2-config-acuris
terraform:
  image: hashicorp/terraform
```

Params
    target_directory: (Optional) defaults to ./target if not defined
