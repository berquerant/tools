#!/bin/sh

cd $(dirname $0)
D=$(pwd)

cd $D/../ddl/dynamodb
for t in $(find . -name "*.json" -maxdepth 1)
do
    filename=$(basename "$t")
    table="${filename%.*}"
    echo "Process $filename"
    aws --region $DYNAMODB_REGION --endpoint-url $DYNAMODB_ENDPOINT dynamodb delete-table --table-name $table > /dev/null
    aws --region $DYNAMODB_REGION --endpoint-url $DYNAMODB_ENDPOINT dynamodb create-table --cli-input-json file://$filename > /dev/null
done

cd seeds
for s in $(find . -name "*.json")
do
    filename=$(basename "$t")
    table="${filename%.*}"
    echo "Process seed $filename"
    cat $filename | python $PROJECT/py/dynamodb.py --region_name $DYNAMODB_REGION --endpoint_url $DYNAMODB_ENDPOINT -t $table write
done

cd $D
