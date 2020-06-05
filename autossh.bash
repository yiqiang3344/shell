#!/usr/bin/expect -f
set timeout 20

if { [llength $argv] < 3} {
    puts "Usage:"
    puts "$argv0 username host passwd"
    exit 1
}
set username [lindex $argv 0]
set host [lindex $argv 1]
set passwd [lindex $argv 2]
set passwderror 0

spawn ssh $username@$host

expect {
    "*assword:*" {
        if { $passwderror == 1 } {
            puts "passwd is error"
            exit 2
        }
        set timeout 1000
        set passwderror 1
        send "$passwd\r"
        # exp_continue
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