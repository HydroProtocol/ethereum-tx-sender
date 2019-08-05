# Intro

用来发送 transaction 的通用组件。

功能

- 托管nonce，调用方无需操心
- 如果不指定gasPrice，会使用合理gasPrice
- 如果不指定gasLimit，会使使用gasEstimate * 1.2
- 交易重试
- 使用nights-watch监测交易状态

有如下使用要求

- 私钥管理使用pkm
- 所发的交易符合pkm的权限检查
- 不能有其他进程同步操作nonce

## client demo

[oracle-price-feeder](https://git.ddex.io/bfd/price-oracle-feeder/blob/master/oracle_writer.go#L64-90)
 
[2link](#todo)

[3link](#todo)