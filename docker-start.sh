docker-compose stop
docker-compose rm -f
docker-compose build
docker-compose up -d
sleep 3
docker exec -i $(docker ps | grep mongo | cut -d ' ' -f1) mongo test_mongo_db < ./mongo/01_createuser.js
docker-compose ps