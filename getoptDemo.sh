#!/usr/bin/env bash
# getopt增强版参数解析示例，包括短标识，长标识

help() {
  cat <<EOF
OPTION
  -h, --help     帮助信息（不接受值）
  -a, --along    必填参数
  -b, --blong    可选参数,短标签传参时值需紧贴参数
DESCRIPTION
  方法说明
EXAMPLES
  getoptDemo.sh -h
  getoptDemo.sh -a 1 -b2 3 4
  getoptDemo.sh --along=1 -b2 3 4
  getoptDemo.sh --along=1 --blong=2 3 4
EOF
}
getopt -T
getoptVersion=$?
if [ "$getoptVersion" != "4" ] && [ "$(uname)" == "Darwin" ]; then
  brew -v >/dev/null
  if [ "$?" != "0" ]; then
    echo 'Please install brew for Mac: ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"'
  fi
  echo 'Please install gnu-getopt for Mac: brew install gnu-getopt'
  exit 1
elif [[ "$getoptVersion" != "4" ]]; then
  echo 'only support getopt version 4'
  exit 1
fi

#处理参数，规范化参数
ARGS=$(getopt -o ha:b:: --long help,along:,blong:: -n "$0" -- "$@")
if [ $? != 0 ]; then
  echo "[ERR]Terminating..."
  exit 1
fi
#重新排列参数顺序
eval set -- "${ARGS}"
#通过shift和while循环处理参数
while true; do
  case $1 in
  -h | --help)
    help
    exit 0
    ;;
  -a | --along)
    echo "Must a, argument $2"
    shift
    ;;
  -b | --blong)
    case "$2" in
    "")
      echo "Option b, no argument"
      shift
      ;;
    *)
      echo "Option b, argument $2"
      shift
      ;;
    esac
    ;;
  --)
    shift
    break
    ;;
  ?)
    echo "[ERR]Unknown option: $1"
    exit 1
    ;;
  esac
  shift
done

#处理剩余的参数
echo remaining parameters=[$@]
echo \$1=[$1]
echo \$2=[$2]
