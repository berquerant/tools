#!/bin/sh

cd $(dirname $0)
D=$(pwd)


echo Truncate
python $PROJECT/py/hbase.py --host $HBASE_HOST --port $HBASE_PORT truncate

echo if import hbase table failes then \"docker-compose exec hbase bash\", \"hbase shell\" and \"create_namespace \'test\'\"
cd $D/../ddl/hbase
for t in $(find . -name "*.json")
do
    filename=$(basename "$t")
    echo "Process $filename"
    cat $filename | python $PROJECT/py/hbase.py --host $HBASE_HOST --port $HBASE_PORT create_table
done
