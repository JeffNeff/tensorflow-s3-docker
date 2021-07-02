This function is still a work in progress.

to run locally:

export K_SINK=
export AWS_ACCESS_KEY=
export AWS_SECRET_KEY=
export AWS_REGION=us-west-2
export TENSORFLOW_ENDPOINT=http://localhost:8501/v1/models/anpr:predict

```
curl -v "localhost:8080" \
       -X POST \
       -H "Ce-Id: 536808d3-88be-4077-9d7a-a3f162705f79" \
       -H "Ce-Specversion: 1.0" \
       -H "Ce-Type: com.amazon.s3.objectcreated" \
       -H "Ce-Source: com.amazon.s3.objectcreated" \
       -H "Content-Type: application/json" \
       -d '{
    "awsRegion": "us-west-2",
    "eventName": "ObjectCreated:Put",
    "eventSource": "aws:s3",
    "eventTime": "2021-07-02T20:24:53.274Z",
    "eventVersion": "2.1",
    "requestParameters": {
      "sourceIPAddress": "216.21.210.46"
    },
    "responseElements": {
      "x-amz-id-2": "ES+BDy24tFJCegrDLv87sP7fbbupp7SybBAPu5byyDvLvEKjcKEEsPCgawyVOWr/ljiCot2pz4NCprFXtrmwSIZfd6gBy9/8",
      "x-amz-request-id": "KCM15WHF4DQRESMB"
    },
    "s3": {
      "bucket": {
        "arn": "arn:aws:s3:::demobkt-triggermesh",
        "name": "demobkt-triggermesh",
        "ownerIdentity": {
          "principalId": "A3L2KFRRF0JY9H"
        }
      },
      "configurationId": "io.triggermesh.awss3sources.dmo.my-bucket",
      "object": {
        "eTag": "fb5ecd9ec398a8cea0882daba0385eee",
        "key": "00000079_kxu8ayg_1.jpg",
        "sequencer": "0060DF7619F9A0B0E0",
        "size": 1093612
      },
      "s3SchemaVersion": "1.0"
    },
    "userIdentity": {
      "principalId": "A3L2KFRRF0JY9H"
    }
  }'