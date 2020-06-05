#!/usr/bin/expect -f
set timeout 20
set passwd 1TGhPx_SPNXDwKGT
set passwderror 0
spawn ssh -i ~/work/new_junqiang.yi.pem junqiang.yi@192.168.59.143
expect {
    "*pem*" {
        if { $passwderror == 1 } {
            puts "passwd is error"
            exit 2
        }
        set timeout 1000
        set passwderror 1
        send "$passwd\r"
        interact #把控制权交给终端
    }
    "*es/no)?*" {
        send "yes\r"
        exp_continue
    }
    timeout {
        puts "connect is timeout"  
        exit 3  
    }
}