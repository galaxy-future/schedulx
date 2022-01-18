use
schedulx;
--
-- Table structure for table `instance`
--

DROP TABLE IF EXISTS `instance`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `instance`
(
    `id`                 bigint(20)   NOT NULL AUTO_INCREMENT,
    `task_id`            bigint(20)   NOT NULL COMMENT '任务 id',
    `service_cluster_id` bigint(20)   NOT NULL COMMENT '服务集群 id',
    `instance_id`        varchar(255) NOT NULL COMMENT '实例id',
    `instance_status`    varchar(16)  NOT NULL COMMENT 'INIT, BASE_ENV,SVC,ALB,NGINX,FAIL,UNALB',
    `ip_inner`           varchar(255) NOT NULL COMMENT '私网',
    `ip_outer`           varchar(255) NOT NULL COMMENT '公网',
    `msg`                varchar(128) NOT NULL COMMENT '失败信息',
    `create_at`          timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `update_at`          timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_instance_id_task_id` (`instance_id`, `task_id`),
    KEY `idx_task_id` (`task_id`),
    KEY `idx_cluster_id` (`service_cluster_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `instr_record`
--

DROP TABLE IF EXISTS `instr_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `instr_record`
(
    `id`           bigint(20) NOT NULL AUTO_INCREMENT,
    `task_id`      bigint(20) NOT NULL COMMENT 'task主键',
    `instr_status` varchar(32)  NOT NULL COMMENT '启动运行结果 成功，失败',
    `msg`          varchar(128) NOT NULL COMMENT '系统信息',
    `instr_id`     bigint(20) NOT NULL COMMENT '指令 id',
    `create_at`    timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `update_at`    timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_task_id_instr_id` (`task_id`,`instr_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='指令记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `instruction`
--

DROP TABLE IF EXISTS `instruction`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `instruction`
(
    `id`           bigint(20) NOT NULL AUTO_INCREMENT,
    `cmd`          varchar(1024) NOT NULL COMMENT '指令集',
    `params`       varchar(256)  NOT NULL DEFAULT '' COMMENT '配套参数{''image_url'':''xx'', ''port'':80}',
    `instr_action` varchar(32)   NOT NULL COMMENT '指令的执行 bridgx.expland, bridgx.shrink, nodeact.initbase, nodeact.initsvc, mount.slb, mount.nginx, umount.slb, umount.nginx',
    `tmpl_id`      bigint(20) NOT NULL COMMENT '所属模板id',
    `is_deleted`   tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否被删除 0 否 1 是',
    `create_at`    timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `update_at`    timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY            `idx_tmpl_id` (`tmpl_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='指令集';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `schedule_template`
--

DROP TABLE IF EXISTS `schedule_template`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `schedule_template`
(
    `id`                    bigint(20) NOT NULL AUTO_INCREMENT,
    `tmpl_name`             varchar(32) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT '模板名称',
    `service_name`          varchar(32)                    NOT NULL COMMENT '服务名',
    `service_cluster_id`    bigint(20) NOT NULL COMMENT '服务集群 id',
    `bridgx_clusname`       varchar(32) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT 'bridgx 集群名称',
    `description`           varchar(32)                    NOT NULL DEFAULT '' COMMENT '模板描述',
    `instr_group`           varchar(64)                    NOT NULL DEFAULT '' COMMENT '指令集含执行步骤[100,101,102,103]',
    `schedule_type`         varchar(32)                    NOT NULL COMMENT 'expand | shrink',
    `reverse_sched_tmpl_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '反向模板的主键id',
    `is_deleted`            tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否被删除 0 否 1 是',
    `create_at`             timestamp                      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `update_at`             timestamp                      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY                     `idx_service_name` (`service_name`),
    KEY                     `idx_service_cluster_id` (`service_cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务模版表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `service`
--

DROP TABLE IF EXISTS `service`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `service`
(
    `id`           bigint(20) NOT NULL AUTO_INCREMENT,
    `service_name` varchar(32) NOT NULL COMMENT '服务名字(全局唯一)',
    `description`  varchar(32) NOT NULL COMMENT '服务描述',
    `language`     varchar(16) NOT NULL COMMENT '服务类型 java,php,golang nginx',
    `is_deleted`   tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否被删除 0 否, 1 是',
    `create_at`    timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `update_at`    timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `server_name` (`service_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务信息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `service_cluster`
--

DROP TABLE IF EXISTS `service_cluster`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `service_cluster`
(
    `id`             bigint(20)  NOT NULL AUTO_INCREMENT,
    `service_name`   varchar(32) NOT NULL COMMENT '所属服务',
    `cluster_name`   varchar(32) NOT NULL COMMENT '服务集群名称',
    `bridgx_cluster` varchar(32) NOT NULL DEFAULT '' comment 'bridgx集群名称',
    `auto_decision`  varchar(3)  NOT NULL DEFAULT 'off' COMMENT '是否开启自动扩缩容决策 on | off',
    `create_at`      timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `update_at`      timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_uservice_name_cluster_name` (`service_name`,`cluster_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务集群信息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `task`
--

DROP TABLE IF EXISTS `task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `task`
(
    `id`               bigint(20) NOT NULL AUTO_INCREMENT,
    `sched_tmpl_id`    bigint(20) NOT NULL COMMENT '关联的模板 id',
    `operator`         varchar(32)  NOT NULL COMMENT '操作人',
    `relation_task_id` varchar(128) NOT NULL COMMENT '{"nodeact_task_id":111,"brigx_task_id":222}关联上下游任务id',
    `task_status`      varchar(32)  NOT NULL COMMENT '任务状态：init, running, succ, fail',
    `task_step`        varchar(32)  NOT NULL COMMENT '任务进度',
    `inst_cnt`         int(11) NOT NULL COMMENT '本次任务操作的实例数量',
    `exec_type`        varchar(8)   NOT NULL COMMENT '执行方式 manual | auto',
    `msg`              varchar(128) NOT NULL COMMENT '系统信息',
    `begin_at`         timestamp NULL DEFAULT NULL COMMENT '任务开始时间',
    `finish_at`        timestamp NULL DEFAULT NULL COMMENT '任务结束时间',
    `create_at`        timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `update_at`        timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY                `idx_sched_tmpl_id` (`sched_tmpl_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='调度任务表';