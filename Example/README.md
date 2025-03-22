# Guide

## Overview

This document illustrates how to create and run a Docker image to reproduce a example for GoLink.
Follow these steps to replicate the environment.

## Prerequisites

- Docker: Ensure you have the latest version of Docker installed.

## Step 1: Create Docker Image

1. **Navigate to the Example Directory**

   Change to the `Example` directory where the example is located:

   ```bash
   cd ./Example
   ```

2. **Build the Docker Image**

   In the example directory, build the image using the Dockerfile:

   ```bash
   docker build --pull --rm -f "./BS_approach/DockerFile" -t bs_approach:latest "./BS_approach"
   docker build --pull --rm -f "./GoLink/DockerFile" -t golink:latest "./GoLink"
   ```

## Step 2: Run Docker Image

Start a container instance with the following command:

   ```bash
   docker run bs_approach:latest
   docker run golink:latest
   ```