后端采用国产化 Solon作为框架，基于Solon开发，使用Maven进行构建， 详细文档请查看 https://solon.noear.org/。
前端采用github开源admin模板soybean-admin， 仓库地址 https://github.com/soybeanjs/soybean-admin

后端需要加入
auto-table-solon-plugin 基于实体类生成数据库表结构，自动创建数据库、表结构、自动初始化数据， 官方文档 https://autotable.tangzc.com/%E6%8C%87%E5%8D%97/%E8%BF%9B%E9%98%B6/%E5%AE%9A%E4%B9%89%E5%88%97.html
mybatis-plus orm框架 https://baomidou.com/
hutool 常用工具类
lombok 注解

运行环境：
JDK 17+
Maven 3.6.3+
Sqlite 3.34.0+