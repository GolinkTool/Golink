# Golink Setup and Build Instructions

This document outlines the steps to build and run the Golink containers for the Go project. Follow the steps carefully to set up the environment and execute the program.

## Prerequisites

Ensure you have Docker installed and running on your machine before proceeding with the following steps.

## Step 1: Build the Golink Data Container

1. Navigate to the `Mysql` directory:
   ```bash
   cd Mysql
   ```
   
2. Build the Golink Data container by running the following Docker command:
   ```bash
   docker build --pull --rm -f "./DockerFile" -t golink_data:latest ./
   ```
   
3. Once the build process is complete, run the Golink Data container:
   ```bash
   docker run -itd golink_data
   ```
   
4. Use docker inspect to retrieve the container's IP address:
   ```bash
   docker inspect golink_data
   ```
   
Look for the IPAddress field in the output to find the container's IP address.

## Step 2: Build the Golink Container
1. Go back to the current directory (where the Golink project is located):

2. Build the Golink container by running the following command:
   ```bash
   docker build --pull --rm -f "./DockerFile" -t golink:latest ./
   ```
   
3. After the build completes, run the Golink container with the following command:
   ```bash
   docker run -it --net=host golink bash
   ```
   
4. Inside the bash shell, run the Golink program with the following command, replacing Container_ip with the IP address of the golink_data container you retrieved earlier:
   ```bash
   ./GoLink -baseName=Go_Example -projectDir=./Go_Example -database_ip=Container_ip
   ```
   
5. Generate the go.mod file in the project folder./Go_Example

6. The logs:
   ```bash
   open DB SUCCESS!
   ProjectDir: ./Go_Example
   BaseName: Go_Example
   Successfully create go.mod file!
   907.945422ms
   ```