#!/bin/bash
# 批量删除指定仓库的指定分支，切换为master分支。

branch=$1       #分支
repos=$2        #仓库名列表，逗号分割
delOrigin=$3    #是否删除远程分支
dir=${4-$(pwd)} #仓库所在目录，默认脚本执行目录

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
  if [ $? -ne 0 ]; then
    continue
  fi
  if [ -z $delOrigin ]; then
    git push origin --delete $branch
  fi
done
