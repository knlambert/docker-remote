# docker-remote

A simple project to be able to spin up a docker service.

Supported hosts:
* ec2

# Usage

## Host

### Start the host.
```bash
export AWS_ACCESS_KEY_ID=xxxx
export AWS_SECRET_ACCESS_KEY=xxxx
export AWS_DEFAULT_REGION="ca-central-1"

docker-remote up  \
    --security-group=sg-1234 \
    --vpc-id=vpc-1234 \
    --default-region="ca-central-1"
```
### Connect to the host.
```bash
docker-remote shell
```

### Kill the host.
```bash
docker-remote down
```

### Port forward

```bash
docker-remote port-forward 80:8080
```

### Configure the docker client

```bash
docker-remote port-forward 80:8080
```


sudo apt install ssh-askpass