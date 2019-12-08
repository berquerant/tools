#!/bin/sh

cd $(dirname $0)
D=$(pwd)

echo Initialize database
echo 'set @@GLOBAL.sql_mode="";' | mysql -h$MYSQL_HOST -P$MYSQL_PORT -u$MYSQL_ROOT_USER
echo 'drop database test;' | mysql -h$MYSQL_HOST -P$MYSQL_PORT -u$MYSQL_ROOT_USER
echo 'create database test;' | mysql -h$MYSQL_HOST -P$MYSQL_PORT -u$MYSQL_ROOT_USER

cd $D/../ddl/mysql
for t in $(find . -name "*.sql")
do
    filename=$(basename "$t")
    echo "Process $filename"
    mysql -h$MYSQL_HOST -P$MYSQL_PORT -u$MYSQL_ROOT_USER < $filename
done

cd $D
