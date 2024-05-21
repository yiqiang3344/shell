#!/usr/bin/env bash
# 批量删除指定目录下仓库的指定分支，切换为master分支。
set -u #有未定义的变量时要报错

branch=${1:?"分支不能为空"} #分支
expectRepos=(${2:-})  #期望处理的git仓库名列表，空格分割
delOrigin=${3:-""}    #是否删除远程分支
dir=${4:-$(pwd)}      #仓库所在目录，默认脚本执行目录

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
    git fetch
    if [ $? -ne 0 ]; then
      continue
    fi
    git checkout master
    if [ $? -ne 0 ]; then
      continue
    fi
    git pull origin master
    if [ $? -ne 0 ]; then
      continue
    fi
    git branch -D $branch
    if [ $? -ne 0 ]; then
      continue
    fi
    if [ -z $delOrigin ]; then
      git push origin --delete $branch
    fi
  done
}

handle $dir
