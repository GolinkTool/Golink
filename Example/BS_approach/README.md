# BS_approach Setup and Build Instructions

This document outlines the steps to build and run the BS_approach containers for the Go project. Follow the steps carefully to set up the environment and execute the program.

## Prerequisites

Ensure you have Docker installed and running on your machine before proceeding with the following steps.

## Build the BS_approach Container

1. Build the BS_approach Data container by running the following Docker command:
   ```bash
   docker build --pull --rm -f "./DockerFile" -t bs_approach:latest ./
   ```

2. After the build completes, run the BS_approach container with the following command:
   ```bash
   docker run -it --net=host bs_approach bash
   ```
   
3. Execute commands
   ```bash
   source myenv/bin/activate &&
   python3 -u BS_approach.py --project_name=Go_Example --working_dir=./ --target_tags_dir=./ --log_dir=./ --target_tpl=github.com/ethereum/go-ethereum
   ```
