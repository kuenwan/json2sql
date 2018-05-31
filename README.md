# json2sql
根据json文件生成sql文件，实现数据库的版本升级管理

使用方法：
1. 在database文件夹中定义json文件，json文件的作用是描述数据库某个版本的表及其表结构
2. 执行程序后，会在sql文件夹下生成相关的create和update的sql文件
3. 使用source命令导入sql文件

注：一般我们在建立数据库时，第一个版本db_v1.json的数据库是我们最初的数据库定义。但是后续随着功能的扩展，数据库会不断有修改的需求，那么每次修改我们的数据库版本递增 db_v2.json db_v3.json ... db_v【n】.json

运行程序时：程序会根据db_v1.json生成create.sql文件，同时根据后续的db_v2.json ...db_v【n】.json生成update.sql文件，以此实现简单的数据库版本管理
