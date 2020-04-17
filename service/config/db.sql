CREATE TABLE websites (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `name` varchar(255) NOT NULL,
    `url` varchar(255) NOT NULL,
    `auth_user` varchar(255) DEFAULT NULL,
    `auth_password` varchar(255) DEFAULT NULL,
    `created_at` datetime DEFAULT NULL,
    `updated_at` datetime DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE tasks (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `website_id` int(11) DEFAULT 0,
    `file` varchar(255) DEFAULT NULL,
    `api_count` int(11) DEFAULT 0,
    `api_error_count` int(11) DEFAULT 0,
    `api_error_rate` float(10,2) DEFAULT 0.00,
    `executed_at` datetime DEFAULT NULL,
    `created_at` datetime DEFAULT NULL,
    `updated_at` datetime DEFAULT NULL,
    INDEX `website_id` (`website_id`),
    INDEX `executed_at` (`executed_at`),
    INDEX `file` (`file`),
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `statuses` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `task_id` int(11) DEFAULT 0,
    `website_id` int(11) DEFAULT 0,
    `position` varchar(50) DEFAULT NULL COMMENT 'QD/TW/QD_CN',
    `url` varchar(255) DEFAULT NULL COMMENT 'api url',
    `label` varchar(250) DEFAULT NULL,
    `timestamp` datetime DEFAULT NULL,
    `filename` varchar(255) DEFAULT NULL,
    `elapsed` int(11) DEFAULT NULL,
    `method` varchar(50) DEFAULT NULL COMMENT 'get/post',
    `response_code` varchar(255) DEFAULT NULL,
    `response_message` varchar(255) DEFAULT NULL,
    `success` varchar(50) DEFAULT NULL,
    `failure_message` varchar(10000) DEFAULT NULL,
    `created_at` datetime DEFAULT NULL,
    `updated_at` datetime DEFAULT NULL,
    `deleted_at` datetime DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;