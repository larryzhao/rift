# PAC 规则简化 —— 最终设计

把代理规则从一份 7985 行的 gfwlist `pac.js`,改为「**个人域名列表 + 自动同步的 gfwlist**」两文件,由 rift 实时合成 PAC。日常只维护一个纯域名列表,墙名单自动跟上游更新。

## 文件布局
```
~/.rift/pac/
├── domains.txt          # 你维护:个人域名列表(纯域名,! 前缀=强制直连)
├── gfwlist.txt          # rift pac update 同步(base64,上游 ~6h 更新一次)
└── pac.js.gfwlist.bak   # 迁移前旧 pac.js 的备份
```
- `domains.txt` 与 `gfwlist.txt` 都在 `~/.rift/pac/` 下,不入 git。
- 仓库内置一份初始域名列表见 [`domains.seed.txt`](./domains.seed.txt)(30 个 gfwlist 未覆盖/需强制的域名)。

## domains.txt 语法
```
# 注释,空行忽略
claude.ai             # 无前缀 = 强制走代理(即使 gfwlist 判直连)
anthropic.com
!fast-cdn.example.com # ! 前缀 = 强制直连(即使 gfwlist 判代理)
!intranet.corp.com
```
- 主域名即覆盖所有子域:`claude.ai` 命中 `api.claude.ai` / `www.claude.ai`。
- 更具体的子域规则优先;同层 `!`(直连)优先于代理。
- 已被 gfwlist 覆盖的域名无需列出,只放 gfwlist 未覆盖或你想强制的。

## 决策顺序(运行时 FindProxyForURL)
1. 命中 `domains.txt` → 用它的判定(直连 / 代理)
2. 否则命中 `gfwlist.txt` → 代理
3. 都没命中 → DIRECT

```js
function FindProxyForURL(url, host) {
  var u = userDecision(host);          // domains.txt,最高优先
  if (u) return u;                     // "DIRECT" 或 proxy
  if (defaultMatcher.matchesAny(url, host) instanceof BlockingFilter) {
    return proxy;                      // gfwlist 兜底
  }
  return "DIRECT";
}
```

## 为什么是「实时生成」而非「生成静态文件」
设计时比较过两种生成时机:
- **静态文件式**:编辑 domains.txt 后跑一次 `gen` 命令落地 pac.js。server 不变,但多一步、易忘。
- **实时生成式(已采用)**:server 每次响应 `/pac/proxy.js` 时即时合成。编辑 domains.txt **保存即生效**,无中间产物。

PAC 请求频率极低,实时读两文件 + 生成的开销可忽略,故选实时生成。

## 为什么仍带 gfwlist 引擎
gfwlist 规则含 `||domain`、`@@` 白名单例外、正则、URL 路径、通配符等多种 AutoProxy 语法,
要忠实匹配必须用其原生引擎。该引擎(Adblock Plus matcher,来自 gfwlist2pac)以
`pac/engine.js`(712 行)`go:embed` 嵌入,**用户永不接触**;
gfwlist.txt 的每行规则在生成时直接喂给 `Filter.fromText()`。
因此「你只维护纯域名」与「忠实保留 gfwlist 全部规则」两者兼得。

> 生成 PAC = 规则原样搬运(正则/lookahead 照抄),由浏览器的 JS 引擎执行,**零损失**。

## 命令
| 命令 | 作用 |
|------|------|
| `rift pac` | 启动本地 PAC server 并设置系统代理 |
| `rift pac update` | 下载最新 gfwlist(上游 ~6h 更新) |
| `rift domain proxy <domain>` | 把域名加入 domains.txt → 走代理 |
| `rift domain direct <domain>` | 把域名加入 domains.txt(`!` 前缀) → 直连 |
| `rift domain status <domain>` | 查当前 domains.txt + gfwlist 下该域名走代理还是直连,并显示命中的规则 |

`domain proxy/direct` 增改即去重、保留注释;`status` 输出命中的具体规则与来源
(domains.txt / gfwlist / gfwlist-whitelist / default)。

## 代码结构
| 文件 | 职责 |
|------|------|
| `pac/generate.go` | `GeneratePAC(domains, gfwlist)`:解析两文件 + 嵌入引擎 + 组装 PAC |
| `pac/engine.js` | 嵌入的 Adblock matcher(来自 gfwlist2pac / Adblock Plus) |
| `pac/sync.go` | `SyncGFWList()`:下载上游 gfwlist.txt |
| `pac/manage.go` | `NormalizeDomain` / `SetDomainRule`:维护 domains.txt |
| `pac/lookup.go` | Go 原生 AutoProxy 匹配器,供 `domain status` 使用 |
| `pac/server.go` | `/pac/proxy.js` 实时调用 `GeneratePAC` |
| `repository.go` | `PACDomainsFile()` / `PACGFWListFile()` 路径 |
| `commands/pac.go` | `rift pac` / `rift pac update` |
| `commands/domain.go` | `rift domain proxy/direct/status` |

## status 与实际路由的两套实现(已知差异)
- **实际上网**:浏览器跑 PAC 里的 `engine.js`,用 **JS 正则**(支持 lookahead),是权威。
- **`rift domain status`**:用 `pac/lookup.go` 的 **Go RE2** 匹配器,**不支持 lookahead**,
  编译失败的规则会被静默丢弃。

因此 `status` 在极冷门正则规则上可能与实际路由有出入(当前 gfwlist 4340 条里仅 1 条
lookahead,且实测不触发)。**实际代理行为始终正确**,差异只影响 status 这个诊断读数。
已通过 node 跑真实 PAC 与 lookup.go 交叉校验(400+ 域名 0 不一致)。

## 验证
- `go build ./cmd/cli/...` / `go test ./...` / `go vet` / `gofmt` 全通过。
- 真机跑过三个 domain 命令;node 实跑生成的 PAC 路由正确(含正则/通配/白名单)。

## 后续优化(TODO,本次未做)
- **autoupdate**:`rift pac update` 目前为手动命令,后续并入整体 autoupdate 逻辑。
- **runner.go 系统代理**:当前硬编码 `Wi-Fi`、用 `command.Start()` 吞错误;
  应枚举所有网络服务、改用 `CombinedOutput()` 捕获错误。
- **PAC 精简**:可在 Go 侧把 gfwlist 域名类规则降维成域名字典,去掉 engine.js,
  PAC 从 ~97KB 降到几 KB;代价是损失部分 URL 路径/正则精度,需处理 JS 正则 vs RE2 方言。
- **status 可见性**:让 status 输出"无法求值的规则数",把静默差异变为可见提示。
- **engine.js 许可标注**:在文件头补充来源(gfwlist2pac / Adblock Plus)与许可说明。
