curl -i --header "Content-Type: application/json" \
  --request POST \
  --data '{
    "machineId": 12345,
    "stats": {
        "cpuTemp": 90,
        "fanSpeed": 400,
        "HDDSpace": 800,
        "unknownField": "abc"
    },
    "lastLoggedIn": "admin/Paul",
    "sysTime": "2022-04-23T18:25:43.511Z"
}' \
  http://localhost:4000/metrics
