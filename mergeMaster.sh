#!/bin/bash

branch=$1
var=${2}
toBranch=master
root=~/docker/code/xjd
OLD_IFS="$IFS"
IFS=","
array=($var)
IFS="$OLD_IFS"
for str in ${array[@]};do
    echo '----------------------------'
    echo $str
    cd $root/$str
    git checkout $toBranch
    git pull origin $toBranch
    git checkout $branch
    git pull origin $branch
    git checkout $toBranch
    git merge $branch
done;