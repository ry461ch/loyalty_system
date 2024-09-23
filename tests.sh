# /bin/bash

curl -X POST -H "Content-Type: application/json" -d '{"login": "example", "password": "example"}' http://127.0.0.1:8081/api/user/register -v
curl -X POST -H "Content-Type: application/json" -d '{"login": "example", "password": "example"}' http://127.0.0.1:8081/api/user/login -v  
curl -X GET -H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjcxODA5MzIsIlVzZXJJRCI6IjQwMzIxMzg5LTU2ZmQtNGFjMi05YzI4LWU0NTMzYTU5MDdmYiIsIkxvZ2luIjoiZXhhbXBsZSJ9.Hu1q7seneAOdraDS4AtlH6tdC4ooltox6kqpjcGMQCc" http://127.0.0.1:8081/api/user/balance -v
curl -X POST -H "Content-Type: text/plain" -H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjcxODA5MzIsIlVzZXJJRCI6IjQwMzIxMzg5LTU2ZmQtNGFjMi05YzI4LWU0NTMzYTU5MDdmYiIsIkxvZ2luIjoiZXhhbXBsZSJ9.Hu1q7seneAOdraDS4AtlH6tdC4ooltox6kqpjcGMQCc" -d '1115' http://127.0.0.1:8081/api/user/orders -v
curl -X GET -H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjcxODA5MzIsIlVzZXJJRCI6IjQwMzIxMzg5LTU2ZmQtNGFjMi05YzI4LWU0NTMzYTU5MDdmYiIsIkxvZ2luIjoiZXhhbXBsZSJ9.Hu1q7seneAOdraDS4AtlH6tdC4ooltox6kqpjcGMQCc" http://127.0.0.1:8081/api/user/orders -v
curl -X GET -H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjcxODA5MzIsIlVzZXJJRCI6IjQwMzIxMzg5LTU2ZmQtNGFjMi05YzI4LWU0NTMzYTU5MDdmYiIsIkxvZ2luIjoiZXhhbXBsZSJ9.Hu1q7seneAOdraDS4AtlH6tdC4ooltox6kqpjcGMQCc" http://127.0.0.1:8081/api/user/withdrawals -v
curl -X POST -H "Content-Type: application/json" -H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjcxODMzOTEsIlVzZXJJRCI6IjQwMzIxMzg5LTU2ZmQtNGFjMi05YzI4LWU0NTMzYTU5MDdmYiIsIkxvZ2luIjoiZXhhbXBsZSJ9.WJaZh2l_E-WtFougSe_acY-5-5Thi4UHLBtkxt0teDI" -d '{"order": "1115", "sum": 10}' http://127.0.0.1:8081/api/user/balance/withdraw -v

curl -X POST -H "Content-Type: application/json" -d '{"match": "test", "reward": 100, "reward_type": "pt"}' http://127.0.0.1:8080/api/goods -v
curl -X POST -H "Content-Type: application/json" -d '{"order": "1115", "goods": [{"description": "test", "price": 3000}]}' http://127.0.0.1:8080/api/orders -v
curl -X GET http://127.0.0.1:8080/api/orders/1115 -v
