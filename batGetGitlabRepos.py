# 批量拉取gitlab仓库到本地
import json
import os
import shlex
import subprocess
import sys
import time
from urllib.request import urlopen

gitlabAddr = sys.argv[1]  # gitlab地址
gitlabToken = sys.argv[2]  # gitlab的token
gitlabCodeRootPath = sys.argv[3]  # 本地存储的目录
branch = sys.argv[4]  # 要切换的分支
expectGroup = sys.argv[5].split(',')  # 期望的仓库组列表，逗号分割
expectRepos = sys.argv[6].split(',')  # 期望的仓库名列表，逗号分割
ignoreGroup = sys.argv[7].split(',')  # 排出的仓库组列表，期望的仓库名优先，逗号分割
ignoreRepos = sys.argv[8].split(',')  # 排出的仓库名列表，期望的仓库名优先，逗号分割

dealNum = 0

for index in range(100):
    index = index + 1
    url = "%s/api/v4/projects?private_token=%s&per_page=100&page=%d&order_by=name" % (gitlabAddr, gitlabToken, index)
    allProjects = urlopen(url)
    allProjectsDict = json.loads(allProjects.read().decode())
    if len(allProjectsDict) == 0:
        break

    for thisProject in allProjectsDict:
        try:
            dealNum = dealNum + 1
            thisProjectURL = thisProject['ssh_url_to_repo']
            group = thisProject['namespace']['name']
            name = thisProject['name']

            if len(expectGroup) > 0 and group not in expectGroup:
                print('不在期望的group中:' + group + ' ' + thisProjectURL)
                continue

            if len(expectRepos) > 0 and name not in expectRepos:
                print('不在期望的repos中:' + name + ' ' + thisProjectURL)
                continue

            if group in ignoreGroup:
                print('忽略的group:' + group + ' ' + thisProjectURL)
                continue

            if name in ignoreRepos:
                print('忽略的repos:' + name + ' ' + thisProjectURL)
                continue

            os.system('mkdir -p ' + gitlabCodeRootPath + '/' + group)
            thisProjectPath = gitlabCodeRootPath + '/' + group + '/' + name

            print(thisProjectPath + ' ' + thisProjectURL)

            if os.path.exists(thisProjectPath):
                command1 = f'git -C {thisProjectPath} checkout {branch}'
                command2 = f'git -C {thisProjectPath} pull'
                print(command1 + "&" + command2)
                resultCode = subprocess.Popen(command1 + "&" + command2, shell=True)
                time.sleep(1)
            else:
                command = shlex.split('git clone %s %s' % (thisProjectURL, thisProjectPath))
                resultCode = subprocess.Popen(command)
                time.sleep(5)
        except Exception as e:
            print("Error on %s: %s" % (thisProject['ssh_url_to_repo'], e.__str__))

print('共处理仓库数:' + str(dealNum))
