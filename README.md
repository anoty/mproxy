# 这货有什么用
应用场景，业务已经极限优化到每个机器都装一个memcache，来减少网络请求和提高性能的地步，mproxy会自动将key和value同步到每一个memcache。

# 怎么用
go get github.com/woodane/mproxy

cd $GOPATH/src/mproxy/mpd

go install

然后就像使用memcache一样使用这个代理就行了。

