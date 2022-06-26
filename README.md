####


#### redis 设置登录密码
进入redis 客户端后, 直接输入
```
config set requirepass 密码
```

以后登陆后, 需要输入下面的命令来验证登录
``` 
auth 密码
```
