# Commands from JOAL UI to server

### Initialization

Endpoint to get the current state of the application

#### HTTP Request

`GET /state`

#### Return

`HTTP 200`

```json
{
  "global": {
    "started": "true",
    "client": {
      "name": "qBittorrent",
      "version": "4.1.0"
    }
  },
  "config": {
    "needRestartToTakeEffect": false,
    "runtimeConfig": {
      "minimumBytesPerSeconds": 50,
      "maximumBytesPerSeconds": 250,
      "client": "qBittorrent-4.1.0"
    }
  },
  "torrents": {
    "MCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCww": {
      "infohash": "MCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCww",
      "name": "ubuntu 20.04",
      "file": "/data/torrents/ubuntu-20.04.torrent",
      "size": 12542111,
      "isAnnouncing": false,
      "seeders": 20,
      "leechers": 5,
      "uploaded": 5942,
      "trackers": {
        "http://tracker.example.com/announce": {
          "url": "http://tracker.example.com/announce",
          "isAnnouncing": false,
          "inUse": true,
          "seeders": 200,
          "leechers": 50,
          "interval": 1800,
          "announceHistory": [
            {
              "wasSuccessful": "true",
              "datetime": "2020/12/22 11:32:25",
              "seeders": 200,
              "leechers": 50,
              "interval": 1800
            },
            {
              "wasSuccessful": "true",
              "datetime": "2020/12/22 11:02:23",
              "seeders": 180,
              "leechers": 70,
              "interval": 1800
            },
            {
              "wasSuccessful": "false",
              "datetime": "2020/12/22 10:32:22",
              "error": "network error"
            }
          ]
        },
        "http://tracker2.example.com/announce": {
          "url": "http://tracker2.example.com/announce",
          "isAnnouncing": false,
          "inUse": true,
          "seeders": 20,
          "leechers": 5,
          "interval": 1800,
          "announceHistory": [
            {
              "wasSuccessful": "true",
              "datetime": "2020/12/22 11:32:25",
              "seeders": 20,
              "leechers": 5,
              "interval": 1800
            },
            {
              "wasSuccessful": "true",
              "datetime": "2020/12/22 11:02:23",
              "seeders": 15,
              "leechers": 10,
              "interval": 1800
            },
            {
              "wasSuccessful": "false",
              "datetime": "2020/12/22 10:32:22",
              "error": "network error"
            }
          ]
        },
        "http://tracker3.example.com/announce": {
          "url": "http://tracker3.example.com/announce",
          "isAnnouncing": false,
          "inUse": false,
          "seeders": 0,
          "leechers": 0,
          "interval": 0,
          "announceHistory": []
        }
      }
    }
  },
  "bandwidth": {
    "currentBandwidth": 200,
    "torrents": {
      "MCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCww": {
        "infohash": "MCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCww",
        "percentOfBandwidth": 100
      }
    }
  }
}
```

---

### Start seeding

Endpoint to start seeding all torrents (basically put the torrent manager in Started state)

#### HTTP Request

`POST /start`

#### Return

`HTTP 200`

---

### Stop seeding

Endpoint to stop seeding all torrents (basically put the torrent manager in Stopped state)

#### HTTP Request

`POST /stop`

#### Return

`HTTP 200`

---

### Get configuration

Endpoint to get the current app configuration

#### HTTP Request

`GET /configuration`

#### Return

`HTTP 200`

```json
{
  "minimumBytesPerSeconds": 50,
  "maximumBytesPerSeconds": 250,
  "client": "qBittorrent-4.1.0"
}
```

---

### Change configuration

Endpoint to stop change update the app configuration

#### HTTP Request

`PUT /configuration`

#### HTTP BODY

```json
{
  "minimumBytesPerSeconds": 50,
  "maximumBytesPerSeconds": 250,
  "client": "qBittorrent-4.1.0"
}
```

#### Return

`HTTP 200`

```json
{
  "minimumBytesPerSeconds": 50,
  "maximumBytesPerSeconds": 250,
  "client": "qBittorrent-4.1.0"
}
```

---

### List all available emulated clients

Endpoint to get the list of all JOAL available clients

#### HTTP Request

`GET /clients/all`

#### Return

`HTTP 200`

```json
[
  "qBittorrent-4.1.1",
  "qBittorrent-4.1.0",
  "qBittorrent-3.2.0",
  "uTorrent-1.2.5"
]
```

--- 

### Add a torrent

Endpoint to upload a torrent

#### HTTP Request

`POST /torrents`

#### Multipart form parameters

| parameter | description   |
|-----------|---------------|
| file  | the .torrent file |

#### Return

`HTTP 201`

--- 

### Remove torrent

Endpoint to remove a torrent

#### HTTP Request

`DELETE /torrents`

#### Query Parameters

| parameter | description                       |
|-----------|-----------------------------------|
| infohash  | torrent infohash (base64 encoded) |

