version=$1
dockerRegistry=${2:-pulkitdocker1234}

echo "============== Building image api-gateway:$version =============="
docker build -t api-gateway:$version .

echo "============== Tagging image $dockerRegistry/api-gateway:$version =============="
docker tag api-gateway:$version $dockerRegistry/api-gateway:$version

echo "============== Pushing image to registry: $dockerRegistry/api-gateway:$version =============="
docker push $dockerRegistry/api-gateway:$version
err=$?
if [ $err -ne 0 ]; then
  echo "Failed to push image to registry"
  exit 1
fi

echo "============== Removing local images =============="
docker rmi $dockerRegistry/api-gateway:$version
docker rmi api-gateway:$version

echo "============== Pushed $dockerRegistry/api-gateway:$version to registry =============="