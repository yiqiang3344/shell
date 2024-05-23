## 常用shell工具

- [autossh.sh](autossh.sh) 自动ssh服务器工具
- [batReplaceFileContent.sh](batReplaceFileContent.sh) 批量替换文件内容工具
- [functions](functions) 常用函数库
- [getoptDemo.sh](getoptDemo.sh) getopt使用示例（兼容mac和linux）
- [gitBatCheckoutBranch.sh](gitBatCheckoutBranch.sh) git批量切分支工具
- [gitBatCreateBranch.sh](gitBatCreateBranch.sh) git批量创建分支工具
- [gitBatDeleteBranch.sh](gitBatDeleteBranch.sh) git批量删除分支工具
- [gitBatMergeBranchToBranch.sh](gitBatMergeBranchToBranch.sh) git批量合并分支工具
- [gitBatPull.sh](gitBatPull.sh) git批量pull工具
- [gitBatResetBranch.sh](gitBatResetBranch.sh) git批量reset工具
- [gitBatShowStatus.sh](gitBatShowStatus.sh) git批量查看仓库状态工具
- [gitBatTag.sh](gitBatTag.sh) git批量打tag工具
- [sysInfo.sh](sysInfo.sh) 查看系统信息
- [go-tools-mac](go-tools-mac) go版本工具集mac系统版，源码详情见[go-tools](go-tools/README.MD)，配置文件为[config.yaml](config.yaml)
  - `./go-tools-mac setGitlabProjectsMember`  给指定用户名的gitlab用户批量设置指定仓库的权限，可通过配置、命令行选项或终端交互方式设置参数。
  - `./go-tools-mac gitClone`  批量clone仓库的代码到指定目录，可通过配置、命令行选项或终端交互方式设置参数。
- [go-tools-linux](go-tools-linux) go版本工具集linux系统版，详情同[go-tools-mac](go-tools-mac)
- [checkMvnDependency.sh](checkMvnDependency.sh) 拉取配置文件中指定的java项目代码仓库到指定目录，并检查maven依赖是否包含指定包