#!/usr/bin/env python

import time
import urllib.request

ENDPOINT = 'http://localhost:3000'

print('Starting Pinkfoot consumer')
print(ENDPOINT)

while True:
    #print('-------------------------------')

    try:
        with urllib.request.urlopen(ENDPOINT) as f:
            print(f.read())

            info = f.info()
            #print('X-Correlation-Id:', info['X-Correlation-Id'])

            req = urllib.request.Request(url=ENDPOINT, data=None, method='PUT')
            req.add_header('X-Correlation-Id', info['X-Correlation-Id'])
            with urllib.request.urlopen(req) as f2:
                #print('Ack:', f2.status, f2.reason)
                pass

    except urllib.error.HTTPError as err:
        #print(err.status, err.reason)
        time.sleep(1)

    time.sleep(0.05)