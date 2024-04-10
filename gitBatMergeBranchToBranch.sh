#!/bin/bash
# 批量合并指定目录的指定仓库列表的指定分支到指定分支

branch=$1             #源分支
var=${2}              #仓库列表，逗号分割
toBranch=${3-develop} #目标分支
dir=${4-$(pwd)}       #仓库所在目录，默认脚本执行目录

OLD_IFS="$IFS"
IFS=","
array=($var)
IFS="$OLD_IFS"
for str in ${array[@]}; do
  echo '----------------------------'
  echo $str
  cd $dir/$str
  if [ ! -d .git ]; then
    echo "非git仓库"
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
