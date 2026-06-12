# 系统配置数据结构

系统配置保存在 `settings` 表中，目前只使用两行：

| key | 说明 |
| --- | --- |
| `public` | 公开配置，前端可以读取 |
| `private` | 私有配置，只给后端和管理员使用 |

## 部署环境变量

多实例部署时建议使用 `deploy/docker-compose.lb.yml`，让多个 app 副本共享同一个 Postgres 和 Redis。

| 变量 | 说明 |
| --- | --- |
| `JWT_SECRET` | 多实例必须固定一致，否则登录态会在不同副本间失效 |
| `STORAGE_DRIVER` | 多实例使用 `postgres`，不要共享 SQLite 文件 |
| `DATABASE_DSN` | Postgres/MySQL 连接字符串 |
| `DATABASE_MAX_OPEN_CONNS` | 单个 app 副本最大数据库连接数，0 表示使用驱动默认值 |
| `DATABASE_MAX_IDLE_CONNS` | 单个 app 副本最大空闲数据库连接数，0 表示使用驱动默认值 |
| `REDIS_ADDR` | Redis 地址，留空表示不启用缓存 |
| `REDIS_PASSWORD` | Redis 密码 |
| `REDIS_DB` | Redis DB 编号 |
| `REDIS_CACHE_TTL_SECONDS` | Redis 缓存过期时间，默认 60 秒 |
| `REDIS_KEY_PREFIX` | Redis key 前缀，默认 `infinite-canvas` |

当前 Redis 主要用于缓存系统配置、画廊图片文件等高频读取数据，并为多实例下的远程提示词定时同步提供分布式锁；注册、邀请码消耗和第三方账号绑定仍以数据库事务和唯一索引保证一致性。

## public.value

```json
{
  "modelChannel": {
    "availableModels": ["gpt-5.5", "gpt-image-2"],
    "modelCosts": [
      { "model": "gpt-5.5", "credits": 1 },
      { "model": "gpt-image-2", "credits": 10 }
    ],
    "defaultModel": "gpt-image-2",
    "defaultImageModel": "gpt-image-2",
    "defaultTextModel": "gpt-5.5",
    "systemPrompt": "",
    "allowCustomChannel": true
  },
  "auth": {
    "allowRegister": true,
    "registerCredits": 0,
    "linuxDo": {
      "enabled": false
    }
  },
  "checkIn": {
    "mode": "fixed",
    "credits": 10,
    "minCredits": 10,
    "maxCredits": 10
  },
  "features": {
    "videoEnabled": true
  }
}
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `modelChannel` | object | 模型渠道公开配置组 |
| `auth` | object | 认证相关公开配置 |
| `checkIn` | object | 每日签到奖励配置 |
| `features` | object | 前台功能开关公开配置 |

`modelChannel` 字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `availableModels` | string[] | 系统可用模型，由管理员手动选择；页面下拉选项可来自私有渠道模型 |
| `modelCosts` | object[] | 模型算力点配置，后端模型接口调用前按模型预扣，上游失败时返还；未配置默认不扣除 |
| `defaultModel` | string | 默认模型，从 `availableModels` 中选择 |
| `defaultImageModel` | string | 默认图片模型，从 `availableModels` 中选择 |
| `defaultTextModel` | string | 默认文本模型，从 `availableModels` 中选择 |
| `systemPrompt` | string | 系统提示词 |
| `allowCustomChannel` | boolean | 是否允许用户在配置弹窗中切换为本地直连渠道，默认允许 |

`modelCosts` 每项字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `model` | string | 模型名称 |
| `credits` | number | 每次后端模型接口调用前预扣的算力点 |

`checkIn` 字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `mode` | string | 奖励模式：`fixed` 固定点数，`random` 随机范围 |
| `credits` | number | 固定模式发放点数，默认 10 |
| `minCredits` | number | 随机模式最小发放点数 |
| `maxCredits` | number | 随机模式最大发放点数 |

用户侧请求模式：

| 模式 | 说明 |
| --- | --- |
| 云端渠道 | 使用后端 `/api/v1/*` 代理接口，请求会按模型名匹配 `private.value.channels` 中的可用渠道 |
| 本地直连 | 默认可选；`allowCustomChannel` 关闭后不可选，用户在浏览器本地配置 `baseUrl`、`apiKey` 和模型列表后直接请求模型接口 |

`auth` 字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `allowRegister` | boolean | 是否允许账号密码公开注册，默认允许；关闭后仍可通过有效注册邀请码创建账号 |
| `registerCredits` | number | 新注册用户获得的固定初始算力点，默认 0 |
| `linuxDo.enabled` | boolean | 是否开启 Linux.do 登录；已绑定账号可直接登录，Linux.do 首次创建账号时必须填写有效注册邀请码并消耗该邀请码 |

`features` 字段：
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `videoEnabled` | boolean | 是否显示视频创作台入口和页面，默认开启 |

## private.value

```json
{
  "channels": [
    {
      "protocol": "openai",
      "name": "默认渠道",
      "baseUrl": "https://api.example.com",
      "apiKey": "sk-xxx",
      "models": ["gpt-5.5", "gpt-image-2"],
      "weight": 1,
      "enabled": true,
      "remark": ""
    }
  ],
  "promptSync": {
    "enabled": true,
    "cron": "*/5 * * * *"
  }
}
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `channels` | object[] | 模型渠道列表 |
| `promptSync` | object | GitHub 远程提示词定时同步配置 |

`channels` 每项字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `protocol` | string | 协议，当前为 `openai` |
| `name` | string | 渠道名称 |
| `baseUrl` | string | OpenAI 兼容接口地址 |
| `apiKey` | string | 渠道密钥 |
| `models` | string[] | 该渠道可用模型 |
| `weight` | number | 渠道权重；同一模型有多个可用渠道时按权重随机 |
| `enabled` | boolean | 是否启用 |
| `remark` | string | 备注 |

后端调用模型时，会从已启用、已配置 `baseUrl` 和 `apiKey`、且 `models` 包含目标模型的渠道中选择一个。

`promptSync` 字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `enabled` | boolean | 是否开启定时同步，默认开启 |
| `cron` | string | Cron 表达式，默认每 5 分钟 |
