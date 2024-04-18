#!/usr/bin/env bash
# 批量拉取gitlab仓库到本地
set -Eeuo pipefail #有未定义的变量时要报错，报错时停止脚本

# 导入配置，需要当前目录自己创建配置文件，格式：
## export gitlabAddr='https://gitlab.cn'
## export gitlabToken='xxxx'
## export gitlabCodeRootPath='/path/to/gitlab/code'
## export branch='master'
## export expectGroup='test,test1'
## export expectRepos='test,test1'
## export ignoreGroup='test,test1'
## export ignoreRepos='test,test1'
. $(dirname $0)/batGetGitlabRepos.cfg

if [ -z $gitlabAddr ]; then
  echo "\$gitlabAddr未配置"
  exit 1
fi
if [ -z $gitlabToken ]; then
  echo "\$gitlabToken未配置"
  exit 1
fi
if [ -z $gitlabToken ]; then
  echo "\$gitlabToken未配置"
  exit 1
fi
if [ -z $gitlabCodeRootPath ]; then
  echo "\$gitlabCodeRootPath未配置"
  exit 1
fi
if [ -z $branch ]; then
  echo "\$branch未配置"
  exit 1
fi

python3 $(dirname $0)/batGetGitlabRepos.py "$gitlabAddr" "$gitlabToken" "$gitlabCodeRootPath" "$branch" "$expectGroup" "$expectRepos" "$ignoreGroup" "$ignoreRepos"
