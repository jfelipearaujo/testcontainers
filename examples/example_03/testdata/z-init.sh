#!/bin/sh

# This script is executed when all other scripts are executed
# It is necessary because LocalStack takes a long time to start the services
# and run other scripts that are executed after the services are started
echo "Initialization complete!"