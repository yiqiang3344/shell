#!/usr/bin/env bash
# 查看centos服务器常用系统信息
echo "服务器产品名"
dmidecode | grep "Product Name"

echo "系统信息"
uname -a

echo "操作系统"
head -n 1 /etc/issue

echo "CPU"
dmidecode -t 4

echo "CPU核心信息"
physicalNumber=0
coreNumber=0
logicalNumber=0
HTNumber=0

logicalNumber=$(grep "processor" /proc/cpuinfo|sort -u|wc -l)
physicalNumber=$(grep "physical id" /proc/cpuinfo|sort -u|wc -l)
coreNumber=$(grep "cpu cores" /proc/cpuinfo|uniq|awk -F':' '{print $2}'|xargs)
HTNumber=$((logicalNumber / (physicalNumber * coreNumber)))

echo "****** CPU Information ******"
echo "Logical CPU Number  : ${logicalNumber}"
echo "Physical CPU Number : ${physicalNumber}"
echo "CPU Core Number     : ${coreNumber}"
echo "HT Number           : ${HTNumber}"
echo "*****************************"

echo "CPU位数"
getconf LONG_BIT

echo "内存信息"
dmidecode -t memory | grep ' Size'

echo "硬盘"
df -h