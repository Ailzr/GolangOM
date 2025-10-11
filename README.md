# 使用方式

1. 使用`git clone`指令克隆本项目

2. 准备golang环境，运行`go build`指令编译项目（windows环境下编译linux的执行程序指令:`$env:CGO_ENABLED="0"; $env:GOOS="linux"; $env:GOARCH="amd64"; go build -o golang-om`，适用于amd64架构的linux系统)

3. 修改config文件中的configs.yaml文件的配置信息

4. 启动项目，访问 服务器ip:25888/golang-om

5. 使用图形化界面进行操作
