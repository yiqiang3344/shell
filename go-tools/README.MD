## 准备

安装`go1.20`以上版本

安装框架工具

```bash
wget -O gf "https://github.com/gogf/gf/releases/latest/download/gf_$(go env GOOS)_$(go env GOARCH)" && chmod +x gf && ./gf install -y && rm ./gf
```

## 开发

- 配置配置文件名环境变量`GF_GCFG_FILE=config.yaml`，配置文件在`hack`和`manifest/config`下。
- 工具命令配置在`go-tools/internal/cmd/init.go`

## 本地调试
```bash
cd go-tools
GF_GCFG_FILE=config.yaml gf run main.go --args="工具命令" #未设置全局配置文件环境变量时
gf run main.go --args="工具命令" #设置了全局配置文件环境变量时
```

## 编译
```bash
cd go-tools
GF_GCFG_FILE=config.yaml gf build main.go #未设置全局配置文件环境变量时
gf build main.go #设置了全局配置文件环境变量时
```
根据`go-tools/hack/config`下的配置生成的可执行文件，默认配置是在`go-tools/bin`目录下，生成了linux和mac系统的可执行文件。
二进制文件执行示例：
```bash
cd go-tools
./bin/darwin_amd64/go-tools -h
./bin/darwin_amd64/go-tools demo
```
