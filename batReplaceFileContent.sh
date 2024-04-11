#!/bin/bash
# 批量替换指定目录下指定文件列表中的文本

from=$1         #原文本
to=$2           #目标文本
includeArr=($3) #指定的文件名列表，为空则处理目录下所有文件,比如："test1 test2"
ignoreArr=($4)  #忽略的文件名列表，优先以指定文件名列表为准，比如："test1 test2"
dir=${5-$(pwd)} #文件目录，默认脚本执行目录

in_array() {
    local array="$1[@]"; shift
    local needle=$1; shift
    local result="1"
    for element in "${!array}"; do
        echo 1111 $element
        if [[ $element == $needle ]]; then
            result="0"
            break
        fi
    done
    return $result
}
in_array includeArr "test"

exit 1


if [[ "$(dirname $0)" == "$(pwd)" ]]; then
  . functions.sh
else
  . $(dirname $0)/functions.sh
fi

handle() {
  for d in $(ls $1); do
    _fromPath=$1/$d

    if [ -f $_fromPath ]; then
      if [[ ${#includeArr[@]} > 0 ]]; then
        if in_array includeArr $d; then
          replace $from $to $_fromPath
        fi
      elif [[ ${#ignoreArr[@]} > 0 ]]; then
        if ! in_array ignoreArr $d; then
          replace $from $to $_fromPath
        fi
      else
        replace $from $to $_fromPath
      fi
    fi
    if [[ -d $_fromPath ]]; then
      handle $_fromPath
    fi
  done
}

handle $dir
