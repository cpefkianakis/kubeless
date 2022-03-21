from math import sin, cos, sqrt, atan2, radians


def main(event, context):
    # input is a string containing four
    # comma-separated float numbers
    # which we assign to the variables below
    # approximate radius of earth in km
    R = 6371000.0
    
    longitude1 = event['data']['longitude1']
    latitude1 = event['data']['latitude1']
    longitude2 = event['data']['longitude2']
    latitude2 = event['data']['latitude2']

    lon1 = float(longitude1)
    lat1 = float(latitude1)
    lon2 = float(longitude2)
    lat2 = float(latitude2)
    
    phi1 = radians(lat1)
    phi2= radians(lat2)
    delta_phi = radians(lat2-lat1)
    delta_lambda = radians(lon2-lon1)

    a = sin(delta_phi/2.0) * sin(delta_phi/2.0) + cos(phi1) * cos(phi2) * sin(delta_lambda/2.0) * sin(delta_lambda/2.0)
    c = 2 * atan2(sqrt(a), sqrt(1 - a))
    distance = R * c

    return str(round(distance/1000, 3))
