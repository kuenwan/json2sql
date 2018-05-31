
USE `db_test`;

DELIMITER ;;
CREATE PROCEDURE db_update()
BEGIN
    SELECT @id:=MAX(id) from db_version;
    SELECT @db_ver:=version from db_version where id = @id;
    
    IF(@db_ver <= 1)THEN
ALTER TABLE user DROP level;



END IF;
IF(@db_ver <= 2)THEN
CREATE TABLE IF NOT EXISTS `item_bag_1` (
`guid` BIGINT UNSIGNED NOT NULL   COMMENT '用户唯一id',
`itemId` INT  NOT NULL   COMMENT '道具id',
`itemCount` INT  NOT NULL   COMMENT '道具数量',
PRIMARY KEY (guid,itemId)
) ENGINE=InnoDB DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci COMMENT='背包表';

CREATE TABLE IF NOT EXISTS `item_bag_2` (
`guid` BIGINT UNSIGNED NOT NULL   COMMENT '用户唯一id',
`itemId` INT  NOT NULL   COMMENT '道具id',
`itemCount` INT  NOT NULL   COMMENT '道具数量',
PRIMARY KEY (guid,itemId)
) ENGINE=InnoDB DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci COMMENT='背包表';

CREATE TABLE IF NOT EXISTS `item_bag_3` (
`guid` BIGINT UNSIGNED NOT NULL   COMMENT '用户唯一id',
`itemId` INT  NOT NULL   COMMENT '道具id',
`itemCount` INT  NOT NULL   COMMENT '道具数量',
PRIMARY KEY (guid,itemId)
) ENGINE=InnoDB DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci COMMENT='背包表';

ALTER TABLE user ADD `level` INT  NOT NULL   COMMENT '等级';



END IF;

    
    INSERT INTO db_version(version, update_time) VALUES(3, utc_timestamp());
END;;

DELIMITER ;

CALL db_update();

DROP PROCEDURE db_update;
