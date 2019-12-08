use test;

set @old_uc=@@unique_checks, unique_checks=0;
set @old_fkc=@@foreign_key_checks, foreign_key_checks=0;
set @old_ac=@@autocommit, autocommit=0;
set @old_sm=@@sql_mode, sql_mode=traditional;

drop table if exists `people`;
create table `people` (
  `id` int(10) unsigned not null auto_increment,
  `email` varchar(255) not null unique,
  `name` varchar(255),
  `created_at` datetime not null default current_timestamp,
  `updated_at` datetime not null default current_timestamp on update current_timestamp,
  primary key (`id`),
  key `email_index` (`email`)
) engine=innodb default charset=utf8 comment='people';

insert into `people` (`id`, `email`, `name`) values
  (1, "alice@example.com", "alice"),
  (2, "bob@example.com", "bob")
;

commit;
set sql_mode=@old_sm;
set autocommit=@old_ac;
set foreign_key_checks=@old_fkc;
set unique_checks=@old_uc;
