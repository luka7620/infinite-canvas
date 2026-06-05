# 后端数据库说明

本文档只记录后端当前已经使用的主要数据表。

## 数据库

后端使用 GORM 管理数据库连接和表结构迁移。

支持的存储驱动：

- `sqlite`
- `mysql`
- `postgresql`

多实例部署时不要共享 SQLite 文件，应使用 `postgresql` 或 `mysql`。Redis 只作为可选缓存层使用，不替代数据库事务和唯一约束。

当前启动时执行 `AutoMigrate`，自动维护以下表：

- `users`
- `user_identities`
- `credit_logs`
- `invite_codes`
- `invite_code_uses`
- `prompts`
- `assets`
- `generated_image_records`
- `gallery_images`
- `gallery_likes`
- `gallery_comments`
- `settings`

后续新增表时再同步补充本文档，未实际使用的规划表不提前写入。

### users

系统用户表。用户基础信息、角色、算力点余额和第三方登录标识放在该表中。

| 字段              | 类型     | 说明                       |
|-----------------|--------|--------------------------|
| `id`            | string | 主键                       |
| `username`      | string | 用户名，唯一索引                 |
| `password`      | string | 密码哈希                     |
| `email`         | string | 邮箱                       |
| `display_name`  | string | 昵称                       |
| `avatar_url`    | string | 头像地址                     |
| `role`          | string | 角色：`user`、`admin`        |
| `credits`       | number | 算力点余额                    |
| `aff_code`      | string | 用户自己的邀请码，唯一索引            |
| `aff_count`     | number | 已邀请用户数量，冗余统计字段           |
| `inviter_id`    | string | 邀请人用户 ID                 |
| `github_id`     | string | GitHub 用户 ID               |
| `linux_do_id`   | string | Linux.do 用户 ID            |
| `wechat_id`     | string | 微信用户 ID                   |
| `status`        | string | 用户状态：`active`、`ban`       |
| `last_login_at` | string | 最近登录时间                   |
| `last_check_in_date` | string | 最近签到日期，格式为 `YYYY-MM-DD` |
| `extra`         | json   | 扩展信息，第三方资料按平台命名空间保存，如 `linuxDo` |
| `created_at`    | string | 创建时间                     |
| `updated_at`    | string | 更新时间                     |

### user_identities

第三方登录身份映射表。用于在并发首次注册时约束同一个第三方账号只能绑定到一个站内用户。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 主键 |
| `user_id` | string | 站内用户 ID |
| `provider` | string | 第三方登录平台，例如 `linux-do` |
| `external_id` | string | 第三方平台用户 ID |
| `created_at` | string | 创建时间 |

`provider` 和 `external_id` 使用联合唯一索引。

### prompts

提示词表。用于保存公开提示词、内置 GitHub 系统提示词、分类和预览内容。

| 字段           | 类型     | 说明                           |
|--------------|--------|------------------------------|
| `id`         | string | 主键                           |
| `title`      | string | 标题                           |
| `cover_url`  | string | 封面图                          |
| `prompt`     | string | 提示词内容                        |
| `tags`       | json   | 标签列表                         |
| `category`   | string | 分类标识                         |
| `preview`    | text   | Markdown 展示内容，可包含文本、图片、视频链接等 |
| `created_at` | string | 创建时间                         |
| `updated_at` | string | 更新时间                         |

`github_url` 仅用于接口返回，不写入数据库。

### assets

素材表。当前用于后台素材库。

| 字段               | 类型     | 说明                            |
|------------------|--------|-------------------------------|
| `id`             | string | 主键                            |
| `title`          | string | 标题                            |
| `type`           | string | 素材类型：`text`、`image`、`video` 等 |
| `cover_url`      | string | 封面图                           |
| `tags`           | json   | 标签列表                          |
| `category`       | string | 分类标识                          |
| `description`    | string | 描述                            |
| `content`        | text   | 文本或 Markdown 内容               |
| `url`            | string | 图片、视频等媒体地址                    |
| `created_at`     | string | 创建时间                          |
| `updated_at`     | string | 更新时间                          |

