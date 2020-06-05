#!/bin/bash

. ~/bash/functions

scanDir(){
    for d in `ls $1`;do
        _fromPath=$1/$d
        if in_array arr $d;then
            replace $from $to $_fromPath;
        fi
        if [[ -d $_fromPath ]];then
            scanDir $_fromPath $_arr
        fi
    done
}

from=$1
to=$2
arr=($3)
scanDir $(pwd)
