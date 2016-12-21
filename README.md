# redis-guestbook


构建好的镜像位于 registry.dataos.io/library/redisguestbook ， 便于从DF页面部署，所以放到了library里。

demo演示了往master写，从slave读。

* 需要一个redis集群的哨兵地址、集群名字；
* 需要导出3000端口；
* 导出一个router；
* 配置环境变量：
  * EnvName_SentinelHost  #哨兵地址的环境变量名
  * EnvName_SentinelPort  #哨兵端口的环境变量名
  * EnvName_ClusterName   #集群名字的环境变量名
  * EnvName_Password (optional) #连接密码的环境变量名
  
例子：
  
  * EnvName_Password=BSI_REDIS_REDIS_PASSWORD
  * EnvName_SentinelHost=BSI_REDIS_REDIS_HOST
  * EnvName_SentinelPort=BSI_REDIS_REDIS_PORT
  * EnvName_ClusterName=BSI_REDIS_REDIS_NAME