### generated_image_records

站内模型通道生成图片记录表。只有通过后端模型接口生成并写入该表的图片，才能继续发布到公开画廊。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 主键 |
| `user_id` | string | 归属用户 ID |
| `image_url` | text | 生成图片地址或 data URL |
| `width` | number | 图片宽度 |
| `height` | number | 图片高度 |
| `mime_type` | string | 图片 MIME 类型 |
| `model` | string | 生成模型 |
| `prompt` | text | 生成提示词 |
| `source` | string | 来源场景，例如 `image-page`、`canvas-node`、`canvas-edit` |
| `is_published` | bool | 是否已发布到画廊 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

### gallery_images

公开画廊图片表。保存公开展示信息和审核状态，不提供本地上传入口。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 主键 |
| `generated_image_id` | string | 关联生成图片记录 ID，唯一索引 |
| `user_id` | string | 发布用户 ID |
| `title` | string | 标题 |
| `description` | text | 描述 |
| `tags` | json | 标签列表 |
| `image_url` | text | 展示图片地址 |
| `width` | number | 图片宽度 |
| `height` | number | 图片高度 |
| `mime_type` | string | 图片 MIME 类型 |
| `model` | string | 生成模型 |
| `prompt` | text | 用户选择上传提示词时保存的生成提示词；前台仅在 `show_prompt` 为 true 时返回展示 |
| `source` | string | 来源场景 |
| `show_prompt` | bool | 是否公开提示词 |
| `status` | string | 状态：`public`、`hidden`、`deleted` |
| `recommended` | bool | 是否推荐 |
| `like_count` | number | 点赞数 |
| `comment_count` | number | 评论数 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

### gallery_likes

画廊点赞记录表。一个用户对同一作品只能保留一条点赞记录。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 主键 |
| `gallery_id` | string | 画廊作品 ID |
| `user_id` | string | 点赞用户 ID |
| `created_at` | string | 创建时间 |

### gallery_comments

画廊评论表。公开画廊作品的评论默认公开展示。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 主键 |
| `gallery_id` | string | 画廊作品 ID |
| `user_id` | string | 评论用户 ID |
| `username` | string | 评论时用户名快照 |
| `display_name` | string | 评论时展示名快照 |
| `avatar_url` | text | 评论时头像快照 |
| `content` | text | 评论内容 |
| `status` | string | 评论状态，当前为 `public` |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

### settings

系统配置表，只保存两行数据：`public` 放前端可读取的公开配置，`private` 放仅后端和管理员可读取的私有配置，配置值都用 JSON。

| 字段           | 类型     | 说明                    |
|--------------|--------|-----------------------|
| `key`        | string | 主键：`public`、`private` |
| `value`      | json   | 配置内容                  |
| `created_at` | string | 创建时间                  |
| `updated_at` | string | 更新时间                  |

`public.value` 常放前端展示和可公开读取的配置，例如模型列表、登录开关等。
`private.value` 常放渠道密钥、登录密钥、后台内部开关等。

当前系统设置接口会按后端结构体序列化和反序列化已知字段；数据库 JSON 中额外存在的旧字段会被忽略。

`public.value` 当前字段：

| 字段                | 类型       | 说明             |
|-------------------|----------|----------------|
| `modelChannel` | object | 模型渠道公开配置组 |
| `auth` | object | 公开登录配置 |
| `checkIn` | object | 每日签到奖励配置 |

`modelChannel` 当前字段：

| 字段                | 类型       | 说明             |
|-------------------|----------|----------------|
| `availableModels` | string[] | 系统可用模型列表       |
| `modelCosts` | object[] | 模型算力点配置       |
| `defaultModel`    | string   | 默认模型           |
| `defaultImageModel` | string | 默认图片模型         |
| `defaultVideoModel` | string | 默认视频模型         |
| `defaultTextModel` | string  | 默认文本模型         |
| `systemPrompt`    | string   | 系统提示词          |
| `allowCustomChannel` | bool    | 是否允许用户自定义渠道，默认允许，关闭后前端只提供走后端渠道的模式 |

