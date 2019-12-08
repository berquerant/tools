#!/bin/sh

cd $(dirname $0)
D=$(pwd)

# python2
echo Truncate test.people
/usr/bin/python /usr/local/bin/asinfo -v "truncate:namespace=test;set=people;"

cd $D/../ddl/aerospike
for t in $(find . -name "*.aql")
do
    filename=$(basename "$t")
    echo "Process $filename"
    aql -h $AEROSPIKE_HOST -p $AEROSPIKE_PORT -f $filename
done

cd $D
