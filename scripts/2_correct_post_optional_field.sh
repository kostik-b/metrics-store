curl -i --header "Content-Type: application/json" \
  --request POST \
  --data '{
    "machineId": 4444,
    "stats": {
        "cpuTemp": 78,
        "fanSpeed": 500,
        "HDDSpace": 100,
        "internalTemp": 23
    },
    "lastLoggedIn": "admin/Ian",
    "sysTime": "2022-04-21T19:25:43.219Z"
}' \
  http://localhost:4000/metrics
