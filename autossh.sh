#!/usr/bin/env bash
# ssh自动输入密码并登录服务器

# 获取服务器标识配置，用于判断读取什么配置
host=${1:?"入参[主机标识]不能为空"}

# 导入配置，需要当前目录自己创建配置文件，格式：
## export sshIp_xxxx=xxxxx
## export sshUser_xxxx=xxxxx
## export sshPass_xxxx=xxxxx
. $(dirname $0)/autossh.cfg

username=$(eval echo '$'"sshUser_$host")
hostname=$(eval echo '$'"sshIp_$host")
password=$(eval echo '$'"sshPass_$host")
port=$(eval echo '$'"sshPort_$host")
if [ -z $port ]; then
  port=22
fi

sw_login() {
  expect -c "
# 每个判断等待两秒
set timeout 2
spawn bash -c \"ssh -p$1 $2@$3\"
# 判断是否需要保存秘钥
expect {
  \"yes/no\"   { send yes\n }
}
# 判断发送密码
expect {
  \"*assword\" { send $4\n }
}
# 停留在当前登录界面
interact
"
}
sw_login $port $username $hostname $password
