#!/bin/bash

# Run integration tests
docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
EXIT_CODE=$?

# Clean up
docker-compose -f docker-compose.test.yml down

exit $EXIT_CODE 