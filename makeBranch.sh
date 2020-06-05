#!/bin/bash

branch=$1
var=$2
root=${3:-~/docker/code/xjd}
OLD_IFS="$IFS"
IFS=","
array=($var)
IFS="$OLD_IFS"
for str in ${array[@]};do
    echo '----------------------------'
    echo $str
    cd $root/$str
    git fetch
    git checkout master
    git pull origin master
    git branch -D $branch
    git push origin --delete $branch
    git checkout -b $branch
    git push origin $branch
done;