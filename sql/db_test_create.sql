DROP DATABASE IF EXISTS `db_test`;
CREATE DATABASE `db_test` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
GRANT ALL PRIVILEGES ON db_test.* TO 'test'@'localhost';
FLUSH PRIVILEGES;
USE `db_test`;

CREATE TABLE IF NOT EXISTS `user` (
`guid` BIGINT UNSIGNED NOT NULL   COMMENT '用户唯一id',
`account` VARCHAR(80)  NOT NULL   COMMENT '账号名称',
`level` INT  NOT NULL   COMMENT '等级',
`createTime` BIGINT  NOT NULL   COMMENT '创建时间',
`money` DECIMAL(23,2)  NOT NULL   COMMENT '持有的货币数量',
PRIMARY KEY (guid)
) ENGINE=InnoDB DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci COMMENT='用户表';

CREATE TABLE IF NOT EXISTS `timer_operate` (
`id` INT  NOT NULL   COMMENT '唯一id',
`operateTime` BIGINT  NOT NULL   COMMENT '操作时间',
PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci COMMENT='时间表';

CREATE TABLE IF NOT EXISTS `db_version` (
`id` INT UNSIGNED NOT NULL AUTO_INCREMENT  COMMENT '自增id',
`version` INT UNSIGNED NOT NULL   COMMENT '版本号',
`update_time` TIMESTAMP  NOT NULL   COMMENT '更新时间',
PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci COMMENT='数据库版本表';


INSERT INTO db_version(version, update_time) VALUES(1, utc_timestamp());
