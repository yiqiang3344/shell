#!/bin/sh

from=$1
to=$2
ignore_arr=(monitor.conf)

in_array() {
    local array="$1[@]"; shift
    local needle=$1; shift
    local result=1
    for element in "${!array}"; do
        if [[ $element == $needle ]]; then
            result=0
            break
        fi
    done
    return $result
}

replace(){
    echo $3
    if [[ $(uname -a | cut -d ' ' -f 1) == "Linux" ]];then
        sed -i "s/$1/$2/g" $3
    else
        sed -i "" "s/$1/$2/g" $3
    fi
}

scanDir(){
    for d in `ls $1`;do
        _fromPath=$1/$d
        if in_array ignore_arr $d;then
            continue;
        fi
        if [[ -d $_fromPath ]];then
            scanDir $_fromPath
            continue;
        fi
        replace $from $to $_fromPath;
    done
}
scanDir $(pwd)
