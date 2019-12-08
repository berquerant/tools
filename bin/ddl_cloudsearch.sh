#!/bin/sh

cd $(dirname $0)
D=$(pwd)

echo Truncate
curl -X DELETE $CLOUDSEARCH_DEV_ENDPOINT

cd $D/../ddl/cloudsearch
for t in $(find . -name "*.json")
do
    filename=$(basename "$t")
    echo "Process $filename"
    cat $t | python $PROJECT/py/cloudsearch.py --region_name $CLOUDSEARCH_REGION --no_mercy upload --endpoint_url $CLOUDSEARCH_DOCUMENT_ENDPOINT
done

