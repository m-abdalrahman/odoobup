# odoobup 

odoobup Command Line Interface to backup from different odoo services, support odoo version 8, 9, and 10, only work on linux.

## Quick Start

 1. download your version from [here](https://github.com/m-abdalrahman/odoobup/releases/tag/v1.0.0-beta2)
 2. unzip the file
 3. move your version to `/usr/bin`

## Installation From Source
`odoobup` requires Go 1.7.3 or later
```
$ go get -u github.com/m-abdalrahman/odoobup
```	

## Usage
```
odoobup                     backup 

optional arguments:
     -h                     show this help message
     -n                     backup by id

subcommand:
     help                   show this help message
     add                    add new configuration setting
     show                   show all configurations
     del                    delete configuration setting by id number
     version                show program version number	
```
