#!/usr/bin/env bash
# 批量替换指定目录下指定文件列表中的文本
set -Eeuo pipefail #有未定义的变量时要报错，报错时停止脚本

from=${1:?"原文本不能为空"} #原文本
to=${2:-""}          #目标文本
includeArr=(${3:-})  #指定的文件名列表，为空则处理目录下所有文件,比如："test1 test2"
ignoreArr=(${4:-})   #忽略的文件名列表，优先以指定文件名列表为准，比如："test1 test2"
dir=${5:-$(pwd)}     #文件目录，默认脚本执行目录

if [[ "$(dirname $0)" == "$(pwd)" ]]; then
  . functions
else
  . $(dirname $0)/functions
fi

handle() {
  for d in $1; do
    _fromPath=$d
    fileName=${i/$dir\//}
    if [ -f $_fromPath ]; then
      if [[ ${#includeArr[@]} -gt 0 ]]; then
        if in_array includeArr $fileName; then
          replace $from $to $_fromPath
        fi
      elif [[ ${#ignoreArr[@]} -gt 0 ]]; then
        if ! in_array ignoreArr $fileName; then
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
