#!/bin/sh
version="1.1"
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[1;34m'
DARK='\033[1;30m'
NC='\033[0m'

echo "${BLUE}notion-md-gen binary installer ${version}${NC}"
unameOut="$(uname -s)"

case "${unameOut}" in
Darwin*)
  arch=macos-amd64
  ;;
*)
  arch=linux-amd64
  ;;
esac
bin_dir="/usr/local/bin"
url=$(curl -s https://api.github.com/repos/bonaysoft/notion-md-gen/releases/latest | grep "browser_download_url.*${arch}.tar.gz\"" | cut -d : -f 2,3 | tr -d '\"[:space:]')

echo "${DARK}"
echo "Configuration: [${arch}]"
echo "Location:      [${url}]"
echo "Directory:     [${bin_dir}]"
echo "${NC}"

test ! -d "${bin_dir}" && mkdir "${bin_dir}"
curl -J -L "${url}" | tar xz -C "${bin_dir}"

if [ $? -eq 0 ]; then
  echo "${GREEN}"
  echo "Installation completed successfully."
  echo "$ notion-md-gen version"
  "${bin_dir}"/notion-md-gen version
else
  echo "${RED}"
  echo "Failed installing notion-md-gen"
fi

echo "${NC}"