#### Return

`HTTP 204`


---

# Events from Server to JOAL-ui

**All STOMP messages are send to the /joal-core-events STOMP endpoint**

### Seed has started

```json
{
  "type": "@STOMP_API/SEED/STARTED",
  "payload": {
    "started": "true",
    "client": {
      "name": "qBittorrent",
      "version": "4.1.0"
    }
  }
}
```

### Seed has stopped

```json
{
  "type": "@STOMP_API/SEED/STOPPED",
  "payload": {
    "started": "false"
  }
}
```

### Config has changed

```json
{
  "type": "@STOMP_API/CONFIG/CHANGED",
  "payload": {
    "needRestartToTakeEffect": false,
    "runtimeConfig": {
      "minimumBytesPerSeconds": 50,
      "maximumBytesPerSeconds": 250,
      "client": "qBittorrent-4.1.0"
    }
  }
}
```

### Torrent has been added

```json
{
  "type": "@STOMP_API/TORRENT/ADDED",
  "payload": {
    "infohash": "MCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCww",
    "name": "ubuntu 20.04",
    "file": "/data/torrents/ubuntu-20.04.torrent",
    "size": 12542111,
    "isAnnouncing": false,
    "uploaded": 0,
    "trackers": [
      {
        "url": "http://tracker.example.com/announce",
        "isAnnouncing": false
      },
      {
        "url": "http://tracker2.example.com/announce",
        "isAnnouncing": false
      },
      {
        "url": "http://tracker3.example.com/announce",
        "isAnnouncing": false
      }
    ]
  }
}
```

### Torrent has changed

```json
{
  "type": "@STOMP_API/TORRENT/CHANGED",
  "payload": {
    "infohash": "MCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCww",
    "name": "ubuntu 20.04",
    "file": "/data/torrents/ubuntu-20.04.torrent",
    "size": 12542111,
    "isAnnouncing": false,
    "seeders": 20,
    "leechers": 5,
    "uploaded": 5942,
    "trackers": {
      "http://tracker.example.com/announce": {
        "url": "http://tracker.example.com/announce",
        "isAnnouncing": false,
        "inUse": true,
        "seeders": 200,
        "leechers": 50,
        "interval": 1800,
        "announceHistory": [
          {
            "wasSuccessful": "true",
            "datetime": "2020/12/22 11:32:25",
            "seeders": 200,
            "leechers": 50,
            "interval": 1800
          },
          {
            "wasSuccessful": "true",
            "datetime": "2020/12/22 11:02:23",
            "seeders": 180,
            "leechers": 70,
            "interval": 1800
          },
          {
            "wasSuccessful": "false",
            "datetime": "2020/12/22 10:32:22",
            "error": "network error"
          }
        ]
      },
      "http://tracker2.example.com/announce": {
        "url": "http://tracker2.example.com/announce",
        "isAnnouncing": false,
        "inUse": true,
        "seeders": 20,
        "leechers": 5,
        "interval": 1800,
        "announceHistory": [
          {
            "wasSuccessful": "true",
            "datetime": "2020/12/22 11:32:25",
            "seeders": 20,
            "leechers": 5,
            "interval": 1800
          },
          {
            "wasSuccessful": "true",
            "datetime": "2020/12/22 11:02:23",
            "seeders": 15,
            "leechers": 10,
            "interval": 1800
          },
          {
            "wasSuccessful": "false",
            "datetime": "2020/12/22 10:32:22",
            "error": "network error"
          }
        ]
      },
      "http://tracker3.example.com/announce": {
        "url": "http://tracker3.example.com/announce",
        "isAnnouncing": false,
        "inUse": false,
        "seeders": 0,
        "leechers": 0,
        "interval": 0,
        "announceHistory": []
      }
    }
  }
}
```

### Torrent has been removed

```json
{
  "type": "@STOMP_API/TORRENT/REMOVED",
  "payload": {
    "infohash": "MCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCww"
  }
}
```

### Dispatcher has changed speed range

```json
{
  "type": "@STOMP_API/BANDWIDTH/RANGE_CHANGED",
  "payload": {
    "currentBandwidth": 200
  }
}
```

### Dispatcher speed distribution has changed

```json
{
  "type": "@STOMP_API/BANDWIDTH/DISTRIBUTION_CHANGED",
  "payload": {
    "MCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCww": {
      "infohash": "MCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCww",
      "percentOfBandwidth": 99.9
    },
    "QldmFhNFjfQQloUQhhdjk": {
      "infohash": "QldmFhNFjfQQloUQhhdjk",
      "percentOfBandwidth": 0.1
    }
  }
}
```

### Unexpected error

```json
{
  "type": "@STOMP_API/UNEXPECTED_ERROR",
  "payload": {
    "error": "Something has happened",
    "datetime": "2020/12/22 11:32:25"
  }
}
```