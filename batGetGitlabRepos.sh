#!/usr/bin/env bash
# 批量拉取gitlab仓库到本地

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

python3 $(dirname $0)/batGetGitlabRepos.py "$gitlabAddr" "$gitlabToken" "$gitlabCodeRootPath" "$branch" "$expectGroup" "$expectRepos" "$ignoreGroup" "$ignoreRepos"
