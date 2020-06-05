#!/bin/bash

branch=$1
var=$2
flag=${3-remove}
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
    echo mode:$flag
    if [[ $flag == "reset" ]];then
		git checkout $branch
	    git pull origin $branch
	    git checkout .
	    git reset HEAD --hard
    fi
    if [[ $flag == "remove" ]];then
    	git checkout develop
    	git branch -D $branch
    	git checkout $branch
    fi
done;