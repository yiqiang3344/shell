#!/bin/bash
# 查看指定目录下所有仓库的新增、变动、删除文件信息
_dir=${1-$(pwd)} #仓库所在目录，默认脚本执行目录

for i in $(ls $_dir); do
  echo ' '
  echo '#'$_dir/$i
  cd $_dir/$i
  if [ ! -d .git ]; then
    echo "非git仓库"
    continue
  fi
  _tmpData=$(git status)
  if [ "$(echo $_tmpData | grep modified:)" != "" ] || [ "$(echo $_tmpData | grep 'Untracked files:')" != "" ] || [ "$(echo $_tmpData | grep deleted:)" != "" ]; then
    echo $(git status | grep 'On branch')
    if [ "$(echo $_tmpData | grep modified:)" != "" ]; then
      echo $(git status | grep modified:)
    fi
    if [ "$(echo $_tmpData | grep deleted:)" != "" ]; then
      echo $(git status | grep deleted:)
    fi
    if [ "$(echo $_tmpData | grep 'Untracked files:')" != "" ]; then
      echo $(git status | grep 'Untracked files:' -A 3)
    fi
  fi
done
