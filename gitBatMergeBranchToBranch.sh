#!/bin/bash
# 批量合并指定目录的指定仓库列表的指定分支到指定分支

branch=$1       #源分支
toBranch=$2     #目标分支
expectRepos=$3  #期望处理的git仓库名列表，空格分割
dir=${4-$(pwd)} #仓库所在目录，默认脚本执行目录

if [[ "$(dirname $0)" == "$(pwd)" ]]; then
  . functions.sh
else
  . $(dirname $0)/functions.sh
fi

handle() {
  for i in $(ls $1); do
    currentPath=$1/$i
    if [ -f $currentPath ]; then
      continue
    fi
    if [ -d $currentPath ] && [ ! -d $currentPath/.git ]; then
      handle $currentPath
      continue
    fi
    if [[ ${#expectRepos[@]} > 0 ]] && ! in_array expectRepos $i; then
      continue
    fi
    cd $currentPath
    echo "#"$currentPath
    git fetch
    if [ $? -ne 0 ]; then
      continue
    fi
    git checkout $toBranch
    if [ $? -ne 0 ]; then
      continue
    fi
    git pull origin $toBranch
    if [ $? -ne 0 ]; then
      continue
    fi
    git checkout $branch
    if [ $? -ne 0 ]; then
      continue
    fi
    git pull origin $branch
    if [ $? -ne 0 ]; then
      continue
    fi
    git checkout $toBranch
    if [ $? -ne 0 ]; then
      continue
    fi
    git merge $branch
    if [ $? -ne 0 ]; then
      continue
    fi
    git push origin $toBranch
    if [ $? -ne 0 ]; then
      continue
    fi
    git merge $fromBranch --squash #合并提交为1次提交，可能有冲突，需要解决冲突
    if [ $? -ne 0 ]; then
      continue
    fi
    echo -e "提交ID\033[32m $(git rev-parse --short HEAD) \033[0m"
  done
}

handle $dir