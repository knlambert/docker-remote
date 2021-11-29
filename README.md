# docker-remote

A simple project to be able to spin up a docker service.

Supported hosts:
* ec2

# Why

Sometimes, you don't want to run docker on your machine. It can be because
you don't have enough CPU / RAM on your Desktop, or because of the
policy of the company you work for.

With docker remote, you only install the docker client on your machine,
and the docker daemon can run on a distant machine, like an AWS EC2.

# Usage

## Setup
You must be connected to the AWS cli first,
then export the AWS REGION variable.

### Linux & Mac
```bash
export AWS_DEFAULT_REGION="ca-central-1"
```

### Windows

```powershell
$env:AWS_REGION = 'ca-central-1'
```

## Host creation

```bash
docker-remote ec2 up  \
    --security-group=sg-1234 \
    --vpc-id=vpc-1234 \
    --default-region="ca-central-1"
```

## Connect to the host.
```bash
docker-remote ec2 shell
```

## Kill the host.
```bash
docker-remote ec2 down
```

## Port forward

```bash
docker-remote ec2 port-forward 80:8080 --remote-host localhost
```



## To use the docker command line from your machine :

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