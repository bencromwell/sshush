# sshush

Takes a YAML defined set of SSH configs and produces an SSH config file from it.

Defaults to `~/.ssh/config.yml` for the source and `~/.ssh/config` for the destination.

#### Premise

I wanted a way to manage my SSH config file that allowed me to define an inheritance based structure. This allows you to group hosts together, apply common options and be able to override the options if needs be. For example, maybe you have everything in DigitalOcean on a particular port with a particular user and SSH key, but everything in AWS is different. Or they share ports but not keys. 

#### Globals

Options that apply to the catch-all `Host *`.

#### Defaults

Basic options such as a default User or IdentityFile. Can be overridden by group or individual host entries.

#### Example

Example config demonstrating the use of the global and default options:

```yaml

---
global:
  UseRoaming: "no"

default:
  User: ben
  IdentityFile: ~/.ssh/id_rsa

web_servers:
  Port: 2201
  IdentityFile: ~/.ssh/digital_ocean
  Hosts:
    projects-do-1:
      HostName: projects-do-1.example.com
    projects-do-2:
      HostName: projects-do-2.example.com
    projects-aws:
      HostName: projects-aws.example.com
      IdentityFile: ~/.ssh/aws

raspberry_pis:
  User: pi
  Hosts:
    pi1:
      HostName: 192.168.0.107
    pi2:
      HostName: 192.168.0.108

local:
  Hosts:
    router:
      HostName: 192.168.0.1
      User: root
    kodi:
      HostName: 192.168.0.200

work:
  User: bcromwell
  Hosts:
    workpc:
      HostName: 10.0.0.80
    gitlab:
      HostName: 10.0.0.30
    jenkins:
      HostName: 10.0.0.20

```
