#!/bin/bash -ex
#########################################################################
# File Name: build.sh
# Author: nian
# Blog: https://whoisnian.com
# Mail: zhuchangbao1998@gmail.com
# Created Time: 2023年04月12日 星期三 00时10分36秒
#########################################################################

export CGO_ENABLED=0
export BUILDTIME=$(date --iso-8601=seconds)
if [[ -z "${GITHUB_REF_NAME}" ]]; then
  export VERSION=$(git describe --tags || echo unknown)
else
  export VERSION=${GITHUB_REF_NAME}
fi

goBuild() {
  GOOS="$1" GOARCH="$2" go build -trimpath \
    -ldflags="-s -w \
    -X 'github.com/whoisnian/virt-launcher/global.Version=${VERSION}' \
    -X 'github.com/whoisnian/virt-launcher/global.BuildTime=${BUILDTIME}'" \
    -o "$3" .
}

if [[ "$1" == '.' ]]; then
  goBuild $(go env GOOS) $(go env GOARCH) virt-launcher
elif [[ "$1" == 'linux-amd64' ]]; then
  goBuild linux amd64 "virt-launcher-linux-amd64-${VERSION}"
elif [[ "$1" == 'linux-arm64' ]]; then
  goBuild linux arm64 "virt-launcher-linux-arm64-${VERSION}"
elif [[ "$1" == 'all' ]]; then
  goBuild linux amd64 "virt-launcher-linux-amd64-${VERSION}"
  goBuild linux arm64 "virt-launcher-linux-arm64-${VERSION}"
else
  echo "Unknown build target"
  exit 1
fi
