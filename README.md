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

docker-remote ec2 up  \
    --security-group=sg-1234 \
    --vpc-id=vpc-1234 \
    --default-region="ca-central-1"
```
### Connect to the host.
```bash
docker-remote ec2 shell
```

### Kill the host.
```bash
docker-remote ec2 down
```

### Port forward

```bash
docker-remote ec2 port-forward 80:8080 --remote-host localhost
```

### Configure the docker client

```bash
docker-remote ec2 port-forward 80:8080
```


### To use the docker command line from your machine :

Docker through ssh does not integrate the keys very well: 

```bash
PATH_TO_SSH_PUBLIC_KEY=${HOME}/kevin.pem
chmod 600 ${PATH_TO_SSH_PUBLIC_KEY}
ssh-add ${PATH_TO_SSH_PUBLIC_KEY}
```

Now the docker command line should work: 

```bash
docker ps
```