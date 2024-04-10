#!/bin/bash
# 批量给指定目录下的指定仓库列表创建指定分支

branch=$1        #要创建的分支名
repos=$2         #仓库列表，逗号分割
dir=${3:-$(pwd)} #仓库所在目录，默认脚本执行目录
OLD_IFS="$IFS"
IFS=","
array=($repos)
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
  git push origin --delete $branch
  git checkout -b $branch
  if [ $? -ne 0 ]; then
    continue
  fi
  git push origin $branch
  if [ $? -ne 0 ]; then
    continue
  fi
done
