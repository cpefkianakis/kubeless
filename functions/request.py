from urllib import response
import urllib.request
import urllib.parse
import json

def main(event, context):
    url = event['data']['url']
    response = urllib.request.urlopen(url)
    json_response = response.read().decode('utf-8')
    return json_response