#!/bin/sh
# 拉取指定目录下所有仓库的代码

dir=${1-$(pwd)} #仓库所在目录，默认脚本执行目录

for str in $(ls $dir); do
  echo ----------------
  echo $str
  if [ ! -d .git ]; then
    echo "非git仓库"
    continue
  fi
  git --git-dir=$dir/$str/.git pull
done
