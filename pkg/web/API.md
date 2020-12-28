# Commands from JOAL UI to server

### Initialization
Endpoint to get the current state of the application

#### HTTP Request
`GET /state`

#### Return
`HTTP 200`
```json
{
  "started": "true",
  "client": {
    "name": "qBittorrent",
    "version": "4.1.0"
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
      "trackers": [
        {
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
        {
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
        {
          "url": "http://tracker3.example.com/announce",
          "isAnnouncing": false,
          "inUse": false,
          "seeders": 0,
          "leechers": 0,
          "interval": 0,
          "announceHistory": []
        }
      ]
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
  "needRestartToTakeEffect": false,
  "runtimeConfig": {
    "minimumBytesPerSeconds": 50,
    "maximumBytesPerSeconds": 250,
    "client": "qBittorrent-4.1.0"
  }
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
  "needRestartToTakeEffect": true
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
`HTTP 200`

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
`HTTP 200`


---

# Events from Server to JOAL-ui

### Seed has started
```json
{
  "started": "true",
  "client": {
    "name": "qBittorrent",
    "version": "4.1.0"
  }
}
```

### Seed has stopped
```json
{
  "started": "stopped"
}
```

### Config has changed
```json
{
  "needRestartToTakeEffect": false,
  "runtimeConfig": {
    "minimumBytesPerSeconds": 50,
    "maximumBytesPerSeconds": 250,
    "client": "qBittorrent-4.1.0"
  }
}
```

### Torrent has been added
```json
{
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
```


### Torrent has changed
```json
{
  "infohash": "MCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCww",
  "name": "ubuntu 20.04",
  "file": "/data/torrents/ubuntu-20.04.torrent",
  "size": 12542111,
  "isAnnouncing": false,
  "seeders": 20,
  "leechers": 5,
  "uploaded": 5942,
  "trackers": [
    {
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
    {
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
    {
      "url": "http://tracker3.example.com/announce",
      "isAnnouncing": false,
      "inUse": false,
      "seeders": 0,
      "leechers": 0,
      "interval": 0,
      "announceHistory": []
    }
  ]
}
```

### Torrent has been removed
```json
{
  "infohash": "MCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCwwLDAsMCww"
}
```

### Unexpected error
```json
{
  "error": "Something has happened",
  "datetime": "2020/12/22 11:32:25"
}
```

### Dispatcher has changed speed range
```json
{
  "currentBandwidth": 200
}
```

### Dispatcher speed distribution has changed
```json
{
  "torrents": {
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

