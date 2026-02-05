-- 创建数据库
CREATE DATABASE IF NOT EXISTS zufen DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE zufen;

-- 参与者表
CREATE TABLE IF NOT EXISTS participants (
    uuid VARCHAR(19) PRIMARY KEY COMMENT '格式 xxxx-xxxx-xxxx-xxxx',
    type ENUM('team', 'person') NOT NULL COMMENT '类型：队伍找人/人找队伍',
    score INT NOT NULL COMMENT '分数',
    match_mode ENUM('exact', 'fuzzy') NOT NULL DEFAULT 'exact' COMMENT '匹配模式',
    qrcode_content TEXT COMMENT '二维码解析内容',
    qrcode_image MEDIUMTEXT COMMENT '二维码图片base64',
    status ENUM('pending', 'matched') NOT NULL DEFAULT 'pending' COMMENT '状态',
    matched_uuid VARCHAR(19) DEFAULT NULL COMMENT '匹配对方UUID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_status_type (status, type),
    INDEX idx_score (score),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='参与者表';

-- 配置表
CREATE TABLE IF NOT EXISTS config (
    `key` VARCHAR(50) PRIMARY KEY COMMENT '配置键',
    `value` TEXT NOT NULL COMMENT '配置值',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='配置表';

-- 初始化默认配置
INSERT INTO config (`key`, `value`) VALUES
    ('target_score', '2026'),
    ('fuzzy_min', '2024'),
    ('fuzzy_max', '2028'),
    ('valid_url_prefix', 'https://example.com/invite')
ON DUPLICATE KEY UPDATE `value` = VALUES(`value`);
