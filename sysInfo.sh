#!/usr/bin/env bash
# 查看centos服务器常用系统信息
set -u #有未定义的变量时要报错

if [[ "$(dirname $0)" == "$(pwd)" ]]; then
  . functions
else
  . $(dirname $0)/functions
fi

echo "服务器产品名"
if program_exists dmidecode; then
  dmidecode | grep "Product Name"
else
  echo "dmidecode命令不存在，无法查看服务器产品名"
fi
echo ""

echo "系统信息"
if program_exists uname; then
  uname -a
else
  echo "uname命令不存在，无法查看系统信息"
fi
echo ""

echo "操作系统"
if program_exists lsb_release; then
  lsb_release -d
else
  echo "lsb_release命令不存在，无法查看操作系统"
fi
echo ""

echo "CPU"
if program_exists dmidecode; then
  dmidecode -t 4
else
  echo "dmidecode命令不存在，无法查看CPU"
fi
echo ""

echo "CPU核心信息"
if [ -f /proc/cpuinfo ]; then
  physicalNumber=0
  coreNumber=0
  logicalNumber=0
  HTNumber=0

  logicalNumber=$(grep "processor" /proc/cpuinfo | sort -u | wc -l)
  physicalNumber=$(grep "physical id" /proc/cpuinfo | sort -u | wc -l)
  coreNumber=$(grep "cpu cores" /proc/cpuinfo | uniq | awk -F':' '{print $2}' | xargs)
  HTNumber=$((logicalNumber / (physicalNumber * coreNumber)))

  echo "****** CPU Information ******"
  echo "Logical CPU Number  : ${logicalNumber}"
  echo "Physical CPU Number : ${physicalNumber}"
  echo "CPU Core Number     : ${coreNumber}"
  echo "HT Number           : ${HTNumber}"
  echo "*****************************"
else
  echo "/proc/cpuinfo文件不存在，无法查看CPU核心信息"
fi
echo ""

echo "CPU位数"
getconf LONG_BIT
echo ""

echo "内存信息"
if program_exists dmidecode; then
  dmidecode -t memory | grep 'Size'
else
  echo "dmidecode命令不存在，无法查看内存信息"
fi
echo ""

echo "硬盘"
if program_exists df; then
  df -h
else
  echo "df命令不存在，无法查看硬盘信息"
fi