`modelCosts` 每项字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `model` | string | 模型名称 |
| `credits` | number | 每次后端模型接口调用前预扣的算力点，未配置默认不扣除 |

`checkIn` 当前字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `mode` | string | 奖励模式：`fixed` 固定点数，`random` 随机范围 |
| `credits` | number | 固定模式发放点数，默认 10 |
| `minCredits` | number | 随机模式最小发放点数 |
| `maxCredits` | number | 随机模式最大发放点数 |

`auth.linuxDo` 当前字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `enabled` | bool | 是否开启 Linux.do 登录 |

`private.value` 当前字段：

| 字段         | 类型       | 说明       |
|------------|----------|----------|
| `channels` | object[] | 模型渠道配置列表 |
| `promptSync` | object | GitHub 远程提示词定时同步配置 |
| `auth` | object | 私有登录配置 |

`channels` 每项字段：

| 字段       | 类型       | 说明       |
|----------|----------|----------|
| `protocol` | string | 协议，当前支持 `openai` |
| `name`   | string   | 渠道名称     |
| `baseUrl` | string  | 渠道接口地址   |
| `apiKey` | string   | 渠道密钥     |
| `models` | string[] | 渠道可用模型列表 |
| `weight` | number   | 渠道权重，同一模型命中多个渠道时按权重随机 |
| `enabled` | bool    | 是否启用     |
| `remark` | string   | 备注       |

`promptSync` 字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `enabled` | bool | 是否开启定时同步，默认开启 |
| `cron` | string | Cron 表达式，默认每 5 分钟 |

`auth.linuxDo` 当前字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `clientId` | string | Linux.do OAuth App Client ID |
| `clientSecret` | string | Linux.do OAuth App Client Secret，后台返回时隐藏 |

后端请求模型时，先按模型名筛选启用且包含该模型的渠道，再按 `weight` 加权随机选择一个渠道。

### credit_logs

用户算力点变更流水表。当前记录后台手动调整、每日签到、邀请码兑换、模型调用预扣和模型调用失败返还。

| 字段           | 类型     | 说明                       |
|--------------|--------|--------------------------|
| `id`         | string | 主键                       |
| `user_id`    | string | 关联用户 ID                  |
| `type`       | string | 类型：`admin_adjust`、`check_in`、`invite_code`、`ai_consume`、`ai_refund` |
| `amount`     | number | 本次变动数量，增加为正，扣减为负         |
| `balance`    | number | 变动后的用户算力点余额              |
| `related_id` | string | 关联业务 ID，可为空                |
| `remark`     | string | 备注                       |
| `extra`      | json   | 扩展信息                     |
| `created_at` | string | 创建时间                     |

`type` 当前取值：

| 值 | 说明 |
| --- | --- |
| `admin_adjust` | 后台手动调整 |
| `check_in` | 每日签到获得 |
| `invite_code` | 邀请码兑换获得 |
| `ai_consume` | 调用后端模型接口消费 |
| `ai_refund` | 后端模型接口调用失败返还 |

### invite_codes

邀请码表。管理员在后台创建，可用于账号密码注册、Linux.do 首次注册，或兑换算力点。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 主键 |
| `code` | string | 邀请码，唯一索引 |
| `type` | string | 类型：`register`、`credits` |
| `credits` | number | 兑换额度；注册码可为 0 |
| `max_uses` | number | 最大使用次数；0 表示不限 |
| `used_count` | number | 已使用次数 |
| `enabled` | bool | 是否启用 |
| `remark` | string | 备注 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

### invite_code_uses

邀请码使用记录表。用于记录注册或兑换行为，并限制同一用户重复使用同一个兑换码。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 主键 |
| `invite_code_id` | string | 邀请码 ID |
| `code` | string | 使用时的邀请码文本 |
| `user_id` | string | 使用用户 ID |
| `type` | string | 使用类型：`register`、`credits` |
| `credits` | number | 本次兑换额度 |
| `created_at` | string | 创建时间 |
