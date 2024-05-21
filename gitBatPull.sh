#!/usr/bin/env bash
# 批量拉取指定目录下所有仓库的代码
set -u #有未定义的变量时要报错

expectRepos=(${1:-}) #期望处理的git仓库名列表，空格分割
dir=${2:-$(pwd)}     #仓库所在目录，默认脚本执行目录

if [[ "$(dirname $0)" == "$(pwd)" ]]; then
  . functions
else
  . $(dirname $0)/functions
fi

handle() {
  for i in $1/*; do
    currentPath=$i
    repoName=${i/$dir\//}
    if [ -f $currentPath ]; then
      continue
    fi
    if [ -d $currentPath ] && [ ! -d $currentPath/.git ]; then
      handle $currentPath
      continue
    fi
    if [[ ${#expectRepos[@]} -gt 0 ]] && ! in_array expectRepos $repoName; then
      continue
    fi
    cd $currentPath || exit 1
    echo "#"$currentPath
    git pull
  done
}

handle $dir
