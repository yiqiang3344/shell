#!/usr/bin/env bash
# 批量把某个目录下的仓库切换到指定分支

branch=$1        #分支名
expectRepos=($2) #期望处理的git仓库名列表，空格分割
dir=${3-$(pwd)}  #仓库所在目录，默认脚本执行目录

if [[ "$(dirname $0)" == "$(pwd)" ]]; then
  . functions
else
  . $(dirname $0)/functions
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
    git checkout $branch
    if [ $? -ne 0 ]; then
      continue
    fi
    git pull origin $branch
    if [ $? -ne 0 ]; then
      continue
    fi
  done
}

handle $dir
