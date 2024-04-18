#!/usr/bin/env bash
# 批量给指定目录下的仓库打指定tag
set -u #有未定义的变量时要报错

branch=${1:?"分支不能为空"} #分支
tag=${2:?"tag不能为空"}   #tag
expectRepos=(${3:-})  #期望处理的git仓库名列表，空格分割
dir=${4:-$(pwd)}      #仓库所在目录，默认脚本执行目录

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
    git tag $tag
    if [ $? -ne 0 ]; then
      continue
    fi
    git push origin $tag
  done
}

handle $dir

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
