#!/bin/bash

branch=$1
var=$2
origin=$3
root=~/htdocs/xjd
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
    if [[ $origin != "" ]];then
        git push origin --delete $branch
    fi
done;