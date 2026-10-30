# Luck5结算实现说明

更新日期：2026-10-30

依据用户指定的D:/project/fg-game-server/fgServer/doc/prod扩展/rpc/luck5.md执行正常发牌、预占和差额结算；用户随后明确补充“不做防止重复扣款、重复派奖”，该补充覆盖原文档的lottery防重、候选结果记忆及资金阶段补写要求。

## 原协议不变

- Deal接收userId、agentId、gameId、currencyType、roundId、betAmount、reserveAmount。
- Settle保持userId、agentId、gameId、currencyType、roundId、winAmount、recordJson、attemptId这8个字段。
- 不新增betAmount或reserveAmount到Settle，也不修改游戏服务器调用。此前助手提出的两个新增字段方案已经撤回。
- attemptId在协议中保留，但lottery不再以它做重复请求判定，也不要求它必填。

## 正常局上下文

现有Settle没有原下注和预占金额，因此Deal成功后仍需保存本局的B、R和原池标识，供Settle计算正常差额。luck5:round只承担这一业务上下文用途，不保存FundsDone、RecordDone、Attempts、Signature等防重或候选状态。

- Deal直接检查本次余额和池额，通过后扣款、预占并保存上下文；不查重、不回放原响应。
- 成功Deal写入luck5:round:*时设置1小时TTL；读取和结算不续期，正常结算完成仍立即删除。超过1小时未结算，上下文过期后按现有逻辑返回参数错误。本次仅对新写入生效，不批量修改历史无TTL的key。
- Deal失败不保存候选拒绝结果。
- Settle从上下文取原下注、预占、汇率和原池，检查P+R是否足够支付W。
- 水池不足保留正常Deal上下文，不改资金、不保存候选拒绝结果。
- 成功时钱包加完整W，原池加R-W，保存正常游戏记录，随后删除本局上下文。
- 原池、金额和玩家身份的正常校验保留；无上下文时返回参数错误，不猜测下注或预占。
- 不保留已完成局作防重，不做候选签名比较、重复请求回放或仅补记录的资金阶段恢复；调用方负责防重。

## 记录余额

按源文档要求，实际落库的cents.endBalanceCents及列表end_chips使用本次真实派奖后的余额，只修改落库副本，不改请求中原有牌面及金额内容。

## 与水池重置分开

水池重置仍按[水池重置简化说明](水池重置简化说明.md)：先清内存水池缓存，再删除四类Redis累计，不扫描或删除Luck5业务上下文。sessions、inflight、reset_rounds和附加seen机制不会恢复。

## 验证和部署

已验证原协议调用、正常Deal与Settle、原组池差额计算、结算后上下文删除、无候选防重状态及不写seen。游戏服务器及其协议文件没有修改，新增两个金额字段的待执行脚本已撤销。

现有历史Redis数据未在现场清理，本次未部署或执行真实资金操作。
