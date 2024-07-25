## ABOUT THE PROJECT

This is a project written in Golang to produce XIR model of the facility in the Mergetest beds. 

This project utilizes Ground-control as the backend to get the LLDP information and the XIR model from MergeTB. The repository of the projects are listed below

Ground-control: https://gitlab.com/mergetb/ops/ground-control
XIR: https://gitlab.com/mergetb/xir

## Pre-requisites

* Golang >= v1.2

## Getting Started

* cd into cmd/mergelab
* Run command `go build` to build the binary. A file called mergelab will be created in the same directory
* Run `./mergelab` to run the binary