#!/bin/sh

shell_dir=$(cd "$(dirname "$0")" || exit;pwd)
if [ ! -d "${shell_dir}" ]; then
  echo "not found bin files"
  exit
fi

project="notion-md-gen"
unameOut="$(uname -s)"
config_dir="/etc/${project}"
test ! -d "${config_dir}" && mkdir "${config_dir}"
cp "${shell_dir}/${project}" /usr/local/bin
