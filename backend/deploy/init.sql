-- Database Init Script
CREATE DATABASE IF NOT EXISTS `interview_agent` DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
USE `interview_agent`;

CREATE TABLE `user` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `password_hash` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `role` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'user',
  `wechat_open_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `wechat_union_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `nickname` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `avatar` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `idx_user_username` (`username`) USING BTREE,
  UNIQUE KEY `idx_user_email` (`email`) USING BTREE,
  UNIQUE KEY `idx_user_wechat_open_id` (`wechat_open_id`) USING BTREE,
  UNIQUE KEY `idx_user_wechat_union_id` (`wechat_union_id`) USING BTREE,
  KEY `idx_user_deleted_at` (`deleted_at`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

CREATE TABLE `user_model` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模型显示名称（用户维度唯一）',
  `model_key` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模型标识（doubao-1.5-vision-lite-250315）',
  `protocol` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '协议类型（openai/ark/claude/gemini/deepseek/ollama/qwen/ernie）',
  `base_url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'API 基础地址',
  `api_key_encrypted` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '加密后的 API 密钥',
  `config_json` json DEFAULT NULL COMMENT '额外配置（如区域、访问密钥等）',
  `secret_hint` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '密钥脱敏提示（如显示末尾4位）',
  `provider_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '提供商名称（如 OpenAI、Ark、DeepSeek）',
  `meta_id` bigint DEFAULT NULL COMMENT '关联全局 model_meta.id（继承能力/图标）',
  `default_params` json DEFAULT NULL COMMENT '默认参数（如 temperature、max_tokens）',
  `scope` bigint NOT NULL DEFAULT '7' COMMENT '使用范围（位掩码：1=智能体, 2=应用, 4=工作流）',
  `status` bigint DEFAULT '1' COMMENT '状态（0=禁用, 1=启用）',
  `created_at` bigint NOT NULL DEFAULT '0' COMMENT '创建时间（毫秒时间戳）',
  `updated_at` bigint NOT NULL DEFAULT '0' COMMENT '更新时间（毫秒时间戳）',
  `deleted` bigint NOT NULL DEFAULT '0' COMMENT '删除状态（0=未删除, 1=已删除）',
  `is_default` bigint NOT NULL DEFAULT '0' COMMENT '是否为默认（0=不是, 1=是）',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_user_model_user_id` (`user_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC;

CREATE TABLE `interview_record` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '面试主表',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '面试标题',
  `type` varchar(255) NOT NULL COMMENT '面试类型(综合面试、专项面试)',
  `difficulty` varchar(128) NOT NULL COMMENT '难度级别（简单、中等、困难）',
  `domain` varchar(255) NOT NULL COMMENT '面试领域(校招、社招；java、golang)',
  `company_name` varchar(128) DEFAULT NULL COMMENT '公司名称',
  `position_name` varchar(128) DEFAULT NULL COMMENT '岗位名称',
  `interview_duration` varchar(128) DEFAULT NULL COMMENT '面试时长',
  `status` varchar(50) NOT NULL DEFAULT 'pending' COMMENT '面试状态（pending/completed）',
  `duration` bigint DEFAULT NULL COMMENT '面试耗时（秒）',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_interview_record_user_id` (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=60 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `interview_dialogues` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `topic_id` bigint unsigned NOT NULL COMMENT '关联的问题主题ID',
  `question` text COMMENT '智能体的提问内容',
  `answer` text COMMENT '用户的回答内容',
  `display_order` int unsigned NOT NULL DEFAULT '0' COMMENT '在主题内的显示顺序',
  `created_at` datetime(3) DEFAULT NULL,
  `report_id` bigint unsigned NOT NULL COMMENT '关联的面试报告ID',
  `parent_id` bigint unsigned NOT NULL COMMENT '父对话记录ID',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_topic_id_order` (`topic_id`,`display_order`),
  KEY `idx_report_id` (`report_id`),
  KEY `idx_parent_id` (`parent_id`)
) ENGINE=InnoDB AUTO_INCREMENT=48 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `interview_evaluation` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `report_id` bigint unsigned NOT NULL COMMENT '关联的面试报告ID',
  `comment` text COMMENT '总体评价',
  `score` decimal(5,2) DEFAULT NULL COMMENT '总体评分',
  `dimensions` json DEFAULT NULL COMMENT '各维度评估',
  `deleted` bigint DEFAULT '0' COMMENT '是否删除',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_report_id` (`report_id`),
  KEY `idx_deleted` (`deleted`)
) ENGINE=InnoDB AUTO_INCREMENT=15 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `answer_report` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `report_id` bigint unsigned NOT NULL COMMENT '关联的面试报告ID',
  `records` json DEFAULT NULL COMMENT '答题记录列表',
  `deleted` bigint DEFAULT '0' COMMENT '是否删除',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_report_id` (`report_id`),
  KEY `idx_deleted` (`deleted`)
) ENGINE=InnoDB AUTO_INCREMENT=18 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `resume` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '简历ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `content` longtext NOT NULL COMMENT '简历内容',
  `file_name` varchar(255) DEFAULT NULL COMMENT '原始文件名',
  `file_size` bigint DEFAULT NULL COMMENT '文件大小（字节）',
  `file_type` varchar(50) DEFAULT NULL COMMENT '文件类型(pdf/doc/txt等)',
  `is_default` bigint DEFAULT '0' COMMENT '是否为默认简历(0-否，1-是)',
  `deleted` bigint DEFAULT '0' COMMENT '删除标记(0-未删除，1-已删除)',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_resume_user_id` (`user_id`),
  KEY `idx_resume_deleted` (`deleted`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `prediction_record` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '押题记录ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `resume_id` bigint unsigned NOT NULL COMMENT '简历ID',
  `type` varchar(20) DEFAULT NULL COMMENT '押题类型(校招/社招)',
  `language` varchar(20) DEFAULT NULL COMMENT '语言类型(java/go)',
  `job_title` varchar(50) DEFAULT NULL COMMENT '岗位名称(前端/后端)',
  `difficulty` varchar(20) DEFAULT NULL COMMENT '难度等级(入门/进阶)',
  `company` varchar(100) DEFAULT NULL COMMENT '公司名称(字节/阿里等)',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_prediction_record_resume_id` (`resume_id`),
  KEY `idx_prediction_record_user_id` (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `prediction_question` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '题目ID',
  `record_id` bigint unsigned NOT NULL COMMENT '押题记录ID',
  `question` text NOT NULL COMMENT '问题',
  `focus` text COMMENT '重点考察',
  `thinking_path` text COMMENT '回答思路',
  `reference_answer` text COMMENT '参考答案',
  `follow_up` text COMMENT '可能追问(JSON或文本)',
  `sort` bigint DEFAULT NULL COMMENT '题目排序',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `content` text COMMENT '重点考察内容',
  PRIMARY KEY (`id`),
  KEY `idx_prediction_question_record_id` (`record_id`),
  CONSTRAINT `fk_prediction_record_questions` FOREIGN KEY (`record_id`) REFERENCES `prediction_record` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=16 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

