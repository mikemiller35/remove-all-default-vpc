# remove-all-default-vpc

Remove default VPCs in all regions discovered

Inspired by [davidobrien1985/delete-aws-default-vpc](https://github.com/davidobrien1985/delete-aws-default-vpc)

## Use

### Locally

Make sure you're auth'd into aws where you're running this

```bash
# Build it with go
make
# Run it
bin/remove-all-default-vpc
```

### Kubernetes

See [Kubernetes](kubernetes/)

### Dockers

```bash
eval "$(aws configure export-credentials --format env)";
docker run --rm -it \
-e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
-e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
-e AWS_SESSION_TOKEN=$AWS_SESSION_TOKEN \
{{.IMAGE_REPO}}:{{.VERSION}} \
remove-default-vpc
```
