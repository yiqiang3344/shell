#!/bin/bash
# 批量给指定目录的指定仓库打指定tag

branch=$1       #分支
repos=$2        #仓库名列表，逗号分割
tag=$3          #tag
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
  git checkout $branch
  if [ $? -ne 0 ]; then
    continue
  fi
  git pull origin $branch
  if [ $? -ne 0 ]; then
    continue
  fi
  git tag $tag
  if [ $? -ne 0 ]; then
    continue
  fi
  git push origin $tag
done
