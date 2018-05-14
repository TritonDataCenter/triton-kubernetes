#!/usr/bin/env bash

# Check gofmt
echo "==> Checking for unchecked errors"
err_files=$(errcheck -ignoretests \
                     -ignore 'bytes:.*' \
                     -ignore 'io:Close|Write' \
                     $(go list ./...| grep -Ev 'vendor|examples|testutils'))

if [[ -n ${err_files} ]]; then
    echo 'Unchecked errors found in the following places:'
    echo "${err_files}"
    echo "Please handle returned errors. You can check directly with \`make errcheck\`"
    exit 1
fi

exit 0
