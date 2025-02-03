#!/usr/bin/env python3

import argparse
import logging
import pathlib
import sqlite3
import time
from enum import IntEnum, auto
from typing import Dict

import requests
from bs4 import BeautifulSoup
from notify import notify, register_notifications


# see ygl-server.go ConfigType
class ConfigType(IntEnum):
    """TODO: _summary_

    :param IntEnum: _description_
    :type IntEnum: _type_
    """
    INTEGER = 0
    BOOLEAN = auto()
    STRING = auto()
    NOTIFICATION = auto()

logger = logging.getLogger(__name__)
logging.basicConfig(
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    level=logging.INFO
)

parser = argparse.ArgumentParser()
parser.add_argument('--db', type=pathlib.Path, required=True, help='Path to sqlite DB file')
parser.add_argument('--notify', action='store_true', help='Enable notifications for new listings')
args = parser.parse_args()

def ygl_listings(url: str):
    '''Generator function for getting all paginated listings from a ygl search query'''
    page = 0

    while True:
        page += 1
        response = requests.get(f'{url}&page={page}', timeout=10)
        response.raise_for_status()
        html_content = response.text

        soup = BeautifulSoup(html_content, 'lxml')

        if soup.find('div', class_='nothing_found') is not None:
            break

        listings = soup.find_all('div', class_='property_item')
        for listing in listings:
            yield listing

def update_db(con: sqlite3.Connection, cur_listings: Dict, ygl_url_base: str):
    '''
    Fill a persistent props dict with listings and their data.

    See ygl-server.go for table schema
    '''
    cursor = con.cursor()

    timestamp = time.time_ns()
    
    # Listings with filters
    filter_names = { # for translating our filter names to YGL parameter names
        "BedsMin": "beds_from",
        "BedsMax": "beds_to",
        "RentMin": "rent_from",
        "RentMax": "rent_to",
        "DateMin": "date_from", # "YYYY-MM-DD"
        "DateMax": "date_to", # "YYYY-MM-DD"
    }
    min_baths = 0
    max_baths = 999
    ygl_params = []
    res = cursor.execute('SELECT * FROM Filter')
    for filter_item in res.fetchall():
        name, val = filter_item
        if "Date" in name: # changing "YYYY-MM-DD" to "MM%2FDD%2FYYYY" for YGL parameter value
            year, month, day = val.split("-")
            dl = "%2F" # delimiter 
            val = f"{month}{dl}{day}{dl}{year}" 
        if name in filter_names:
            name = filter_names[name]
            ygl_params.append(f"{name}={val}") # "name=val"
        elif name == "BathsMin":
            min_baths = float(val)
        elif name == "BathsMax":
            max_baths = float(val)

    for listing in ygl_listings(f'{ygl_url_base}?{"&".join(ygl_params)}'):
        listing_element = listing.find('a', class_='item_title')
        listing_addr = listing_element.get_text()
        listing_url = listing_element['href']

        listing_props_elements = listing.find_all('div', class_='column')
        listing_props = list(map(lambda tag: tag.text.strip(), listing_props_elements))
                    
        # the listing properties are well-ordered, so we parse them directly
        listing_baths = float(listing_props[2].split(' ')[0])
        if listing_baths < min_baths or listing_baths > max_baths:
            continue
        listing_price = int(''.join(filter(lambda char: char.isdigit(), listing_props[0])))
        listing_beds = float(listing_props[1].split(' ')[0])
        listing_date = listing_props[3].split(' ')[1]

        # TODO: omg remove this I forgot it was here
        # ignore 1 Baths.. and 4 Beds over $4,600
        if True or (listing_baths >= 1.5 and listing_price/listing_beds <= 1150):
            if listing_addr not in cur_listings:
                if args.notify:
                    notify(listing_url)

                new_listing = {
                    'addr': listing_addr,
                    'refs': listing_url,
                    "price": listing_price,
                    'beds': listing_beds,
                    'baths': listing_baths,
                    'date': listing_date,
                    'notes': '',
                    'favorite': 0,
                    'dismissed': 0,
                    'timestamp': timestamp
                }
                cur_listings[listing_addr] = new_listing

                cursor.execute('''
                    INSERT INTO Listing 
                    VALUES(:addr, :refs, :price, :beds, :baths, :date, :notes, :favorite, :dismissed, :timestamp)
                ''', new_listing)

            # always check if this is a new copy of the listing
            if listing_url not in cur_listings[listing_addr]['refs']:
                cur_listings[listing_addr]['refs'] += f',{listing_url}'
                cursor.execute('''
                    UPDATE Listing 
                    SET refs = ? 
                    WHERE addr == ? 
                ''', (cur_listings[listing_addr]['refs'], listing_addr))

if __name__ == "__main__":
    con = sqlite3.connect(args.db, autocommit=True)
    cursor = con.cursor()

    cur_listings = {}
    res = cursor.execute('SELECT * FROM Listing')
    for listing in res.fetchall():
        # we only ever use the address and the refs when looking at existing entries
        # so we don't need to store the rest of the attributes here
        cur_listings[listing[0]] = {"refs": listing[1]}

    notifs = []
    res = cursor.execute('SELECT * FROM Notification')
    for notif in res.fetchall():
        notifs.append(notif[0])
    register_notifications(notifs)

    brokers = []
    res = cursor.execute('SELECT * FROM Broker')
    for broker in res.fetchall():
        brokers.append(broker)

    for broker in brokers:
        update_db(con, cur_listings, broker[0])

    con.close()
