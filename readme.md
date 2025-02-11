# TensorFlow Inference Serving with Docker and S3

## Export Amazon Creds

```sh
export AWS_ACCESS_KEY_ID=xxxxxxxxxxxxxxxxxx
export AWS_SECRET_ACCESS_KEY=xxxxxxxxxxxxxx
```

## Edit and Upload `configs/` and `models/` dir to your S3 bucket

Edit `models.config` to point to your bucket:

```sh
model_config_list {
  config {
    name: 'half_plus_two'
    base_path: 's3://cnr-knative-tfserving-models/models/half_plus_two/'
      model_platform: "tensorflow"
  },
  config {
    name: 'resnet'
    base_path: 's3://cnr-knative-tfserving-models/models/resnet/'
      model_platform: "tensorflow"
  },
  config {
    name: 'anpr'
    base_path: 's3://cnr-knative-tfserving-models/models/anpr/'
      model_platform: "tensorflow"
  }
}
```

Upload the `configs/` and `models/` dirs to the S3 bucket.

## Config container to pull config and model from S3 and run container

```sh
docker run \
    -p 8500:8500 \
    -p 8501:8501 \
    -e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
    -e AWS_REGION=${REGION} \
    -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
    -e S3_ENDPOINT=$S3_URL \
    harbor-repo.vmware.com/dockerhub-proxy-cache/tensorflow/serving \
    --model_config_file=s3://${BUCKETNAME}/${FOLDER/FILE_PATH}
```

Example:

```sh
docker run \
    -p 8500:8500 \
    -p 8501:8501 \
    -e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
    -e AWS_REGION=eu-west-1 \
    -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
    -e S3_ENDPOINT=s3.eu-west-1.amazonaws.com \
    harbor-repo.vmware.com/dockerhub-proxy-cache/tensorflow/serving \
    --model_config_file=s3://cnr-knative-tfserving-models/configs/models.config \
    --monitoring_config_file=s3://cnr-knative-tfserving-models/configs/monitoring_config.txt
```

## Test the container is pulling the model and inference works

You can check the status of the model pull from S3 by checking the container logs with `docker logs -f <container-id>`.
### Half plus two

```sh
# Send three values to the model and get it to half them and add two
curl -s -X POST -d '{"instances": [1.0, 2.0, 5.0]}' http://localhost:8501/v1/models/half_plus_two:predict | jq 
```

### Resnet

Cat:

```sh
# Encode the cat image in base64 for the model
INPUT_IMG=$(cat test/cat.jpg| base64)
# Query the inference server with the cat image and see what it predicts
curl -s -X POST -d '{"instances": [{"b64": "'$(echo $INPUT_IMG)'"}]}' http://localhost:8501/v1/models/resnet:predict | jq '.predictions[0].classes'

# If the output is `286` then we've been successful in mapping the image as a cat: https://gist.github.com/yrevar/942d3a0ac09ec9e5eb3a
```

Dog:

```sh
# Encode the dog image in base64 for the model
INPUT_IMG=$(cat test/dog.jpg| base64)
# Query the inference server with the cat image and see what it predicts
curl -s -X POST -d '{"instances": [{"b64": "'$(echo $INPUT_IMG)'"}]}' http://localhost:8501/v1/models/resnet:predict | jq '.predictions[0].classes'

# If the output is `209` then we've been successful in mapping the image as a golden retriever: https://gist.github.com/yrevar/942d3a0ac09ec9e5eb3a
```

### ANPR

Car with plate:

```sh
# Encode the car image in base64 for the model
INPUT_IMG=$(cat test/cars/car.png| base64)

# Query the inference server with the car image and see what it predicts
curl -s -X POST -d '{"instances": [{"b64": "'$(echo $INPUT_IMG)'"}]}' http://localhost:8501/v1/models/anpr:predict | jq '.predictions[0].detection_classes'

#This will ouput the raw TensorFlow model output and it needs a little post-processing to line up the labels with the predictions, I wrote a small CLI to do that:
$ python client/predict_images_client.py -s http://localhost:8501/v1/models/anpr:predict -i test/cars/ -l test/classes.pbtxt
    Found plate:  35nn72
    Found plate:  6y0m172
```

