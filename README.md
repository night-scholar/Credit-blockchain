使用PBFT共识算法，以及UTXO设计模型
毕业论文，设计思路如下：
（1）数据查询方向数据提供方发起数据查询请求，数据提供方获取用户授权。
（2）数据提供方将用户征信数据存储在区块链中。
（3）将添加好的数据经过对数据的加密、签名后发送给主节点。
（4）主节点向其他节点广播交易，对交易进行校验并返回结果。
（5）校验成功后数据提供方将数据发送至数据查询方，数据查询方将数据保存在自己的中心数据库中。