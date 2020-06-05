#!/bin/bash

branch=$1
var=$2
method=$3
root=~/docker/code/xjd
OLD_IFS="$IFS"
IFS=","
array=($var)
IFS="$OLD_IFS"
for str in ${array[@]};do
    echo '----------------------------'
    echo $str
    cd $root/$str
    git fetch
    if [[ $method == "" ]];then
        git checkout $branch
        git pull origin $branch
    fi
    if [[ $method == "rebase" ]];then
        git checkout master
        git pull origin master --rebase
        git checkout $branch
        git pull origin $branch --rebase
        git rebase master
        git push origin $branch
    fi
    if [[ $method == "merge" ]];then
        git checkout master
        git pull origin master
        git checkout $branch
        git pull origin $branch
        git merge master
        git push origin $branch
    fi
done;