## Build dedicated docker container with baked-in images

```sh
docker run --rm -d --name serving_base tensorflow/serving
docker cp models/ serving_base:/
docker cp configs/ serving_base:/
docker commit --author "Myles Gray" serving_base anpr-serving
docker kill serving_base

$ docker image list
REPOSITORY                                                        TAG       IMAGE ID       CREATED         SIZE
anpr-serving                                                      latest    0d0f48077d9a   8 seconds ago   702MB
```

Tag and push:

```sh
docker tag anpr-serving harbor-repo.vmware.com/vspheretmm/anpr-serving
docker push harbor-repo.vmware.com/vspheretmm/anpr-serving
```

Run:

```sh
docker run -d -p 8500:8500 -p 8501:8501 anpr-serving \
  --model_config_file=/configs/models-local.config \
  --monitoring_config_file=/configs/monitoring_config.txt
```

## Run on KNative

```sh
$ kn service create tf-inference-server -n default --autoscale-window 300s \
  --request "memory=2Gi" \
  -p 8501 --image harbor-repo.vmware.com/vspheretmm/anpr-serving \
  --arg --model_config_file=/configs/models-local.config \
  --arg --monitoring_config_file=/configs/monitoring_config.txt

Creating service 'tf-inference-server' in namespace 'default':

 55.580s Configuration "tf-inference-server" is waiting for a Revision to become ready.
 55.672s Ingress has not yet been reconciled.
 55.757s Waiting for Envoys to receive Endpoints data.
 55.971s Waiting for load balancer to be ready
 56.207s Ready to serve.

Service 'tf-inference-server' created to latest revision 'tf-inference-server-00001' is available at URL:
http://tf-inference-server.default.10.198.53.135.sslip.io
```

### Half plus two

```json
$ curl -s -X POST -d '{"instances": [1.0, 2.0, 5.0]}' http://tf-inference-server.default.10.198.53.135.sslip.io/v1/models/half_plus_two:predict | jq
{
  "predictions": [
    2.5,
    3,
    4.5
  ]
}
```

### Resnet

Cat:

```sh
# Encode the cat image in base64 for the model
INPUT_IMG=$(cat test/cat.jpg| base64)
# Query the inference server with the cat image and see what it predicts
curl -s -X POST -d '{"instances": [{"b64": "'$(echo $INPUT_IMG)'"}]}' http://tf-inference-server.default.10.198.53.135.sslip.io/v1/models/resnet:predict | jq '.predictions[0].classes'

# If the output is `286` then we've been successful in mapping the image as a cat: https://gist.github.com/yrevar/942d3a0ac09ec9e5eb3a
```

Dog:

```sh
# Encode the dog image in base64 for the model
INPUT_IMG=$(cat test/dog.jpg| base64)
# Query the inference server with the cat image and see what it predicts
curl -s -X POST -d '{"instances": [{"b64": "'$(echo $INPUT_IMG)'"}]}' http://tf-inference-server.default.10.198.53.135.sslip.io/v1/models/resnet:predict | jq '.predictions[0].classes'

# If the output is `209` then we've been successful in mapping the image as a golden retriever: https://gist.github.com/yrevar/942d3a0ac09ec9e5eb3a
```

### ANPR

Car with plate:

```sh
# Encode the car image in base64 for the model
INPUT_IMG=$(cat test/cars/car.png| base64)

# Query the inference server with the car image and see what it predicts
curl -s -X POST -d '{"instances": [{"b64": "'$(echo $INPUT_IMG)'"}]}' http://tf-inference-server.default.10.198.53.135.sslip.io/v1/models/anpr:predict | jq '.predictions[0].detection_classes'

#This will ouput the raw TensorFlow model output and it needs a little post-processing to line up the labels with the predictions, I wrote a small CLI to do that:
$ python client/predict_images_client.py -s http://tf-inference-server.default.10.198.53.135.sslip.io/v1/models/anpr:predict -i test/cars/ -l test/classes.pbtxt
    Found plate:  35nn72
    Found plate:  6y0m172
```
