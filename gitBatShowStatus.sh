#!/usr/bin/env bash
# 查看指定目录下所有仓库的新增、变动、删除文件信息
set -u #有未定义的变量时要报错

dir=${1:-$(pwd)} #仓库所在目录，默认脚本执行目录

handle() {
  for i in $1/*; do
    currentPath=$i
    if [ -f $currentPath ]; then
      continue
    fi
    if [ -d $currentPath ] && [ ! -d $currentPath/.git ]; then
      handle $currentPath
      continue
    fi
    cd $currentPath || exit 1
    _tmpData=$(git status)
    if [ "$(echo $_tmpData | grep "无文件要提交，干净的工作区")" != "" ]; then
      continue
    fi
    echo '#'$currentPath
    git status
  done
}

handle $dir
