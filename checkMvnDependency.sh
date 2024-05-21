#!/usr/bin/env bash
# 拉取配置文件中指定的java项目代码仓库到指定目录，并检查maven依赖是否包含指定包
set -Euo pipefail #有未定义的变量时要报错

grepCmd=""
codeDir="./code" #默认存放代码仓库的地址
readonly needJavaVersion="1.8" #依赖的java版本
readonly needMavenVersion="3.9.6" #依赖的maven版本

help() {
  cat <<EOF
OPTION
  -h, --help     帮助信息（不接受值）
  -g, --grep     [必填]要检查的依赖grep命令，如grep Hikari
  -d, --dir      [选填]代码存放目录,默认./code,短标签传参时值需紧贴参数
DESCRIPTION
  本地环境依赖
    - java${needJavaVersion}
    - maven${needMavenVersion}
    - 配置好gitlab ssh密钥
  脚本同目录下配置好gitlab仓库地址列表配置文件checkMvnDependency.cfg,格式：
  git@gitlab.xinyongfei.cn:infrastructure/techplayprod.git
  git@gitlab.xinyongfei.cn:infrastructure/techplaycore.git
EXAMPLES
  checkMvnDependency.sh -h
  checkMvnDependency.sh -g "grep Hikari" -d/tmp/code
  checkMvnDependency.sh -g "grep Hikari" --dir="/tmp/code"
  checkMvnDependency.sh --grep="grep Hikari" --dir="/tmp/code"
EOF
}

getopt -T
getoptVersion=$?
if [ "$getoptVersion" != "4" ] && [ "$(uname)" == "Darwin" ]; then
  brew -v >/dev/null
  if [ "$?" -ne "0" ]; then
    echo 'Please install brew for Mac: ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"'
  fi
  echo 'Please install gnu-getopt for Mac: brew install gnu-getopt'
  exit 1
elif [[ "$getoptVersion" != "4" ]]; then
  echo 'only support getopt version 4'
  exit 1
fi

#处理参数，规范化参数
ARGS=$(getopt -o hg:d:: --long help,grep:,dir:: -n "$0" -- "$@")
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
  -g | --grep)
    grepCmd=$2
    shift
    ;;
  -d | --dir)
    case "$2" in
    "")
      shift
      ;;
    *)
      codeDir=$2
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

#检查配置
if [ -z "$grepCmd" ]; then
  echo '[ERR]-g or --grep 参数值不能为空'
  exit 1
fi
if [ ! -f checkMvnDependency.cfg ]; then
  echo "仓库配置文件checkMvnDependency.cfg不存在"
  exit 1
fi
repos=($(cat checkMvnDependency.cfg))
if [ "${#repos[@]}" -eq "0" ]; then
  echo "仓库配置文件内容为空"
  exit 1
fi
javaVersion=$(java -version 2>&1 | sed '1!d' | sed -e 's/"//g' | awk '{print $3}' | awk -F'.' '{print $1"."$2}')
if [ "$javaVersion" != "$needJavaVersion" ]; then
  echo "请安装java${needJavaVersion}"
  exit 1
fi
mvnVersion=$(mvn -version 2>&1 | sed '1!d' | sed -e 's/"//g' | awk '{print $3}' | awk -F'.' '{print $1"."$2"."$3}')
if [ "$mvnVersion" != "$needMavenVersion" ]; then
  echo "请安装maven${needMavenVersion}"
  exit 1
fi

#准备代码目录
mkdir -p "$codeDir"
if [ $? -ne 0 ]; then
  echo "目录代码创建失败:$codeDir"
  exit 1
fi

#拉取代码
echo "拉取代码:"
for i in ${repos[*]}; do
  #提取group和name
  pathWithNameSpace=${i##*:}
  pathWithNameSpace=${pathWithNameSpace%.git}
  #  echo pathWithNameSpace:$pathWithNameSpace
  group=${pathWithNameSpace%/*}
  #  echo group:$group
  # 创建group目录
  mkdir -p "$codeDir/$group"
  if [ $? -ne 0 ]; then
    echo "仓库group目录创建失败:$codeDir/$group"
    exit 1
  fi
  name=${pathWithNameSpace##*/}
  #  echo name:$name
  #检查代码目录是否存在
  if [ -d "$codeDir/$group/$name" ]; then
    #存在则pull
    echo "git -C $codeDir/$group/$name pull"
    git -C "$codeDir/$group/$name" pull
  else
    #不存在则clone
    echo "git clone $i $codeDir/$group/$name"
    git clone $i $codeDir/$group/$name
  fi
done

echo "------------"

#开始检查
echo "开始检查:"
for i in ${repos[*]}; do
  #提取group和name
  pathWithNameSpace=${i##*:}
  pathWithNameSpace=${pathWithNameSpace%.git}
  #  echo pathWithNameSpace:$pathWithNameSpace
  group=${pathWithNameSpace%/*}
  #  echo group:$group
  name=${pathWithNameSpace##*/}
  #  echo name:$name
  #判断是否有pom.xml文件
  if [ ! -f $codeDir/$group/$name/pom.xml ]; then
    echo "[result:failed]$codeDir/$group/$name pom.xml不存在"
    continue
  fi
  grepRet=$(cd $codeDir/$group/$name && mvn dependency:tree | $grepCmd)
  if [ -z "$grepRet" ]; then
    echo "[result:failed]$codeDir/$group/$name"
  else
    echo "[result:success]$codeDir/$group/$name"
  fi
  echo "$grepRet"
done

echo "done"
