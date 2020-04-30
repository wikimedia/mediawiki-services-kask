#!/usr/bin/env bash

set +eu

TEST_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $TEST_DIR

export API_TESTING_CONFIG_FILE=quibble.json

npm install --no-package-lock
npm test
