docker-compose -f compose-files/solo.yaml kill \
&& docker-compose -f compose-files/solo.yaml down


docker rm $(docker ps -f "name=.test.com-" -aq)
docker rmi $(docker images -f "reference=dev-*test.com*" -q)

docker exec fabric_host_test.com rm -rf config fabric crypto-config
 

docker kill fabric_host_test.com 
docker rm fabric_host_test.com

