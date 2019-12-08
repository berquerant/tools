#!/bin/sh

cd $(dirname $0)
D=$(pwd)

echo Truncate
for b in $(aws s3 --region $S3_REGION --endpoint-url $S3_ENDPOINT ls | awk '{print $3}' | tr '\n' ' ')
do
    aws s3 --region $S3_REGION --endpoint-url $S3_ENDPOINT rb s3://$b --force
done

cd $D/../ddl/s3
for b in $(find * -type d -maxdepth 1)
do
    echo "Process $b make bucket"
    aws s3 --region $S3_REGION --endpoint-url $S3_ENDPOINT mb s3://$b
    for f in $(find $b -type f)
    do
        echo "Process $f upload"
        aws s3 --region $S3_REGION --endpoint-url $S3_ENDPOINT cp $f s3://$f
    done
done

cd $D
