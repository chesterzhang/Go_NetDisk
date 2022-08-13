### 进入 docker mysql-master
``` 
docker exec -it mysql-master bash
```

登录mysql
``` 
mysql -u root -p
```
### 进入 docker  resdis-test

``` 
 docker exec -it redis-test /bin/bash
```
进入 redis 客户端
``` 
redis-cli
```
### redis 设置登录密码
进入redis 客户端后, 直接输入
```
config set requirepass 密码
```

以后登陆后, 需要输入下面的命令来验证登录
``` 
auth 密码
```

### docker 安装 rabbitmq 
#### 拉取镜像
``` 
docker pull rabbitmq:management
```

#### 运行镜像
``` 
docker run -d -p 5672:5672 -p 15672:15672 --name rabbitmq rabbitmq:management
```
#### 登录浏览器图形界面
``` 
http://127.0.0.1:15672
```
或者直接
``` 
localhost:15672
```