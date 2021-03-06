#!/usr/bin/env bash
set -uo pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )/.."
source ./scripts/install_tools.sh

# TODO: change default to `cfbuildpacks/cflinuxfs3-cnb-experimental:build` when pack cli can use it
export CNB_BUILD_IMAGE=${CNB_BUILD_IMAGE:-packs/samples:v3alpha2}

# TODO: change default to `cfbuildpacks/cflinuxfs3-cnb-experimental:run` when pack cli can use it
export CNB_RUN_IMAGE=${CNB_RUN_IMAGE:-packs/run:v3alpha2}

# Always pull latest images
# Most helpful for local testing consistency with CI (which would already pull the latest)
docker pull $CNB_BUILD_IMAGE
docker pull $CNB_RUN_IMAGE

echo "Run Buildpack Runtime Integration Tests"
set +e
go test ./integration/... -v -run Integration
exit_code=$?

if [ "$exit_code" != "0" ]; then
    echo -e "\n\033[0;31m** GO Test Failed **\033[0m"
else
    echo -e "\n\033[0;32m** GO Test Succeeded **\033[0m"
fi

exit $exit_code
