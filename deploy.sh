#!/bin/bash

# Remove old binary if it exists
rm -f notescanner-linux-amd64

# Pull latest code
git pull

# Set executable permissions
sudo chmod +x notescanner-linux-amd64

 