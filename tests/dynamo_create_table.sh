aws dynamodb create-table \
    --table-name ld-table \
    --attribute-definitions \
        AttributeName=namespace,AttributeType=S \
        AttributeName=key,AttributeType=S \
    --key-schema \
        AttributeName=namespace,KeyType=HASH \
        AttributeName=key,KeyType=RANGE \
    --provisioned-throughput \
        ReadCapacityUnits=10,WriteCapacityUnits=5 \
    --endpoint-url http://localhost:8000