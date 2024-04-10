#!/bin/bash
# 批量把某个目录下指定的git仓库切换到指定分支

branch=$1       #分支名
var=$2          #git仓库名列表，逗号分割
method=$3       #如果需要合并时的合并方式:rebase,merge
dir=${4-$(pwd)} #仓库所在目录，默认脚本执行目录

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
  git fetch
  if [[ $method == "" ]]; then
    git checkout $branch
    if [ $? -ne 0 ]; then
      continue
    fi
    git pull origin $branch
    if [ $? -ne 0 ]; then
      continue
    fi
  fi
  if [[ $method == "rebase" ]]; then
    git checkout master
    if [ $? -ne 0 ]; then
      continue
    fi
    git pull origin master --rebase
    if [ $? -ne 0 ]; then
      continue
    fi
    git checkout $branch
    if [ $? -ne 0 ]; then
      continue
    fi
    git pull origin $branch --rebase
    if [ $? -ne 0 ]; then
      continue
    fi
    git rebase master
    if [ $? -ne 0 ]; then
      continue
    fi
    git push origin $branch
    if [ $? -ne 0 ]; then
      continue
    fi
  fi
  if [[ $method == "merge" ]]; then
    git checkout master
    if [ $? -ne 0 ]; then
      continue
    fi
    git pull origin master
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
    git merge master
    if [ $? -ne 0 ]; then
      continue
    fi
    git push origin $branch
    if [ $? -ne 0 ]; then
      continue
    fi
  fi
done
