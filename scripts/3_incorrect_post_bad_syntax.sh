curl -i --header "Content-Type: application/json" \
  --request POST \
  --data '{
    "machineId": 61616,
    "stats": {
        "cpuTemp": 78,
        "fanSpeed": 500,
        "HDDSpace": 100
        "internalTemp": 23
    },
    "lastLoggedIn": "admin/Tim",
    "sysTime": "Wed 2021-07-28 14:16:27"
}' \
  http://localhost:4000/metrics
