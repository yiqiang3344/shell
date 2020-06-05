#!/bin/bash

branch=$1
var=${2-'xinyongfei-app-api,xinyongfei-mobile,shoufuyou-task,shoufuyou-notify,shoufuyou-php-lib,shoufuyou-bill-system,shoufuyou-console,xinyongfei-cuishou,xinyongfei-cps,shoufuyou-sms-system,shoufuyou-fund-system'}
toBranch=${3-develop}
root=~/htdocs/xjd
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
    git push origin $toBranch
done;