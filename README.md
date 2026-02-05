# 拼好分 - 互助集分匹配系统

一个用于 ZFB 芝麻集分活动的自动匹配系统，帮助队伍和个人快速找到合适的互助伙伴。

## 功能特性

- **双向匹配**：支持"队伍找人"和"人找队伍"两种模式
- **精确/模糊匹配**：精确匹配要求分数之和等于目标值，模糊匹配允许一定范围浮动
- **二维码识别**：自动解析上传的邀请二维码内容
- **实时状态查询**：通过 UUID 随时查询匹配状态
- **安全限制**：IP 每日提交限制、二维码去重、图片大小/类型验证

## 技术栈

- **后端**：Go 1.21 + Gin + GORM
- **数据库**：MySQL
- **前端**：原生 HTML/CSS/JavaScript

## 快速开始

### 1. 环境要求

- Go 1.21+
- MySQL 5.7+

### 2. 创建数据库

```sql
CREATE DATABASE IF NOT EXISTS huzhu CHARACTER SET utf8mb4;
```

### 3. 配置文件

编辑 `config.yaml`：

```yaml
server:
  port: 8080

database:
  host: localhost
  port: 3306
  user: root
  password: "your_password"
  name: huzhu

match:
  target_score: 2026
  fuzzy_min: 2024
  fuzzy_max: 2028

qrcode:
  valid_url_prefix: "https://u.alipay.cn/"
```

### 4. 运行

```bash
# 直接运行
go run main.go

# 或编译后运行
go build -o pinhaofen.exe
./pinhaofen.exe
```

访问 http://localhost:8080 即可使用。

## 动态配置

系统启动时会自动创建 `config` 表并初始化默认配置，可在数据库中动态修改：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| target_score | 2026 | 目标分数 |
| fuzzy_min | 2024 | 模糊匹配最小值 |
| fuzzy_max | 2028 | 模糊匹配最大值 |
| valid_url_prefix | https://u.alipay.cn/ | 二维码 URL 前缀验证 |
| upload_dir | ./uploads | 图片上传目录 |
| max_per_day_ip | 3 | 每 IP 每日最大提交次数 |

## API 接口

### 注册匹配

```
POST /api/register
Content-Type: application/json

{
  "type": "team",           // team: 队伍找人, person: 人找队伍
  "score": 1000,            // 当前分数 (0-2026)
  "match_mode": "exact",    // exact: 精确匹配, fuzzy: 模糊匹配
  "qrcode_image": "data:image/png;base64,..."  // 二维码图片 base64
}
```

### 查询状态

```
GET /api/status/:uuid
```

### 获取配置

```
GET /api/config
```

## 项目结构

```
.
├── main.go              # 程序入口
├── config/              # 配置加载
├── database/            # 数据库连接
├── handler/             # API 接口处理
├── model/               # 数据模型
├── service/             # 业务逻辑
├── static/              # 前端静态文件
├── uploads/             # 上传的二维码图片
├── config.yaml          # 配置文件
└── README.md
```

## 安全特性

- IP 限制：每个 IP 每天最多提交 3 次
- 二维码去重：同一二维码不允许重复提交
- 图片验证：仅支持 PNG/JPG/GIF，最大 2MB
- 分数验证：前后端双重验证，范围 0-2026
- SQL 注入防护：使用 GORM 参数化查询

## 开源协议

MIT License

## 免责声明

本项目仅供学习交流使用，请勿用于任何商业用途。使用本系统产生的任何后果由使用者自行承担。
