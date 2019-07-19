#!/usr/bin/env python

import time
import urllib.request

ENDPOINT = 'http://localhost:3000'

print('Starting Pinkfoot producer')
print(ENDPOINT)

while True:
    msg = b'This is a test'
    print(msg)

    req = urllib.request.Request(url=ENDPOINT, data=msg, method='POST')
    with urllib.request.urlopen(req) as f:
        #print(f.status, f.reason)
        pass

    time.sleep(0.05)